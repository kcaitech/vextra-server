package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/authservice/controllers"
	"protodesign.cn/kcserver/common"
)

func LoadRoutes(router *gin.Engine) {
	router.GET("/health_check", controllers.HealthCheck)
	apiGroup := router.Group(common.ApiVersionPath)
	loadUserRoutes(apiGroup)
}
