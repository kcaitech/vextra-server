package document

import (
	"errors"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
)

// GetUserDocumentFavoritesList 获取用户收藏的文档列表
func GetUserDocumentFavoritesList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := c.Query("project_id")
	favoritesList := services.NewDocumentService().FindFavoritesByUserId(userId, projectId)
	// 获取用户信息
	userIds := make([]string, 0)
	for _, item := range *favoritesList {
		userIds = append(userIds, item.Document.UserId)
	}

	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	result := make([]services.AccessRecordAndFavoritesQueryResItem, 0)
	for _, item := range *favoritesList {
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

type SetUserDocumentFavoriteStatusReq struct {
	DocId  string `json:"doc_id" binding:"required"`
	Status bool   `json:"status"`
}

// SetUserDocumentFavoriteStatus 设置用户对某份文档的收藏状态
func SetUserDocumentFavoriteStatus(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req SetUserDocumentFavoriteStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := (req.DocId)
	if documentId == "" {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	documentService := services.NewDocumentService()
	if exist, err := documentService.Exist("id = ?", documentId); !exist || err != nil {
		response.BadRequest(c, "文档不存在")
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
			response.ServerError(c, "查询错误")
			return
		}
		documentFavorites.UserId = userId
		documentFavorites.DocumentId = documentId
		documentFavorites.IsFavorite = req.Status
		if err := documentFavoritesService.Create(&documentFavorites); err != nil {
			response.ServerError(c, "创建失败")
			return
		}
		response.Success(c, "")
		return
	}
	documentFavorites.IsFavorite = req.Status
	if _, err := documentFavoritesService.UpdatesById(documentFavorites.Id, &documentFavorites); err != nil {
		response.ServerError(c, "更新失败")
		return
	}
	response.Success(c, "")
}
