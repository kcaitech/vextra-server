package tables

import (
	"log"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func MigrateUserKVStorage(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error)) {
	// 迁移用户键值存储表
	var oldUserKVStorages []struct {
		ID        int64      `gorm:"column:id"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		UpdatedAt time.Time  `gorm:"column:updated_at"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
		UserId    int64      `gorm:"column:user_id"`
		Key       string     `gorm:"column:key"`
		Value     string     `gorm:"column:value"`
	}

	if err := sourceDB.Table("user_kv_storage").Find(&oldUserKVStorages).Error; err != nil {
		log.Fatalf("Error querying user kv storages: %v", err)
	}

	for _, oldKV := range oldUserKVStorages {
		userId, err := getUserID(oldKV.UserId)
		if err != nil {
			log.Printf("Error getting user ID for user kv storage %d: %v", oldKV.ID, err)
			continue
		}

		newKV := models.UserKVStorage{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldKV.CreatedAt,
				UpdatedAt: oldKV.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId: userId,
			Key:    oldKV.Key,
			Value:  oldKV.Value,
		}

		if oldKV.DeletedAt != nil {
			newKV.DeletedAt.Time = *oldKV.DeletedAt
			newKV.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "user_kv_storage", "user_id = ? AND `key` = ?", []interface{}{newKV.UserId, newKV.Key}, newKV); err != nil {
			log.Printf("Error migrating user kv storage %d: %v", oldKV.ID, err)
		}
	}
}
