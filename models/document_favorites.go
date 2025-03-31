package models

import "gorm.io/gorm"

// DocumentFavorites 文档收藏
type DocumentFavorites struct {
	BaseModelStruct
	UserId     string `gorm:"uniqueIndex:idx_user_document,length:64" json:"user_id"`
	DocumentId int64  `gorm:"uniqueIndex:idx_user_document" json:"document_id"`
	IsFavorite bool   `gorm:"" json:"is_favorite"`
}

func (model DocumentFavorites) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model DocumentFavorites) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model DocumentFavorites) GetId() int64 {
	return model.Id
}

// tablename
func (model DocumentFavorites) TableName() string {
	return "document_favorites"
}
