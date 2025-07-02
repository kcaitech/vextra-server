package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
	utilTime "kcaitech.com/kcserver/utils/time"
)

func MigrateTeamJoinRequestMessageShow(sourceDB *gorm.DB, targetDB *gorm.DB) {

	// 迁移团队加入申请消息表
	var oldTeamMessageShows []struct {
		ID                int64      `gorm:"column:id"`
		CreatedAt         time.Time  `gorm:"column:created_at"`
		UpdatedAt         time.Time  `gorm:"column:updated_at"`
		DeletedAt         *time.Time `gorm:"column:deleted_at"`
		TeamJoinRequestId int64      `gorm:"column:team_join_request_id"`
		UserId            int64      `gorm:"column:user_id"`
		TeamId            int64      `gorm:"column:team_id"`
		FirstDisplayedAt  time.Time  `gorm:"column:first_displayed_at"`
	}

	if err := sourceDB.Table("team_join_request_message_show").Find(&oldTeamMessageShows).Error; err != nil {
		log.Fatalf("Error querying team join request messages: %v", err)
	}

	for _, oldMessage := range oldTeamMessageShows {
		// 转换时间类型
		customFirstDisplayedAt := utilTime.Time(oldMessage.FirstDisplayedAt)

		newMessage := models.TeamJoinRequestMessageShow{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldMessage.CreatedAt,
				UpdatedAt: oldMessage.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			// TeamJoinRequestId 保持 int64 类型
			TeamJoinRequestId: oldMessage.TeamJoinRequestId,
			// 转换 UserId 和 TeamId 为 string 类型
			UserId: strconv.FormatInt(oldMessage.UserId, 10),
			TeamId: func() string {
				if oldMessage.TeamId == 0 {
					return ""
				}
				return strconv.FormatInt(oldMessage.TeamId, 10)
			}(),
			FirstDisplayedAt: customFirstDisplayedAt,
		}

		if oldMessage.DeletedAt != nil {
			newMessage.DeletedAt.Time = *oldMessage.DeletedAt
			newMessage.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "team_join_request_message_show", "team_join_request_id = ? AND team_id = ?", []interface{}{newMessage.TeamJoinRequestId, newMessage.TeamId}, newMessage); err != nil {
			log.Printf("Error migrating team join request message %d: %v", oldMessage.ID, err)
		}
	}

}
