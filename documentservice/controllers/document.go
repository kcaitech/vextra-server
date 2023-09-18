package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/mongo"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/snowflake"
	"protodesign.cn/kcserver/common/storage"
	"protodesign.cn/kcserver/utils/sliceutil"
	"protodesign.cn/kcserver/utils/str"
)

// GetUserDocumentList 获取用户的文档列表
func GetUserDocumentList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := str.DefaultToInt(c.Query("project_id"), 0)
	var result any
	if projectId > 0 {
		result = services.NewDocumentService().FindDocumentByProjectId(projectId, userId)
	} else {
		result = services.NewDocumentService().FindDocumentByUserId(userId)
	}
	response.Success(c, result)
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
	documentService := services.NewDocumentService()
	document := models.Document{}
	if documentService.GetById(documentId, &document) != nil {
		response.BadRequest(c, "文档不存在")
		return
	}
	if document.ProjectId == 0 {
		if _, err := documentService.Delete(
			"user_id = ? and id = ?", userId, documentId,
		); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
			response.Fail(c, "删除错误")
			return
		}
		_, _ = documentService.UpdateColumns(map[string]any{"delete_by": userId}, "deleted_at is not null and id = ?", documentId, &services.Unscoped{})
	} else {
		projectService := services.NewProjectService()
		projectPermType, err := projectService.GetProjectPermTypeByForUser(document.ProjectId, userId)
		if err != nil || projectPermType == nil || (*projectPermType) < models.ProjectPermTypeEditable {
			response.Forbidden(c, "")
			return
		}
		if _, err := documentService.Delete(
			"id = ?", documentId,
		); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
			response.Fail(c, "删除错误")
			return
		}
		_, _ = documentService.UpdateColumns(map[string]any{"delete_by": userId}, "deleted_at is not null and id = ?", documentId, &services.Unscoped{})
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
	documentService := services.NewDocumentService()
	document := models.Document{}
	if documentService.Get(&document, "id = ?", documentId) != nil {
		response.BadRequest(c, "文档不存在")
		return
	}
	if document.ProjectId <= 0 {
		if document.UserId != userId {
			response.Forbidden(c, "")
			return
		}
	} else {
		projectService := services.NewProjectService()
		projectPermType, err := projectService.GetProjectPermTypeByForUser(document.ProjectId, userId)
		// 管理员以上权限或文档创建者且可编辑权限
		if !(err == nil && projectPermType != nil && ((*projectPermType) >= models.ProjectPermTypeAdmin || ((*projectPermType) == models.ProjectPermTypeEditable && document.UserId == userId))) {
			response.Forbidden(c, "")
			return
		}
	}
	if _, err = services.NewDocumentService().UpdateColumns(
		map[string]any{"name": req.Name},
		"id = ?", documentId,
	); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "更新错误")
		return
	}
	if errors.Is(err, services.ErrRecordNotFound) {
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
	if err = documentService.Get(&sourceDocument, "id = ?", documentId); err != nil {
		response.Forbidden(c, "")
		return
	}
	if sourceDocument.ProjectId <= 0 {
		if sourceDocument.UserId != userId {
			response.Forbidden(c, "")
			return
		}
	} else {
		projectService := services.NewProjectService()
		projectPermType, err := projectService.GetProjectPermTypeByForUser(sourceDocument.ProjectId, userId)
		// 管理员以上权限或文档创建者且可编辑权限
		if !(err == nil && projectPermType != nil && ((*projectPermType) >= models.ProjectPermTypeAdmin || ((*projectPermType) == models.ProjectPermTypeEditable && sourceDocument.UserId == userId))) {
			response.Forbidden(c, "")
			return
		}
	}
	// 复制目录
	targetDocumentPath := uuid.New().String()
	if _, err := storage.Bucket.CopyDirectory(sourceDocument.Path, targetDocumentPath); err != nil {
		log.Println("复制目录失败：", err)
		response.Fail(c, "复制失败")
		return
	}
	// 复制文档
	targetDocument := models.Document{
		UserId:    userId,
		Path:      targetDocumentPath,
		DocType:   sourceDocument.DocType,
		Name:      sourceDocument.Name + "_副本",
		Size:      sourceDocument.Size,
		TeamId:    sourceDocument.TeamId,
		ProjectId: sourceDocument.ProjectId,
	}
	if err := documentService.Create(&targetDocument); err != nil {
		response.Fail(c, "创建失败")
		return
	}
	// 复制cmd
	type DocumentCmd struct {
		Id         int64  `json:"id" bson:"_id"`
		DocumentId int64  `json:"document_id" bson:"document_id"`
		UserId     int64  `json:"user_id" bson:"user_id"`
		UnitId     string `json:"unit_id" bson:"unit_id"`
		Cmd        bson.M `json:"cmd" bson:"cmd"`
		LastId     string `json:"last_id" bson:"last_id"`
		VersionId  string `json:"version_id" bson:"version_id"`
	}
	documentCmdList := make([]DocumentCmd, 0)
	reqParams := bson.M{
		"document_id": documentId,
		"version_id":  sourceDocument.VersionId,
	}
	documentCollection := mongo.DB.Collection("document")
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"_id", 1}})
	if cur, err := documentCollection.Find(nil, reqParams, findOptions); err == nil {
		_ = cur.All(nil, &documentCmdList)
	}
	lastId := ""
	newDocumentCmdList := sliceutil.MapT(func(item DocumentCmd) any {
		item.Id = snowflake.NextId()
		item.DocumentId = targetDocument.Id
		item.LastId = lastId
		lastId = str.IntToString(item.Id)
		return item
	}, documentCmdList...)
	_, err = documentCollection.InsertMany(nil, newDocumentCmdList)
	if err != nil {
		log.Println("cmd复制失败：", err)
	}
	// 添加最近访问
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
