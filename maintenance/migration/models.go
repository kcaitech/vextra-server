package migration

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"migration/config"
)

var DB *gorm.DB

func Init(config *config.BaseConfiguration) {
	var err error
	DB, err = gorm.Open(mysql.Open(config.DB.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
}
