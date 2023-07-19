package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"protodesign.cn/kcserver/common/gin/start"
	myInit "protodesign.cn/kcserver/common/init"
	"protodesign.cn/kcserver/common/mongo"
	"protodesign.cn/kcserver/common/redis"
	"protodesign.cn/kcserver/common/storage"
	httpApi "protodesign.cn/kcserver/documentservice/api/v1/http"
	myConfig "protodesign.cn/kcserver/documentservice/config"
)

func Init() {
	if err := storage.Init(""); err != nil {
		log.Fatalln("storage初始化失败：" + err.Error())
	}
	if err := mongo.Init(""); err != nil {
		log.Fatalln("mongo初始化失败：" + err.Error())
	}
	if err := redis.Init(""); err != nil {
		log.Fatalln("redis初始化失败：" + err.Error())
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
