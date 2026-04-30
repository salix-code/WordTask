package model

import "time"

// UserProgress tracks how many words a user has completed in a given wordbook.
// The primary key is (user_id, wordbook) so each user has an independent offset per wordbook.
type UserProgress struct {
	UserID               string    `gorm:"primaryKey"`
	Wordbook             string    `gorm:"primaryKey"`
	CompletedCount       int       `gorm:"default:0;not null"`
	LastSessionDate      time.Time // date of the last daily batch
	LastSessionCompleted int       `gorm:"default:0;not null"` // known-word count at start of last daily batch
}
