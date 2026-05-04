package model

import "time"

// WordCycle links a word to a cycle and tracks its learning status within that cycle.
// The composite (word_id, cycle_id) is unique.
type WordCycle struct {
	WordID     uint   `gorm:"primaryKey;uniqueIndex:idx_word_cycle"`
	CycleID    uint   `gorm:"primaryKey;uniqueIndex:idx_word_cycle"`
	SortOrder  int    `gorm:"not null;default:0"`   // input order within this cycle
	Status     string `gorm:"default:new;not null"` // new | known
	ReviewedAt *time.Time
}
