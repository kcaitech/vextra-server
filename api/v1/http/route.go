package http

import (
	"context"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	handlers "kcaitech.com/kcserver/handlers"
	"kcaitech.com/kcserver/middlewares"
	"kcaitech.com/kcserver/services"
)

var StaticFileSuffix = []string{".html", ".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf"}

func isStaticFile(path string) bool {
	// 先获取path后缀
	suffix := filepath.Ext(path)
	if suffix == "" {
		return false
	}
	return slices.Contains(StaticFileSuffix, suffix)
}

const staticFilePath = "/app/html"

func onNotFound(c *gin.Context) {
	path := c.Request.URL.Path
	if path == "/api" || strings.HasPrefix(path, "/api/") {
		c.JSON(http.StatusNotFound, gin.H{"error": "auth endpoint not found"})
		return
	}

	// 检查是否是对静态文件的请求（HTML、JS、CSS等）
	if isStaticFile(path) {
		// 确保路径拼接正确
		filePath := staticFilePath + path
		if !strings.HasPrefix(path, "/") {
			filePath = staticFilePath + "/" + path
		}
		c.File(filePath)
		return
	}

	// 设置缓存时间为15分钟
	c.Header("Cache-Control", "public, max-age=900")
	c.File(staticFilePath + "/index.html")
}

func LoadRoutes(router *gin.Engine) {
	router.RedirectTrailingSlash = false
	router.GET("/health_check", handlers.HealthCheck)
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(static.Serve("/", static.LocalFile(staticFilePath, false))) // 前端工程
	config := services.GetConfig()
	router.NoRoute(onNotFound)

	if config.DetailedLog {
		router.Use(middlewares.AccessDetailedLogMiddleware())
	} else {
		router.Use(middlewares.AccessLogMiddleware())
	}
	if config.AllowCors {
		router.Use(middlewares.CORSMiddleware()) // 测试时需要
	}

	router.Use(middlewares.NewRateLimiter(&middlewares.RedisStore{
		Client: services.GetRedisDB().Client,
		Ctx:    context.Background(),
	}, middlewares.DefaultRateLimiterConfig()).RateLimitMiddleware())

	apiGroup := router.Group("/api")
	loadWsRoutes(apiGroup)    // 单独鉴权
	loadLoginRoutes(apiGroup) // 从refreshToken获取信息
	apiGroup.Use(services.GetKCAuthClient().AuthRequired())
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
	loadFeedbackRoutes(apiGroup)
	apiGroup.POST("/batch_request", handlers.BatchRequestHandler(router))
}
