package mysql

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"kcaitech.com/kcserver/maintenance/migration/mysql/config"
)

var DB *gorm.DB

func Init(config *config.BaseConfiguration) {
	var err error
	DB, err = gorm.Open(mysql.Open(config.DB.DSN), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数形式的表名
		},
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
}
