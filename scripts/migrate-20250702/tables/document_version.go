package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func MigrateDocumentVersion(sourceDB *gorm.DB, targetDB *gorm.DB) {

	// 迁移文档版本表
	var oldVersions []struct {
		DeletedAt  *time.Time `gorm:"column:deleted_at"`
		CreatedAt  time.Time  `gorm:"column:created_at"`
		UpdatedAt  time.Time  `gorm:"column:updated_at"`
		ID         int64      `gorm:"column:id"`
		DocumentId int64      `gorm:"column:document_id"`
		VersionId  string     `gorm:"column:version_id"`
		LastCmdId  int64      `gorm:"column:last_cmd_id"`
		// 其他BaseModel字段
	}

	if err := sourceDB.Table("document_version").Find(&oldVersions).Error; err != nil {
		log.Fatalf("Error querying document versions: %v", err)
	}
	for _, oldVer := range oldVersions {
		newVer := models.DocumentVersion{
			BaseModelStruct: models.BaseModelStruct{
				DeletedAt: models.DeletedAt{},
				CreatedAt: oldVer.CreatedAt,
				UpdatedAt: oldVer.UpdatedAt,
			},
			DocumentId:   strconv.FormatInt(oldVer.DocumentId, 10), // 转为string
			VersionId:    oldVer.VersionId,
			LastCmdVerId: uint(oldVer.LastCmdId), // 注意这里字段名和类型都改变
		}

		if oldVer.DeletedAt != nil {
			newVer.DeletedAt.Time = *oldVer.DeletedAt
			newVer.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "document_version", "document_id = ? AND version_id = ?", []interface{}{newVer.DocumentId, newVer.VersionId}, newVer); err != nil {
			log.Printf("Error migrating version %d: %v", oldVer.ID, err)
		}
	}
}
