package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyProjectMember(db *gorm.DB, userIds []string) []models.ProjectMember {

	var deletedRecords []models.ProjectMember
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("project_member").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("project_member").Where("user_id NOT IN (?)", userIds).Delete(&models.ProjectMember{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.ProjectMember{}
	}

	return deletedRecords
}
