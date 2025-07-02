package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
	utilTime "kcaitech.com/kcserver/utils/time"
)

func MigrateTeamJoinRequest(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error)) {

	// 迁移团队加入申请表
	var oldTeamJoinRequests []struct {
		ID               int64      `gorm:"column:id"`
		CreatedAt        time.Time  `gorm:"column:created_at"`
		UpdatedAt        time.Time  `gorm:"column:updated_at"`
		DeletedAt        *time.Time `gorm:"column:deleted_at"`
		UserId           int64      `gorm:"column:user_id"`
		TeamId           int64      `gorm:"column:team_id"`
		PermType         uint8      `gorm:"column:perm_type"`
		Status           uint8      `gorm:"column:status"`
		FirstDisplayedAt time.Time  `gorm:"column:first_displayed_at"`
		ProcessedAt      time.Time  `gorm:"column:processed_at"`
		ProcessedBy      int64      `gorm:"column:processed_by"`
		ApplicantNotes   string     `gorm:"column:applicant_notes"`
		ProcessorNotes   string     `gorm:"column:processor_notes"`
	}

	if err := sourceDB.Table("team_join_request").Find(&oldTeamJoinRequests).Error; err != nil {
		log.Fatalf("Error querying team join requests: %v", err)
	}

	for _, oldRequest := range oldTeamJoinRequests {
		// 转换时间类型
		customFirstDisplayedAt := utilTime.Time(oldRequest.FirstDisplayedAt)
		customProcessedAt := utilTime.Time(oldRequest.ProcessedAt)

		userId, err := getUserID(oldRequest.UserId)
		if err != nil {
			log.Printf("Error getting user ID for team join request %d: %v", oldRequest.ID, err)
			continue
		}

		processedBy, _ := getUserID(oldRequest.ProcessedBy)

		newRequest := models.TeamJoinRequest{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldRequest.CreatedAt,
				UpdatedAt: oldRequest.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:           userId,
			TeamId:           strconv.FormatInt(oldRequest.TeamId, 10),
			PermType:         models.TeamPermType(oldRequest.PermType),
			Status:           models.TeamJoinRequestStatus(oldRequest.Status),
			FirstDisplayedAt: customFirstDisplayedAt,
			ProcessedAt:      customProcessedAt,
			ProcessedBy:      processedBy,
			ApplicantNotes:   oldRequest.ApplicantNotes,
			ProcessorNotes:   oldRequest.ProcessorNotes,
		}

		// 处理空ID
		if oldRequest.ProcessedBy == 0 {
			newRequest.ProcessedBy = ""
		}

		// 设置DeletedAt
		if oldRequest.DeletedAt != nil {
			newRequest.DeletedAt.Time = *oldRequest.DeletedAt
			newRequest.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "team_join_request", "user_id = ? AND team_id = ?", []interface{}{newRequest.UserId, newRequest.TeamId}, newRequest); err != nil {
			log.Printf("Error migrating team join request %d: %v", oldRequest.ID, err)
		}
	}
}
