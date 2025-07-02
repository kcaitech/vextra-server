package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func MigrateDocumentFavorites(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error)) {

	// 迁移文档收藏表
	var oldFavorites []struct {
		ID         int64      `gorm:"column:id"`
		CreatedAt  time.Time  `gorm:"column:created_at"`
		UpdatedAt  time.Time  `gorm:"column:updated_at"`
		DeletedAt  *time.Time `gorm:"column:deleted_at"`
		UserId     int64      `gorm:"column:user_id"`
		DocumentId int64      `gorm:"column:document_id"`
		IsFavorite bool       `gorm:"column:is_favorite"`
	}

	if err := sourceDB.Table("document_favorites").Find(&oldFavorites).Error; err != nil {
		log.Fatalf("Error querying document favorites: %v", err)
	}
	for _, oldFav := range oldFavorites {

		userId, err := getUserID(oldFav.UserId)
		if err != nil {
			log.Printf("Error getting user ID for document favorite %d: %v", oldFav.ID, err)
			continue
		}

		newFav := models.DocumentFavorites{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldFav.CreatedAt,
				UpdatedAt: oldFav.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:     userId,
			DocumentId: strconv.FormatInt(oldFav.DocumentId, 10),
			IsFavorite: oldFav.IsFavorite,
		}

		if oldFav.DeletedAt != nil {
			newFav.DeletedAt.Time = *oldFav.DeletedAt
			newFav.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "document_favorites", "user_id = ? AND document_id = ?", []interface{}{newFav.UserId, newFav.DocumentId}, newFav); err != nil {
			log.Printf("Error migrating favorite %d: %v", oldFav.ID, err)
		}
	}
}
