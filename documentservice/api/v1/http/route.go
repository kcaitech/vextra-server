package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common"
	"protodesign.cn/kcserver/documentservice/controllers"
)

func LoadRoutes(router *gin.Engine) {
	router.GET("/health_check", controllers.HealthCheck)
	apiGroup := router.Group(common.ApiVersionPath)
	loadDocumentRoutes(apiGroup)
}
