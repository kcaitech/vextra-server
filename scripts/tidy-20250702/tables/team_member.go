package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyTeamMember(db *gorm.DB, userIds []string) []models.TeamMember {

	var deletedRecords []models.TeamMember
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("team_member").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("team_member").Where("user_id NOT IN (?)", userIds).Delete(&models.TeamMember{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.TeamMember{}
	}

	return deletedRecords
}
