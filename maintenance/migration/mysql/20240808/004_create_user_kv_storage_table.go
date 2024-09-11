package main

import (
	"gorm.io/gorm"
)

// UserKVStorage 配置表
type UserKVStorage struct {
	BaseModel
	UserId int64  `gorm:"not null;index;uniqueIndex:idx_user_key" json:"user_id"`
	Key    string `gorm:"not null;index;uniqueIndex:idx_user_key" json:"key"`
	Value  string `gorm:"not null" json:"value"`
}

func UserKVStorageUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&UserKVStorage{}); err != nil {
		return err
	}
	return nil
}

func UserKVStorageDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&UserKVStorage{}); err != nil {
		return err
	}
	return nil
}
