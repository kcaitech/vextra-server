package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyUserKVStorage(db *gorm.DB, userIds []string) []models.UserKVStorage {

	var deletedRecords []models.UserKVStorage
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("user_kv_storage").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("user_kv_storage").Where("user_id NOT IN (?)", userIds).Delete(&models.UserKVStorage{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.UserKVStorage{}
	}

	return deletedRecords
}
