package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"protodesign.cn/kcserver/common/config"
	"time"
)

var DB *gorm.DB

func Init(config *config.BaseConfiguration) {
	var err error
	DB, err = gorm.Open(mysql.Open(config.DB.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
}

type BaseModel struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
