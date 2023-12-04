package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"protodesign.cn/kcserver/common/gin/start"
	myInit "protodesign.cn/kcserver/common/init"
	"protodesign.cn/kcserver/common/safereview"
	"protodesign.cn/kcserver/common/storage"
	httpApi "protodesign.cn/kcserver/userservice/api/v1/http"
	myConfig "protodesign.cn/kcserver/userservice/config"
)

func Init() {
	if err := storage.Init(""); err != nil {
		log.Fatalln("storage初始化失败：" + err.Error())
	}
	if err := safereview.Init(""); err != nil {
		log.Fatalln("safereview初始化失败：" + err.Error())
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
