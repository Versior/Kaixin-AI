package repository

import (
	"sync"
	"testing"

	"github.com/basketikun/infinite-canvas/config"
	"github.com/basketikun/infinite-canvas/model"
)

func setupCreditTestDB(t *testing.T) {
	t.Helper()
	config.Cfg.StorageDriver = "sqlite"
	config.Cfg.DatabaseDSN = ":memory:"
	db = nil
	dbErr = nil
	dbOnce = sync.Once{}
	t.Cleanup(func() {
		db = nil
		dbErr = nil
		dbOnce = sync.Once{}
	})
}

func TestConsumeUserCreditsWithLogRollsBackWhenLogSaveFails(t *testing.T) {
	setupCreditTestDB(t)
	user := model.User{ID: "user-1", Username: "alice", Credits: 10}
	if _, err := SaveUser(user); err != nil {
		t.Fatalf("save user: %v", err)
	}

	_, ok, err := ConsumeUserCreditsWithLog("user-1", 4, "now", model.CreditLog{ID: "bad-log", UserID: "missing-user", Amount: -4})
	if err == nil {
		t.Fatal("expected log save error")
	}
	if ok {
		t.Fatal("expected operation to fail")
	}

	saved, ok, err := GetUserByID("user-1")
	if err != nil || !ok {
		t.Fatalf("get user: ok=%v err=%v", ok, err)
	}
	if saved.Credits != 10 {
		t.Fatalf("credits changed without log, got %d", saved.Credits)
	}
}

func TestConsumeUserCreditsWithLogCommitsBalanceAndLogTogether(t *testing.T) {
	setupCreditTestDB(t)
	user := model.User{ID: "user-1", Username: "alice", Credits: 10}
	if _, err := SaveUser(user); err != nil {
		t.Fatalf("save user: %v", err)
	}

	updated, ok, err := ConsumeUserCreditsWithLog("user-1", 4, "now", model.CreditLog{ID: "credit-1", UserID: "user-1", Amount: -4, Balance: 6})
	if err != nil || !ok {
		t.Fatalf("consume: ok=%v err=%v", ok, err)
	}
	if updated.Credits != 6 {
		t.Fatalf("updated credits = %d", updated.Credits)
	}

	logs, total, err := ListCreditLogs(model.Query{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list logs: %v", err)
	}
	if total != 1 || len(logs) != 1 {
		t.Fatalf("expected one log, total=%d len=%d", total, len(logs))
	}
	if logs[0].Balance != 6 || logs[0].Amount != -4 {
		t.Fatalf("unexpected log: %+v", logs[0])
	}
}
