package service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
	wb "wordtask-server/internal/wordbook"
)

// MarkWordKnown marks a single word as known within the user's active cycle.
// If all words in the cycle become known, the cycle is completed and
// UserProgress.CompletedCount is advanced by the cycle size.
func MarkWordKnown(userID, wordbook string, wordID uint) error {
	if err := wb.ValidateEnabled(wordbook); err != nil {
		return err
	}
	wordTable, err := db.ResolveWordTableName(wordbook)
	if err != nil {
		return err
	}

	// 1. Find the active cycle.
	var cycle model.Cycle
	if err := db.DB.
		Where("user_id = ? AND wordbook = ? AND status = ?", userID, wordbook, "ongoing").
		First(&cycle).Error; err != nil {
		return fmt.Errorf("no active cycle for user %s wordbook %s: %w", userID, wordbook, err)
	}

	// 2. Validate word belongs to this cycle.
	var word model.Word
	if err := db.DB.
		Table(wordTable).
		Where("id = ? AND wordbook = ? AND cycle = ?", wordID, wordbook, cycle.CycleNo).
		First(&word).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: word does not belong to current cycle", ErrInvalidCycleInput)
		}
		return err
	}

	// 3. Seed review progress as "known in learning cycle" if it does not exist.
	if err := db.DB.FirstOrCreate(
		&model.ReviewProgress{},
		model.ReviewProgress{
			UserID:       userID,
			Wordbook:     wordbook,
			WordID:       wordID,
			Repetition:   0,
			IntervalDays: 0,
			EaseFactor:   defaultEaseFactor,
		},
	).Error; err != nil {
		return err
	}

	// 4. Check whether this cycle is fully known for current user.
	var cycleWordCount int64
	if err := db.DB.Table(wordTable).
		Where("wordbook = ? AND cycle = ?", wordbook, cycle.CycleNo).
		Count(&cycleWordCount).Error; err != nil {
		return err
	}
	if cycleWordCount == 0 {
		return nil
	}

	var knownCount int64
	if err := db.DB.
		Model(&model.ReviewProgress{}).
		Distinct("review_progresses.word_id").
		Joins(fmt.Sprintf("JOIN %s ON %s.id = review_progresses.word_id", wordTable, wordTable)).
		Where(fmt.Sprintf("review_progresses.user_id = ? AND review_progresses.wordbook = ? AND %s.wordbook = ? AND %s.cycle = ?", wordTable, wordTable), userID, wordbook, wordbook, cycle.CycleNo).
		Count(&knownCount).Error; err != nil {
		return err
	}

	if knownCount >= cycleWordCount {
		if err := db.DB.Transaction(func(tx *gorm.DB) error {
			// Mark cycle as completed.
			if err := tx.Model(&cycle).Update("status", "completed").Error; err != nil {
				return err
			}

			// Advance UserProgress.CompletedCount.
			if err := tx.Model(&model.UserProgress{}).
				Where("user_id = ? AND wordbook = ?", userID, wordbook).
				UpdateColumn("completed_count", gorm.Expr("completed_count + ?", cycleWordCount)).Error; err != nil {
				return err
			}

			// Auto-activate the next queued cycle (oldest first).
			var nextCycle model.Cycle
			nextErr := tx.
				Where("user_id = ? AND wordbook = ? AND status = ?", userID, wordbook, "queued").
				Order("created_at ASC, id ASC").
				First(&nextCycle).Error
			if nextErr != nil && !errors.Is(nextErr, gorm.ErrRecordNotFound) {
				return nextErr
			}
			if nextErr == nil {
				if err := tx.Model(&nextCycle).Update("status", "ongoing").Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}
