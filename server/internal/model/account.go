package model

import "gorm.io/gorm"

const AdminUserID = "00000000-0000-0000-0000-000000000000"

// Account 定义了登录账号实体
type Account struct {
	gorm.Model        // 包含 ID (作为 UserId), CreatedAt, UpdatedAt
	UserID     string `gorm:"size:36;index"`   // 对外通信使用的 GUID
	Name       string `gorm:"unique;not null"` // 账号名，设为唯一且非空
}
