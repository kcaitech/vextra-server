package controllers

import (
	"encoding/base64"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/gin/auth"
	"kcaitech.com/kcserver/common/gin/response"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/safereview"
	safereviewBase "kcaitech.com/kcserver/common/safereview/base"
	"kcaitech.com/kcserver/common/services"
	config "kcaitech.com/kcserver/controllers"
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
		user.Avatar = config.Config.StorageUrl.Attatch + user.Avatar
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

	reviewResponse, err := safereview.Client.ReviewText(req.Nickname)
	if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
		log.Println("昵称审核不通过", req.Nickname, err, reviewResponse)
		response.Fail(c, "审核不通过")
		return
	}

	userService := services.NewUserService()
	if _, err := userService.UpdateColumnsById(userId, map[string]any{
		"nickname": req.Nickname,
	}); err != nil {
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
	if fileHeader.Size > 2<<20 {
		response.BadRequest(c, "文件大小不能超过2MB")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.BadRequest(c, "获取文件失败")
		return
	}
	defer file.Close()
	fileBytes := make([]byte, fileHeader.Size)
	if _, err := file.Read(fileBytes); err != nil {
		response.BadRequest(c, "读取文件失败")
		return
	}
	contentType := fileHeader.Header.Get("Content-Type")
	base64Str := base64.StdEncoding.EncodeToString(fileBytes)
	reviewResponse, err := safereview.Client.ReviewPictureFromBase64(base64Str)
	if err != nil {
		log.Println("头像审核失败", err)
		response.Fail(c, "头像审核失败")
		return

	} else if reviewResponse.Status != safereviewBase.ReviewImageResultPass {
		log.Println("头像审核不通过", err, reviewResponse)
		response.Fail(c, "头像审核不通过")
		return
	}
	avatarPath, err := services.NewUserService().UploadUserAvatar(user, fileBytes, contentType)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, map[string]any{
		"avatar": config.Config.StorageUrl.Attatch + avatarPath,
	})
}
