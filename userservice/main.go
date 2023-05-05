package main

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/start"
	myInit "protodesign.cn/kcserver/common/init"
	httpApi "protodesign.cn/kcserver/userservice/api/v1/http"
	myConfig "protodesign.cn/kcserver/userservice/config"
)

func main() {
	conf := &myConfig.LoadConfig().BaseConfiguration
	myInit.Init(conf)
	start.Run(conf, func(router *gin.Engine) {
		httpApi.LoadRoutes(router)
	})
}
