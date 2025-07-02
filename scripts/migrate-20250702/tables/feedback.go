package tables

import (
	"log"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
)

func MigrateFeedback(sourceDB *gorm.DB, targetDB *gorm.DB, getUserID func(int64) (string, error)) {
	// 迁移反馈表
	var oldFeedbacks []struct {
		ID            int64      `gorm:"column:id"`
		CreatedAt     time.Time  `gorm:"column:created_at"`
		UpdatedAt     time.Time  `gorm:"column:updated_at"`
		DeletedAt     *time.Time `gorm:"column:deleted_at"`
		UserId        int64      `gorm:"column:user_id"`
		Type          uint8      `gorm:"column:type"`
		Content       string     `gorm:"column:content"`
		ImagePathList string     `gorm:"column:image_path_list"`
		PageUrl       string     `gorm:"column:page_url"`
	}

	if err := sourceDB.Table("feedback").Find(&oldFeedbacks).Error; err != nil {
		log.Fatalf("Error querying feedbacks: %v", err)
	}

	for _, oldFeedback := range oldFeedbacks {
		userId, err := getUserID(oldFeedback.UserId)
		if err != nil {
			log.Printf("Error getting user ID for feedback %d: %v", oldFeedback.ID, err)
			continue
		}

		newFeedback := models.Feedback{
			BaseModelStruct: models.BaseModelStruct{
				CreatedAt: oldFeedback.CreatedAt,
				UpdatedAt: oldFeedback.UpdatedAt,
				DeletedAt: models.DeletedAt{},
			},
			UserId:        userId,
			Type:          models.FeedbackType(oldFeedback.Type),
			Content:       oldFeedback.Content,
			ImagePathList: oldFeedback.ImagePathList,
			PageUrl:       oldFeedback.PageUrl,
		}

		if oldFeedback.DeletedAt != nil {
			newFeedback.DeletedAt.Time = *oldFeedback.DeletedAt
			newFeedback.DeletedAt.Valid = true
		}

		if err := CheckAndUpdate(targetDB, "feedback", "user_id = ?", []interface{}{newFeedback.UserId}, newFeedback); err != nil {
			log.Printf("Error migrating feedback %d: %v", oldFeedback.ID, err)
		}
	}
}
