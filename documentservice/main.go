package main

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/start"
	httpApi "protodesign.cn/kcserver/documentservice/api/v1/http"
	"protodesign.cn/kcserver/documentservice/config"
	"protodesign.cn/kcserver/documentservice/models/migrations"
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
