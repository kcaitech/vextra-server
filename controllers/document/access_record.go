package document

import (
	"errors"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
	"kcaitech.com/kcserver/utils/str"
)

// GetUserDocumentAccessRecordsList 获取用户的文档访问记录列表
func GetUserDocumentAccessRecordsList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	response.Success(c, services.NewDocumentService().FindAccessRecordsByUserId(userId))
}

// DeleteUserDocumentAccessRecord 删除用户的某条文档访问记录
func DeleteUserDocumentAccessRecord(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	accessRecordId := str.DefaultToInt(c.Query("access_record_id"), 0)
	if accessRecordId <= 0 {
		response.BadRequest(c, "参数错误：access_record_id")
		return
	}
	if _, err := services.NewDocumentService().DocumentAccessRecordService.Delete(
		"user_id = ? and id = ?", userId, accessRecordId,
	); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "删除错误")
		return
	}
	response.Success(c, "")
}
