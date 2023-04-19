package main

import (
	"github.com/gin-gonic/gin"
	_ "gorm.io/driver/mysql"
	"log"
	"protodesign.cn/kcserver/authservice/api/v1/http"
	"protodesign.cn/kcserver/authservice/config"
	"protodesign.cn/kcserver/authservice/models"
	"protodesign.cn/kcserver/authservice/models/migrations"
	"protodesign.cn/kcserver/utils/gin/middlewares"
)

func Init() {
	models.Init()
	migrations.Migrate(models.DB)
}

func main() {
	log.Println("开始运行")

	config.LoadConfig()

	Init()

	router := gin.Default()
	router.Use(middlewares.ErrorHandler())
	http.LoadRoutes(router)

	// 启动 HTTP 服务器
	err := router.Run(":" + config.Config.Server.Port)
	if err != nil {
		if err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}
}
