package service

import (
	"fmt"
	"sync"
	"testing"

	"github.com/basketikun/infinite-canvas/config"
	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/repository"
)

func TestRegisterStoresRegisterIPAndBlocksDuplicateIP(t *testing.T) {
	setupAuthTestDB(t)
	first, err := Register("alice", "password-1", "203.0.113.8")
	if err != nil {
		t.Fatalf("first register: %v", err)
	}
	user, ok, err := repository.GetUserByID(first.User.ID)
	if err != nil || !ok {
		t.Fatalf("load first user ok=%v err=%v", ok, err)
	}
	if user.RegisterIP != "203.0.113.8" {
		t.Fatalf("expected register ip stored, got %q", user.RegisterIP)
	}
	if _, err := Register("bob", "password-2", "203.0.113.8"); err == nil || err.Error() != "同一 IP 只允许注册一个账号" {
		t.Fatalf("expected duplicate ip error, got %v", err)
	}
}

func TestRegisterDuplicateIPIsBlockedConcurrently(t *testing.T) {
	setupAuthTestDB(t)
	const workers = 8
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := Register(fmt.Sprintf("ip-user-%d", index), "password-1", "203.0.113.88")
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	succeeded := 0
	for err := range errs {
		if err == nil {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Fatalf("expected exactly one successful registration for same IP, got %d", succeeded)
	}
}

func setupAuthTestDB(t *testing.T) {
	t.Helper()
	config.Cfg.StorageDriver = "sqlite"
	config.Cfg.DatabaseDSN = "file::memory:?cache=shared"
	repository.ResetDBForTest(t)
	if _, err := repository.DB(); err != nil {
		t.Fatalf("db: %v", err)
	}
	if _, err := repository.SaveSettings(model.Settings{Public: model.PublicSetting{Auth: model.PublicAuthSetting{AllowRegister: boolPtr(true)}}}, "2026-05-28T00:00:00Z"); err != nil {
		t.Fatalf("save settings: %v", err)
	}
}

func boolPtr(v bool) *bool { return &v }
