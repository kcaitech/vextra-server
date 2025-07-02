package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyProject(db *gorm.DB, removedTeams []string) []models.Project {

	var deletedRecords []models.Project
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("project").Where("team_id IN (?)", removedTeams).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("project").Where("team_id IN (?)", removedTeams).Delete(&models.Project{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.Project{}
	}

	return deletedRecords
}
