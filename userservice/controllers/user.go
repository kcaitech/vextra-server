package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"strings"
)

type userInfoResp struct {
	models.DefaultModelData
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// GetUserInfo 获取用户信息
func GetUserInfo(c *gin.Context) {
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
	if strings.HasPrefix(user.Avatar, "/") {
		user.Avatar = common.StorageHost + user.Avatar
	}
	response.Success(c, user)
}

// SetNickname 设置昵称
func SetNickname(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		Nickname string `json:"nickname" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userService := services.NewUserService()
	if userService.UpdateColumnsById(userId, map[string]any{
		"nickname": req.Nickname,
	}) != nil {
		response.BadRequest(c, "")
		return
	}
	response.Success(c, "")
}

// SetAvatar 设置头像
func SetAvatar(c *gin.Context) {
	user, err := auth.GetUser(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "参数错误：file")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.BadRequest(c, "获取文件失败")
		return
	}
	defer file.Close()
	fileSize := fileHeader.Size
	contentType := fileHeader.Header.Get("Content-Type")
	avatarPath, err := services.NewUserService().UploadUserAvatar(user, file, fileSize, contentType)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, map[string]any{
		"avatar": common.StorageHost + avatarPath,
	})
}
