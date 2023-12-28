package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
)

// PostFeedback 提交反馈
func PostFeedback(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		Type    models.FeedbackType `json:"type" form:"type"`
		Content string              `json:"content" form:"content" binding:"required"`
		PageUrl string              `json:"page_url" form:"page_url" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Type >= models.FeedbackTypeLast || req.Content == "" || req.PageUrl == "" {
		response.BadRequest(c, "")
		return
	}
	feedbackService := services.NewFeedbackService()
	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, "参数错误：files")
	}
	fileHeaderList := form.File["files"]
	for _, fileHeader := range fileHeaderList {
		if fileHeader.Size > 2<<20 {
			response.BadRequest(c, "文件大小不能超过2MB")
			return
		}
	}
	imagePathList := []string{}
	for _, fileHeader := range fileHeaderList {
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
		imagePath, err := feedbackService.UploadImage(userId, fileBytes, contentType)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
		imagePathList = append(imagePathList, imagePath)
	}
	imagePathListJson := ""
	imagePathListJsonByte, err := json.Marshal(imagePathList)
	if err == nil {
		imagePathListJson = string(imagePathListJsonByte)
	}
	if err := feedbackService.Create(&models.Feedback{
		UserId:        userId,
		Type:          req.Type,
		Content:       req.Content,
		ImagePathList: imagePathListJson,
		PageUrl:       req.PageUrl,
	}); err != nil {
		response.Fail(c, "")
	}
	response.Success(c, nil)
}
