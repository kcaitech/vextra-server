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
	"kcaitech.com/kcserver/handlers/common"
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

func joinPath(dir, path string) string {
	if !strings.HasPrefix(path, "/") {
		return dir + "/" + path
	}
	return dir + path
}

func onNotFound(c *gin.Context) {
	path := c.Request.URL.Path
	if path == "/api" || strings.HasPrefix(path, "/api/") {
		c.JSON(http.StatusNotFound, gin.H{"error": "auth endpoint not found"})
		return
	}

	if isStaticFile(path) {
		c.File(joinPath(staticFilePath, path))
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
	apiGroup.GET("/version.json", func(c *gin.Context) {
		c.File("/app/version.json")
	})
	loadWsRoutes(apiGroup)    // 单独鉴权
	loadLoginRoutes(apiGroup) // 从refreshToken获取信息
	apiGroup.Use(services.GetKCAuthClient().AuthRequired())
	apiGroup.Use(common.Sha1SaveData)
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
	loadShareRoutes(apiGroup)
	loadTeamRoutes(apiGroup)
	loadFeedbackRoutes(apiGroup)
}
