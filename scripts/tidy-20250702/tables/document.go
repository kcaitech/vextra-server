package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyDocument(db *gorm.DB, userIds []string) []models.Document {

	var deletedRecords []models.Document
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("document").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("document").Where("user_id NOT IN (?)", userIds).Delete(&models.Document{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.Document{}
	}

	return deletedRecords
}
