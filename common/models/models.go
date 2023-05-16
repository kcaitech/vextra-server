package models

import (
	"encoding/json"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"protodesign.cn/kcserver/common/config"
	myReflect "protodesign.cn/kcserver/utils/reflect"
	"protodesign.cn/kcserver/utils/str"
	"protodesign.cn/kcserver/utils/time"
	"reflect"
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

func MarshalJSON(model ModelData) ([]byte, error) {
	modelMap := make(map[string]any)
	myReflect.StructToMap(model, modelMap)
	for key, value := range modelMap {
		if valueInt, ok := value.(int64); ok {
			modelMap[key] = str.IntToString(valueInt)
		}
	}
	return json.Marshal(modelMap)
}

func mapToStruct(mapData map[string]any, structData any) {
	structDataValue := reflect.ValueOf(structData)
	if !structDataValue.IsValid() {
		return
	}
	if structDataValue.Kind() == reflect.Ptr {
		structDataValue = myReflect.EnterPointer(structDataValue)
	}
	if !structDataValue.IsValid() || structDataValue.Kind() != reflect.Struct {
		return
	}
	structDataType := structDataValue.Type()
	for i, num := 0, structDataValue.NumField(); i < num; i++ {
		name := structDataType.Field(i).Name
		if name == "Id" {
			name = "id"
		}
		value, ok := mapData[name]
		if !ok {
			continue
		}
		if name == "id" {
			var valueString string
			var ok bool
			if valueString, ok = value.(string); !ok {
				continue
			}
			value = str.DefaultToInt(valueString, 0)
		}
		structDataValue.Field(i).Set(reflect.ValueOf(value))
	}
}

func UnmarshalJSON(model ModelData, data []byte) error {
	var modelMap map[string]any
	err := json.Unmarshal(data, &modelMap)
	if err != nil {
		return err
	}
	mapToStruct(modelMap, model)
	return nil
}

type DefaultModelData struct{}

func (data *DefaultModelData) GetId() int64 {
	return 0
}

func (data *DefaultModelData) SetId(id int64) {}

// ModelListData 指向具体Model数组的指针，例如：&[]User{}
type ModelListData any
