package models

import (
	"kcaitech.com/kcserver/utils/time"
)

// AppVersion App版本
type AppVersion struct {
	BaseModel
	WebAppChannel string    `gorm:"index;not null;size:64" json:"web_app_channel"`
	Type          string    `gorm:"index;not null;size:32" json:"type"`
	Version       string    `gorm:"index;not null;size:64" json:"version"`
	Code          int64     `gorm:"index;not null" json:"code"`
	CmdVersion    int64     `gorm:"index;not null" json:"cmd_version"`
	PublishTime   time.Time `gorm:"index;not null;type:datetime(6)" json:"publish_time"`
	Desc          string    `gorm:"not null" json:"desc"`
	Detail        string    `gorm:"not null" json:"detail"`
}

func (model AppVersion) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
