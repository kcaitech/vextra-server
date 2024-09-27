// 内部使用的服务，没有用户鉴权

package inner

import (
	"github.com/gin-gonic/gin"
	"log"
	"kcaitech.com/kcserver/common/gin/start"
	"kcaitech.com/kcserver/common/mongo"
	"kcaitech.com/kcserver/common/redis"
	"kcaitech.com/kcserver/common/safereview"
	"kcaitech.com/kcserver/common/storage"
	// httpApi "kcaitech.com/kcserver/api/v1/http"
	config "kcaitech.com/kcserver/controllers"
	"kcaitech.com/kcserver/common/jwt"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/snowflake"
	"kcaitech.com/kcserver/controllers"
	"kcaitech.com/kcserver/controllers/document"
	// "kcaitech.com/kcserver/middlewares"
	"kcaitech.com/kcserver/common"
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

func LoadRoutes(router *gin.Engine) {
	router.GET("/health_check", controllers.HealthCheck)
	apiGroup := router.Group(common.ApiVersionPath)
	// documents/document_upload 用于版本更新
	{
		router1 := apiGroup.Group("/documents")
		router1.GET("/document_upload", document.UploadDocument)
	}
}

func main() {
	conf := Init()
	start.Run(&conf.BaseConfiguration, func(router *gin.Engine) {
		LoadRoutes(router)
	})
}
