package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/userservice/services"
	"protodesign.cn/kcserver/utils/gin/response"
)

// UserInfo 获取用户信息
func UserInfo(c *gin.Context) {
	id, exists := c.Get("Id")
	_, ok := id.(uint)
	if !exists || !ok {
		response.Unauthorized(c)
		return
	}

	userService := services.NewUserService()
	user, err := userService.GetUser(id.(uint))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, user)
}
