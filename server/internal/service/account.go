package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
)

type AccountService struct{}

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
	var account model.Account
	if err := db.DB.Where("user_id = ?", model.AdminUserID).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// IsAdmin checks whether a user ID is the canonical administrator.
func (s *AccountService) IsAdmin(userID string) (bool, error) {
	admin, err := s.GetAdmin()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return admin.UserID == userID, nil
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

	account := &model.Account{
		UserID: uuid.NewString(),
		Name:   trimmedName,
	}
	if err := db.DB.Create(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}
