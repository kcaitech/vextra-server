package controllers

import (
	"kcaitech.com/kcserver/common/response"

	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	// todo 需要检查数据库连接是否正常
	response.Success(c, map[string]string{
		"status": "healthy",
	})
}
