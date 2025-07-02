package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/scripts/migrate-20250702/config"
	"kcaitech.com/kcserver/scripts/migrate-20250702/storage"
)

func MigrateDocument(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error), config config.Config) {
	// 迁移文档表
	var oldDocuments []struct {
		ID           int64      `gorm:"column:id"`
		CreatedAt    time.Time  `gorm:"column:created_at"`
		UpdatedAt    time.Time  `gorm:"column:updated_at"`
		DeletedAt    *time.Time `gorm:"column:deleted_at"`
		UserId       int64      `gorm:"column:user_id"`
		Path         string     `gorm:"column:path"`
		DocType      uint8      `gorm:"column:doc_type"`
		Name         string     `gorm:"column:name"`
		Size         uint64     `gorm:"column:size"`
		PurgedAt     time.Time  `gorm:"column:purged_at"`
		DeleteBy     int64      `gorm:"column:delete_by"`
		VersionId    string     `gorm:"column:version_id"`
		TeamId       int64      `gorm:"column:team_id"`
		ProjectId    int64      `gorm:"column:project_id"`
		LockedAt     time.Time  `gorm:"column:locked_at"`
		LockedReason string     `gorm:"column:locked_reason"`
		LockedWords  string     `gorm:"column:locked_words"`
	}
	if err := sourceDB.Table("document").Where("deleted_at IS NULL AND version_id IS NOT NULL AND version_id != ''").Find(&oldDocuments).Error; err != nil {
		log.Fatalf("Error querying documents: %v", err)
	}

	log.Printf("实际查询到符合条件的文档数: %d", len(oldDocuments))

	// var documentIds []int64
	for _, oldDoc := range oldDocuments {
		// documentIds = append(documentIds, oldDoc.ID)
		// 创建新文档记录
		// if oldDoc.Name == "腾讯TDesign 桌面端组件.sketch" || oldDoc.Name == "Ant Design Open Source (Community).fig" ||
		// 	oldDoc.Name == "腾讯TDesign 桌面端组件" || oldDoc.Name == "Ant Design Open Source (Community)" { //这个文件会导致服务崩溃
		// 	continue
		// }

		userId, err := getUserID(oldDoc.UserId)
		if err != nil {
			// log.Printf("Error getting user ID for document %d: %v", oldDoc.ID, err)
			continue
		}

		deleteBy, _ := getUserID(oldDoc.DeleteBy)
		// if err != nil {
		// 	log.Printf("Error getting delete by user ID for document %d: %v", oldDoc.ID, err)
		// 	continue
		// }

		newDoc := models.Document{
			Id:        strconv.FormatInt(oldDoc.ID, 10),
			CreatedAt: oldDoc.CreatedAt,
			UpdatedAt: oldDoc.UpdatedAt,
			DeletedAt: models.DeletedAt{},
			UserId:    userId,
			Path:      oldDoc.Path,
			DocType:   models.DocType(oldDoc.DocType),
			Name:      oldDoc.Name,
			Size:      oldDoc.Size,
			DeleteBy:  deleteBy,
			VersionId: oldDoc.VersionId,
			TeamId: func() string {
				if oldDoc.TeamId == 0 {
					return ""
				}
				return strconv.FormatInt(oldDoc.TeamId, 10)
			}(),
			ProjectId: func() string {
				if oldDoc.ProjectId == 0 {
					return ""
				}
				return strconv.FormatInt(oldDoc.ProjectId, 10)
			}(),
		}

		// 设置DeletedAt
		if oldDoc.DeletedAt != nil {
			newDoc.DeletedAt.Time = *oldDoc.DeletedAt
			newDoc.DeletedAt.Valid = true
		}

		// 检查并更新文档记录
		if err := CheckAndUpdate(targetDB, "document", "id = ?", newDoc.Id, newDoc); err != nil {
			log.Printf("Error migrating document %d: %v", oldDoc.ID, err)
			continue
		}

		// 迁移文档数据
		err = storage.MigrateDocumentStorage(oldDoc.ID, config.Source.GenerateApiUrl, config, oldDoc.Path)
		if err != nil {
			log.Println("migrateDocumentStorage failed", err)
			continue
		}

		// 如果有锁定信息，检查并更新DocumentLock记录
		if !oldDoc.LockedAt.IsZero() || oldDoc.LockedReason != "" || oldDoc.LockedWords != "" {
			docLock := models.DocumentLock{
				DocumentId:   newDoc.Id,
				LockedReason: oldDoc.LockedReason,
				LockedType:   models.LockedTypeText,
				LockedTarget: "",
				LockedWords:  oldDoc.LockedWords,
			}
			if err := CheckAndUpdate(targetDB, "document_lock", "document_id = ?", docLock.DocumentId, docLock); err != nil {
				log.Printf("Error creating/updating document lock for document %s: %v", newDoc.Id, err)
			}
		}
	}

	// 删除不存在的用户的文档
	// todo

}
