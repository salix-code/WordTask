package model

// SystemConfig stores system-level key/value settings.
type SystemConfig struct {
	Key   string `gorm:"primaryKey;size:64"`
	Value string `gorm:"not null"`
}
