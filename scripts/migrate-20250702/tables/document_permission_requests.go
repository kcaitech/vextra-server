package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
	utilTime "kcaitech.com/kcserver/utils/time"
)

func MigrateDocumentPermissionRequests(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error)) {
	// 迁移文档权限申请表
	var oldDocPermRequests []struct {
		ID               int64      `gorm:"column:id"`
		CreatedAt        time.Time  `gorm:"column:created_at"`
		UpdatedAt        time.Time  `gorm:"column:updated_at"`
		DeletedAt        *time.Time `gorm:"column:deleted_at"`
		UserId           int64      `gorm:"column:user_id"`
		DocumentId       int64      `gorm:"column:document_id"`
		PermType         uint8      `gorm:"column:perm_type"`
		Status           uint8      `gorm:"column:status"`
		FirstDisplayedAt time.Time  `gorm:"column:first_displayed_at"`
		ProcessedAt      time.Time  `gorm:"column:processed_at"`
		ProcessedBy      int64      `gorm:"column:processed_by"`
		ApplicantNotes   string     `gorm:"column:applicant_notes"`
		ProcessorNotes   string     `gorm:"column:processor_notes"`
	}

	if err := sourceDB.Table("document_permission_requests").Find(&oldDocPermRequests).Error; err != nil {
		log.Fatalf("Error querying document permission requests: %v", err)
	}

	for _, oldRequest := range oldDocPermRequests {
		// 转换时间类型
		customFirstDisplayedAt := utilTime.Time(oldRequest.FirstDisplayedAt)
		customProcessedAt := utilTime.Time(oldRequest.ProcessedAt)

		userId, err := getUserID(oldRequest.UserId)
		if err != nil {
			log.Printf("Error getting user ID for document permission request %d: %v", oldRequest.ID, err)
			continue
		}

		processedBy, _ := getUserID(oldRequest.ProcessedBy)
		// if err != nil {
		// 	log.Printf("Error getting processed by user ID for document permission request %d: %v", oldRequest.ID, err)
		// 	continue
		// }

		newRequest := models.DocumentPermissionRequests{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldRequest.CreatedAt,
				UpdatedAt: oldRequest.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:           userId,
			DocumentId:       strconv.FormatInt(oldRequest.DocumentId, 10),
			PermType:         models.PermType(oldRequest.PermType),
			Status:           models.StatusType(oldRequest.Status),
			FirstDisplayedAt: customFirstDisplayedAt,
			ProcessedAt:      customProcessedAt,
			ProcessedBy:      processedBy,
			ApplicantNotes:   oldRequest.ApplicantNotes,
			ProcessorNotes:   oldRequest.ProcessorNotes,
		}

		if oldRequest.DeletedAt != nil {
			newRequest.DeletedAt.Time = *oldRequest.DeletedAt
			newRequest.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "document_permission_requests", "document_id = ? AND user_id = ? AND perm_type = ?",
			[]interface{}{newRequest.DocumentId, newRequest.UserId, newRequest.PermType}, newRequest); err != nil {
			log.Printf("Error migrating document permission request %d: %v", oldRequest.ID, err)
		}
	}

}
