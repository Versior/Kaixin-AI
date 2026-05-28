package repository

import (
	"strings"

	"github.com/basketikun/infinite-canvas/model"
)

func SaveGenerationTask(task model.GenerationTask) (model.GenerationTask, error) {
	db, err := DB()
	if err != nil {
		return task, err
	}
	return task, db.Save(&task).Error
}

func UpdateGenerationTaskStatus(id string, status model.GenerationTaskStatus, startedAt string, completedAt string, errMessage string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	updates := map[string]any{"status": status}
	if startedAt != "" {
		updates["started_at"] = startedAt
	}
	if completedAt != "" {
		updates["completed_at"] = completedAt
	}
	if errMessage != "" {
		updates["error"] = errMessage
	}
	return db.Model(&model.GenerationTask{}).Where("id = ?", id).Updates(updates).Error
}

func LinkGenerationTaskArtifacts(id string, creditLogID string, generationLogID string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	updates := map[string]any{}
	if creditLogID != "" {
		updates["credit_log_id"] = creditLogID
	}
	if generationLogID != "" {
		updates["generation_log_id"] = generationLogID
	}
	if len(updates) == 0 {
		return nil
	}
	return db.Model(&model.GenerationTask{}).Where("id = ?", id).Updates(updates).Error
}

func ListGenerationTasks(q model.Query) ([]model.GenerationTask, int64, error) {
	db, err := DB()
	if err != nil {
		return nil, 0, err
	}
	q.Normalize()
	tx := db.Model(&model.GenerationTask{}).Select("generation_tasks.*, COALESCE(NULLIF(users.display_name, ''), users.username, '-') AS username").Joins("LEFT JOIN users ON users.id = generation_tasks.user_id")
	if keyword := strings.TrimSpace(q.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("generation_tasks.id LIKE ? OR generation_tasks.user_id LIKE ? OR users.username LIKE ? OR users.display_name LIKE ? OR model LIKE ? OR path LIKE ? OR status LIKE ? OR error LIKE ?", like, like, like, like, like, like, like, like)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.GenerationTask
	err = tx.Order("generation_tasks.created_at desc").Offset(q.Offset()).Limit(q.PageSize).Find(&items).Error
	return items, total, err
}
