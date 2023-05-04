package models

import "protodesign.cn/kcserver/utils/time"

// DocumentAccessRecord 文档访问记录
type DocumentAccessRecord struct {
	BaseModel
	UserId         int64     `gorm:"" json:"user_id"`          // 用户ID
	DocumentId     int64     `gorm:"" json:"document_id"`      // 文档ID
	LastAccessTime time.Time `gorm:"" json:"last_access_time"` // 上次访问时间
}
