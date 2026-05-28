package repository

import (
	"testing"

	"github.com/basketikun/infinite-canvas/model"
)

func TestGenerationTaskStatsCountsImagesAndRanksUsers(t *testing.T) {
	setupCreditTestDB(t)
	if _, err := SaveUser(model.User{ID: "user-1", Username: "alice", DisplayName: "Alice", AffCode: "aff-1"}); err != nil {
		t.Fatalf("save user-1: %v", err)
	}
	if _, err := SaveUser(model.User{ID: "user-2", Username: "bob", DisplayName: "Bob", AffCode: "aff-2"}); err != nil {
		t.Fatalf("save user-2: %v", err)
	}
	logs := []model.GenerationLog{
		{ID: "old-success", UserID: "user-1", Kind: model.GenerationLogKindImage, Status: "success", Images: []string{"a", "b"}, CreatedAt: "2026-05-27T23:59:59Z"},
		{ID: "today-success", UserID: "user-1", Kind: model.GenerationLogKindImage, Status: "success", Images: []string{"c", "d", "e"}, CreatedAt: "2026-05-28T01:00:00Z"},
		{ID: "today-partial", UserID: "user-2", Kind: model.GenerationLogKindImage, Status: "partial_success", Images: []string{"f"}, CreatedAt: "2026-05-28T02:00:00Z"},
		{ID: "today-failed", UserID: "user-2", Kind: model.GenerationLogKindImage, Status: "failed", Images: nil, CreatedAt: "2026-05-28T03:00:00Z"},
		{ID: "chat-log", UserID: "user-2", Kind: model.GenerationLogKindChat, Status: "success", Images: []string{"ignored"}, CreatedAt: "2026-05-28T04:00:00Z"},
	}
	for _, log := range logs {
		if _, err := SaveGenerationLog(log); err != nil {
			t.Fatalf("save log %s: %v", log.ID, err)
		}
	}

	stats, err := GenerationImageStats("2026-05-28", 10)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.TotalImages != 6 || stats.TodayImages != 4 || stats.SuccessImages != 6 || stats.FailedImages != 0 {
		t.Fatalf("unexpected totals: %+v", stats)
	}
	if len(stats.UserRanks) != 2 {
		t.Fatalf("expected two ranked users: %+v", stats.UserRanks)
	}
	if stats.UserRanks[0].UserID != "user-1" || stats.UserRanks[0].Images != 5 || stats.UserRanks[0].Tasks != 2 {
		t.Fatalf("unexpected first rank: %+v", stats.UserRanks[0])
	}
	if stats.UserRanks[1].UserID != "user-2" || stats.UserRanks[1].Images != 1 || stats.UserRanks[1].Tasks != 1 {
		t.Fatalf("unexpected second rank: %+v", stats.UserRanks[1])
	}
}
