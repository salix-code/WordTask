package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
	wb "wordtask-server/internal/wordbook"
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

// ValidateAndCreateCycle validates and creates a cycle definition.
// Selected words are assigned by setting the dictionary table's cycle column.
func ValidateAndCreateCycle(userID, wordbook string, terms []string) (*SetupCycleResult, error) {
	if err := wb.ValidateEnabled(wordbook); err != nil {
		return nil, err
	}
	wordTable, err := db.ResolveWordTableName(wordbook)
	if err != nil {
		return nil, err
	}

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

	// Check whether there is already an ongoing cycle.
	var existingOngoing model.Cycle
	err = db.DB.
		Where("user_id = ? AND wordbook = ? AND status = ?", userID, wordbook, "ongoing").
		First(&existingOngoing).Error
	hasOngoing := err == nil
	if !hasOngoing && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("query ongoing cycle: %w", err)
	}

	// Normalize terms to lowercase for matching.
	lowerTerms := make([]string, len(terms))
	for i, t := range terms {
		lowerTerms[i] = strings.ToLower(strings.TrimSpace(t))
	}

	// 1. Batch-query words table (case-insensitive).
	var matchedWords []model.Word
	if err := db.DB.Table(wordTable).
		Where("wordbook = ? AND LOWER(term) IN ?", wordbook, lowerTerms).
		Find(&matchedWords).Error; err != nil {
		return nil, fmt.Errorf("query words: %w", err)
	}

	// Build a map: lower(term) → Word for fast lookup.
	wordMap := make(map[string]model.Word, len(matchedWords))
	for _, w := range matchedWords {
		wordMap[strings.ToLower(w.Term)] = w
	}

	// 2. Collect all word IDs already assigned to any cycle.
	usedWordIDs := make(map[uint]struct{})
	var usedIDs []uint
	if err := db.DB.
		Table(wordTable).
		Where("wordbook = ? AND cycle > 0", wordbook).
		Pluck("id", &usedIDs).Error; err != nil {
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

	// 4. All words are valid — create cycle row and assign words.
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		status := "ongoing"
		if hasOngoing {
			status = "queued"
		}

		var maxWordCycle int
		if err := tx.Table(wordTable).
			Where("wordbook = ?", wordbook).
			Select("COALESCE(MAX(cycle), 0)").
			Scan(&maxWordCycle).Error; err != nil {
			return fmt.Errorf("query max word cycle: %w", err)
		}
		nextCycleNo := maxWordCycle + 1

		newCycle := model.Cycle{
			UserID:   userID,
			Wordbook: wordbook,
			CycleNo:  nextCycleNo,
			Status:   status,
		}
		if err := tx.Create(&newCycle).Error; err != nil {
			return fmt.Errorf("create cycle: %w", err)
		}

		wordIDs := make([]uint, 0, len(validWords))
		for _, w := range validWords {
			wordIDs = append(wordIDs, w.ID)
		}
		if err := tx.Table(wordTable).
			Where("wordbook = ? AND id IN ?", wordbook, wordIDs).
			Update("cycle", nextCycleNo).Error; err != nil {
			return fmt.Errorf("assign words to cycle: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &SetupCycleResult{Ok: true, Results: results}, nil
}

// ensureUserCyclesFromAdmin clones admin-defined cycles for a non-admin user on first use.
// Each user gets an independent copy so progress and review status remain isolated.
func ensureUserCyclesFromAdmin(userID, wordbook string) error {
	if err := wb.ValidateEnabled(wordbook); err != nil {
		return err
	}

	accountService := AccountService{}
	isAdmin, err := accountService.IsAdmin(userID)
	if err != nil {
		return err
	}
	if isAdmin {
		return nil
	}

	var existingCount int64
	if err := db.DB.Model(&model.Cycle{}).
		Where("user_id = ? AND wordbook = ?", userID, wordbook).
		Count(&existingCount).Error; err != nil {
		return err
	}
	if existingCount > 0 {
		return nil
	}

	admin, err := accountService.GetAdmin()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	var templateCycles []model.Cycle
	if err := db.DB.
		Where("user_id = ? AND wordbook = ?", admin.UserID, wordbook).
		Order("created_at ASC, id ASC").
		Find(&templateCycles).Error; err != nil {
		return err
	}
	if len(templateCycles) == 0 {
		return nil
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		for i, templateCycle := range templateCycles {
			status := "queued"
			if i == 0 {
				status = "ongoing"
			}

			newCycle := model.Cycle{
				UserID:   userID,
				Wordbook: wordbook,
				CycleNo:  templateCycle.CycleNo,
				Status:   status,
			}
			if err := tx.Create(&newCycle).Error; err != nil {
				return fmt.Errorf("create mirrored cycle: %w", err)
			}
		}
		return nil
	})
}

// ClearAllCycles deletes every cycle (and their word_cycle rows) for the given
// user + wordbook, and resets the user's completed_count to 0. This allows the
// user to start fresh from scratch.
func ClearAllCycles(userID, wordbook string) error {
	if err := wb.ValidateEnabled(wordbook); err != nil {
		return err
	}
	wordTable, err := db.ResolveWordTableName(wordbook)
	if err != nil {
		return err
	}

	accountService := AccountService{}
	isAdmin, err := accountService.IsAdmin(userID)
	if err != nil {
		return err
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		var cycleIDs []uint
		var cycleNos []int
		if err := tx.Model(&model.Cycle{}).
			Where("user_id = ? AND wordbook = ?", userID, wordbook).
			Pluck("id", &cycleIDs).Error; err != nil {
			return fmt.Errorf("query cycle ids: %w", err)
		}
		if err := tx.Model(&model.Cycle{}).
			Where("user_id = ? AND wordbook = ?", userID, wordbook).
			Pluck("cycle_no", &cycleNos).Error; err != nil {
			return fmt.Errorf("query cycle nos: %w", err)
		}

		if len(cycleIDs) > 0 {
			if err := tx.Where("id IN ?", cycleIDs).
				Delete(&model.Cycle{}).Error; err != nil {
				return fmt.Errorf("delete cycles: %w", err)
			}
		}
		if isAdmin && len(cycleNos) > 0 {
			if err := tx.Table(wordTable).
				Where("wordbook = ? AND cycle IN ?", wordbook, cycleNos).
				Update("cycle", 0).Error; err != nil {
				return fmt.Errorf("clear cycle assignment: %w", err)
			}
		}

		if err := tx.Model(&model.UserProgress{}).
			Where("user_id = ? AND wordbook = ?", userID, wordbook).
			Updates(map[string]interface{}{
				"completed_count":        0,
				"last_session_completed": 0,
			}).Error; err != nil {
			return fmt.Errorf("reset user progress: %w", err)
		}

		if err := tx.Where("user_id = ? AND wordbook = ?", userID, wordbook).
			Delete(&model.ReviewProgress{}).Error; err != nil {
			return fmt.Errorf("clear review progress: %w", err)
		}

		return nil
	})
}

