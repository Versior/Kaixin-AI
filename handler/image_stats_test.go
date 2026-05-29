package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/basketikun/infinite-canvas/config"
	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/repository"
)

func TestImageStatsMasksRankingUsernames(t *testing.T) {
	setupHandlerStatsTestDB(t)
	if _, err := repository.SaveUser(model.User{ID: "user-1", Username: "alice-secret", AffCode: "aff-1"}); err != nil {
		t.Fatalf("save user: %v", err)
	}
	if _, err := repository.SaveGenerationLog(model.GenerationLog{ID: "log-1", UserID: "user-1", Kind: model.GenerationLogKindImage, Status: "success", Images: []string{"a", "b"}, CreatedAt: "2026-05-28T01:00:00Z"}); err != nil {
		t.Fatalf("save log: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/stats", nil)
	res := httptest.NewRecorder()
	AIImageStats(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	if !containsAll(body, "totalImages", "userRanks", "alic") {
		t.Fatalf("unexpected body: %s", body)
	}
	if strings.Contains(body, "alice-secret") {
		t.Fatalf("ranking username was not masked: %s", body)
	}
}

func setupHandlerStatsTestDB(t *testing.T) {
	t.Helper()
	config.Cfg.StorageDriver = "sqlite"
	config.Cfg.DatabaseDSN = ":memory:"
	repository.ResetDBForTest(t)
	if _, err := repository.DB(); err != nil {
		t.Fatalf("db: %v", err)
	}
}

func containsAll(text string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(text, part) {
			return false
		}
	}
	return true
}
