package service

import (
	"wordtask-server/internal/db"
	"wordtask-server/internal/model"
)

type AccountService struct{}

// FindByName finds an account by its name.
// It returns the account if found, otherwise returns an error.
func (s *AccountService) FindByName(name string) (*model.Account, error) {
	var account model.Account
	if err := db.DB.Where("name = ?", name).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}
