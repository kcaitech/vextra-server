package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyDocumentPermissionRequests(db *gorm.DB, userIds []string) []models.DocumentPermissionRequests {
	var deletedRecords []models.DocumentPermissionRequests

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("document_permission_requests").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("document_permission_requests").Where("user_id NOT IN (?)", userIds).Delete(&models.DocumentPermissionRequests{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.DocumentPermissionRequests{}
	}

	return deletedRecords
}
