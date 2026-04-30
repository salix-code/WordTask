package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
)

// MarkWordKnown marks a single word as 'known' within the user's active cycle.
// If all words in the cycle become known, the cycle is completed and
// UserProgress.CompletedCount is advanced by the actual cycle word count.
func MarkWordKnown(userID, wordbook string, wordID uint) error {
	// 1. Find the active cycle.
	var cycle model.Cycle
	if err := db.DB.
		Where("user_id = ? AND wordbook = ? AND status = ?", userID, wordbook, "ongoing").
		First(&cycle).Error; err != nil {
		return fmt.Errorf("no active cycle for user %s wordbook %s: %w", userID, wordbook, err)
	}

	// 2. Mark the word as known.
	now := time.Now()
	result := db.DB.Model(&model.WordCycle{}).
		Where("cycle_id = ? AND word_id = ?", cycle.ID, wordID).
		Updates(map[string]interface{}{"status": "known", "reviewed_at": now})
	if result.Error != nil {
		return result.Error
	}

	// 3. Check if all words in the cycle are now known.
	var newCount int64
	if err := db.DB.Model(&model.WordCycle{}).
		Where("cycle_id = ? AND status = ?", cycle.ID, "new").
		Count(&newCount).Error; err != nil {
		return err
	}

	if newCount == 0 {
		// Count actual words in this cycle to advance CompletedCount correctly.
		var cycleWordCount int64
		db.DB.Model(&model.WordCycle{}).Where("cycle_id = ?", cycle.ID).Count(&cycleWordCount)

		// Mark cycle as completed.
		if err := db.DB.Model(&cycle).Update("status", "completed").Error; err != nil {
			return err
		}

		// Advance UserProgress.CompletedCount.
		if err := db.DB.Model(&model.UserProgress{}).
			Where("user_id = ? AND wordbook = ?", userID, wordbook).
			UpdateColumn("completed_count", gorm.Expr("completed_count + ?", cycleWordCount)).Error; err != nil {
			return err
		}
	}

	return nil
}
