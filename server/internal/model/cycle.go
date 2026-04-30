package model

import "time"

// Cycle represents a learning cycle for a user in a given wordbook.
// Each cycle contains CycleSize words. Status transitions: ongoing → completed.
type Cycle struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	UserID    string `gorm:"index;not null"`
	Wordbook  string `gorm:"not null"`
	Status    string `gorm:"default:ongoing;not null"` // ongoing | completed
	CreatedAt time.Time
}
