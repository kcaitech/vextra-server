package models

type DocType uint8 // 文档权限

const (
	DocTypePrivate        DocType = 0 // 私有
	DocTypePublicReadable DocType = 1 // 公共可读
	DocTypePublicWritable DocType = 2 // 公共可写
)

type Document struct {
	BaseModel
	ID      uint    `gorm:"primary_key" json:"id"`
	UserId  uint    `gorm:"" json:"user_id"`
	Path    string  `gorm:"size:64" json:"path"`
	DocType DocType `gorm:"default:0" json:"doc_type"`
	Name    string  `gorm:"size:64" json:"name"`
	Size    uint    `gorm:"" json:"size"`
}

type PermType uint8 // 文档授权类型

const (
	PermTypeNone     PermType = 0 // 无权限
	PermTypeReadOnly PermType = 1 // 只读
	PermTypeWritable PermType = 2 // 可写
)

type DocumentUser struct {
	BaseModel
	ID         uint     `gorm:"primary_key" json:"id"`
	DocumentId uint     `gorm:"" json:"document_id"`
	UserId     uint     `gorm:"" json:"user_id"`
	PermType   PermType `gorm:"default:0" json:"perm_type"`
}
