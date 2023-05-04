package main

import (
	"log"
	"migration"
	"migration/config"
)

func main() {
	log.Println("开始运行")
	var conf config.BaseConfiguration
	config.LoadConfig("./config/config.yaml", &conf)
	migration.Init(&conf)
	var err error
	log.Println("开始迁移User表")
	if err = UserUp(migration.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}
	log.Println("开始迁移Document表")
	if err = DocumentUp(migration.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}
	log.Println("开始迁移DocumentPermission表")
	if err = DocumentPermissionUp(migration.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}
	log.Println("开始迁移DocumentAccessRecord表")
	if err = DocumentAccessRecordUp(migration.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}
	log.Println("迁移结束")
}
