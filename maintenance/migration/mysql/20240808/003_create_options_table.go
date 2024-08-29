package main

import (
	"gorm.io/gorm"
)

// Options 配置表
type Options struct {
	BaseModel
	Type   string `gorm:"index;not null;size:255" json:"type"`
	Detail string `gorm:"not null" json:"detail"`
}

func OptionsUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&Options{}); err != nil {
		return err
	}
	return nil
}

func OptionsDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&Options{}); err != nil {
		return err
	}
	return nil
}
