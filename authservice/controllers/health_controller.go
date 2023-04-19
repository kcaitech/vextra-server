package controllers

import (
	"protodesign.cn/kcserver/utils/gin/response"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	response.Success(c, map[string]string{
		"status": "healthy",
	})
}
