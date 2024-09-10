package main

import (
	"gorm.io/gorm"
	"protodesign.cn/kcserver/utils/time"
)

// User 用户
type User struct {
	BaseModel

	Nickname      string `gorm:"size:64" json:"nickname"`
	Avatar        string `gorm:"size:256" json:"avatar"`
	Uid           string `gorm:"unique;size:64" json:"uid"`
	IsActivated   bool   `gorm:"default:false" json:"is_activated"`
	WebAppChannel string `gorm:"size:64" json:"web_app_channel"`

	// 微信开放平台网页应用
	WxOpenId                 string    `gorm:"index;uniqueIndex:wx_openid_unique;size:64" json:"wx_open_id"`
	WxAccessToken            string    `gorm:"size:255" json:"wx_access_token"`
	WxAccessTokenCreateTime  time.Time `gorm:"type:datetime(6)" json:"wx_access_token_create_time"`
	WxRefreshToken           string    `gorm:"size:255" json:"wx_refresh_token"`
	WxRefreshTokenCreateTime time.Time `gorm:"type:datetime(6)" json:"wx_refresh_token_create_time"`
	WxLoginCode              string    `gorm:"size:64" json:"wx_login_code"`

	// 微信小程序
	WxMpOpenId               string    `gorm:"index;uniqueIndex:wx_openid_unique;size:64" json:"wx_mp_open_id"`
	WxMpSessionKey           string    `gorm:"size:255" json:"wx_mp_session_key"`
	WxMpSessionKeyCreateTime time.Time `gorm:"type:datetime(6)" json:"wx_mp_session_key_create_time"`
	WxMpLoginCode            string    `gorm:"size:64" json:"wx_mp_login_code"`

	// 微信开放平台UnionId
	WxUnionId string `gorm:"unique;size:64" json:"wx_union_id"`
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
