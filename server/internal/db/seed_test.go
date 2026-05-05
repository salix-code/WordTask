package db

import (
	"strconv"
	"testing"

	"wordtask-server/internal/model"
)

func TestSeedAccounts_RenameAdminKeepsIdentity(t *testing.T) {
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
	adminIDBefore := before.ID

	if err := SeedAccounts("abc"); err != nil {
		t.Fatalf("seed abc: %v", err)
	}

	var after model.Account
	if err := DB.First(&after, adminIDBefore).Error; err != nil {
		t.Fatalf("find admin by old id: %v", err)
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

	var cfg model.SystemConfig
	if err := DB.Where("key = ?", adminAccountIDConfigKey).First(&cfg).Error; err != nil {
		t.Fatalf("find admin config: %v", err)
	}
	expectedAdminID := strconv.FormatUint(uint64(adminIDBefore), 10)
	if cfg.Value != expectedAdminID {
		t.Fatalf("expected admin_account_id to stay %s, got %s", expectedAdminID, cfg.Value)
	}
}
