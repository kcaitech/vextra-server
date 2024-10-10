package main

import (
	"log"

	"kcaitech.com/kcserver/maintenance/migration/mysql"
	"kcaitech.com/kcserver/maintenance/migration/mysql/config"
)

func main() {
	log.Println("开始运行")
	var conf config.BaseConfiguration
	config.LoadConfig("./config/config.yaml", &conf)
	mysql.Init(&conf)

	var err error
	log.Println("开始迁移User表")
	if err = UserUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("开始迁移Document表")
	if err = DocumentUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("开始迁移DocumentPermission表")
	if err = DocumentPermissionUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("开始迁移DocumentAccessRecord表")
	if err = DocumentAccessRecordUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("开始迁移DocumentFavorites表")
	if err = DocumentFavoritesUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("开始迁移DocumentPermissionRequests表")
	if err = DocumentPermissionRequestsUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("开始迁移InviteCode表")
	if err = InviteCodeUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("开始迁移DocumentVersion表")
	if err = DocumentVersionUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("开始迁移Team表")
	if err = TeamUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("开始迁移Project表")
	if err = ProjectUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("开始迁移Feedback表")
	if err = FeedbackUp(mysql.DB); err != nil {
		log.Fatalln("数据库迁移错误：" + err.Error())
	}

	log.Println("迁移结束")
}
