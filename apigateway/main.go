package main

import (
	"github.com/gin-gonic/gin"
	"log"
	httpApi "protodesign.cn/kcserver/apigateway/api/v1/http"
	"protodesign.cn/kcserver/apigateway/common/k8s_api"
	myConfig "protodesign.cn/kcserver/apigateway/config"
	"protodesign.cn/kcserver/common/gin/start"
	myInit "protodesign.cn/kcserver/common/init"
	"protodesign.cn/kcserver/common/mongo"
	"protodesign.cn/kcserver/common/redis"
)

func Init() {
	if err := mongo.Init(""); err != nil {
		log.Fatalln("mongo初始化失败：" + err.Error())
	}
	if err := redis.Init(""); err != nil {
		log.Fatalln("redis初始化失败：" + err.Error())
	}
	if err := k8s_api.Init(); err != nil {
		log.Fatalln("k8s初始化失败：" + err.Error())
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
