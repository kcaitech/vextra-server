package models

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/utils/time"
)

// DocumentAccessRecord 文档访问记录
type DocumentAccessRecord struct {
	BaseModelStruct
	UserId         string    `gorm:"uniqueIndex:idx_user_document,length:64" json:"user_id"`     // 用户ID
	DocumentId     string    `gorm:"uniqueIndex:idx_user_document,length:32" json:"document_id"` // 文档ID
	LastAccessTime time.Time `gorm:"autoCreateTime;type:datetime(6)" json:"last_access_time"`    // 上次访问时间
}

func (model DocumentAccessRecord) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model DocumentAccessRecord) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model DocumentAccessRecord) GetId() interface{} {
	return model.Id
}

// tablename
func (model DocumentAccessRecord) TableName() string {
	return "document_access_record"
}
