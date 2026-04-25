package model

// Word represents a vocabulary entry stored in the database.
type Word struct {
	ID             uint   `gorm:"primaryKey;autoIncrement"`
	Wordbook       string `gorm:"index;not null"`
	SourceOrder    int    `gorm:"not null"`
	Term           string `gorm:"not null"`
	Phonetic       string
	PartOfSpeech   string
	Translation    string
	Examples       string // JSON-encoded []string
	Theme          string
	Priority       string
	LearningTarget string
}
