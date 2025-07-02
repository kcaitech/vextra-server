package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
	utilTime "kcaitech.com/kcserver/utils/time"
)

func MigrateProject(sourceDB *gorm.DB, targetDB *gorm.DB) {
	// 迁移项目表
	var oldProjects []struct {
		ID            int64      `gorm:"column:id"`
		CreatedAt     time.Time  `gorm:"column:created_at"`
		UpdatedAt     time.Time  `gorm:"column:updated_at"`
		DeletedAt     *time.Time `gorm:"column:deleted_at"`
		TeamId        int64      `gorm:"column:team_id"`
		Name          string     `gorm:"column:name"`
		Description   string     `gorm:"column:description"`
		PublicSwitch  bool       `gorm:"column:public_switch"`
		PermType      uint8      `gorm:"column:perm_type"`
		InvitedSwitch bool       `gorm:"column:invited_switch"`
		NeedApproval  bool       `gorm:"column:need_approval"`
	}

	if err := sourceDB.Table("project").Find(&oldProjects).Error; err != nil {
		log.Fatalf("Error querying projects: %v", err)
	}
	log.Println("oldProjects 长度", len(oldProjects))
	for _, oldProject := range oldProjects {
		// 转换时间类型
		customCreatedAt := utilTime.Time(oldProject.CreatedAt)
		customUpdatedAt := utilTime.Time(oldProject.UpdatedAt)

		newProject := models.Project{
			Id:        strconv.FormatInt(oldProject.ID, 10),
			CreatedAt: customCreatedAt,
			UpdatedAt: customUpdatedAt,
			DeletedAt: models.DeletedAt{},
			TeamId: func() string {
				if oldProject.TeamId == 0 {
					return ""
				}
				return strconv.FormatInt(oldProject.TeamId, 10)
			}(),
			Name:         oldProject.Name,
			Description:  oldProject.Description,
			IsPublic:     oldProject.PublicSwitch,
			PermType:     models.ProjectPermType(oldProject.PermType),
			OpenInvite:   oldProject.InvitedSwitch,
			NeedApproval: oldProject.NeedApproval,
		}

		if oldProject.DeletedAt != nil {
			newProject.DeletedAt.Time = *oldProject.DeletedAt
			newProject.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "project", "id = ?", newProject.Id, newProject); err != nil {
			log.Printf("Error migrating project %d: %v", oldProject.ID, err)
		}
	}
}
