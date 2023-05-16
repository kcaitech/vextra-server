package models

import "protodesign.cn/kcserver/utils/time"

// DocumentAccessRecord 文档访问记录
type DocumentAccessRecord struct {
	BaseModel
	UserId         int64     `gorm:"uniqueIndex:user_document" json:"user_id"`                // 用户ID
	DocumentId     int64     `gorm:"uniqueIndex:user_document" json:"document_id"`            // 文档ID
	LastAccessTime time.Time `gorm:"autoCreateTime;type:datetime(6)" json:"last_access_time"` // 上次访问时间
}

func (model *DocumentAccessRecord) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model *DocumentAccessRecord) UnmarshalJSON(data []byte) error {
	return UnmarshalJSON(model, data)
}
