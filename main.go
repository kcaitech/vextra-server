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

func Init() *config.Configuration {

	configDir := "config/"
	conf := config.LoadConfig(configDir + "config.yaml")

	jwtConfig := configDir + conf.Jwt.Ref
	snowflakeConfig := configDir + conf.Snowflake.Ref
	storageConfig := configDir + conf.Storage.Ref
	mongoConfig := configDir + conf.MongoDb.Ref
	redisConfig := configDir + conf.Redis.Ref
	safereviewConfig := configDir + conf.SafeReiew.Ref

	jwt.Init(jwtConfig)
	snowflake.Init(snowflakeConfig)
	models.Init(&conf.BaseConfiguration)

	if err := storage.Init(storageConfig); err != nil {
		log.Fatalln("storage init fail:" + err.Error())
	}
	if err := mongo.Init(mongoConfig); err != nil {
		log.Fatalln("mongo init fail:" + err.Error())
	}
	if err := redis.Init(redisConfig); err != nil {
		log.Fatalln("redis init fail:" + err.Error())
	}
	if err := safereview.Init(safereviewConfig); err != nil {
		log.Fatalln("safereview init fail:" + err.Error())
	}

	return conf
}

func main() {
	conf := Init()
	start.Run(&conf.BaseConfiguration, func(router *gin.Engine) {
		httpApi.LoadRoutes(router)
	})
}
