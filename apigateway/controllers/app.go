package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/services"
)

// GetAppVersionList 获取APP版本列表
func GetAppVersionList(c *gin.Context) {
	appVersionService := services.NewAppVersionService()
	result := appVersionService.FindAll()
	response.Success(c, result)
}

// GetLatestAppVersion 获取最新的版本信息
func GetLatestAppVersion(c *gin.Context) {
	userId, _ := auth.GetUserId(c)

	appVersionService := services.NewAppVersionService()
	result := appVersionService.GetLatest(userId)
	response.Success(c, result)
}
