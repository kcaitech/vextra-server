package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyProjectJoinRequest(db *gorm.DB, userIds []string) []models.ProjectJoinRequest {

	var deletedRecords []models.ProjectJoinRequest
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("project_join_request").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("project_join_request").Where("user_id NOT IN (?)", userIds).Delete(&models.ProjectJoinRequest{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.ProjectJoinRequest{}
	}

	return deletedRecords
}
