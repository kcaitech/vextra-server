package main

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/utils/time"
)

// User 用户
type User struct {
	BaseModel
	Nickname                 string    `gorm:"size:64" json:"nickname"`
	WxOpenId                 string    `gorm:"unique;size:64" json:"wx_open_id"`
	WxAccessToken            string    `gorm:"size:255" json:"wx_access_token"`
	WxAccessTokenCreateTime  time.Time `gorm:"type:datetime(6)" json:"wx_access_token_create_time"`
	WxRefreshToken           string    `gorm:"size:255" json:"wx_refresh_token"`
	WxRefreshTokenCreateTime time.Time `gorm:"type:datetime(6)" json:"wx_refresh_token_create_time"`
	Avatar                   string    `gorm:"size:256" json:"avatar"`
	Uid                      string    `gorm:"unique;size:64" json:"uid"`
	IsActivated              bool      `gorm:"default:false" json:"is_activated"`
	WxLoginCode              string    `gorm:"size:64" json:"wx_login_code"`
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
