/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package handlers

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
)

// PostFeedback 提交反馈
func PostFeedback(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req struct {
		Type    models.FeedbackType `json:"type" form:"type"`
		Content string              `json:"content" form:"content" binding:"required"`
		PageUrl string              `json:"page_url" form:"page_url" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}
	if req.Type >= models.FeedbackTypeLast || req.Content == "" || req.PageUrl == "" {
		common.BadRequest(c, "")
		return
	}
	feedbackService := services.NewFeedbackService()
	form, err := c.MultipartForm()
	if err != nil {
		common.BadRequest(c, "参数错误：files")
		return
	}
	fileHeaderList := form.File["files"]
	for _, fileHeader := range fileHeaderList {
		if fileHeader.Size > 2<<20 {
			common.BadRequest(c, "文件大小不能超过2MB")
			return
		}
	}
	imagePathList := []string{}
	for _, fileHeader := range fileHeaderList {
		file, err := fileHeader.Open()
		if err != nil {
			common.BadRequest(c, "获取文件失败")
			return
		}
		defer file.Close()
		fileBytes := make([]byte, fileHeader.Size)
		if _, err := file.Read(fileBytes); err != nil {
			common.BadRequest(c, "读取文件失败")
			return
		}
		contentType := fileHeader.Header.Get("Content-Type")
		imagePath, err := feedbackService.UploadImage(userId, fileBytes, contentType)
		if err != nil {
			common.ServerError(c, err.Error())
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
		common.ServerError(c, "")
	}
	common.Success(c, nil)
}
