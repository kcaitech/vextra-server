package main

import (
	"github.com/gin-gonic/gin"
	"log"
	httpApi "protodesign.cn/kcserver/authservice/api/v1/http"
	myConfig "protodesign.cn/kcserver/authservice/config"
	"protodesign.cn/kcserver/common/gin/start"
	myInit "protodesign.cn/kcserver/common/init"
	"protodesign.cn/kcserver/common/storage"
)

func Init() {
	if err := storage.Init(""); err != nil {
		log.Fatalln("storage初始化失败：" + err.Error())
	}
}

func main() {
	conf := &myConfig.LoadConfig().BaseConfiguration
	myInit.Init(conf)
	Init()
	start.Run(conf, func(router *gin.Engine) {
		httpApi.LoadRoutes(router)
	})
}
