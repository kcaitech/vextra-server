package start

import (
	"github.com/gin-gonic/gin"
	"log"
	. "kcaitech.com/kcserver/common/config"
	"kcaitech.com/kcserver/common/gin/middlewares"
)

func Run(config *BaseConfiguration, afterInit func(router *gin.Engine)) {
	log.Println("开始运行")

	//gin.SetMode(gin.DebugMode)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.MaxMultipartMemory = 10 << 20 // 10 MiB
	router.Use(gin.Recovery())
	router.Use(middlewares.ErrorHandler())

	afterInit(router)

	err := router.Run(":" + config.Server.Port)
	if err != nil {
		if err != nil {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}
}
