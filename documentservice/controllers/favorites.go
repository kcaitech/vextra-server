package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
)

// GetUserDocumentFavoritesList 获取用户收藏的文档列表
func GetUserDocumentFavoritesList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	response.Success(c, services.NewDocumentService().FindFavoritesByUserId(userId))
}

type SetUserDocumentFavoriteStatusReq struct {
	DocId  string `json:"doc_id" binding:"required"`
	Status bool   `json:"status" binding:"required"`
}

// SetUserDocumentFavoriteStatus 设置用户对某份文档的收藏状态
func SetUserDocumentFavoriteStatus(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req SetUserDocumentFavoriteStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := str.DefaultToInt(req.DocId, 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	documentFavoritesService := services.NewDocumentService().DocumentFavoritesService
	documentFavorites := models.DocumentFavorites{}
	if err := documentFavoritesService.Get(
		&documentFavorites,
		"document_favorites.user_id = ? and document_favorites.document_id = ? and document.deleted_at is null", userId, documentId,
		services.JoinArgs{Join: "inner join document on document.id = document_favorites.document_id"},
		services.SelectArgs{Select: "document_favorites.*"},
	); err != nil {
		if err != services.ErrRecordNotFound {
			response.Fail(c, "查询错误")
			return
		}
		documentFavorites.UserId = userId
		documentFavorites.DocumentId = documentId
		documentFavorites.IsFavorite = req.Status
		if err := documentFavoritesService.Create(&documentFavorites); err != nil {
			response.Fail(c, "创建失败")
			return
		}
		response.Success(c, "")
		return
	}
	documentFavorites.IsFavorite = req.Status
	if err := documentFavoritesService.UpdatesById(documentFavorites.Id, &documentFavorites); err != nil {
		response.Fail(c, "更新失败")
		return
	}
	response.Success(c, "")
}
