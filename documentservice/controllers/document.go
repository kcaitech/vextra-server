package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
)

// DocumentUserList 获取用户文档列表
func DocumentUserList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	documentService := services.NewDocumentService()
	var documentList []models.Document
	if err := documentService.Find(&documentList, "user_id = ?", userId); err != nil {
		response.Fail(c, "")
		return
	}

	response.Success(c, documentList)
}

// DocumentUserAccessRecordsList 获取用户的文档访问记录
func DocumentUserAccessRecordsList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	documentService := services.NewDocumentService()
	result := documentService.FindAccessRecordsByUserId(userId)

	response.Success(c, result)
}
