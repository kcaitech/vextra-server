package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
)

// UserInfo 获取用户信息
func UserInfo(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	userService := services.NewUserService()
	user := &models.User{}
	if userService.GetById(userId, user) != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, user)
}
