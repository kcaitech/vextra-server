package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
)

type UserDocumentPermResp struct {
	PermType models.PermType `json:"perm_type"`
}

// GetUserDocumentPerm 获取文档权限
func GetUserDocumentPerm(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	documentId := str.DefaultToInt(c.Query("doc_id"), 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	result := UserDocumentPermResp{}
	if err := services.NewDocumentService().GetDocumentPermissionByDocumentAndUserId(&result.PermType, documentId, userId); err != nil {
		response.Fail(c, "")
		return
	}
	response.Success(c, result)
}
