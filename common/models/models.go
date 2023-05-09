package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"protodesign.cn/kcserver/common/config"
	"protodesign.cn/kcserver/utils/time"
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

type BaseModel struct {
	Id        int64          `gorm:"primaryKey;autoIncrement:false" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime;type:datetime(6)" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;type:datetime(6)" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// ModelData 指向具体Model的指针，例如：&User{}
type ModelData interface {
	GetId() int64
	SetId(id int64)
}

func (data *BaseModel) GetId() int64 {
	return data.Id
}

func (data *BaseModel) SetId(id int64) {
	data.Id = id
}

type DefaultModelData struct{}

func (data *DefaultModelData) GetId() int64 {
	return 0
}

func (data *DefaultModelData) SetId(id int64) {}

// ModelListData 指向具体Model数组的指针，例如：&[]User{}
type ModelListData interface{}
