package main

import (
	"gorm.io/gorm"
	"protodesign.cn/kcserver/utils/time"
)

// DocumentAccessRecord 文档访问记录
type DocumentAccessRecord struct {
	BaseModel
	Id             int64     `gorm:"primary_key" json:"id"`    // Id
	UserId         int64     `gorm:"" json:"user_id"`          // 用户ID
	DocumentId     int64     `gorm:"" json:"document_id"`      // 文档ID
	LastAccessTime time.Time `gorm:"" json:"last_access_time"` // 上次访问时间
}

func DocumentAccessRecordUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&DocumentAccessRecord{}); err != nil {
		return err
	}
	return nil
}

func DocumentAccessRecordDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&DocumentAccessRecord{}); err != nil {
		return err
	}
	return nil
}
