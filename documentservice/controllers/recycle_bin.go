package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
	"time"
)

// GetUserRecycleBinDocumentList 获取用户回收站文档列表
func GetUserRecycleBinDocumentList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	response.Success(c, services.NewDocumentService().FindRecycleBinByUserId(userId))
}

type RestoreUserRecycleBinDocumentReq struct {
	DocId string `json:"doc_id" binding:"required"`
}

// RestoreUserRecycleBinDocument 恢复用户回收站内的某份文档
func RestoreUserRecycleBinDocument(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req RestoreUserRecycleBinDocumentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := str.DefaultToInt(req.DocId, 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	if err := services.NewDocumentService().UpdateColumns(
		map[string]any{"deleted_at": nil},
		"user_id = ? and id = ? and deleted_at is not null and purged_at is null", userId, documentId,
		services.Unscoped{},
	); err != nil && err != services.ErrRecordNotFound {
		response.Fail(c, "更新错误")
		return
	}
	response.Success(c, "")
}

// DeleteUserRecycleBinDocument 彻底删除用户回收站内的某份文档
func DeleteUserRecycleBinDocument(c *gin.Context) {
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
	if err := services.NewDocumentService().UpdateColumns(
		map[string]any{"purged_at": myTime.Time(time.Now())},
		"user_id = ? and id = ? and deleted_at is not null", userId, documentId,
		services.Unscoped{},
	); err != nil && err != services.ErrRecordNotFound {
		response.Fail(c, "更新错误")
		return
	}
	response.Success(c, "")
}
