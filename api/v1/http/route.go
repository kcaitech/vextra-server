package http

import (
	"context"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/redis"
	"kcaitech.com/kcserver/controllers"
	"kcaitech.com/kcserver/middlewares"
)

func LoadRoutes(router *gin.Engine) {
	router.RedirectTrailingSlash = false
	router.GET("/health_check", controllers.HealthCheck)
	// router.GET("/version", controllers.GetAppVersion) // 由前端文件提供

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(static.Serve("/", static.LocalFile("/app/html", false))) // 前端工程
	// 如果不是直接使用，不使用noroute，不然代理不好处理错误路径
	// 未知的路由交由前端vue router处理
	if controllers.Config.DefaultRoute {
		router.NoRoute(func(c *gin.Context) {
			c.File("/app/html/index.html")
		})
	}

	if controllers.Config.DetailedLog {
		router.Use(middlewares.AccessDetailedLogMiddleware())
	} else {
		router.Use(middlewares.AccessLogMiddleware())
	}
	if controllers.Config.AllowCors {
		router.Use(middlewares.CORSMiddleware()) // 测试时需要
	}

	router.Use(middlewares.NewRateLimiter(&middlewares.RedisStore{
		Client: redis.Client,
		Ctx:    context.Background(),
	}, middlewares.DefaultRateLimiterConfig()).RateLimitMiddleware())

	apiGroup := router.Group("/api")
	loadLoginRoutes(apiGroup)
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
	loadApiGatewayRoutes(apiGroup)
	apiGroup.POST("/batch_request", controllers.BatchRequestHandler(router))
}
