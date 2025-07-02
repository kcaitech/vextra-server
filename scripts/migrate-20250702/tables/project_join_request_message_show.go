package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
	utilTime "kcaitech.com/kcserver/utils/time"
)

func MigrateProjectJoinRequestMessageShow(sourceDB *gorm.DB, targetDB *gorm.DB) {

	// 迁移项目申请消息表
	var oldMessageShows []struct {
		ID                   int64      `gorm:"column:id"`
		CreatedAt            time.Time  `gorm:"column:created_at"`
		UpdatedAt            time.Time  `gorm:"column:updated_at"`
		DeletedAt            *time.Time `gorm:"column:deleted_at"`
		ProjectJoinRequestId int64      `gorm:"column:project_join_request_id"`
		UserId               int64      `gorm:"column:user_id"`
		ProjectId            int64      `gorm:"column:project_id"`
		FirstDisplayedAt     time.Time  `gorm:"column:first_displayed_at"`
	}

	if err := sourceDB.Table("project_join_request_message_show").Find(&oldMessageShows).Error; err != nil {
		log.Fatalf("Error querying project join request messages: %v", err)
	}

	for _, oldMessage := range oldMessageShows {
		// 转换时间类型
		customFirstDisplayedAt := utilTime.Time(oldMessage.FirstDisplayedAt)

		newMessage := models.ProjectJoinRequestMessageShow{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldMessage.CreatedAt,
				UpdatedAt: oldMessage.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			// 保持 ProjectJoinRequestId 为 int64 类型
			ProjectJoinRequestId: oldMessage.ProjectJoinRequestId,
			// 转换 UserId 和 ProjectId 为 string 类型
			UserId: strconv.FormatInt(oldMessage.UserId, 10),
			ProjectId: func() string {
				if oldMessage.ProjectId == 0 {
					return ""
				}
				return strconv.FormatInt(oldMessage.ProjectId, 10)
			}(),
			FirstDisplayedAt: customFirstDisplayedAt,
		}

		if oldMessage.DeletedAt != nil {
			newMessage.DeletedAt.Time = *oldMessage.DeletedAt
			newMessage.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "project_join_request_message_show", "project_join_request_id = ? AND project_id = ?", []interface{}{newMessage.ProjectJoinRequestId, newMessage.ProjectId}, newMessage); err != nil {
			log.Printf("Error migrating project join request message %d: %v", oldMessage.ID, err)
		}
	}
}
