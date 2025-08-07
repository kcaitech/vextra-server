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

// GetUserDocumentFavoritesList 获取用户收藏的文档列表
func GetUserDocumentFavoritesList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}

	projectId := c.Query("project_id")
	cursor := c.Query("cursor")
	limit := utils.QueryInt(c, "limit", 20) // 默认每页20条

	documentService := services.NewDocumentService()
	favoritesList, hasMore := documentService.FindFavoritesByUserIdWithCursor(userId, projectId, cursor, limit)

	// 获取用户信息
	userIds := make([]string, 0)
	for _, item := range *favoritesList {
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

	result := make([]services.AccessRecordAndFavoritesQueryResItem, 0)
	for _, item := range *favoritesList {
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

	common.SuccessWithCursor(c, result, hasMore, nextCursor)
}

type SetUserDocumentFavoriteStatusReq struct {
	DocId  string `json:"doc_id" binding:"required"`
	Status bool   `json:"status"`
}

// SetUserDocumentFavoriteStatus 设置用户对某份文档的收藏状态
func SetUserDocumentFavoriteStatus(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req SetUserDocumentFavoriteStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "")
		return
	}
	documentId := (req.DocId)
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}
	documentService := services.NewDocumentService()
	if exist, err := documentService.Exist("id = ?", documentId); !exist || err != nil {
		common.BadRequest(c, "文档不存在")
		return
	}
	documentFavoritesService := documentService.DocumentFavoritesService
	documentFavorites := models.DocumentFavorites{}
	if err := documentFavoritesService.Get(
		&documentFavorites,
		"document_favorites.user_id = ? and document_favorites.document_id = ? and document.deleted_at is null", userId, documentId,
		services.JoinArgsRaw{Join: "inner join document on document.id = document_favorites.document_id"},
		services.SelectArgs{Select: "document_favorites.*"},
	); err != nil {
		if !errors.Is(err, services.ErrRecordNotFound) {
			common.ServerError(c, "查询错误")
			return
		}
		documentFavorites.UserId = userId
		documentFavorites.DocumentId = documentId
		documentFavorites.IsFavorite = req.Status
		if err := documentFavoritesService.Create(&documentFavorites); err != nil {
			common.ServerError(c, "创建失败")
			return
		}
		common.Success(c, "")
		return
	}
	documentFavorites.IsFavorite = req.Status
	if _, err := documentFavoritesService.UpdatesById(documentFavorites.Id, &documentFavorites); err != nil {
		common.ServerError(c, "更新失败")
		return
	}
	common.Success(c, "")
}
