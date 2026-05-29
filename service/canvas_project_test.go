package service

import (
	"testing"

	"github.com/basketikun/infinite-canvas/config"
	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/repository"
)

func TestCanvasProjectsAreScopedByUserAndUpserted(t *testing.T) {
	setupCanvasProjectTestDB(t)
	if _, err := repository.SaveUser(model.User{ID: "user-1", Username: "alice", AffCode: "aff-1"}); err != nil {
		t.Fatalf("save user: %v", err)
	}
	if _, err := repository.SaveUser(model.User{ID: "user-2", Username: "bob", AffCode: "aff-2"}); err != nil {
		t.Fatalf("save other user: %v", err)
	}

	project := model.CanvasProject{
		ID:        "canvas-1",
		UserID:    "user-1",
		Title:     "灵感 A",
		Payload:   `{"nodes":[{"id":"node-1"}]}`,
		CreatedAt: "2026-05-29T00:00:00Z",
		UpdatedAt: "2026-05-29T00:00:00Z",
	}
	if _, err := SaveCanvasProject("user-1", project); err != nil {
		t.Fatalf("save project: %v", err)
	}
	project.Title = "灵感 A+"
	project.Payload = `{"nodes":[{"id":"node-2"}]}`
	if _, err := SaveCanvasProject("user-1", project); err != nil {
		t.Fatalf("upsert project: %v", err)
	}

	items, err := ListCanvasProjects("user-1")
	if err != nil {
		t.Fatalf("list user projects: %v", err)
	}
	if len(items) != 1 || items[0].Title != "灵感 A+" || items[0].Payload != project.Payload {
		t.Fatalf("unexpected user projects: %#v", items)
	}

	other, err := ListCanvasProjects("user-2")
	if err != nil {
		t.Fatalf("list other projects: %v", err)
	}
	if len(other) != 0 {
		t.Fatalf("project leaked to another user: %#v", other)
	}
}

func TestDeleteCanvasProjectIsScopedByUser(t *testing.T) {
	setupCanvasProjectTestDB(t)
	if _, err := repository.SaveUser(model.User{ID: "user-1", Username: "alice", AffCode: "aff-1"}); err != nil {
		t.Fatalf("save user: %v", err)
	}
	if _, err := repository.SaveUser(model.User{ID: "user-2", Username: "bob", AffCode: "aff-2"}); err != nil {
		t.Fatalf("save other user: %v", err)
	}
	if _, err := SaveCanvasProject("user-1", model.CanvasProject{ID: "canvas-1", Title: "A", Payload: `{}`}); err != nil {
		t.Fatalf("save user project: %v", err)
	}
	if _, err := SaveCanvasProject("user-2", model.CanvasProject{ID: "canvas-1", Title: "B", Payload: `{}`}); err != nil {
		t.Fatalf("save other project: %v", err)
	}
	if err := DeleteCanvasProject("user-1", "canvas-1"); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	items, err := ListCanvasProjects("user-2")
	if err != nil {
		t.Fatalf("list other projects: %v", err)
	}
	if len(items) != 1 || items[0].UserID != "user-2" {
		t.Fatalf("other user's project should remain: %#v", items)
	}
}

func setupCanvasProjectTestDB(t *testing.T) {
	t.Helper()
	config.Cfg.StorageDriver = "sqlite"
	config.Cfg.DatabaseDSN = ":memory:"
	repository.ResetDBForTest(t)
	if _, err := repository.DB(); err != nil {
		t.Fatalf("db: %v", err)
	}
}
