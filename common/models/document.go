package models

type DocType uint8 // 文档权限

const (
	Private        DocType = 0 // 私有
	PublicReadable DocType = 1 // 公共可读
	PublicWritable DocType = 2 // 公共可写
)

type Document struct {
	BaseModel
	ID      uint    `gorm:"primary_key" json:"id"`
	UserId  uint    `gorm:"" json:"user_id"`
	Path    string  `gorm:"size:64" json:"path"`
	DocType DocType `gorm:"default:0" json:"doc_type"`
}

type PermType uint8 // 文档授权类型

const (
	None     PermType = 0 // 无权限
	ReadOnly PermType = 1 // 只读
	Writable PermType = 2 // 可写
)

type DocumentUser struct {
	BaseModel
	ID         uint     `gorm:"primary_key" json:"id"`
	DocumentId uint     `gorm:"" json:"document_id"`
	UserId     uint     `gorm:"" json:"user_id"`
	PermType   PermType `gorm:"default:0" json:"perm_type"`
}
