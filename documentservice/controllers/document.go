package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/storage"
	"protodesign.cn/kcserver/utils/str"
)

// GetUserDocumentList 获取用户的文档列表
func GetUserDocumentList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	response.Success(c, services.NewDocumentService().FindDocumentByUserId(userId))
}

// DeleteUserDocument 删除用户的某份文档
func DeleteUserDocument(c *gin.Context) {
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
	if _, err := services.NewDocumentService().Delete(
		"user_id = ? and id = ?", userId, documentId,
	); err != nil && err != services.ErrRecordNotFound {
		response.Fail(c, "删除错误")
		return
	}
	response.Success(c, "")
}

// GetUserDocumentInfo 获取用户某份文档的信息
func GetUserDocumentInfo(c *gin.Context) {
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
	permType := models.PermType(str.DefaultToInt(c.Query("perm_type"), 0))
	if permType < models.PermTypeReadOnly || permType > models.PermTypeEditable {
		permType = models.PermTypeNone
	}
	if result := services.NewDocumentService().GetDocumentInfoByDocumentAndUserId(documentId, userId, permType); result == nil {
		response.BadRequest(c, "文档不存在")
	} else {
		response.Success(c, result)
	}
}

type SetDocumentNameReq struct {
	DocId string `json:"doc_id" binding:"required"`
	Name  string `json:"name" binding:"required"`
}

// SetDocumentName 设置文档名称
func SetDocumentName(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req SetDocumentNameReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := str.DefaultToInt(req.DocId, 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	if _, err = services.NewDocumentService().UpdateColumns(
		map[string]any{"name": req.Name},
		"user_id = ? and id = ?", userId, documentId,
	); err != nil && err != services.ErrRecordNotFound {
		response.Fail(c, "删除错误")
		return
	}
	if err == services.ErrRecordNotFound {
		response.Forbidden(c, "")
		return
	}
	response.Success(c, "")
}

type CopyDocumentReq struct {
	DocId string `json:"doc_id" binding:"required"`
}

// CopyDocument 复制文档
func CopyDocument(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req CopyDocumentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := str.DefaultToInt(req.DocId, 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	documentService := services.NewDocumentService()
	sourceDocument := models.Document{}
	if err = documentService.Get(&sourceDocument, "id = ? and user_id = ?", documentId, userId); err != nil {
		response.Forbidden(c, "")
		return
	}
	targetDocumentPath := uuid.New().String()
	if _, err := storage.Bucket.CopyDirectory(sourceDocument.Path, targetDocumentPath); err != nil {
		log.Println("复制目录失败：", err)
		response.Fail(c, "复制失败")
		return
	}
	targetDocument := models.Document{
		UserId:  userId,
		Path:    targetDocumentPath,
		DocType: sourceDocument.DocType,
		Name:    sourceDocument.Name + "_副本",
		Size:    sourceDocument.Size,
	}
	if err := documentService.Create(&targetDocument); err != nil {
		response.Fail(c, "创建失败")
		return
	}
	documentAccessRecord := models.DocumentAccessRecord{
		UserId:     userId,
		DocumentId: targetDocument.Id,
	}
	_ = documentService.DocumentAccessRecordService.Create(&documentAccessRecord)
	var resultList []services.AccessRecordAndFavoritesQueryResItem
	_ = documentService.DocumentAccessRecordService.Find(
		&resultList,
		&services.ParamArgs{"?user_id": userId},
		&services.WhereArgs{Query: "document_access_record.id = ? and document.deleted_at is null", Args: []any{documentAccessRecord.Id}},
	)
	var result *services.AccessRecordAndFavoritesQueryResItem
	if len(resultList) > 0 {
		result = &resultList[0]
	}
	response.Success(c, result)
}
