package model

import "gorm.io/gorm"

// Account 定义了登录账号实体
type Account struct {
	gorm.Model        // 包含 ID (作为 UserId), CreatedAt, UpdatedAt
	Name       string `gorm:"unique;not null"` // 账号名，设为唯一且非空
}
