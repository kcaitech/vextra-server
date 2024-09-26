package controllers

import (
	"kcaitech.com/kcserver/common/gin/response"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	response.Success(c, map[string]string{
		"status": "healthy",
	})
}
