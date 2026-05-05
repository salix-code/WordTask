package db

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
	"wordtask-server/internal/model"
)

var wordTableNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func BuildWordTableName(wordbookCode string) string {
	return "word_" + strings.ToLower(strings.TrimSpace(wordbookCode))
}

func ResolveWordTableName(wordbookCode string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(wordbookCode))
	if code == "" {
		return "", fmt.Errorf("wordbook code is required")
	}

	var info model.WordInfo
	err := DB.Where("wordbook = ?", code).First(&info).Error
	switch {
	case err == nil:
		if !wordTableNamePattern.MatchString(info.TableName) {
			return "", fmt.Errorf("invalid word table name %q for wordbook %s", info.TableName, code)
		}
		return info.TableName, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return "", err
	default:
		return BuildWordTableName(code), nil
	}
}

func EnsureWordTable(wordbookCode string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(wordbookCode))
	tableName := BuildWordTableName(code)
	if !wordTableNamePattern.MatchString(tableName) {
		return "", fmt.Errorf("invalid word table name %q", tableName)
	}
	if err := DB.Table(tableName).AutoMigrate(&model.Word{}); err != nil {
		return "", err
	}
	return tableName, nil
}
