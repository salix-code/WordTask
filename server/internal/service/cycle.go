package service

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
)

// ErrInvalidCycleInput indicates that the caller supplied invalid parameters
// (wrong word count, duplicate ongoing cycle, etc). API layer should map this
// to HTTP 400.
var ErrInvalidCycleInput = errors.New("invalid cycle input")

// WordValidationResult holds the validation status of a single submitted word.
type WordValidationResult struct {
	Term   string `json:"term"`
	Status string `json:"status"` // valid | not_in_dict | already_used
}

// SetupCycleResult is the response returned by ValidateAndCreateCycle.
type SetupCycleResult struct {
	Ok      bool                   `json:"ok"`
	Results []WordValidationResult `json:"results"`
}

// ValidateAndCreateCycle validates the provided word list and, if all words
// pass validation, creates a new cycle with those words for the given user.
//
// Validation rules:
//  1. len(words) must be >= 1.
//  2. The user must NOT already have an ongoing cycle for this wordbook.
//  3. Each word must exist in the words table (case-insensitive match on term).
//  4. Each word must not have been used in ANY existing cycle
//     (ongoing or completed) for this user + wordbook.
//
// Returns ErrInvalidCycleInput for user-facing validation errors that should
// surface as HTTP 400.
func ValidateAndCreateCycle(userID, wordbook string, terms []string) (*SetupCycleResult, error) {
	if len(terms) < 1 {
		return nil, fmt.Errorf("%w: at least 1 word is required", ErrInvalidCycleInput)
	}

	// Reject duplicate terms within the submitted list (case-insensitive).
	seen := make(map[string]struct{}, len(terms))
	for _, t := range terms {
		k := strings.ToLower(strings.TrimSpace(t))
		if _, dup := seen[k]; dup {
			return nil, fmt.Errorf("%w: duplicate word %q in submission", ErrInvalidCycleInput, t)
		}
		seen[k] = struct{}{}
	}

	// Reject if user already has an ongoing cycle for this wordbook.
	var existingOngoing model.Cycle
	err := db.DB.
		Where("user_id = ? AND wordbook = ? AND status = ?", userID, wordbook, "ongoing").
		First(&existingOngoing).Error
	if err == nil {
		return nil, fmt.Errorf("%w: user already has an ongoing cycle for wordbook %s", ErrInvalidCycleInput, wordbook)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("query ongoing cycle: %w", err)
	}

	// Normalize terms to lowercase for matching.
	lowerTerms := make([]string, len(terms))
	for i, t := range terms {
		lowerTerms[i] = strings.ToLower(strings.TrimSpace(t))
	}

	// 1. Batch-query words table (case-insensitive).
	var matchedWords []model.Word
	if err := db.DB.
		Where("wordbook = ? AND LOWER(term) IN ?", wordbook, lowerTerms).
		Find(&matchedWords).Error; err != nil {
		return nil, fmt.Errorf("query words: %w", err)
	}

	// Build a map: lower(term) → Word for fast lookup.
	wordMap := make(map[string]model.Word, len(matchedWords))
	for _, w := range matchedWords {
		wordMap[strings.ToLower(w.Term)] = w
	}

	// 2. Collect all word IDs used in ANY cycle (ongoing or completed) of this user + wordbook.
	usedWordIDs := make(map[uint]struct{})
	var usedIDs []uint
	if err := db.DB.
		Table("word_cycles").
		Select("word_cycles.word_id").
		Joins("JOIN cycles ON cycles.id = word_cycles.cycle_id").
		Where("cycles.user_id = ? AND cycles.wordbook = ?", userID, wordbook).
		Scan(&usedIDs).Error; err != nil {
		return nil, fmt.Errorf("query used word ids: %w", err)
	}
	for _, id := range usedIDs {
		usedWordIDs[id] = struct{}{}
	}

	// 3. Build per-term results.
	results := make([]WordValidationResult, len(terms))
	allValid := true
	validWords := make([]model.Word, 0, len(terms))

	for i, lt := range lowerTerms {
		w, found := wordMap[lt]
		if !found {
			results[i] = WordValidationResult{Term: terms[i], Status: "not_in_dict"}
			allValid = false
			continue
		}
		if _, used := usedWordIDs[w.ID]; used {
			results[i] = WordValidationResult{Term: terms[i], Status: "already_used"}
			allValid = false
			continue
		}
		results[i] = WordValidationResult{Term: terms[i], Status: "valid"}
		validWords = append(validWords, w)
	}

	if !allValid {
		return &SetupCycleResult{Ok: false, Results: results}, nil
	}

	// 4. All words are valid — create the new cycle and word_cycle rows in a transaction.
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		newCycle := model.Cycle{
			UserID:   userID,
			Wordbook: wordbook,
			Status:   "ongoing",
		}
		if err := tx.Create(&newCycle).Error; err != nil {
			return fmt.Errorf("create cycle: %w", err)
		}

		wordCycles := make([]model.WordCycle, len(validWords))
		for i, w := range validWords {
			wordCycles[i] = model.WordCycle{
				WordID:  w.ID,
				CycleID: newCycle.ID,
				Status:  "new",
			}
		}
		if err := tx.Create(&wordCycles).Error; err != nil {
			return fmt.Errorf("create word_cycles: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &SetupCycleResult{Ok: true, Results: results}, nil
}

// CurrentCycleWord holds a word's term and its status within the ongoing cycle.
type CurrentCycleWord struct {
	Term   string `json:"term"`
	Status string `json:"status"` // new | known
}

// GetCurrentCycleWords returns all words in the user's ongoing cycle.
// Returns nil slice (no error) when no ongoing cycle exists.
func GetCurrentCycleWords(userID, wordbook string) ([]CurrentCycleWord, error) {
	var cycle model.Cycle
	err := db.DB.
		Where("user_id = ? AND wordbook = ? AND status = ?", userID, wordbook, "ongoing").
		First(&cycle).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query ongoing cycle: %w", err)
	}

	type row struct {
		Term   string
		Status string
	}
	var rows []row
	if err := db.DB.
		Table("word_cycles").
		Select("words.term, word_cycles.status").
		Joins("JOIN words ON words.id = word_cycles.word_id").
		Where("word_cycles.cycle_id = ?", cycle.ID).
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("query cycle words: %w", err)
	}

	result := make([]CurrentCycleWord, len(rows))
	for i, r := range rows {
		result[i] = CurrentCycleWord{Term: r.Term, Status: r.Status}
	}
	return result, nil
}

