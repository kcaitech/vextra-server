package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"protodesign.cn/kcserver/common/config"
	"protodesign.cn/kcserver/utils/time"
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
	Id        int64          `gorm:"primaryKey;autoIncrement:false" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime;type:datetime(6)" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;type:datetime(6)" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type ModelData interface{}     // 指向具体Model的指针，例如：&User{}
type ModelListData interface{} // 指向具体Model数组的指针，例如：&[]User{}
