package tables

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func TidyDocumentVersion(db *gorm.DB, removdDocuments []string) []models.DocumentVersion {

	var deletedRecords []models.DocumentVersion
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("document_version").Where("document_id IN (?)", removdDocuments).Find(&deletedRecords).Error; err != nil {
			return err
		}

		if err := tx.Table("document_version").Where("document_id IN (?)", removdDocuments).Delete(&models.DocumentVersion{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return []models.DocumentVersion{}
	}

	return deletedRecords
}
