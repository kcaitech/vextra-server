package main

import (
	"github.com/gin-gonic/gin"
	_ "gorm.io/driver/mysql"
	"protodesign.cn/kcserver/common/gin/start"
	httpApi "protodesign.cn/kcserver/userservice/api/v1/http"
	"protodesign.cn/kcserver/userservice/config"
	"protodesign.cn/kcserver/userservice/models/migrations"
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