// CurrentCycleWord holds a word's term and its status within the ongoing cycle.
type CurrentCycleWord struct {
	Term   string `json:"term"`
	Status string `json:"status"` // new | known
}

// CycleSummary is a lightweight item for cycle list page.
type CycleSummary struct {
	CycleID    uint      `json:"cycleId"`
	Status     string    `json:"status"` // ongoing | completed
	CreatedAt  time.Time `json:"createdAt"`
	TotalWords int64     `json:"totalWords"`
	KnownWords int64     `json:"knownWords"`
}

// CycleDetail includes cycle metadata and all words under this cycle.
type CycleDetail struct {
	CycleID   uint               `json:"cycleId"`
	Status    string             `json:"status"` // ongoing | completed
	CreatedAt time.Time          `json:"createdAt"`
	Words     []CurrentCycleWord `json:"words"`
}

// ListCycles returns all cycles of the given user+wordbook for management UI.
func ListCycles(userID, wordbook string) ([]CycleSummary, error) {
	if err := ensureUserCyclesFromAdmin(userID, wordbook); err != nil {
		return nil, err
	}
	wordTable, err := db.ResolveWordTableName(wordbook)
	if err != nil {
		return nil, err
	}

	var cycles []model.Cycle
	if err = db.DB.
		Where("user_id = ? AND wordbook = ?", userID, wordbook).
		Order("CASE status WHEN 'ongoing' THEN 0 WHEN 'queued' THEN 1 ELSE 2 END, created_at DESC").
		Find(&cycles).Error; err != nil {
		return nil, fmt.Errorf("list cycles: %w", err)
	}

	result := make([]CycleSummary, 0, len(cycles))
	for _, cycle := range cycles {
		var totalWords int64
		if err := db.DB.Table(wordTable).
			Where("wordbook = ? AND cycle = ?", wordbook, cycle.CycleNo).
			Count(&totalWords).Error; err != nil {
			return nil, fmt.Errorf("count cycle words: %w", err)
		}

		var knownWords int64
		if err := db.DB.Model(&model.ReviewProgress{}).
			Distinct("review_progresses.word_id").
			Joins(fmt.Sprintf("JOIN %s ON %s.id = review_progresses.word_id", wordTable, wordTable)).
			Where(fmt.Sprintf("review_progresses.user_id = ? AND review_progresses.wordbook = ? AND %s.wordbook = ? AND %s.cycle = ?", wordTable, wordTable), userID, wordbook, wordbook, cycle.CycleNo).
			Count(&knownWords).Error; err != nil {
			return nil, fmt.Errorf("count known cycle words: %w", err)
		}

		result = append(result, CycleSummary{
			CycleID:    cycle.ID,
			Status:     cycle.Status,
			CreatedAt:  cycle.CreatedAt,
			TotalWords: totalWords,
			KnownWords: knownWords,
		})
	}
	return result, nil
}

