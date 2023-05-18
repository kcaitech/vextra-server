package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
)

// GetUserDocumentAccessRecordsList 获取用户的文档访问记录列表
func GetUserDocumentAccessRecordsList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	response.Success(c, services.NewDocumentService().FindAccessRecordsByUserId(userId))
}

// DeleteUserDocumentAccessRecord 删除用户的某条文档访问记录
func DeleteUserDocumentAccessRecord(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	accessRecordId := str.DefaultToInt(c.Query("access_record_id"), 0)
	if accessRecordId <= 0 {
		response.BadRequest(c, "参数错误：access_record_id")
		return
	}
	if err := services.NewDocumentService().DocumentAccessRecordService.HardDelete(
		"user_id = ? and id = ?", userId, accessRecordId,
	); err != nil && err != services.ErrRecordNotFound {
		response.Fail(c, "删除错误")
		return
	}
	response.Success(c, "")
}
