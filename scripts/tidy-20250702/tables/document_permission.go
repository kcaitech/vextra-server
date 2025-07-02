package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyDocumentPermission(db *gorm.DB, removdDocuments []string) []models.DocumentPermission {

	// 分文档跟目录
	// 被删除的文档需要清空

	// todo 目录

	var deletedRecords []models.DocumentPermission
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("document_permission").Where("resource_type = ? AND resource_id IN (?)", models.ResourceTypeDoc, removdDocuments).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("document_permission").Where("resource_type = ? AND resource_id IN (?)", models.ResourceTypeDoc, removdDocuments).Delete(&models.DocumentPermission{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.DocumentPermission{}
	}

	return deletedRecords
}
