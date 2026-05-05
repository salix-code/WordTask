package model

import "time"

// Cycle represents a learning cycle for a user in a given wordbook.
// Status transitions:
//   - ongoing -> completed
//   - queued -> ongoing (when current ongoing cycle is completed)
type Cycle struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	UserID    string `gorm:"index;not null"`
	Wordbook  string `gorm:"not null"`
	CycleNo   int    `gorm:"column:cycle_no;index;not null;default:0"`
	Status    string `gorm:"default:ongoing;not null"` // ongoing | queued | completed
	CreatedAt time.Time
}
