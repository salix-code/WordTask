package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
)

type AccountService struct{}

const adminAccountIDConfigKey = "admin_account_id"

var (
	ErrAccountNameRequired  = errors.New("account name is required")
	ErrAccountAlreadyExists = errors.New("account already exists")
)

// FindByName finds an account by its name.
// It returns the account if found, otherwise returns an error.
func (s *AccountService) FindByName(name string) (*model.Account, error) {
	var account model.Account
	if err := db.DB.Where("name = ?", name).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// GetAdmin returns the canonical administrator account.
func (s *AccountService) GetAdmin() (*model.Account, error) {
	var cfg model.SystemConfig
	if err := db.DB.Where("key = ?", adminAccountIDConfigKey).First(&cfg).Error; err != nil {
		return nil, err
	}

	adminID, err := strconv.ParseUint(cfg.Value, 10, 64)
	if err != nil || adminID == 0 {
		return nil, fmt.Errorf("invalid %s: %q", adminAccountIDConfigKey, cfg.Value)
	}

	var account model.Account
	if err := db.DB.First(&account, adminID).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// IsAdmin checks whether a user ID is the canonical administrator.
func (s *AccountService) IsAdmin(userID uint) (bool, error) {
	admin, err := s.GetAdmin()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return admin.ID == userID, nil
}

// Create creates a new account with a unique name.
func (s *AccountService) Create(name string) (*model.Account, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, ErrAccountNameRequired
	}

	var existing model.Account
	err := db.DB.Where("name = ?", trimmedName).First(&existing).Error
	if err == nil {
		return nil, ErrAccountAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	account := &model.Account{Name: trimmedName}
	if err := db.DB.Create(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}
