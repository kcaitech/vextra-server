package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
)

type userInfoResp struct {
	models.DefaultModelData
	Id       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// UserInfo 获取用户信息
func UserInfo(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	userService := services.NewUserService()
	user := &userInfoResp{}
	if userService.GetById(userId, user) != nil {
		response.BadRequest(c, "")
		return
	}

	response.Success(c, user)
}
