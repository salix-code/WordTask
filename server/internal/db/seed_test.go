package db

import (
	"testing"

	"wordtask-server/internal/model"
)

func TestSeedAccounts_RenameAdminKeepsGUID(t *testing.T) {
	if err := Init(":memory:"); err != nil {
		t.Fatalf("init db: %v", err)
	}

	if err := SeedAccounts("admin"); err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	var before model.Account
	if err := DB.Where("name = ?", "admin").First(&before).Error; err != nil {
		t.Fatalf("find admin: %v", err)
	}
	if before.UserID != model.AdminUserID {
		t.Fatalf("expected admin user_id to be %s, got %s", model.AdminUserID, before.UserID)
	}

	if err := SeedAccounts("abc"); err != nil {
		t.Fatalf("seed abc: %v", err)
	}

	var after model.Account
	if err := DB.Where("user_id = ?", model.AdminUserID).First(&after).Error; err != nil {
		t.Fatalf("find admin by user_id: %v", err)
	}
	if after.Name != "abc" {
		t.Fatalf("expected renamed admin to be abc, got %s", after.Name)
	}

	var oldNameCount int64
	if err := DB.Model(&model.Account{}).Where("name = ?", "admin").Count(&oldNameCount).Error; err != nil {
		t.Fatalf("count old admin name: %v", err)
	}
	if oldNameCount != 0 {
		t.Fatalf("expected no account named admin after rename, got %d", oldNameCount)
	}
}

func TestSeedAccounts_TesterHasGUID(t *testing.T) {
	if err := Init(":memory:"); err != nil {
		t.Fatalf("init db: %v", err)
	}
	if err := SeedAccounts("admin"); err != nil {
		t.Fatalf("seed accounts: %v", err)
	}

	var tester model.Account
	if err := DB.Where("name = ?", "tester").First(&tester).Error; err != nil {
		t.Fatalf("find tester: %v", err)
	}
	if tester.UserID == "" {
		t.Fatalf("expected tester user_id to be generated")
	}
	if tester.UserID == model.AdminUserID {
		t.Fatalf("expected tester user_id not to equal admin guid")
	}
}
