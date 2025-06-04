package models

import "gorm.io/gorm"

type ResourceDocument struct {
	BaseModelStruct
	UserId      string `gorm:"uniqueIndex:idx_user_document,length:64" json:"user_id"`     // 用户ID
	DocumentId  string `gorm:"uniqueIndex:idx_user_document,length:32" json:"document_id"` // 文档ID
	References  int    `gorm:"" json:"references"`                                         // 引用次数
	Description string `gorm:"" json:"description"`                                        // 描述
}

func (model ResourceDocument) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model ResourceDocument) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model ResourceDocument) GetId() interface{} {
	return model.Id
}

func (model ResourceDocument) TableName() string {
	return "resource_document"
}
