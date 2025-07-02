package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
	utilTime "kcaitech.com/kcserver/utils/time"
)

func MigrateProjectJoinRequest(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error)) {
	// 迁移项目申请表
	var oldProjectJoinRequests []struct {
		ID               int64      `gorm:"column:id"`
		CreatedAt        time.Time  `gorm:"column:created_at"`
		UpdatedAt        time.Time  `gorm:"column:updated_at"`
		DeletedAt        *time.Time `gorm:"column:deleted_at"`
		UserId           int64      `gorm:"column:user_id"`
		ProjectId        int64      `gorm:"column:project_id"`
		PermType         uint8      `gorm:"column:perm_type"`
		Status           uint8      `gorm:"column:status"`
		FirstDisplayedAt time.Time  `gorm:"column:first_displayed_at"`
		ProcessedAt      time.Time  `gorm:"column:processed_at"`
		ProcessedBy      int64      `gorm:"column:processed_by"`
		ApplicantNotes   string     `gorm:"column:applicant_notes"`
		ProcessorNotes   string     `gorm:"column:processor_notes"`
	}

	if err := sourceDB.Table("project_join_request").Find(&oldProjectJoinRequests).Error; err != nil {
		log.Fatalf("Error querying project join requests: %v", err)
	}

	for _, oldRequest := range oldProjectJoinRequests {
		// 转换时间类型
		customFirstDisplayedAt := utilTime.Time(oldRequest.FirstDisplayedAt)
		customProcessedAt := utilTime.Time(oldRequest.ProcessedAt)

		userId, err := getUserID(oldRequest.UserId)
		if err != nil {
			log.Printf("Error getting user ID for project join request %d: %v", oldRequest.ID, err)
			continue
		}

		processedBy, _ := getUserID(oldRequest.ProcessedBy)

		newRequest := models.ProjectJoinRequest{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldRequest.CreatedAt,
				UpdatedAt: oldRequest.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId: userId,
			ProjectId: func() string {
				if oldRequest.ProjectId == 0 {
					return ""
				}
				return strconv.FormatInt(oldRequest.ProjectId, 10)
			}(),
			PermType:         models.ProjectPermType(oldRequest.PermType),
			Status:           models.ProjectJoinRequestStatus(oldRequest.Status),
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

		if err := CheckAndUpdate(targetDB, "project_join_request", "user_id = ? AND project_id = ?", []interface{}{newRequest.UserId, newRequest.ProjectId}, newRequest); err != nil {
			log.Printf("Error migrating project join request %d: %v", oldRequest.ID, err)
		}
	}
}
