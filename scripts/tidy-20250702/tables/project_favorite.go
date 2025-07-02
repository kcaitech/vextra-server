package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyProjectFavorite(db *gorm.DB, userIds []string) []models.ProjectFavorite {

	var deletedRecords []models.ProjectFavorite
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("project_favorite").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("project_favorite").Where("user_id NOT IN (?)", userIds).Delete(&models.ProjectFavorite{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.ProjectFavorite{}
	}

	return deletedRecords
}
