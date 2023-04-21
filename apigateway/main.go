package main

import (
	"github.com/gin-gonic/gin"
	httpApi "protodesign.cn/kcserver/apigateway/api/v1/http"
	"protodesign.cn/kcserver/apigateway/config"
	"protodesign.cn/kcserver/apigateway/models/migrations"
	"protodesign.cn/kcserver/common/gin/start"
)

func main() {
	start.Run(
		&config.LoadConfig().BaseConfiguration,
		func() {
			migrations.Migrate()
		},
		func(router *gin.Engine) {
			httpApi.LoadRoutes(router)
		},
	)
}
