package main

import (
	"gorm.io/gorm"
)

type DocType uint8 // 文档权限

const (
	Private        DocType = 0 // 私有
	PublicReadable DocType = 1 // 公共可读
	PublicWritable DocType = 2 // 公共可写
)

type Document struct {
	BaseModel
	ID      uint    `gorm:"primary_key" json:"id"`
	UserId  uint    `gorm:"" json:"user_id"`
	Path    string  `gorm:"size:64" json:"path"`
	DocType DocType `gorm:"default:0" json:"doc_type"`
}

func DocumentUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&Document{}); err != nil {
		return err
	}
	return nil
}

func DocumentDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&Document{}); err != nil {
		return err
	}
	return nil
}
