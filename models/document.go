package models

import (
	"time"

	"gorm.io/gorm"
)

type DocumentLock struct {
	BaseModelStruct
	DocumentId   string    `gorm:"index" json:"document_id"`
	LockedAt     time.Time `gorm:"" json:"locked_at"`
	LockedReason string    `gorm:"size:255" json:"locked_reason"`
	LockedWords  string    `gorm:"size:255" json:"locked_words"`
}

func (model DocumentLock) GetId() int64 {
	return model.Id
}

func (model DocumentLock) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model DocumentLock) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

// tablename
func (model DocumentLock) TableName() string {
	return "document_lock"
}

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
	// BaseModelStruct
	Id        string    `gorm:"primaryKey" json:"id,string"` // 主键
	CreatedAt time.Time `gorm:"autoCreateTime;type:datetime(6)" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;type:datetime(6)" json:"updated_at"`
	DeletedAt DeletedAt `gorm:"index" json:"deleted_at"`

	UserId    string    `gorm:"index" json:"user_id"`
	Path      string    `gorm:"size:64" json:"path"`
	DocType   DocType   `gorm:"default:0" json:"doc_type"`
	Name      string    `gorm:"size:64" json:"name"`
	Size      uint64    `gorm:"" json:"size"`
	PurgedAt  time.Time `gorm:"" json:"purged_at"`
	DeleteBy  int64     `gorm:"" json:"delete_by"` // 删除人ID
	VersionId string    `gorm:"size:64" json:"version_id"`
	TeamId    string    `gorm:"index" json:"team_id"`
	ProjectId string    `gorm:"index" json:"project_id"`
	// LockedAt     time.Time `gorm:"" json:"locked_at"`
	// LockedReason string    `gorm:"size:255" json:"locked_reason"`
	// LockedWords  string    `gorm:"size:255" json:"locked_words"`
}

func (model Document) GetId() interface{} {
	return model.Id
}

func (model Document) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

//func (model *Document) UnmarshalJSON(data []byte) error {
//	return UnmarshalJSON(model, data)
//}

func (model Document) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

// tablename
func (model Document) TableName() string {
	return "document"
}
