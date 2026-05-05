package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"wordtask-server/internal/model"
	"wordtask-server/internal/wordbook"
)

// ketEnrichedWord mirrors KET enriched JSON structure for decoding.
type ketEnrichedWord struct {
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

type wordImporter func(wordbookCode string, data []byte) ([]model.Word, error)

var importers = map[string]wordImporter{
	"ket_enriched_v1": importKETEnrichedV1,
}

// SeedWords seeds words table per enabled wordbook.
// wordbookDir is the directory containing the wordbook JSON files.
func SeedWords(wordbookDir string) error {
	for _, wb := range wordbook.GetEnabled() {
		tableName, err := EnsureWordTable(wb.Code)
		if err != nil {
			return fmt.Errorf("ensure word table for %s: %w", wb.Code, err)
		}
		if err := upsertWordInfo(wb.Code, tableName); err != nil {
			return fmt.Errorf("upsert word info for %s: %w", wb.Code, err)
		}

		var count int64
		if err := DB.Table(tableName).Where("wordbook = ?", wb.Code).Count(&count).Error; err != nil {
			return fmt.Errorf("count words for %s: %w", wb.Code, err)
		}
		if count > 0 {
			continue
		}

		importer, ok := importers[wb.Importer]
		if !ok {
			return fmt.Errorf("wordbook %s uses unsupported importer %s", wb.Code, wb.Importer)
		}

		path := wb.JSONPath
		if !filepath.IsAbs(path) {
			path = filepath.Join(wordbookDir, path)
		}
		if err := seedWordbook(tableName, wb.Code, path, importer); err != nil {
			return fmt.Errorf("seed %s: %w", wb.Code, err)
		}
	}
	return nil
}

func upsertWordInfo(wordbookCode, tableName string) error {
	var info model.WordInfo
	err := DB.Where("wordbook = ?", wordbookCode).First(&info).Error
	switch {
	case err == nil:
		if info.TableName == tableName {
			return nil
		}
		return DB.Model(&info).Update("table_name", tableName).Error
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return err
	default:
		return DB.Create(&model.WordInfo{
			Wordbook:  wordbookCode,
			TableName: tableName,
		}).Error
	}
}

func seedWordbook(tableName, wordbookCode, path string, importer wordImporter) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	words, err := importer(wordbookCode, data)
	if err != nil {
		return err
	}

	// Insert in batches to avoid hitting SQLite variable limits.
	return DB.Table(tableName).CreateInBatches(words, 200).Error
}

func importKETEnrichedV1(wordbookCode string, data []byte) ([]model.Word, error) {
	var raw []ketEnrichedWord
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	words := make([]model.Word, 0, len(raw))
	for _, jw := range raw {
		exJSON, _ := json.Marshal(jw.Examples)
		words = append(words, model.Word{
			Wordbook:       wordbookCode,
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
	return words, nil
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
	adminAccountIDConfigKey  = "admin_account_id" // legacy compatibility for numeric admin ID
)

func ensureAccountByName(name string) error {
	var existing model.Account
	err := DB.Where("name = ?", name).First(&existing).Error
	if err == nil {
		if existing.UserID == "" {
			return DB.Model(&existing).Update("user_id", uuid.NewString()).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return DB.Create(&model.Account{
		UserID: uuid.NewString(),
		Name:   name,
	}).Error
}

func ensureAdminAccount(adminAccountName string) error {
	var byGUID model.Account
	if err := DB.Where("user_id = ?", model.AdminUserID).First(&byGUID).Error; err == nil {
		if byGUID.Name != adminAccountName {
			return DB.Model(&byGUID).Update("name", adminAccountName).Error
		}
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	legacyAdmin, err := getLegacyAdminByNumericID()
	if err != nil {
		return err
	}
	if legacyAdmin != nil {
		return DB.Model(legacyAdmin).Updates(map[string]interface{}{
			"user_id": model.AdminUserID,
			"name":    adminAccountName,
		}).Error
	}

	var byName model.Account
	if err := DB.Where("name = ?", adminAccountName).First(&byName).Error; err == nil {
		return DB.Model(&byName).Update("user_id", model.AdminUserID).Error
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return DB.Create(&model.Account{
		UserID: model.AdminUserID,
		Name:   adminAccountName,
	}).Error
}

func getLegacyAdminByNumericID() (*model.Account, error) {
	var cfg model.SystemConfig
	if err := DB.Where("key = ?", adminAccountIDConfigKey).First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var account model.Account
	if err := DB.Where("id = ?", cfg.Value).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// EnsureAccountUserIDs fills missing account.user_id and enforces uniqueness at DB level.
func EnsureAccountUserIDs() error {
	var accounts []model.Account
	if err := DB.Where("user_id IS NULL OR user_id = ''").Find(&accounts).Error; err != nil {
		return err
	}

	for _, account := range accounts {
		if err := DB.Model(&account).Update("user_id", uuid.NewString()).Error; err != nil {
			return err
		}
	}

	return DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_user_id_unique ON accounts(user_id)").Error
}
