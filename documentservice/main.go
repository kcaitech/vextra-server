package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"protodesign.cn/kcserver/common/gin/start"
	"protodesign.cn/kcserver/common/storage"
	httpApi "protodesign.cn/kcserver/documentservice/api/v1/http"
	"protodesign.cn/kcserver/documentservice/config"
)

func Init() {
	_, err := storage.Init()
	if err != nil {
		log.Fatalln("storage初始化失败：" + err.Error())
	}
}

func main() {
	start.Run(
		&config.LoadConfig().BaseConfiguration,
		Init,
		func(router *gin.Engine) {
			httpApi.LoadRoutes(router)
		},
	)
}
