/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package api

import (
	"context"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	v1 "kcaitech.com/kcserver/api/v1"
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

func joinPath(dir, path string) string {
	if !strings.HasPrefix(path, "/") {
		return dir + "/" + path
	}
	return dir + path
}

func onNotFound(webFilePath string) func(c *gin.Context) {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/api" || strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "auth endpoint not found"})
			return
		}

		if isStaticFile(path) {
			c.File(joinPath(webFilePath, path))
			return
		}

		// 设置缓存时间为15分钟
		c.Header("Cache-Control", "public, max-age=900")
		c.File(webFilePath + "/index.html")
	}
}

func LoadRoutes(router *gin.Engine, webFilePath string) {
	router.RedirectTrailingSlash = false
	router.GET("/health", handlers.HealthCheck)
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(static.Serve("/", static.LocalFile(webFilePath, false))) // 前端工程
	router.NoRoute(onNotFound(webFilePath))

	config := services.GetConfig()
	if config.Middleware.DebugLog {
		router.Use(middlewares.AccessDetailedLogMiddleware())
	} else {
		router.Use(middlewares.AccessLogMiddleware())
	}
	if config.Middleware.Cors {
		router.Use(middlewares.CORSMiddleware()) // 测试时需要
	}

	router.Use(middlewares.NewRateLimiter(&middlewares.RedisStore{
		Client: services.GetRedisDB().Client,
		Ctx:    context.Background(),
	}, middlewares.DefaultRateLimiterConfig()).RateLimitMiddleware())

	apiGroup := router.Group("/api")
	apiGroup.GET("/version.json", func(c *gin.Context) {
		c.File("version.json") // 当前目录
	})

	v1.LoadRoutes(apiGroup.Group("/v1"))
}
