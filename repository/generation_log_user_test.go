package repository

import (
	"testing"

	"github.com/basketikun/infinite-canvas/model"
)

func TestListUserGenerationLogsFiltersByCurrentUser(t *testing.T) {
	setupCreditTestDB(t)
	if _, err := SaveUser(model.User{ID: "user-1", Username: "alice", AffCode: "aff-1"}); err != nil {
		t.Fatalf("save user-1: %v", err)
	}
	if _, err := SaveUser(model.User{ID: "user-2", Username: "bob", AffCode: "aff-2"}); err != nil {
		t.Fatalf("save user-2: %v", err)
	}
	_, _ = SaveGenerationLog(model.GenerationLog{ID: "log-old", UserID: "user-1", Kind: model.GenerationLogKindImage, Prompt: "old", Images: []string{"https://example.com/old.png"}, Status: "success", CreatedAt: "2026-05-27T00:00:00Z"})
	_, _ = SaveGenerationLog(model.GenerationLog{ID: "log-mine", UserID: "user-1", Kind: model.GenerationLogKindImage, Prompt: "mine", Images: []string{"https://example.com/mine.png"}, Status: "success", CreatedAt: "2026-05-28T00:00:00Z"})
	_, _ = SaveGenerationLog(model.GenerationLog{ID: "log-other", UserID: "user-2", Kind: model.GenerationLogKindImage, Prompt: "other", Images: []string{"https://example.com/other.png"}, Status: "success", CreatedAt: "2026-05-29T00:00:00Z"})

	logs, total, err := ListUserGenerationLogs("user-1", model.Query{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list user logs: %v", err)
	}
	if total != 2 || len(logs) != 2 {
		t.Fatalf("expected two logs for user-1, total=%d len=%d logs=%+v", total, len(logs), logs)
	}
	if logs[0].ID != "log-mine" || logs[1].ID != "log-old" {
		t.Fatalf("expected user logs newest first, got %+v", logs)
	}
	for _, item := range logs {
		if item.UserID != "user-1" {
			t.Fatalf("leaked another user's log: %+v", item)
		}
		if item.Request != "" || item.Response != "" {
			t.Fatalf("user history must not expose raw request/response: %+v", item)
		}
	}
}
