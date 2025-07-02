package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
	utilTime "kcaitech.com/kcserver/utils/time"
)

func MigrateTeam(sourceDB *gorm.DB, targetDB *gorm.DB) {

	// 迁移团队表
	var oldTeams []struct {
		ID              int64      `gorm:"column:id"`
		CreatedAt       time.Time  `gorm:"column:created_at"`
		UpdatedAt       time.Time  `gorm:"column:updated_at"`
		DeletedAt       *time.Time `gorm:"column:deleted_at"`
		Name            string     `gorm:"column:name"`
		Description     string     `gorm:"column:description"`
		Avatar          string     `gorm:"column:avatar"`
		Uid             string     `gorm:"column:uid"`
		InvitedPermType uint8      `gorm:"column:invited_perm_type"`
		InvitedSwitch   bool       `gorm:"column:invited_switch"`
	}

	if err := sourceDB.Table("team").Find(&oldTeams).Error; err != nil {
		log.Fatalf("Error querying teams: %v", err)
	}
	log.Println("oldTeams 长度", len(oldTeams))
	for _, oldTeam := range oldTeams {
		// 转换时间类型
		customCreatedAt := utilTime.Time(oldTeam.CreatedAt)
		customUpdatedAt := utilTime.Time(oldTeam.UpdatedAt)

		newTeam := models.Team{
			Id:              strconv.FormatInt(oldTeam.ID, 10),
			CreatedAt:       customCreatedAt,
			UpdatedAt:       customUpdatedAt,
			DeletedAt:       models.DeletedAt{},
			Name:            oldTeam.Name,
			Description:     oldTeam.Description,
			Avatar:          oldTeam.Avatar,
			InvitedPermType: models.TeamPermType(oldTeam.InvitedPermType),
			OpenInvite:      oldTeam.InvitedSwitch,
		}

		if oldTeam.DeletedAt != nil {
			newTeam.DeletedAt.Time = *oldTeam.DeletedAt
			newTeam.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "team", "id = ?", newTeam.Id, newTeam); err != nil {
			log.Printf("Error migrating team %d: %v", oldTeam.ID, err)
		}
	}
}
