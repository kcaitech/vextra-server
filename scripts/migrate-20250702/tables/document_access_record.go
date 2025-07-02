package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func MigrateDocumentAccessRecord(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error)) {

	// 迁移文档访问记录表
	var oldAccessRecords []struct {
		ID             int64      `gorm:"column:id"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
		UpdatedAt      time.Time  `gorm:"column:updated_at"`
		DeletedAt      *time.Time `gorm:"column:deleted_at"`
		UserId         int64      `gorm:"column:user_id"`
		DocumentId     int64      `gorm:"column:document_id"`
		LastAccessTime time.Time  `gorm:"column:last_access_time"`
	}

	if err := sourceDB.Table("document_access_record").Find(&oldAccessRecords).Error; err != nil {
		log.Fatalf("Error querying document access records: %v", err)
	}

	for _, oldRecord := range oldAccessRecords {

		userId, err := getUserID(oldRecord.UserId)
		if err != nil {
			log.Printf("Error getting user ID for document access record %d: %v", oldRecord.ID, err)
			continue
		}

		newRecord := models.DocumentAccessRecord{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldRecord.CreatedAt,
				UpdatedAt: oldRecord.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:         userId,
			DocumentId:     strconv.FormatInt(oldRecord.DocumentId, 10),
			LastAccessTime: oldRecord.LastAccessTime,
		}

		if oldRecord.DeletedAt != nil {
			newRecord.DeletedAt.Time = *oldRecord.DeletedAt
			newRecord.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "document_access_record", "user_id = ? AND document_id = ?", []interface{}{newRecord.UserId, newRecord.DocumentId}, newRecord); err != nil {
			log.Printf("Error migrating access record %d: %v", oldRecord.ID, err)
		}
	}
}
