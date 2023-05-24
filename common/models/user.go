package models

import (
	"protodesign.cn/kcserver/utils/time"
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
}

func (model *User) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
