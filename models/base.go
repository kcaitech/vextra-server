/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"strings"

	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	myReflect "kcaitech.com/kcserver/utils/reflect"
	"kcaitech.com/kcserver/utils/str"
	myTime "kcaitech.com/kcserver/utils/time"
)

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

// type DefaultModelData struct {
// 	BaseModelStruct
// }

type BaseModel interface {
	GetId() interface{}
}

type BaseModelStruct struct {
	Id        int64     `gorm:"primaryKey;autoIncrement" json:"id,string"` // 主键，自增
	CreatedAt time.Time `gorm:"autoCreateTime;type:datetime(6)" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;type:datetime(6)" json:"updated_at"`
	DeletedAt DeletedAt `gorm:"index" json:"deleted_at"`
}

func (data BaseModelStruct) GetId() interface{} {
	return data.Id
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
			// 检查是否有 inline 标签
			isInline := false
			for _, tag := range jsonNameSplitRes[1:] {
				if strings.TrimSpace(tag) == "inline" {
					isInline = true
					break
				}
			}
			if isInline && field.Kind() == reflect.Struct {
				StructToMap(field.Interface(), mapData)
				continue
			}
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

// type DefaultModelData struct{}

// func (data *DefaultModelData) GetId() int64 {
// 	return 0
// }

// func (data *DefaultModelData) SetId(id int64) {}

// ModelListData 指向具体Model数组的指针，例如：&[]User{}
type ModelListData any
