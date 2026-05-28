package service

import (
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/repository"
)

type GenerationTaskList struct {
	Items []model.GenerationTask `json:"items"`
	Total int                    `json:"total"`
}

func CreateGenerationTask(userID string, kind model.GenerationLogKind, modelName string, path string, batchCount int, credits int) (model.GenerationTask, error) {
	if batchCount < 1 {
		batchCount = 1
	}
	task := model.GenerationTask{ID: newID("task"), UserID: userID, Kind: kind, Model: modelName, Path: path, BatchCount: batchCount, Credits: credits, Status: model.GenerationTaskStatusQueued, CreatedAt: now()}
	return repository.SaveGenerationTask(task)
}

func ListGenerationTasks(q model.Query) (GenerationTaskList, error) {
	items, total, err := repository.ListGenerationTasks(q)
	if err != nil {
		return GenerationTaskList{}, err
	}
	return GenerationTaskList{Items: items, Total: int(total)}, nil
}

func MarkGenerationTaskRunning(id string) error {
	return repository.UpdateGenerationTaskStatus(id, model.GenerationTaskStatusRunning, now(), "", "")
}

func CompleteGenerationTask(id string, ok bool, generationLogID string, errMessage string) error {
	status := model.GenerationTaskStatusSucceeded
	if !ok {
		status = model.GenerationTaskStatusFailed
	} else if isPartialSuccessError(errMessage) {
		status = model.GenerationTaskStatusPartialSuccess
	}
	if generationLogID != "" {
		_ = repository.LinkGenerationTaskArtifacts(id, "", generationLogID)
	}
	return repository.UpdateGenerationTaskStatus(id, status, "", now(), errMessage)
}

func isPartialSuccessError(errMessage string) bool {
	return strings.Contains(errMessage, "少返回图片")
}

func CancelGenerationTask(id string, errMessage string) error {
	return repository.UpdateGenerationTaskStatus(id, model.GenerationTaskStatusCancelled, "", now(), errMessage)
}
