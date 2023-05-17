package models

import (
	"protodesign.cn/kcserver/utils/time"
)

// DocType 文档类型
type DocType uint8

const (
	DocTypePrivate           DocType = iota // 私有
	DocTypeShareable                        // 可分享（默认无权限，需申请）
	DocTypePublicReadable                   // 公共可读
	DocTypePublicCommentable                // 公共可评论
	DocTypePublicEditable                   // 公共可编辑
)

// Document 文档
type Document struct {
	BaseModel
	UserId   int64     `gorm:"index" json:"user_id"`
	Path     string    `gorm:"size:64" json:"path"`
	DocType  DocType   `gorm:"default:0" json:"doc_type"`
	Name     string    `gorm:"size:64" json:"name"`
	Size     uint64    `gorm:"" json:"size"`
	PurgedAt time.Time `gorm:"" json:"purged_at"`
}

func (model *Document) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

//func (model *Document) UnmarshalJSON(data []byte) error {
//	return UnmarshalJSON(model, data)
//}