// GetCycleWordsByID returns all words for a specific cycle owned by user+wordbook.
func GetCycleWordsByID(userID, wordbook string, cycleID uint) (*CycleDetail, error) {
	if err := ensureUserCyclesFromAdmin(userID, wordbook); err != nil {
		return nil, err
	}
	wordTable, err := db.ResolveWordTableName(wordbook)
	if err != nil {
		return nil, err
	}

	var cycle model.Cycle
	if err = db.DB.
		Where("id = ? AND user_id = ? AND wordbook = ?", cycleID, userID, wordbook).
		First(&cycle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: cycle not found", ErrInvalidCycleInput)
		}
		return nil, fmt.Errorf("query cycle: %w", err)
	}

	var wordsInCycle []model.Word
	if err := db.DB.Table(wordTable).
		Where("wordbook = ? AND cycle = ?", wordbook, cycle.CycleNo).
		Order("source_order ASC").
		Find(&wordsInCycle).Error; err != nil {
		return nil, fmt.Errorf("query cycle words: %w", err)
	}

	wordIDs := make([]uint, 0, len(wordsInCycle))
	for _, w := range wordsInCycle {
		wordIDs = append(wordIDs, w.ID)
	}
	knownSet := make(map[uint]struct{}, len(wordIDs))
	if len(wordIDs) > 0 {
		var knownIDs []uint
		if err := db.DB.Model(&model.ReviewProgress{}).
			Where("user_id = ? AND wordbook = ? AND word_id IN ?", userID, wordbook, wordIDs).
			Pluck("word_id", &knownIDs).Error; err != nil {
			return nil, fmt.Errorf("query known words: %w", err)
		}
		for _, id := range knownIDs {
			knownSet[id] = struct{}{}
		}
	}

	words := make([]CurrentCycleWord, 0, len(wordsInCycle))
	for _, w := range wordsInCycle {
		status := "new"
		if _, ok := knownSet[w.ID]; ok {
			status = "known"
		}
		words = append(words, CurrentCycleWord{Term: w.Term, Status: status})
	}

	return &CycleDetail{
		CycleID:   cycle.ID,
		Status:    cycle.Status,
		CreatedAt: cycle.CreatedAt,
		Words:     words,
	}, nil
}

