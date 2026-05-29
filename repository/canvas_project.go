package repository

import (
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ListCanvasProjects(userID string) ([]model.CanvasProject, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var items []model.CanvasProject
	if err := db.Where("user_id = ?", userID).Order("updated_at desc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func SaveCanvasProject(item model.CanvasProject) (model.CanvasProject, error) {
	db, err := DB()
	if err != nil {
		return item, err
	}
	return item, db.Transaction(func(tx *gorm.DB) error {
		var existing model.CanvasProject
		err := tx.Where("id = ? AND user_id = ?", item.ID, item.UserID).First(&existing).Error
		if err == nil {
			item.CreatedAt = existing.CreatedAt
			return tx.Model(&model.CanvasProject{}).Where("id = ? AND user_id = ?", item.ID, item.UserID).Updates(map[string]any{
				"title":      item.Title,
				"payload":    item.Payload,
				"updated_at": item.UpdatedAt,
			}).Error
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if strings.TrimSpace(item.CreatedAt) == "" {
			item.CreatedAt = item.UpdatedAt
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&item).Error
	})
}

func DeleteCanvasProject(userID string, id string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Where("user_id = ? AND id = ?", userID, id).Delete(&model.CanvasProject{}).Error
}
