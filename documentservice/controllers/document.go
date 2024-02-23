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
	"protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/mongo"
	"protodesign.cn/kcserver/common/safereview"
	safereviewBase "protodesign.cn/kcserver/common/safereview/base"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/snowflake"
	"protodesign.cn/kcserver/common/storage"
	"protodesign.cn/kcserver/utils/sliceutil"
	"protodesign.cn/kcserver/utils/str"
	"strings"
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
	} else if !result.Document.LockedAt.IsZero() && result.Document.UserId != userId {
		response.Forbidden(c, "审核不通过")
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

	reviewResponse, err := safereview.Client.ReviewText(req.Name)
	if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
		log.Println("名称审核不通过", req.Name, err, reviewResponse)
		response.Fail(c, "审核不通过")
		return
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

func copyDocument(userId int64, documentId int64, c *gin.Context, documentName string) (result *services.AccessRecordAndFavoritesQueryResItem) {
	documentService := services.NewDocumentService()
	sourceDocument := models.Document{}
	if err := documentService.Get(&sourceDocument, "id = ?", documentId); err != nil {
		response.Forbidden(c, "")
		return
	}
	if !sourceDocument.LockedAt.IsZero() {
		response.Forbidden(c, "审核不通过")
		return
	}
	if sourceDocument.ProjectId <= 0 {
		var permType models.PermType
		if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeEditable {
			response.Forbidden(c, "")
			return
		}
	} else {
		projectService := services.NewProjectService()
		projectPermType, err := projectService.GetProjectPermTypeByForUser(sourceDocument.ProjectId, userId)
		if !(err == nil && projectPermType != nil && (*projectPermType) >= models.ProjectPermTypeEditable) { // 需有可编辑权限
			response.Forbidden(c, "")
			return
		}
	}

	if documentName == "" {
		documentName = "%s_副本"
	}
	documentName = strings.ReplaceAll(documentName, "%s", sourceDocument.Name)

	documentVersion := models.DocumentVersion{}
	if err := documentService.DocumentVersionService.Get(&documentVersion, "document_id = ? and version_id = ?", documentId, sourceDocument.VersionId); err != nil {
		response.Fail(c, "获取文档版本失败")
		return
	}

	// 复制目录
	targetDocumentPath := uuid.New().String()
	if _, err := storage.Bucket.CopyDirectory(sourceDocument.Path, targetDocumentPath); err != nil {
		log.Println("复制目录失败：", err)
		response.Fail(c, "复制失败")
		return
	}

	documentMetaBytes, err := storage.Bucket.GetObject(targetDocumentPath + "/document-meta.json")
	if err != nil {
		log.Println("获取document-meta.json失败：", targetDocumentPath+"/document-meta.json", err)
		response.Fail(c, "复制失败")
		return
	}
	documentMeta := map[string]any{}
	if err := bson.UnmarshalExtJSON(documentMetaBytes, true, &documentMeta); err != nil {
		log.Println("documentMetaBytes转json失败：", err)
		response.Fail(c, "复制失败")
		return
	}

	for _, page := range documentMeta["pagesList"].(bson.A) {
		pageItem, ok := page.(map[string]any)
		if !ok {
			log.Println("pageItem转map失败：", err)
			response.Fail(c, "复制失败")
			return
		}
		pageObjectInfo, err := storage.Bucket.GetObjectInfo(targetDocumentPath + "/pages/" + pageItem["id"].(string) + ".json")
		if err != nil {
			log.Println("获取pageObjectInfo失败：", err)
			response.Fail(c, "复制失败")
			return
		}
		pageItem["versionId"] = pageObjectInfo.VersionID
	}

	freeSymbolsOjectInfo, err := storage.Bucket.GetObjectInfo(targetDocumentPath + "/freesymbols.json")
	if err != nil {
		log.Println("获取freeSymbolsOjectInfo失败：", err)
		response.Fail(c, "复制失败")
		return
	}
	documentMeta["freesymbolsVersionId"] = freeSymbolsOjectInfo.VersionID

	documentMetaBytes, err = bson.MarshalExtJSON(documentMeta, true, true)
	if err != nil {
		log.Println("documentMeta转json失败：", err)
		response.Fail(c, "复制失败")
		return
	}
	documentMetaUploadInfo, err := storage.Bucket.PutObjectByte(targetDocumentPath+"/document-meta.json", documentMetaBytes)
	if err != nil {
		log.Println("documentMeta上传失败：", err)
		response.Fail(c, "复制失败")
		return
	}

	// 复制文档
	targetDocument := models.Document{
		UserId:    userId,
		Path:      targetDocumentPath,
		DocType:   sourceDocument.DocType,
		Name:      documentName,
		Size:      sourceDocument.Size,
		TeamId:    sourceDocument.TeamId,
		ProjectId: sourceDocument.ProjectId,
		VersionId: documentMetaUploadInfo.VersionID,
	}
	if err := documentService.Create(&targetDocument); err != nil {
		response.Fail(c, "创建失败")
		return
	}
	// 创建文档版本记录
	if err := documentService.DocumentVersionService.Create(&models.DocumentVersion{
		DocumentId: targetDocument.Id,
		VersionId:  documentMetaUploadInfo.VersionID,
		LastCmdId:  0,
	}); err != nil {
		log.Println("创建文档版本记录失败：", err)
		return
	}

	// 复制cmd
	type DocumentCmd struct {
		Id           int64  `json:"id" bson:"_id"`
		PreviousId   int64  `json:"previous_id" bson:"previous_id"`
		BatchStartId int64  `json:"batch_start_id" bson:"batch_start_id"`
		BatchEndId   int64  `json:"batch_end_id" bson:"batch_end_id"`
		BatchLength  int    `json:"batch_length" bson:"batch_length"`
		DocumentId   int64  `json:"document_id" bson:"document_id"`
		UserId       int64  `json:"user_id" bson:"user_id"`
		CmdId        string `json:"cmd_id" bson:"cmd_id"`
		Cmd          bson.M `json:"cmd" bson:"cmd"`
	}
	documentCmdList := make([]DocumentCmd, 0)
	reqParams := bson.M{
		"document_id": documentId,
		"_id": bson.M{
			"$gt": documentVersion.LastCmdId,
		},
	}
	documentCollection := mongo.DB.Collection("document1")
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"_id", 1}})
	if cur, err := documentCollection.Find(nil, reqParams, findOptions); err == nil {
		_ = cur.All(nil, &documentCmdList)
	}
	lastId := int64(0)
	newDocumentCmdList := sliceutil.MapT(func(item DocumentCmd) DocumentCmd {
		item.Id = snowflake.NextId()
		item.PreviousId = lastId
		item.DocumentId = targetDocument.Id
		lastId = item.Id
		return item
	}, documentCmdList...)
	idMap := map[int64]int64{
		0: 0,
	}
	for i := 0; i < len(newDocumentCmdList); i++ {
		idMap[documentCmdList[i].Id] = newDocumentCmdList[i].Id
	}
	for i := 0; i < len(newDocumentCmdList); i++ {
		item := &newDocumentCmdList[i]
		item.BatchStartId = idMap[item.BatchStartId]
		item.BatchEndId = idMap[item.BatchEndId]
	}
	if len(newDocumentCmdList) > 0 {
		newDocumentCmdList[0].Cmd["baseVer"] = ""
		_, err = documentCollection.InsertMany(nil, sliceutil.ConvertToAnySlice(newDocumentCmdList))
		if err != nil {
			log.Println("cmd复制失败：", err)
		}
	}

	// 复制评论数据
	documentCommentList := make([]UserComment, 0)
	reqParams = bson.M{
		"document_id": str.IntToString(documentId),
	}
	commentCollection := mongo.DB.Collection("comment")
	findOptions = options.Find()
	findOptions.SetSort(bson.D{{"record_created_at", -1}, {"_id", -1}})
	if cur, err := commentCollection.Find(nil, reqParams, findOptions); err == nil {
		_ = cur.All(nil, &documentCommentList)
	}
	commentIdMap := map[string]string{}
	for i := 0; i < len(documentCommentList); i++ {
		item := &documentCommentList[i]
		newId := str.IntToString(snowflake.NextId())
		commentIdMap[item.Id] = newId
		item.Id = newId
		item.DocumentId = str.IntToString(targetDocument.Id)
	}
	newDocumentCommentList := sliceutil.MapT(func(item UserComment) any {
		item.ParentId = commentIdMap[item.ParentId]
		item.RootId = commentIdMap[item.RootId]
		return item
	}, documentCommentList...)
	if len(newDocumentCommentList) > 0 {
		_, err = commentCollection.InsertMany(nil, newDocumentCommentList)
		if err != nil {
			log.Println("评论复制失败：", err)
		}
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
	if len(resultList) > 0 {
		result = &resultList[0]
	}
	return
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
	result := copyDocument(userId, documentId, c, "%s_副本")
	if result != nil {
		response.Success(c, result)
	}
}

