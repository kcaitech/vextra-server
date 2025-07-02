package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyFeedback(db *gorm.DB, userIds []string) []models.Feedback {

	var deletedRecords []models.Feedback
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("feedback").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("feedback").Where("user_id NOT IN (?)", userIds).Delete(&models.Feedback{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.Feedback{}
	}

	return deletedRecords
}
