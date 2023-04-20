package main

import (
	"github.com/gin-gonic/gin"
	_ "gorm.io/driver/mysql"
	httpApi "protodesign.cn/kcserver/authservice/api/v1/http"
	"protodesign.cn/kcserver/authservice/config"
	"protodesign.cn/kcserver/authservice/models/migrations"
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
