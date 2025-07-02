package main

import (
	"log"
	"reflect"

	"gorm.io/gorm"
)

// checkAndUpdate 辅助函数：检查记录是否存在并更新
func checkAndUpdate(db *gorm.DB, table string, whereClause string, whereArgs interface{}, newRecord interface{}) error {
	var count int64
	query := db.Table(table)

	// 处理多个参数的情况
	if args, ok := whereArgs.([]interface{}); ok {
		query = query.Where(whereClause, args...)
	} else {
		query = query.Where(whereClause, whereArgs)
	}

	if err := query.Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		// 记录存在，执行更新
		if err := query.Updates(newRecord).Error; err != nil {
			return err
		}
		log.Printf("Updated existing record in %s with %s = %v", table, whereClause, whereArgs)
	} else {
		// 记录不存在，执行创建
		recordPtr := reflect.New(reflect.TypeOf(newRecord)).Interface()
		reflect.ValueOf(recordPtr).Elem().Set(reflect.ValueOf(newRecord))
		if err := db.Table(table).Create(recordPtr).Error; err != nil {
			return err
		}
		log.Printf("Created new record in %s with %s = %v", table, whereClause, whereArgs)
	}
	return nil
}
