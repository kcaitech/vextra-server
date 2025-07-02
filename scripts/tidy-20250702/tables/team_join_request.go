package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyTeamJoinRequest(db *gorm.DB, userIds []string) []models.TeamJoinRequest {

	var deletedRecords []models.TeamJoinRequest
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("team_join_request").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("team_join_request").Where("user_id NOT IN (?)", userIds).Delete(&models.TeamJoinRequest{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.TeamJoinRequest{}
	}

	return deletedRecords
}
