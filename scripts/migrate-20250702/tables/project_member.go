package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func MigrateProjectMember(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error)) {
	// 迁移项目成员表
	var oldProjectMembers []struct {
		ID             int64      `gorm:"column:id"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
		UpdatedAt      time.Time  `gorm:"column:updated_at"`
		DeletedAt      *time.Time `gorm:"column:deleted_at"`
		ProjectId      int64      `gorm:"column:project_id"`
		UserId         int64      `gorm:"column:user_id"`
		PermType       uint8      `gorm:"column:perm_type"`
		PermSourceType uint8      `gorm:"column:perm_source_type"`
	}

	if err := sourceDB.Table("project_member").Find(&oldProjectMembers).Error; err != nil {
		log.Fatalf("Error querying project members: %v", err)
	}

	for _, oldMember := range oldProjectMembers {
		userId, err := getUserID(oldMember.UserId)
		if err != nil {
			log.Printf("Error getting user ID for project member %d: %v", oldMember.ID, err)
			continue
		}

		newMember := models.ProjectMember{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldMember.CreatedAt,
				UpdatedAt: oldMember.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			ProjectId: func() string {
				if oldMember.ProjectId == 0 {
					return ""
				}
				return strconv.FormatInt(oldMember.ProjectId, 10)
			}(),
			UserId:         userId,
			PermType:       models.ProjectPermType(oldMember.PermType),
			PermSourceType: models.ProjectPermSourceType(oldMember.PermSourceType),
		}

		if oldMember.DeletedAt != nil {
			newMember.DeletedAt.Time = *oldMember.DeletedAt
			newMember.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "project_member", "project_id = ? AND user_id = ?", []interface{}{newMember.ProjectId, newMember.UserId}, newMember); err != nil {
			log.Printf("Error migrating project member %d: %v", oldMember.ID, err)
		}
	}
}
