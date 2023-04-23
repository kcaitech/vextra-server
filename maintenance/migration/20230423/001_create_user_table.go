package main

import (
	"gorm.io/gorm"
)

type User struct {
	BaseModel
	ID       uint   `gorm:"primary_key" json:"id"`
	Nickname string `gorm:"size:64" json:"nickname"`
	WxOpenId string `gorm:"unique;size:64" json:"wx_open_id"`
}

func UserUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&User{}); err != nil {
		return err
	}
	return nil
}

func UserDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&User{}); err != nil {
		return err
	}
	return nil
}
