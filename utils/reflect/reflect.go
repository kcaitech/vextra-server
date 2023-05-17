package reflect

import (
	"reflect"
)

// EnterPointer 进入指针指向的对象
func EnterPointer(pointer reflect.Value) reflect.Value {
	if pointer.IsValid() && pointer.Kind() == reflect.Ptr {
		value := pointer.Elem()
		if !value.IsValid() {
			return pointer
		}
		pointer = value
	}
	return pointer
}

// FieldByName 获取结构体或结构体指针中特定字段的指针
func FieldByName(structData any, fieldName string) any {
	modelDataValue := reflect.ValueOf(structData)
	if !modelDataValue.IsValid() {
		return nil
	}
	if modelDataValue.Kind() == reflect.Ptr {
		modelDataValue = EnterPointer(modelDataValue)
	}
	if !modelDataValue.IsValid() || modelDataValue.Kind() != reflect.Struct {
		return nil
	}
	fieldValue := modelDataValue.FieldByName(fieldName)
	if !fieldValue.IsValid() {
		return nil
	}
	return fieldValue.Addr().Interface()
}

// StructToMap Struct转map
func StructToMap(structData any, mapData map[string]any) {
	structDataValue := reflect.ValueOf(structData)
	if !structDataValue.IsValid() {
		return
	}
	if structDataValue.Kind() == reflect.Ptr {
		structDataValue = EnterPointer(structDataValue)
	}
	if !structDataValue.IsValid() || structDataValue.Kind() != reflect.Struct {
		return
	}
	for i, num := 0, structDataValue.NumField(); i < num; i++ {
		field := structDataValue.Field(i)
		typeField := structDataValue.Type().Field(i)
		// 跳过非公开字段
		if typeField.PkgPath != "" {
			continue
		}
		// if typeField.Anonymous
		if field.Kind() == reflect.Struct {
			StructToMap(field.Interface(), mapData)
			continue
		}
		name := typeField.Name
		if jsonName := typeField.Tag.Get("json"); jsonName != "" {
			name = jsonName
		}
		mapData[name] = field.Interface()
	}
}

// MapToStruct map转Struct
func MapToStruct(mapData map[string]any, structData any) {
	structDataValue := reflect.ValueOf(structData)
	if !structDataValue.IsValid() {
		return
	}
	if structDataValue.Kind() == reflect.Ptr {
		structDataValue = EnterPointer(structDataValue)
	}
	if !structDataValue.IsValid() || structDataValue.Kind() != reflect.Struct {
		return
	}
	structDataType := structDataValue.Type()
	for i, num := 0, structDataValue.NumField(); i < num; i++ {
		value, ok := mapData[structDataType.Field(i).Name]
		if !ok {
			continue
		}
		structDataValue.Field(i).Set(reflect.ValueOf(value))
	}
}
