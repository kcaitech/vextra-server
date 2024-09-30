package http

import (
	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/controllers"
	"kcaitech.com/kcserver/middlewares"
	"kcaitech.com/kcserver/common"
)

func LoadRoutes(router *gin.Engine) {
	router.RedirectTrailingSlash = false
	router.GET("/health_check", controllers.HealthCheck)
	router.GET("/version", controllers.GetAppVersion)
	apiGroup := router.Group(common.ApiVersionPath)
	apiGroup.Use(middlewares.CORSMiddleware())
	loadLoginRoutes(apiGroup)
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
	loadApiGatewayRoutes(apiGroup)

}
