package main

import (
	"gorm.io/gorm"
)

// DocumentVersion 文档版本
type DocumentVersion struct {
	BaseModel
	DocumentId int64  `gorm:"index" json:"document_id"`
	VersionId  string `gorm:"index;size:36" json:"version_id"`
	LastCmdId  int64  `gorm:"" json:"last_cmd_id"` // 此版本最后一个cmd的id
}

func DocumentVersionUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&DocumentVersion{}); err != nil {
		return err
	}
	return nil
}

func DocumentVersionDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&DocumentVersion{}); err != nil {
		return err
	}
	return nil
}
