package models

// DocumentFavorites 文档收藏
type DocumentFavorites struct {
	BaseModel
	UserId     int64 `gorm:"index;uniqueIndex:idx_userid_documentid" json:"user_id"`
	DocumentId int64 `gorm:"index;uniqueIndex:idx_userid_documentid" json:"document_id"`
	IsFavorite bool  `gorm:"" json:"is_favorite"`
}

func (model *DocumentFavorites) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
