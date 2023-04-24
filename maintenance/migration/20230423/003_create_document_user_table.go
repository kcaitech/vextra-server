package main

import (
	"gorm.io/gorm"
)

type PermType uint8 // 文档授权类型

const (
	PermTypeNone     PermType = 0 // 无权限
	PermTypeReadOnly PermType = 1 // 只读
	PermTypeWritable PermType = 2 // 可写
)

type DocumentUser struct {
	BaseModel
	ID         uint     `gorm:"primary_key" json:"id"`
	DocumentId uint     `gorm:"" json:"document_id"`
	UserId     uint     `gorm:"" json:"user_id"`
	PermType   PermType `gorm:"default:0" json:"perm_type"`
}

func DocumentUserUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&DocumentUser{}); err != nil {
		return err
	}
	return nil
}

func DocumentUserDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&DocumentUser{}); err != nil {
		return err
	}
	return nil
}
