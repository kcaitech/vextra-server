package document

import (
	"errors"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
)

// GetUserDocumentAccessRecordsList 获取用户的文档访问记录列表
func GetUserDocumentAccessRecordsList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	accessRecordsList := services.NewDocumentService().FindAccessRecordsByUserId(userId)
	// 获取用户信息
	userIds := make([]string, 0)
	for _, item := range *accessRecordsList {
		userIds = append(userIds, item.Document.UserId)
	}

	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	result := make([]services.AccessRecordAndFavoritesQueryResItem, 0)
	for _, item := range *accessRecordsList {
		userId := item.Document.UserId
		userInfo, exists := userMap[userId]
		if exists {
			item.User = &models.UserProfile{
				Id:       userInfo.UserID,
				Nickname: userInfo.Profile.Nickname,
				Avatar:   userInfo.Profile.Avatar,
			}
			result = append(result, item)
		}
	}
	response.Success(c, result)
}

// DeleteUserDocumentAccessRecord 删除用户的某条文档访问记录
func DeleteUserDocumentAccessRecord(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	accessRecordId := c.Query("access_record_id")
	if accessRecordId == "" {
		response.BadRequest(c, "参数错误：access_record_id")
		return
	}
	if _, err := services.NewDocumentService().DocumentAccessRecordService.Delete(
		"user_id = ? and id = ?", userId, accessRecordId,
	); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.ServerError(c, "删除错误")
		return
	}
	response.Success(c, "")
}
