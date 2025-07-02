package tables

import (
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func MigrateProjectFavorite(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error)) {

	// 迁移项目收藏表
	var oldProjectFavorites []struct {
		ID        int64      `gorm:"column:id"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		UpdatedAt time.Time  `gorm:"column:updated_at"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
		UserId    int64      `gorm:"column:user_id"`
		ProjectId int64      `gorm:"column:project_id"`
		IsFavor   bool       `gorm:"column:is_favor"`
	}

	if err := sourceDB.Table("project_favorite").Find(&oldProjectFavorites).Error; err != nil {
		log.Fatalf("Error querying project favorites: %v", err)
	}

	for _, oldFav := range oldProjectFavorites {
		userId, err := getUserID(oldFav.UserId)
		if err != nil {
			log.Printf("Error getting user ID for project favorite %d: %v", oldFav.ID, err)
			continue
		}

		newFav := models.ProjectFavorite{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldFav.CreatedAt,
				UpdatedAt: oldFav.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:    userId,
			ProjectId: strconv.FormatInt(oldFav.ProjectId, 10),
			IsFavor:   oldFav.IsFavor,
		}

		if oldFav.DeletedAt != nil {
			newFav.DeletedAt.Time = *oldFav.DeletedAt
			newFav.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "project_favorite", "user_id = ? AND project_id = ?", []interface{}{newFav.UserId, newFav.ProjectId}, newFav); err != nil {
			log.Printf("Error migrating project favorite %d: %v", oldFav.ID, err)
		}
	}
}
