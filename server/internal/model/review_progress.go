package model

import "time"

// ReviewProgress stores SM-2 scheduling state for one word of a user/wordbook.
type ReviewProgress struct {
	UserID         string    `gorm:"primaryKey"`
	Wordbook       string    `gorm:"primaryKey"`
	WordID         uint      `gorm:"primaryKey"`
	Repetition     int       `gorm:"default:0;not null"`
	IntervalDays   int       `gorm:"default:0;not null"`
	EaseFactor     float64   `gorm:"default:2.5;not null"`
	DueAt          time.Time `gorm:"index;not null"`
	LastReviewedAt *time.Time
}
