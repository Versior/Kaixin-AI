package repository

import (
	"strings"

	"github.com/basketikun/infinite-canvas/model"
)

func SaveGenerationLog(log model.GenerationLog) (model.GenerationLog, error) {
	db, err := DB()
	if err != nil {
		return log, err
	}
	return log, db.Save(&log).Error
}

func ListGenerationLogs(q model.Query) ([]model.GenerationLog, int64, error) {
	db, err := DB()
	if err != nil {
		return nil, 0, err
	}
	q.Normalize()
	tx := db.Model(&model.GenerationLog{}).Select("generation_logs.*, COALESCE(NULLIF(users.display_name, ''), users.username, '-') AS username").Joins("LEFT JOIN users ON users.id = generation_logs.user_id")
	if keyword := strings.TrimSpace(q.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("generation_logs.user_id LIKE ? OR users.username LIKE ? OR users.display_name LIKE ? OR kind LIKE ? OR model LIKE ? OR path LIKE ? OR prompt LIKE ? OR status LIKE ? OR error LIKE ?", like, like, like, like, like, like, like, like, like)
	}
	if isActiveAssetOption(q.Type) {
		tx = tx.Where("kind = ?", q.Type)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.GenerationLog
	err = tx.Order("generation_logs.created_at desc").Offset(q.Offset()).Limit(q.PageSize).Find(&items).Error
	return items, total, err
}

func ListUserGenerationLogs(userID string, q model.Query) ([]model.GenerationLog, int64, error) {
	db, err := DB()
	if err != nil {
		return nil, 0, err
	}
	q.Normalize()
	tx := db.Model(&model.GenerationLog{}).Select("generation_logs.*, COALESCE(NULLIF(users.display_name, ''), users.username, '-') AS username").Joins("LEFT JOIN users ON users.id = generation_logs.user_id").Where("generation_logs.user_id = ?", userID)
	if keyword := strings.TrimSpace(q.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("kind LIKE ? OR model LIKE ? OR path LIKE ? OR prompt LIKE ? OR status LIKE ? OR error LIKE ?", like, like, like, like, like, like)
	}
	if isActiveAssetOption(q.Type) {
		tx = tx.Where("kind = ?", q.Type)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.GenerationLog
	err = tx.Order("generation_logs.created_at desc").Offset(q.Offset()).Limit(q.PageSize).Find(&items).Error
	return items, total, err
}

func DeleteGenerationLog(id string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Delete(&model.GenerationLog{}, "id = ?", id).Error
}

func DeleteGenerationLogs(ids []string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	return db.Delete(&model.GenerationLog{}, "id IN ?", ids).Error
}
