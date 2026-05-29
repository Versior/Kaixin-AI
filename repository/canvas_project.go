package repository

import "github.com/basketikun/infinite-canvas/model"

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
	return item, db.Save(&item).Error
}

func DeleteCanvasProject(userID string, id string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Where("user_id = ? AND id = ?", userID, id).Delete(&model.CanvasProject{}).Error
}
