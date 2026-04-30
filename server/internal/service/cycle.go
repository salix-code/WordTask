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
//  1. len(words) must equal CycleSize (30).
//  2. The user must NOT already have an ongoing cycle for this wordbook.
//  3. Each word must exist in the words table (case-insensitive match on term).
//  4. Each word must not have been used in ANY existing cycle
//     (ongoing or completed) for this user + wordbook.
//
// Returns ErrInvalidCycleInput for user-facing validation errors that should
// surface as HTTP 400.
func ValidateAndCreateCycle(userID, wordbook string, terms []string) (*SetupCycleResult, error) {
	if len(terms) != CycleSize {
		return nil, fmt.Errorf("%w: expected %d words, got %d", ErrInvalidCycleInput, CycleSize, len(terms))
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
