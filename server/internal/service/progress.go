package service

import (
	"fmt"

	"gorm.io/gorm"

	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
)

// AdvanceProgress increments the user's completed_count for a wordbook by count.
// If no progress row exists yet, it is created.
func AdvanceProgress(userID, wordbook string, count int) error {
	if count <= 0 {
		return fmt.Errorf("count must be positive, got %d", count)
	}

	result := db.DB.
		Model(&model.UserProgress{}).
		Where("user_id = ? AND wordbook = ?", userID, wordbook).
		UpdateColumn("completed_count", gorm.Expr("completed_count + ?", count))

	if result.Error != nil {
		return result.Error
	}

	// Row didn't exist yet — create it with the initial count.
	if result.RowsAffected == 0 {
		return db.DB.Create(&model.UserProgress{
			UserID:         userID,
			Wordbook:       wordbook,
			CompletedCount: count,
		}).Error
	}

	return nil
}