// GetCurrentCycleWords returns all words in the user's ongoing cycle.
// Returns nil slice (no error) when no ongoing cycle exists.
func GetCurrentCycleWords(userID, wordbook string) ([]CurrentCycleWord, error) {
	if err := ensureUserCyclesFromAdmin(userID, wordbook); err != nil {
		return nil, err
	}
	wordTable, err := db.ResolveWordTableName(wordbook)
	if err != nil {
		return nil, err
	}

	var cycle model.Cycle
	err = db.DB.
		Where("user_id = ? AND wordbook = ? AND status = ?", userID, wordbook, "ongoing").
		First(&cycle).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query ongoing cycle: %w", err)
	}

	var wordsInCycle []model.Word
	if err := db.DB.Table(wordTable).
		Where("wordbook = ? AND cycle = ?", wordbook, cycle.CycleNo).
		Order("source_order ASC").
		Find(&wordsInCycle).Error; err != nil {
		return nil, fmt.Errorf("query cycle words: %w", err)
	}

	wordIDs := make([]uint, 0, len(wordsInCycle))
	for _, w := range wordsInCycle {
		wordIDs = append(wordIDs, w.ID)
	}
	knownSet := make(map[uint]struct{}, len(wordIDs))
	if len(wordIDs) > 0 {
		var knownIDs []uint
		if err := db.DB.Model(&model.ReviewProgress{}).
			Where("user_id = ? AND wordbook = ? AND word_id IN ?", userID, wordbook, wordIDs).
			Pluck("word_id", &knownIDs).Error; err != nil {
			return nil, fmt.Errorf("query known words: %w", err)
		}
		for _, id := range knownIDs {
			knownSet[id] = struct{}{}
		}
	}

	result := make([]CurrentCycleWord, 0, len(wordsInCycle))
	for _, w := range wordsInCycle {
		status := "new"
		if _, ok := knownSet[w.ID]; ok {
			status = "known"
		}
		result = append(result, CurrentCycleWord{Term: w.Term, Status: status})
	}
	return result, nil
}

// UpdateCurrentCycleResult is returned by UpdateCurrentCycle.
type UpdateCurrentCycleResult struct {
	Ok      bool                   `json:"ok"`
	Results []WordValidationResult `json:"results"`
}

// UpdateCurrentCycle replaces the word list of the user's ongoing cycle by rewriting word table cycle values.
func UpdateCurrentCycle(userID, wordbook string, terms []string) (*UpdateCurrentCycleResult, error) {
	if err := wb.ValidateEnabled(wordbook); err != nil {
		return nil, err
	}
	wordTable, err := db.ResolveWordTableName(wordbook)
	if err != nil {
		return nil, err
	}

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
	err = db.DB.
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
	if err := db.DB.Table(wordTable).
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
		Table(wordTable).
		Where("wordbook = ? AND cycle > 0 AND cycle != ?", wordbook, cycle.CycleNo).
		Pluck("id", &usedIDs).Error; err != nil {
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

	newWordIDs := make([]uint, 0, len(validWords))
	for _, w := range validWords {
		newWordIDs = append(newWordIDs, w.ID)
	}

	// Apply changes in a transaction.
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(wordTable).
			Where("wordbook = ? AND cycle = ?", wordbook, cycle.CycleNo).
			Update("cycle", 0).Error; err != nil {
			return fmt.Errorf("clear current cycle words: %w", err)
		}
		if len(newWordIDs) > 0 {
			if err := tx.Table(wordTable).
				Where("wordbook = ? AND id IN ?", wordbook, newWordIDs).
				Update("cycle", cycle.CycleNo).Error; err != nil {
				return fmt.Errorf("assign updated cycle words: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// If all remaining words are known, complete the cycle.
	var totalWords int64
	if err := db.DB.Table(wordTable).
		Where("wordbook = ? AND cycle = ?", wordbook, cycle.CycleNo).
		Count(&totalWords).Error; err != nil {
		return nil, err
	}
	if totalWords > 0 {
		var knownWords int64
		if err := db.DB.Model(&model.ReviewProgress{}).
			Distinct("review_progresses.word_id").
			Joins(fmt.Sprintf("JOIN %s ON %s.id = review_progresses.word_id", wordTable, wordTable)).
			Where(fmt.Sprintf("review_progresses.user_id = ? AND review_progresses.wordbook = ? AND %s.wordbook = ? AND %s.cycle = ?", wordTable, wordTable), userID, wordbook, wordbook, cycle.CycleNo).
			Count(&knownWords).Error; err != nil {
			return nil, err
		}
		if knownWords >= totalWords {
			db.DB.Model(&cycle).Update("status", "completed")
			db.DB.Model(&model.UserProgress{}).
				Where("user_id = ? AND wordbook = ?", userID, wordbook).
				UpdateColumn("completed_count", gorm.Expr("completed_count + ?", totalWords))
		}
	}

	return &UpdateCurrentCycleResult{Ok: true, Results: results}, nil
}
