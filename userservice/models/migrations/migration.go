package migrations

import (
	"protodesign.cn/kcserver/userservice/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		return
	}
}