// UpdateCurrentCycleResult is returned by UpdateCurrentCycle.
type UpdateCurrentCycleResult struct {
	Ok      bool                   `json:"ok"`
	Results []WordValidationResult `json:"results"`
}

// UpdateCurrentCycle replaces the word list of the user's ongoing cycle.
//   - Words kept in the new list retain their existing status (new / known).
//   - Words added for the first time are inserted with status "new".
//   - Words removed are deleted from word_cycles.
//   - If after the update all remaining words are "known", the cycle is completed.
//
// "already_used" check only applies to OTHER cycles, not the current one.
func UpdateCurrentCycle(userID, wordbook string, terms []string) (*UpdateCurrentCycleResult, error) {
	if len(terms) < 1 {
		return nil, fmt.Errorf("%w: at least 1 word is required", ErrInvalidCycleInput)
	}

	// Reject duplicates within the submitted list.
	seen := make(map[string]struct{}, len(terms))
	for _, t := range terms {
		k := strings.ToLower(strings.TrimSpace(t))
		if _, dup := seen[k]; dup {
			return nil, fmt.Errorf("%w: duplicate word %q in submission", ErrInvalidCycleInput, t)
		}
		seen[k] = struct{}{}
	}

	// Find the ongoing cycle.
	var cycle model.Cycle
	err := db.DB.
		Where("user_id = ? AND wordbook = ? AND status = ?", userID, wordbook, "ongoing").
		First(&cycle).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: no ongoing cycle to update", ErrInvalidCycleInput)
	}
	if err != nil {
		return nil, fmt.Errorf("query ongoing cycle: %w", err)
	}

	lowerTerms := make([]string, len(terms))
	for i, t := range terms {
		lowerTerms[i] = strings.ToLower(strings.TrimSpace(t))
	}

	// Batch-query words table.
	var matchedWords []model.Word
	if err := db.DB.
		Where("wordbook = ? AND LOWER(term) IN ?", wordbook, lowerTerms).
		Find(&matchedWords).Error; err != nil {
		return nil, fmt.Errorf("query words: %w", err)
	}
	wordMap := make(map[string]model.Word, len(matchedWords))
	for _, w := range matchedWords {
		wordMap[strings.ToLower(w.Term)] = w
	}

	// Words used in OTHER cycles (not the current one).
	usedWordIDs := make(map[uint]struct{})
	var usedIDs []uint
	if err := db.DB.
		Table("word_cycles").
		Select("word_cycles.word_id").
		Joins("JOIN cycles ON cycles.id = word_cycles.cycle_id").
		Where("cycles.user_id = ? AND cycles.wordbook = ? AND cycles.id != ?", userID, wordbook, cycle.ID).
		Scan(&usedIDs).Error; err != nil {
		return nil, fmt.Errorf("query used word ids: %w", err)
	}
	for _, id := range usedIDs {
		usedWordIDs[id] = struct{}{}
	}

	// Validate each term.
	results := make([]WordValidationResult, len(terms))
	allValid := true
	validWords := make([]model.Word, 0, len(terms))
	for i, lt := range lowerTerms {
		w, found := wordMap[lt]
		if !found {
			results[i] = WordValidationResult{Term: terms[i], Status: "not_in_dict"}
			allValid = false
			continue
		}
		if _, used := usedWordIDs[w.ID]; used {
			results[i] = WordValidationResult{Term: terms[i], Status: "already_used"}
			allValid = false
			continue
		}
		results[i] = WordValidationResult{Term: terms[i], Status: "valid"}
		validWords = append(validWords, w)
	}
	if !allValid {
		return &UpdateCurrentCycleResult{Ok: false, Results: results}, nil
	}

	// Load existing word_cycles to preserve status for unchanged words.
	type existingWC struct {
		WordID uint
		Status string
	}
	var existingWCs []existingWC
	if err := db.DB.
		Table("word_cycles").
		Select("word_id, status").
		Where("cycle_id = ?", cycle.ID).
		Scan(&existingWCs).Error; err != nil {
		return nil, fmt.Errorf("query existing word_cycles: %w", err)
	}
	existingMap := make(map[uint]string, len(existingWCs))
	for _, wc := range existingWCs {
		existingMap[wc.WordID] = wc.Status
	}

	// New word ID set.
	newWordIDSet := make(map[uint]struct{}, len(validWords))
	for _, w := range validWords {
		newWordIDSet[w.ID] = struct{}{}
	}

	// Apply changes in a transaction.
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Remove words that are no longer in the list.
		toRemove := make([]uint, 0)
		for id := range existingMap {
			if _, keep := newWordIDSet[id]; !keep {
				toRemove = append(toRemove, id)
			}
		}
		if len(toRemove) > 0 {
			if err := tx.Where("cycle_id = ? AND word_id IN ?", cycle.ID, toRemove).
				Delete(&model.WordCycle{}).Error; err != nil {
				return fmt.Errorf("delete removed word_cycles: %w", err)
			}
		}

		// Add words that are new to this cycle.
		toAdd := make([]model.WordCycle, 0)
		for _, w := range validWords {
			if _, exists := existingMap[w.ID]; !exists {
				toAdd = append(toAdd, model.WordCycle{
					WordID:  w.ID,
					CycleID: cycle.ID,
					Status:  "new",
				})
			}
		}
		if len(toAdd) > 0 {
			if err := tx.Create(&toAdd).Error; err != nil {
				return fmt.Errorf("insert new word_cycles: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// If all remaining words are known, complete the cycle.
	var newCount int64
	if err := db.DB.Model(&model.WordCycle{}).
		Where("cycle_id = ? AND status = ?", cycle.ID, "new").
		Count(&newCount).Error; err != nil {
		return nil, err
	}
	if newCount == 0 {
		var cycleWordCount int64
		db.DB.Model(&model.WordCycle{}).Where("cycle_id = ?", cycle.ID).Count(&cycleWordCount)
		db.DB.Model(&cycle).Update("status", "completed")
		db.DB.Model(&model.UserProgress{}).
			Where("user_id = ? AND wordbook = ?", userID, wordbook).
			UpdateColumn("completed_count", gorm.Expr("completed_count + ?", cycleWordCount))
	}

	return &UpdateCurrentCycleResult{Ok: true, Results: results}, nil
}
