package models

// DocType 文档类型
type DocType uint8 // 文档权限
const (
	DocTypePrivate           DocType = 0 // 私有
	DocTypeShareable         DocType = 1 // 可分享（默认无权限，需申请）
	DocTypePublicReadable    DocType = 2 // 公共可读
	DocTypePublicCommentable DocType = 3 // 公共可评论
	DocTypePublicEditable    DocType = 4 // 公共可编辑
)

// Document 文档
type Document struct {
	BaseModel
	UserId  int64   `gorm:"" json:"user_id"`
	Path    string  `gorm:"size:64" json:"path"`
	DocType DocType `gorm:"default:0" json:"doc_type"`
	Name    string  `gorm:"size:64" json:"name"`
	Size    uint    `gorm:"" json:"size"`
}
