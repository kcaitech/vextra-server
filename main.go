package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"kcaitech.com/kcserver/common/gin/start"
	"kcaitech.com/kcserver/common/mongo"
	"kcaitech.com/kcserver/common/redis"
	"kcaitech.com/kcserver/common/safereview"
	"kcaitech.com/kcserver/common/storage"
	httpApi "kcaitech.com/kcserver/api/v1/http"
	config "kcaitech.com/kcserver/controllers"
	"kcaitech.com/kcserver/common/jwt"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/snowflake"
)

func Init() {

	jwt.Init("config/jwt.yaml")

	snowflake.Init("config/snowflake.yaml")

	conf := &config.LoadConfig("config/config.yaml").BaseConfiguration
	models.Init(conf)

	if err := storage.Init("config/storage.yaml"); err != nil {
		log.Fatalln("storage init fail:" + err.Error())
	}
	if err := mongo.Init("config/mongodb.yaml"); err != nil {
		log.Fatalln("mongo init fail:" + err.Error())
	}
	if err := redis.Init("config/redis.yaml"); err != nil {
		log.Fatalln("redis init fail:" + err.Error())
	}
	if err := safereview.Init("config/safereview.yaml"); err != nil {
		log.Fatalln("safereview init fail:" + err.Error())
	}
}

func main() {
	Init()
	conf := &config.Config.BaseConfiguration
	start.Run(conf, func(router *gin.Engine) {
		httpApi.LoadRoutes(router)
	})
}
