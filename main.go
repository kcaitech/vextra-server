package main

import (
	"log"

	"github.com/gin-gonic/gin"
	httpApi "kcaitech.com/kcserver/api/v1/http"
	commonConf "kcaitech.com/kcserver/common/config"
	"kcaitech.com/kcserver/common/gin/start"
	"kcaitech.com/kcserver/common/jwt"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/mongo"
	"kcaitech.com/kcserver/common/redis"
	"kcaitech.com/kcserver/common/safereview"
	"kcaitech.com/kcserver/common/snowflake"
	"kcaitech.com/kcserver/common/storage"
	config "kcaitech.com/kcserver/controllers"
)

func Init() *config.Configuration {

	configDir := "config/"
	conf, err := config.LoadConfig(configDir + "config.json")
	if err != nil {
		conf, _ = config.LoadConfigEnv("kcconfig")
	}

	// jwtConfig := configDir + conf.Jwt.Ref
	// snowflakeConfig := configDir + conf.Snowflake.Ref
	// storageConfig := configDir + conf.Storage.Ref
	// mongoConfig := configDir + conf.MongoDb.Ref
	// redisConfig := configDir + conf.Redis.Ref
	// safereviewConfig := configDir + conf.SafeReiew.Ref

	commonConf.Init(&conf.BaseConfiguration)

	jwt.Init(&conf.Jwt)
	snowflake.Init(&conf.Snowflake)
	models.Init(&conf.BaseConfiguration)

	if err := storage.Init(&conf.Storage); err != nil {
		log.Fatalln("storage init fail:" + err.Error())
	}
	if err := mongo.Init(&conf.Mongo); err != nil {
		log.Fatalln("mongo init fail:" + err.Error())
	}
	if err := redis.Init(&conf.Redis); err != nil {
		log.Fatalln("redis init fail:" + err.Error())
	}
	if err := safereview.Init(&conf.SafeReiew); err != nil {
		log.Fatalln("safereview init fail:" + err.Error())
	}

	return conf
}

const port = 80

func main() {
	conf := Init()
	start.Run(&conf.BaseConfiguration, func(router *gin.Engine) {
		httpApi.LoadRoutes(router)
	}, port)
}
