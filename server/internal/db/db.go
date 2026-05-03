package db

import (
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"wordtask-server/internal/model"
)

// DB is the global GORM database instance.
var DB *gorm.DB

// Init opens the SQLite database at dsn, creates the directory if needed,
// and auto-migrates all tables.
func Init(dsn string) error {
	if err := os.MkdirAll(filepath.Dir(dsn), 0755); err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&model.Account{}, &model.Word{}, &model.UserProgress{}, &model.Cycle{}, &model.WordCycle{}); err != nil {
		return err
	}

	DB = db
	return nil
}
