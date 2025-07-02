package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyDocumentAccessRecord(db *gorm.DB, userIds []string) []models.DocumentAccessRecord {
	var deletedRecords []models.DocumentAccessRecord

	// 使用事务确保查询和删除操作的原子性
	err := db.Transaction(func(tx *gorm.DB) error {
		// 先查询要删除的数据
		if err := tx.Table("document_access_record").Where("user_id NOT IN (?)", userIds).Find(&deletedRecords).Error; err != nil {
			return err
		}

		// 然后删除数据
		if err := tx.Table("document_access_record").Where("user_id NOT IN (?)", userIds).Delete(&models.DocumentAccessRecord{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// 如果事务失败，返回空切片
		return []models.DocumentAccessRecord{}
	}

	return deletedRecords
}
