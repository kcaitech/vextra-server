package main

import (
	"log"
	"mysql"
	"mysql/config"
)

func main() {
	log.Println("开始运行")
	var conf config.BaseConfiguration
	config.LoadConfig("./config/config.yaml", &conf)
	mysql.Init(&conf)

	var err error
	log.Println("开始迁移AppVersion表")
	if err = AppVersionUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}
	log.Println("开始迁移User表")
	if err = UserUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("迁移结束")
}
