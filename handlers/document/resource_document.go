package document

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
)

type ResourceDocumentQueryResItem struct {
	services.ResourceDocumentQueryResItem
	Thumbnail *common.ThumbnailResponse `json:"thumbnail,omitempty"`
}

func GetResourceDocumentList(c *gin.Context) {
	cursor := c.Query("cursor")
	limit := utils.QueryInt(c, "limit", 20)

	documentService := services.NewDocumentService()
	resourceDocuments, hasMore := documentService.FindResourceDocuments(cursor, limit)

	userIds := make([]string, 0)

	for _, item := range *resourceDocuments {
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

	result := make([]ResourceDocumentQueryResItem, 0)

	for _, item := range *resourceDocuments {
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
		result = append(result, ResourceDocumentQueryResItem{
			ResourceDocumentQueryResItem: item,
			Thumbnail:                    thumbnail,
		})
	}

	var nextCursor string
	if hasMore && len(result) > 0 {
		lastItem := result[len(result)-1]
		nextCursor = lastItem.Document.CreatedAt.Format(time.RFC3339)
	}

	common.SuccessWithCursor(c, result, hasMore, nextCursor)
}

type CreateResourceDocumentReq struct {
	DocId       string `json:"doc_id" binding:"required"`
	Description string `json:"description"`
}

func CreateResourceDocument(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}

	var req CreateResourceDocumentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "")
		return
	}

	documentId := req.DocId
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}

	// 复制文档
	documentService := services.NewDocumentService()
	result := copyDocument(userId, documentId, c, "%s_资源", services.GetStorageClient(), false)
	if result == nil {
		return
	}

	// 创建资源文档记录
	resourceDocument := models.ResourceDocument{
		UserId:      userId,
		DocumentId:  result.CopyId,
		References:  0,
		Description: req.Description,
	}

	if err := documentService.ResourceDocumentService.Create(&resourceDocument); err != nil {
		common.ServerError(c, "创建资源文档失败")
		return
	}

	common.Success(c, result)
}
