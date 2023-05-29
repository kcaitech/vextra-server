package main

import "gorm.io/gorm"

// InviteCode 邀请码
type InviteCode struct {
	BaseModel
	Code   string `gorm:"size:8;unique" json:"code"`
	UserId int64  `gorm:"" json:"user_id"`
}

func InviteCodeUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&InviteCode{}); err != nil {
		return err
	}
	return nil
}

func InviteCodeDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&InviteCode{}); err != nil {
		return err
	}
	return nil
}