// CreateTest 创建测试文档环境
func CreateTest(c *gin.Context) {
	var req struct {
		VerifyCode   string `json:"verify_code" binding:"required"`
		UserId       string `json:"user_id" binding:"required"`
		DocumentId   string `json:"document_id" binding:"required"`
		DocumentId2  string `json:"document_id2"`
		DocumentName string `json:"document_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}

	if req.VerifyCode != "123456" {
		response.Forbidden(c, "")
		return
	}

	result := map[string]any{}

	if req.DocumentId2 == "" {
		userId := str.DefaultToInt(req.UserId, 0)
		if userId <= 0 {
			response.BadRequest(c, "参数错误：user_id")
			return
		}

		documentId := str.DefaultToInt(req.DocumentId, 0)
		if documentId <= 0 {
			response.BadRequest(c, "参数错误：document_id")
			return
		}

		result1 := copyDocument(userId, documentId, c, req.DocumentName)
		if result1 == nil {
			response.Fail(c, "创建文档失败")
			return
		}
		models.StructToMap(result1, result)
	} else {
		result["document"] = map[string]any{
			"id": req.DocumentId2,
		}
	}

	// 创建JWT
	token, err := jwt.CreateJwt(&jwt.Data{
		Id:       req.UserId,
		Nickname: "测试用户",
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	result["token"] = token

	response.Success(c, result)
}
