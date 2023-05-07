package reflect

import (
	"reflect"
)

// FieldByName 获取结构体或结构体指针中特定字段的指针
func FieldByName(structData interface{}, fieldName string) interface{} {
	modelDataValue := reflect.ValueOf(structData)
	if !(modelDataValue.IsValid() && (modelDataValue.Kind() == reflect.Ptr || modelDataValue.Kind() == reflect.Struct)) {
		return nil
	}
	if modelDataValue.Kind() == reflect.Ptr {
		modelDataValue = modelDataValue.Elem()
		if !(modelDataValue.IsValid() && modelDataValue.Kind() == reflect.Struct) {
			return nil
		}
	}
	fieldValue := modelDataValue.FieldByName(fieldName)
	if !fieldValue.IsValid() {
		return nil
	}
	return fieldValue.Addr().Interface()
}
