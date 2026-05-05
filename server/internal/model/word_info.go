package model

// WordInfo stores dictionary-to-table mapping.
type WordInfo struct {
	Wordbook  string `gorm:"primaryKey;size:32"`
	TableName string `gorm:"not null"`
}
