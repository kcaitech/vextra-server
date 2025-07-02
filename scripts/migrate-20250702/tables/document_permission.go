package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func MigrateDocumentPermission(sourceDB *gorm.DB, targetDB *gorm.DB) {

	// 迁移文档权限表
	var oldPermissions []struct {
		ID             int64      `gorm:"column:id"`
		CreatedAt      time.Time  `gorm:"column:created_at"`
		UpdatedAt      time.Time  `gorm:"column:updated_at"`
		DeletedAt      *time.Time `gorm:"column:deleted_at"`
		ResourceType   uint8      `gorm:"column:resource_type"`
		ResourceId     int64      `gorm:"column:resource_id"`
		GranteeType    uint8      `gorm:"column:grantee_type"`
		GranteeId      int64      `gorm:"column:grantee_id"`
		PermType       uint8      `gorm:"column:perm_type"`
		PermSourceType uint8      `gorm:"column:perm_source_type"`
	}

	if err := sourceDB.Table("document_permission").Find(&oldPermissions).Error; err != nil {
		log.Fatalf("Error querying document permissions: %v", err)
	}

	for _, oldPerm := range oldPermissions {
		newPerm := models.DocumentPermission{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldPerm.CreatedAt,
				UpdatedAt: oldPerm.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			ResourceType:   models.ResourceType(oldPerm.ResourceType),
			ResourceId:     strconv.FormatInt(oldPerm.ResourceId, 10),
			GranteeType:    models.GranteeType(oldPerm.GranteeType),
			GranteeId:      strconv.FormatInt(oldPerm.GranteeId, 10),
			PermType:       models.PermType(oldPerm.PermType),
			PermSourceType: models.PermSourceType(oldPerm.PermSourceType),
		}

		if oldPerm.DeletedAt != nil {
			newPerm.DeletedAt.Time = *oldPerm.DeletedAt
			newPerm.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "document_permission", "resource_id = ? AND grantee_id = ?", []interface{}{newPerm.ResourceId, newPerm.GranteeId}, newPerm); err != nil {
			log.Printf("Error migrating permission %d: %v", oldPerm.ID, err)
		}
	}
}
