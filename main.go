package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	api "kcaitech.com/kcserver/api"
	config "kcaitech.com/kcserver/config"
	"kcaitech.com/kcserver/middlewares"
	"kcaitech.com/kcserver/services"
)

func start(afterInit func(router *gin.Engine), port int32) {
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

func Init() *config.Configuration {

	configDir := "config/"
	conf, err := config.LoadYamlFile(configDir + "config.yaml")
	if err != nil {
		fmt.Println("err", err)
		panic(err)
	}
	fmt.Println("conf", conf)

	// 初始化services
	err = services.InitAllBaseServices(conf)
	if err != nil {
		log.Fatalf("kcserver服务初始化失败: %v", err)
	}

	return conf
}

const port = 80

func main() {
	Init()
	start(func(router *gin.Engine) {
		api.LoadRoutes(router)
	}, port)
}
