package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gorm.io/gorm"
	"wordtask-server/internal/model"
)

// wordbookFiles maps wordbook names to their JSON filenames under wordbookDir.
var wordbookFiles = map[string]string{
	"KET": "ket-a2-key.enriched.json",
}

// jsonWord mirrors the enriched JSON structure for decoding.
type jsonWord struct {
	SourceOrder    int      `json:"sourceOrder"`
	Term           string   `json:"term"`
	PartOfSpeech   string   `json:"partOfSpeech"`
	Examples       []string `json:"examples"`
	Theme          string   `json:"theme"`
	Priority       string   `json:"priority"`
	LearningTarget string   `json:"learningTarget"`
	Translation    string   `json:"translation"`
	Phonetic       string   `json:"phonetic"`
}

// SeedWords seeds the words table from JSON files if the table is empty.
// wordbookDir is the directory containing the wordbook JSON files.
func SeedWords(wordbookDir string) error {
	var count int64
	DB.Model(&model.Word{}).Count(&count)
	if count > 0 {
		return nil // already seeded
	}

	for wordbook, filename := range wordbookFiles {
		path := filepath.Join(wordbookDir, filename)
		if err := seedWordbook(wordbook, path); err != nil {
			return fmt.Errorf("seed %s: %w", wordbook, err)
		}
	}
	return nil
}

func seedWordbook(wordbook, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var raw []jsonWord
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	words := make([]model.Word, 0, len(raw))
	for _, jw := range raw {
		exJSON, _ := json.Marshal(jw.Examples)
		words = append(words, model.Word{
			Wordbook:       wordbook,
			SourceOrder:    jw.SourceOrder,
			Term:           jw.Term,
			Phonetic:       jw.Phonetic,
			PartOfSpeech:   jw.PartOfSpeech,
			Translation:    jw.Translation,
			Examples:       string(exJSON),
			Theme:          jw.Theme,
			Priority:       jw.Priority,
			LearningTarget: jw.LearningTarget,
		})
	}

	// Insert in batches to avoid hitting SQLite variable limits.
	return DB.CreateInBatches(words, 200).Error
}

// SeedAccounts pre-populates the accounts table with some default accounts.
func SeedAccounts(adminAccountName string) error {
	if adminAccountName == "" {
		adminAccountName = defaultAdminAccountName
	}

	if err := ensureAccountByName(defaultTesterAccountName); err != nil {
		return err
	}

	if err := ensureAdminAccount(adminAccountName); err != nil {
		return err
	}

	return nil
}

const (
	defaultTesterAccountName = "tester"
	defaultAdminAccountName  = "admin"
	adminAccountIDConfigKey  = "admin_account_id"
)

func ensureAccountByName(name string) error {
	var existing model.Account
	err := DB.Where("name = ?", name).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return DB.Create(&model.Account{Name: name}).Error
}

func ensureAdminAccount(adminAccountName string) error {
	adminID, err := getAdminAccountID()
	if err != nil {
		return err
	}

	if adminID > 0 {
		var existing model.Account
		err := DB.First(&existing, adminID).Error
		switch {
		case err == nil:
			if existing.Name != adminAccountName {
				if err := DB.Model(&existing).Update("name", adminAccountName).Error; err != nil {
					return err
				}
			}
			return nil
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		}
	}

	var adminAccount model.Account
	err = DB.Where("name = ?", adminAccountName).First(&adminAccount).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		adminAccount = model.Account{Name: adminAccountName}
		if err := DB.Create(&adminAccount).Error; err != nil {
			return err
		}
	}

	return upsertSystemConfig(adminAccountIDConfigKey, strconv.FormatUint(uint64(adminAccount.ID), 10))
}

func getAdminAccountID() (uint, error) {
	var cfg model.SystemConfig
	if err := DB.Where("key = ?", adminAccountIDConfigKey).First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	parsed, err := strconv.ParseUint(cfg.Value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("invalid %s: %q", adminAccountIDConfigKey, cfg.Value)
	}
	return uint(parsed), nil
}

func upsertSystemConfig(key, value string) error {
	var cfg model.SystemConfig
	err := DB.Where("key = ?", key).First(&cfg).Error
	switch {
	case err == nil:
		if cfg.Value == value {
			return nil
		}
		cfg.Value = value
		return DB.Save(&cfg).Error
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return err
	default:
		return DB.Create(&model.SystemConfig{Key: key, Value: value}).Error
	}
}
