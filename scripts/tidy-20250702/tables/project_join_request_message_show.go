package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyProjectJoinRequestMessageShow(db *gorm.DB, userIds []string) []models.ProjectJoinRequestMessageShow {

	var deletedRecords []models.ProjectJoinRequestMessageShow
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("project_join_request_message_show").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("project_join_request_message_show").Where("user_id NOT IN (?)", userIds).Delete(&models.ProjectJoinRequestMessageShow{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.ProjectJoinRequestMessageShow{}
	}

	return deletedRecords
}
