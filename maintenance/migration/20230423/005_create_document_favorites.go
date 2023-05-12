package main

import (
	"gorm.io/gorm"
)

// DocumentFavorites 文档收藏
type DocumentFavorites struct {
	BaseModel
	UserId     int64 `gorm:"index;uniqueIndex:idx_userid_documentid" json:"user_id"`
	DocumentId int64 `gorm:"index;uniqueIndex:idx_userid_documentid" json:"document_id"`
	IsFavorite bool  `gorm:"" json:"is_favorite"`
}

func DocumentFavoritesUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&DocumentFavorites{}); err != nil {
		return err
	}
	return nil
}

func DocumentFavoritesDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&DocumentFavorites{}); err != nil {
		return err
	}
	return nil
}
