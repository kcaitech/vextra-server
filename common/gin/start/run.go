package start

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	. "kcaitech.com/kcserver/common/config"
	"kcaitech.com/kcserver/common/gin/middlewares"
)

func Run(config *BaseConfiguration, afterInit func(router *gin.Engine), port int32) {
	log.Printf("kcserver服务已启动 %d", port)

	//gin.SetMode(gin.DebugMode)
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.MaxMultipartMemory = 10 << 20 // 10 MiB
	router.Use(gin.Recovery())
	router.Use(middlewares.ErrorHandler())

	afterInit(router)

	err := router.Run(":" + fmt.Sprint(port))
	if err != nil {
		log.Fatalf("kcserver服务启动失败: %v", err)
	}
}
