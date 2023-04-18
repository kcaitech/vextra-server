package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"protodesign.cn/kcserver/loginservice/config"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(mysql.Open(config.Config.DB.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
}
