package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyTeamJoinRequestMessageShow(db *gorm.DB, userIds []string) []models.TeamJoinRequestMessageShow {

	var deletedRecords []models.TeamJoinRequestMessageShow
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("team_join_request_message_show").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("team_join_request_message_show").Where("user_id NOT IN (?)", userIds).Delete(&models.TeamJoinRequestMessageShow{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.TeamJoinRequestMessageShow{}
	}

	return deletedRecords
}
