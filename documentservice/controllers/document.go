package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
)

// GetUserDocumentList 获取用户的文档列表
func GetUserDocumentList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var resp []services.DocumentQueryResItem
	if err := services.NewDocumentService().Find(&resp, "document.user_id = ?", userId,
		services.JoinArgs{Join: "inner join user on user.id = document.user_id"},
		services.SelectArgs{Select: "document.*, user.*"},
	); err != nil {
		response.Fail(c, "")
		return
	}

	response.Success(c, resp)
}

// DeleteUserDocument 删除用户的某份文档
func DeleteUserDocument(c *gin.Context) {
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
	if err := services.NewDocumentService().Delete(
		"user_id = ? and id = ?", userId, documentId,
	); err != nil && err != services.ErrRecordNotFound {
		response.Fail(c, "删除错误")
		return
	}
	response.Success(c, "")
}

// GetUserDocumentInfo 获取用户某份文档的信息
func GetUserDocumentInfo(c *gin.Context) {
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
	permType := models.PermType(str.DefaultToInt(c.Query("perm_type"), 0))
	if permType < models.PermTypeReadOnly || permType > models.PermTypeEditable {
		response.BadRequest(c, "参数错误：perm_type")
		return
	}
	response.Success(c, services.NewDocumentService().GetDocumentInfoByDocumentAndUserId(documentId, userId, permType))
}
