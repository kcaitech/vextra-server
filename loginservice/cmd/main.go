package main

import (
	"github.com/gin-gonic/gin"
	_ "gorm.io/driver/mysql"
	"log"
	"protodesign.cn/kcserver/loginservice/api/v1/http"
	"protodesign.cn/kcserver/loginservice/config"
	"protodesign.cn/kcserver/loginservice/models"
	"protodesign.cn/kcserver/loginservice/models/migrations"
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
