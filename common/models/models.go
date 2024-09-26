package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"kcaitech.com/kcserver/common/config"
	myReflect "kcaitech.com/kcserver/utils/reflect"
	"kcaitech.com/kcserver/utils/str"
	myTime "kcaitech.com/kcserver/utils/time"
	"reflect"
	"strings"
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
	DB.Logger = DB.Logger.LogMode(logger.Silent)
	//DB = DB.Debug()
	//DB.Logger = DB.Logger.LogMode(logger.Info)
}

type DeletedAt gorm.DeletedAt

func (n *DeletedAt) Scan(value interface{}) error {
	return (*gorm.DeletedAt)(n).Scan(value)
}

func (n DeletedAt) Value() (driver.Value, error) {
	return gorm.DeletedAt(n).Value()
}

func (n DeletedAt) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(myTime.Time(n.Time))
	}
	return gorm.DeletedAt(n).MarshalJSON()
}

func (n *DeletedAt) UnmarshalJSON(b []byte) error {
	var t myTime.Time
	err := t.UnmarshalJSON(b)
	if err == nil {
		if newB, err := t.MarshalJSON(); err == nil {
			b = newB
		}
	}
	return (*gorm.DeletedAt)(n).UnmarshalJSON(b)
}

func (n DeletedAt) QueryClauses(f *schema.Field) []clause.Interface {
	return gorm.DeletedAt(n).QueryClauses(f)
}

func (n DeletedAt) UpdateClauses(f *schema.Field) []clause.Interface {
	return gorm.DeletedAt(n).UpdateClauses(f)
}

func (n DeletedAt) DeleteClauses(f *schema.Field) []clause.Interface {
	return gorm.DeletedAt(n).DeleteClauses(f)
}

type BaseModel struct {
	Id        int64       `gorm:"primaryKey;autoIncrement:false" json:"id"`
	CreatedAt myTime.Time `gorm:"autoCreateTime;type:datetime(6)" json:"created_at"`
	UpdatedAt myTime.Time `gorm:"autoUpdateTime;type:datetime(6)" json:"updated_at"`
	DeletedAt DeletedAt   `gorm:"index" json:"deleted_at"`
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

func StructToMap(structData any, mapData map[string]any) {
	structDataValue := reflect.ValueOf(structData)
	if !structDataValue.IsValid() {
		return
	}
	if structDataValue.Kind() == reflect.Ptr {
		structDataValue = myReflect.EnterPointerByValue(structDataValue)
	}
	if !structDataValue.IsValid() || structDataValue.Kind() != reflect.Struct {
		return
	}
	for i, num := 0, structDataValue.NumField(); i < num; i++ {
		field := structDataValue.Field(i)
		typeField := structDataValue.Type().Field(i)
		if !typeField.IsExported() {
			continue
		}
		name := typeField.Name
		if jsonNameSplitRes := strings.Split(typeField.Tag.Get("json"), ","); len(jsonNameSplitRes) > 0 {
			name = strings.TrimSpace(jsonNameSplitRes[0])
		}
		anonymous := typeField.Tag.Get("anonymous")
		if (typeField.Anonymous || anonymous == "true") && field.Kind() == reflect.Struct {
			StructToMap(field.Interface(), mapData)
			continue
		}
		// 如果是int64，则转换为字符串
		if field.Kind() == reflect.Int64 {
			mapData[name] = str.IntToString(field.Int())
			continue
		}
		mapData[name] = field.Interface()
	}
}

func MarshalJSON(model any) ([]byte, error) {
	modelMap := make(map[string]any)
	StructToMap(model, modelMap)
	return json.Marshal(modelMap)
}

func mapToStruct(mapData map[string]any, structData any) {
	structDataValue := reflect.ValueOf(structData)
	if !structDataValue.IsValid() {
		return
	}
	if structDataValue.Kind() == reflect.Ptr {
		structDataValue = myReflect.EnterPointerByValue(structDataValue)
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

func UnmarshalJSON(model any, data []byte) error {
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

func GetTableName(model any) string {
	db := DB.Model(model)
	_ = db.Statement.Parse(&model)
	return db.Statement.Table
}

func GetTableFieldNames(model any) []string {
	db := DB.Model(model)
	_ = db.Statement.Parse(&model)
	fieldNames := make([]string, 0, len(db.Statement.Schema.Fields))
	for _, field := range db.Statement.Schema.Fields {
		fieldNames = append(fieldNames, field.DBName)
	}
	return fieldNames
}

func GetTableFieldNamesStr(model any) []string {
	db := DB.Model(model)
	_ = db.Statement.Parse(&model)
	tableName := db.Statement.Table
	fieldNames := make([]string, 0, len(db.Statement.Schema.Fields))
	for _, field := range db.Statement.Schema.Fields {
		fieldNames = append(fieldNames, fmt.Sprintf("%s.%s", tableName, field.DBName))
	}
	return fieldNames
}

func GetTableFieldNamesAliasByPrefix(model any, prefix string) []string {
	db := DB.Model(model)
	_ = db.Statement.Parse(&model)
	tableName := db.Statement.Table
	fieldNames := make([]string, 0, len(db.Statement.Schema.Fields))
	for _, field := range db.Statement.Schema.Fields {
		fieldNames = append(fieldNames, fmt.Sprintf("%s.%s as %s", tableName, field.DBName, prefix+field.DBName))
	}
	return fieldNames
}

func GetTableFieldNamesStrAliasByPrefix(model any, prefix string) string {
	return strings.Join(GetTableFieldNamesAliasByPrefix(model, prefix), ",")
}

func GetTableFieldNamesStrAliasByDefaultPrefix(model any, connector string) string {
	if connector == "" {
		connector = "__"
	}
	db := DB.Model(model)
	_ = db.Statement.Parse(&model)
	tableName := db.Statement.Table
	fieldNames := make([]string, 0, len(db.Statement.Schema.Fields))
	for _, field := range db.Statement.Schema.Fields {
		fieldNames = append(fieldNames, fmt.Sprintf("%s.%s as %s", tableName, field.DBName, tableName+connector+field.DBName))
	}
	return strings.Join(fieldNames, ",")
}
