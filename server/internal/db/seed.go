package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
func SeedAccounts() error {
	accounts := []model.Account{
		{Name: "tester"},
		{Name: "admin"},
	}

	for _, acc := range accounts {
		// Check if account already exists
		var existing model.Account
		if err := DB.Where("name = ?", acc.Name).First(&existing).Error; err == nil {
			// Account exists, skip
			continue
		}

		if err := DB.Create(&acc).Error; err != nil {
			return err
		}
	}
	return nil
}
