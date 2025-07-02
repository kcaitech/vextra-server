package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyDocumentFavorites(db *gorm.DB, userIds []string) []models.DocumentFavorites {

	var deletedRecords []models.DocumentFavorites

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("document_favorites").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("document_favorites").Where("user_id NOT IN (?)", userIds).Delete(&models.DocumentFavorites{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.DocumentFavorites{}
	}

	return deletedRecords
}
