package models

import (
	"gorm.io/gorm"
	"protodesign.cn/kcserver/utils/gin/models"
)

type User struct {
	models.BaseModel
	ID       uint   `gorm:"primary_key" json:"id"`
	Nickname string `gorm:"" json:"nickname"`
	WxOpenId string `gorm:"unique" json:"wx_open_id"`
}

func (u *User) BeforeCreate(db *gorm.DB) error {
	return nil
}
