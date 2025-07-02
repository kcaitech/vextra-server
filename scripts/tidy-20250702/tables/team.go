package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyTeam(db *gorm.DB) []models.Team {

	var deletedRecords []models.Team
	err := db.Transaction(func(tx *gorm.DB) error {
		// 查找没有成员的团队
		// 先找到所有有成员的团队ID
		var teamIdsWithMembers []string
		if err := tx.Table("team_member").
			Where("deleted_at IS NULL").
			Select("DISTINCT team_id").
			Pluck("team_id", &teamIdsWithMembers).Error; err != nil {
			return err
		}

		// 查找没有成员的团队
		var query *gorm.DB
		if len(teamIdsWithMembers) > 0 {
			query = tx.Table("team").Where("id NOT IN (?)", teamIdsWithMembers)
		} else {
			// 如果没有任何团队成员，则删除所有团队
			query = tx.Table("team")
		}

		// 先查询要删除的记录
		if err := query.Find(&deletedRecords).Error; err != nil {
			return err
		}

		// 删除没有成员的团队
		if len(deletedRecords) > 0 {
			var teamIdsToDelete []string
			for _, team := range deletedRecords {
				teamIdsToDelete = append(teamIdsToDelete, team.Id)
			}
			if err := tx.Table("team").Where("id IN (?)", teamIdsToDelete).Delete(&models.Team{}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return []models.Team{}
	}

	return deletedRecords
}
