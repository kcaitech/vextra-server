package main

import (
	"github.com/gin-gonic/gin"
	httpApi "protodesign.cn/kcserver/authservice/api/v1/http"
	myConfig "protodesign.cn/kcserver/authservice/config"
	"protodesign.cn/kcserver/common/gin/start"
	myInit "protodesign.cn/kcserver/common/init"
)

func main() {
	conf := &myConfig.LoadConfig().BaseConfiguration
	myInit.Init(conf)
	start.Run(conf, func(router *gin.Engine) {
		httpApi.LoadRoutes(router)
	})
}
