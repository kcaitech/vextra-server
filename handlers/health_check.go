package handlers

import (
	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	// todo 需要检查数据库连接是否正常
	common.Success(c, map[string]string{
		"status": "healthy",
	})
}
