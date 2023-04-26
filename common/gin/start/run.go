package start

import (
	"github.com/gin-gonic/gin"
	"log"
	. "protodesign.cn/kcserver/common/config"
	"protodesign.cn/kcserver/common/gin/middlewares"
	myInit "protodesign.cn/kcserver/common/init"
	"protodesign.cn/kcserver/common/models"
)

func Run(config *BaseConfiguration, initFunc func(), afterInit func(router *gin.Engine)) {
	log.Println("开始运行")

	models.Init(config)

	myInit.Init()
	initFunc()

	router := gin.Default()
	router.Use(middlewares.ErrorHandler())

	afterInit(router)

	err := router.Run(":" + config.Server.Port)
	if err != nil {
		if err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}
}
