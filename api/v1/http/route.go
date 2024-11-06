package http

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/controllers"
	"kcaitech.com/kcserver/middlewares"
)

func LoadRoutes(router *gin.Engine) {
	router.RedirectTrailingSlash = false
	router.GET("/health_check", controllers.HealthCheck)
	router.GET("/version", controllers.GetAppVersion)

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(static.Serve("/", static.LocalFile("/app/html", false))) // 前端工程
	// 未知的路由交由前端vue router处理
	router.NoRoute(func(c *gin.Context) {
		c.File("/app/html/index.html")
	})

	router.Use(middlewares.AccessLogMiddleware())
	apiGroup := router.Group("/api") // router.Group(common.ApiVersionPath)
	// apiGroup.Use(middlewares.CORSMiddleware())
	loadLoginRoutes(apiGroup)
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
	loadApiGatewayRoutes(apiGroup)
}
