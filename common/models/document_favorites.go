package models

// DocumentFavorites 文档收藏
type DocumentFavorites struct {
	BaseModel
	UserId     string `gorm:"uniqueIndex:idx_user_document" json:"user_id"`
	DocumentId int64  `gorm:"uniqueIndex:idx_user_document" json:"document_id"`
	IsFavorite bool   `gorm:"" json:"is_favorite"`
}

func (model DocumentFavorites) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
