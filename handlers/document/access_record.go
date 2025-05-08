package document

import (
	"errors"
	"time"

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

	cursor := c.Query("cursor")
	limit := utils.QueryInt(c, "limit", 20) // 默认每页20条

	documentService := services.NewDocumentService()
	accessRecordsList, hasMore := documentService.FindAccessRecordsByUserIdWithCursor(userId, cursor, limit)

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
				Nickname: userInfo.Nickname,
				Avatar:   userInfo.Avatar,
			}
			result = append(result, item)
		}
	}

	// 构建包含下一页游标信息的响应
	var nextCursor string
	if hasMore && len(result) > 0 {
		// 使用最后一条记录的访问时间作为下一页的游标
		lastItem := result[len(result)-1]
		nextCursor = lastItem.DocumentAccessRecord.LastAccessTime.Format(time.RFC3339)
	}

	response.Success(c, gin.H{
		"list":        result,
		"has_more":    hasMore,
		"next_cursor": nextCursor,
	})
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
