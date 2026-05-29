package service

import (
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/repository"
)

func ListCanvasProjects(userID string) ([]model.CanvasProject, error) {
	return repository.ListCanvasProjects(userID)
}

func SaveCanvasProject(userID string, project model.CanvasProject) (model.CanvasProject, error) {
	project.UserID = userID
	project.ID = strings.TrimSpace(project.ID)
	if project.ID == "" {
		project.ID = newID("canvas")
	}
	if strings.TrimSpace(project.CreatedAt) == "" {
		project.CreatedAt = now()
	}
	project.UpdatedAt = now()
	return repository.SaveCanvasProject(project)
}

func DeleteCanvasProject(userID string, id string) error {
	return repository.DeleteCanvasProject(userID, id)
}
