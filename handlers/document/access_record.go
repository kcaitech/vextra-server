/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package document

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
)

type AccessRecordQueryResItem struct {
	services.AccessRecordAndFavoritesQueryResItem
	Thumbnail *common.ThumbnailResponse `json:"thumbnail,omitempty"`
}

// GetUserDocumentAccessRecordsList 获取用户的文档访问记录列表
func GetUserDocumentAccessRecordsList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
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

	userMap, err, statusCode := GetUsersInfo(c, userIds)
	if err != nil {
		if statusCode == http.StatusUnauthorized {
			common.Unauthorized(c)
			return
		}
		common.ServerError(c, err.Error())
		return
	}

	result := make([]AccessRecordQueryResItem, 0)
	for _, item := range *accessRecordsList {
		userId := item.Document.UserId
		userInfo, exists := userMap[userId]
		if exists {
			item.User = &models.UserProfile{
				Id:       userInfo.UserID,
				Nickname: userInfo.Nickname,
				Avatar:   userInfo.Avatar,
			}
		}
		thumbnail, _ := common.GetDocumentThumbnailAccessKeyNoCheckAuth(c, item.Document.Id, services.GetStorageClient())
		result = append(result, AccessRecordQueryResItem{
			AccessRecordAndFavoritesQueryResItem: item,
			Thumbnail:                            thumbnail,
		})
	}

	// 构建包含下一页游标信息的响应
	var nextCursor string
	if hasMore && len(result) > 0 {
		// 使用最后一条记录的访问时间作为下一页的游标
		lastItem := result[len(result)-1]
		nextCursor = lastItem.DocumentAccessRecord.LastAccessTime.Format(time.RFC3339)
	}

	common.SuccessWithCursor(c, result, hasMore, nextCursor)
}

// DeleteUserDocumentAccessRecord 删除用户的某条文档访问记录
func DeleteUserDocumentAccessRecord(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	accessRecordId := c.Query("access_record_id")
	if accessRecordId == "" {
		common.BadRequest(c, "参数错误：access_record_id")
		return
	}
	if _, err := services.NewDocumentService().DocumentAccessRecordService.Delete(
		"user_id = ? and id = ?", userId, accessRecordId,
	); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		common.ServerError(c, "删除错误")
		return
	}
	common.Success(c, "")
}
