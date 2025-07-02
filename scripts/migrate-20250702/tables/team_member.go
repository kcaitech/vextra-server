package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func MigrateTeamMember(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error)) {

	// 迁移团队成员表
	var oldTeamMembers []struct {
		ID        int64      `gorm:"column:id"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		UpdatedAt time.Time  `gorm:"column:updated_at"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
		TeamId    int64      `gorm:"column:team_id"`
		UserId    int64      `gorm:"column:user_id"`
		PermType  uint8      `gorm:"column:perm_type"`
		Nickname  string     `gorm:"column:nickname"`
	}

	if err := sourceDB.Table("team_member").Find(&oldTeamMembers).Error; err != nil {
		log.Fatalf("Error querying team members: %v", err)
	}

	for _, oldMember := range oldTeamMembers {

		userId, err := getUserID(oldMember.UserId)
		if err != nil {
			log.Printf("Error getting user ID for team member %d: %v", oldMember.ID, err)
			continue
		}

		newMember := models.TeamMember{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldMember.CreatedAt,
				UpdatedAt: oldMember.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			TeamId: func() string {
				if oldMember.TeamId == 0 {
					return ""
				}
				return strconv.FormatInt(oldMember.TeamId, 10)
			}(),
			UserId:   userId,
			PermType: models.TeamPermType(oldMember.PermType),
			Nickname: oldMember.Nickname,
		}

		if oldMember.DeletedAt != nil {
			newMember.DeletedAt.Time = *oldMember.DeletedAt
			newMember.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "team_member", "team_id = ? AND user_id = ?", []interface{}{newMember.TeamId, newMember.UserId}, newMember); err != nil {
			log.Printf("Error migrating team member %d: %v", oldMember.ID, err)
		}
	}
}
