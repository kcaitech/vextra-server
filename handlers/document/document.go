package document

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/mongo"
	safereviewBase "kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
	"kcaitech.com/kcserver/utils/str"
)

// GetUserDocumentList 获取用户的文档列表
func GetUserDocumentList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := c.Query("project_id")
	var result *[]services.AccessRecordAndFavoritesQueryResItem
	if projectId != "" {
		result = services.NewDocumentService().FindDocumentByProjectId(projectId, userId)
	} else {
		result = services.NewDocumentService().FindDocumentByUserId(userId)
	}
	// 获取文档相关用户信息
	userIds := make([]string, 0)
	for _, item := range *result {
		userIds = append(userIds, item.Document.UserId)
	}
	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	for i := range *result {
		item := &(*result)[i]
		userInfo, ok := userMap[item.Document.UserId]
		if ok {
			item.User = &models.UserProfile{
				Id:       userInfo.UserID,
				Nickname: userInfo.Profile.Nickname,
				Avatar:   userInfo.Profile.Avatar,
			}
		}
	}
	response.Success(c, result)
}

// DeleteUserDocument 删除用户的某份文档
func DeleteUserDocument(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	documentId := (c.Query("doc_id"))
	if documentId == "" {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	documentService := services.NewDocumentService()
	document := models.Document{}
	if documentService.GetById(documentId, &document) != nil {
		response.BadRequest(c, "文档不存在")
		return
	}
	if document.ProjectId == "" {
		if _, err := documentService.Delete(
			"user_id = ? and id = ?", userId, documentId,
		); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
			response.ServerError(c, "删除错误")
			return
		}
		_, err = documentService.UpdateColumns(map[string]any{"delete_by": userId}, "deleted_at is not null and id = ?", documentId, &services.Unscoped{})
		if err != nil {
			log.Println("更新文档删除者失败", err.Error())
		}
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
			response.ServerError(c, "删除错误")
			return
		}
		_, err = documentService.UpdateColumns(map[string]any{"delete_by": userId}, "deleted_at is not null and id = ?", documentId, &services.Unscoped{})
		if err != nil {
			log.Println("更新文档删除者失败", err.Error())
		}
	}
	response.Success(c, "")
}

// GetUserDocumentInfo 获取用户某份文档的信息
func GetUserDocumentInfo(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		log.Println("获取userId失败", err.Error())
		response.Unauthorized(c)
		return
	}
	documentId := (c.Query("doc_id"))
	if documentId == "" {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	permType := models.PermType(str.DefaultToInt(c.Query("perm_type"), 0))
	if permType < models.PermTypeReadOnly || permType > models.PermTypeEditable {
		permType = models.PermTypeNone
	}
	result, errmsg := GetUserDocumentInfo1(userId, documentId, permType)
	if errmsg != "" {
		response.BadRequest(c, errmsg)
		return
	}
	// 获取文档对应的user信息
	docUserId := result.Document.UserId
	authClient := services.GetKCAuthClient()
	token, err := utils.GetAccessToken(c)
	if err != nil {
		log.Println("获取token失败")
		response.Unauthorized(c)
		return
	}
	userInfo, err := authClient.GetUserInfoById(token, docUserId)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"user": models.UserProfile{
			Id:       userInfo.UserID,
			Nickname: userInfo.Profile.Nickname,
			Avatar:   userInfo.Profile.Avatar,
		},
		"document":                     result.Document,
		"team":                         result.Team,
		"project":                      result.Project,
		"document_favorites":           result.DocumentFavorites,
		"document_access_record":       result.DocumentAccessRecord,
		"document_permission":          result.DocumentPermission,
		"document_permission_requests": result.DocumentPermissionRequests,
		"shares_count":                 result.SharesCount,
		"application_count":            result.ApplicationCount,
		"locked_info":                  result.LockedInfo,
	})
}

func GetUserDocumentInfo1(userId string, documentId string, permType models.PermType) (*services.DocumentInfoQueryRes, string) {

	// permType := models.PermType(str.DefaultToInt(c.Query("perm_type"), 0))
	// if permType < models.PermTypeReadOnly || permType > models.PermTypeEditable {
	// 	permType = models.PermTypeNone
	// }

	result := services.NewDocumentService().GetDocumentInfoByDocumentAndUserId(documentId, userId, permType)
	if result == nil {
		return nil, "文档不存在"
	} else if result.LockedInfo != nil && !result.LockedInfo.LockedAt.IsZero() && result.Document.UserId != userId {
		return nil, "审核不通过"
	} else {
		return result, ""
	}
}

type SetDocumentNameReq struct {
	DocId string `json:"doc_id" binding:"required"`
	Name  string `json:"name" binding:"required"`
}

// SetDocumentName 设置文档名称
func SetDocumentName(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req SetDocumentNameReq
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
	document := models.Document{}
	if documentService.Get(&document, "id = ?", documentId) != nil {
		response.BadRequest(c, "文档不存在")
		return
	}
	if document.ProjectId == "" {
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

	reviewClient := services.GetSafereviewClient()
	if reviewClient != nil {
		reviewResponse, err := (reviewClient).ReviewText(req.Name)
		if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
			log.Println("名称审核不通过", req.Name, err, reviewResponse)
			response.ReviewFail(c, "审核不通过")
			return
		}
	}

	if _, err = documentService.UpdateColumns(
		map[string]any{"name": req.Name},
		"id = ?", documentId,
	); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.ServerError(c, "更新错误")
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

func copyDocument(userId string, documentId string, c *gin.Context, documentName string, dbModule *models.DBModule, _storage *storage.StorageClient, mongo *mongo.MongoDB) (result *services.AccessRecordAndFavoritesQueryResItem) {
	documentService := services.NewDocumentService()
	sourceDocument := models.Document{}
	if err := documentService.Get(&sourceDocument, "id = ?", documentId); err != nil {
		response.Forbidden(c, "")
		return
	}
	lockedInfo, err := documentService.GetLocked(documentId)
	if err != nil {
		response.Forbidden(c, "")
		return
	}
	if lockedInfo != nil && !lockedInfo.LockedAt.IsZero() {
		response.Forbidden(c, "审核不通过")
		return
	}
	if sourceDocument.ProjectId == "" {
		var permType models.PermType
		if err := documentService.GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeEditable {
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
		response.ServerError(c, "获取文档版本失败")
		return
	}

	// 获取文档cmd
	type DocumentCmd = models.CmdItem

	cmdServices := services.GetCmdService()
	documentCmdList, err := cmdServices.GetCmdItemsFromStart(documentId, documentVersion.LastCmdVerId)
	if err != nil {
		response.ServerError(c, "获取文档cmd失败")
		return
	}

	minBaseVer := utils.Reduce(documentCmdList, func(acc uint, item DocumentCmd) uint {
		if item.Cmd.BaseVer < acc {
			return item.Cmd.BaseVer
		}
		return acc
	}, math.MaxUint)

	// 获取minBaseVer到documentVersion.lastCmdVerId之间的cmd
	baseVerCmdList, err := cmdServices.GetCmdItems(documentId, minBaseVer, documentVersion.LastCmdVerId)
	if err != nil {
		response.ServerError(c, "获取文档cmd失败")
		return
	}

	lastCmdVerId := uint(0)
	if len(baseVerCmdList) > 0 {
		lastCmdVerId = baseVerCmdList[len(baseVerCmdList)-1].VerId
	}

	// 复制目录
	targetDocumentPath := uuid.New().String()
	if _, err := _storage.Bucket.CopyDirectory(sourceDocument.Path, targetDocumentPath); err != nil {
		log.Println("复制目录失败：", err)
		response.ServerError(c, "复制失败")
		return
	}

	documentMetaBytes, err := _storage.Bucket.GetObject(targetDocumentPath + "/document-meta.json")
	if err != nil {
		log.Println("获取document-meta.json失败：", targetDocumentPath+"/document-meta.json", err)
		response.ServerError(c, "复制失败")
		return
	}
	documentMeta := map[string]any{}
	if err := json.Unmarshal(documentMetaBytes, &documentMeta); err != nil {
		log.Println("documentMetaBytes转json失败：", err)
		response.ServerError(c, "复制失败")
		return
	}

	for _, page := range documentMeta["pagesList"].([]any) {
		pageItem, ok := page.(map[string]any)
		if !ok {
			log.Println("pageItem转map失败：", err)
			response.ServerError(c, "复制失败")
			return
		}
		pageObjectInfo, err := _storage.Bucket.GetObjectInfo(targetDocumentPath + "/pages/" + pageItem["id"].(string) + ".json")
		if err != nil {
			log.Println("获取pageObjectInfo失败：", err)
			response.ServerError(c, "复制失败")
			return
		}
		pageItem["versionId"] = pageObjectInfo.VersionID
	}

	documentMeta["lastCmdVerId"] = (lastCmdVerId)

	documentMetaBytes, err = json.Marshal(documentMeta)
	if err != nil {
		log.Println("documentMeta转json失败：", err)
		response.ServerError(c, "复制失败")
		return
	}
	documentMetaUploadInfo, err := _storage.Bucket.PutObjectByte(targetDocumentPath+"/document-meta.json", documentMetaBytes)
	if err != nil {
		log.Println("documentMeta上传失败：", err)
		response.ServerError(c, "复制失败")
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
		response.ServerError(c, "创建失败")
		return
	}

	documentCmdList1 := append(baseVerCmdList, documentCmdList...)
	// 生成新id
	// idMap := map[string]string{}
	offset := uint(0)
	if len(documentCmdList1) > 0 {
		offset = documentCmdList1[0].VerId
	}
	for i := 0; i < len(documentCmdList1); i++ { // todo 这不对。会导致id重复
		// newId := uuid.NewString()
		cmd := documentCmdList1[i]
		// idMap[cmd.Cmd.Id] = newId //不需要改cmd.id
		// cmd.Cmd.Id = newId
		cmd.DocumentId = targetDocument.Id
		cmd.VerId -= offset
		cmd.BatchStartId -= offset
		cmd.Cmd.BaseVer -= offset
	}

	cmdServices.SaveCmdItems(documentCmdList1)

	lastCmdVerId -= offset

	// 创建文档版本
	if err := documentService.DocumentVersionService.Create(&models.DocumentVersion{
		DocumentId:   targetDocument.Id,
		VersionId:    documentMetaUploadInfo.VersionID,
		LastCmdVerId: lastCmdVerId,
	}); err != nil {
		log.Println("创建文档版本记录失败：", err)
		return
	}

	// 复制评论数据
	commentService := services.GetUserCommentService()
	documentCommentList, err := commentService.GetUserComment(documentId)
	if err != nil {
		log.Println("获取评论数据失败：", err)
		return
	}

	//
	// commentIdMap := map[string]string{}
	for i := 0; i < len(documentCommentList); i++ {
		item := &documentCommentList[i]
		item.DocumentId = (targetDocument.Id)
	}
	_, err = commentService.SaveCommentItems(documentCommentList)
	if err != nil {
		log.Println("评论复制失败：", err)
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
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req CopyDocumentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	documentId := (req.DocId)
	if documentId == "" {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}
	result := copyDocument(userId, documentId, c, "%s_副本", services.GetDBModule(), services.GetStorageClient(), services.GetMongoDB())
	if result != nil {
		response.Success(c, result)
	}
}

// CreateTest 创建测试文档环境
// func CreateTest(c *gin.Context) {
// 	var req struct {
// 		VerifyCode   string `json:"verify_code" binding:"required"`
// 		UserId       string `json:"user_id" binding:"required"`
// 		DocumentId   string `json:"document_id" binding:"required"`
// 		DocumentId2  string `json:"document_id2"`
// 		DocumentName string `json:"document_name" binding:"required"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		response.BadRequest(c, "")
// 		return
// 	}

// 	if req.VerifyCode != "123456" {
// 		response.Forbidden(c, "")
// 		return
// 	}

// 	result := map[string]any{}

// 	if req.DocumentId2 == "" {
// 		userId := req.UserId
// 		if userId == "" {
// 			response.BadRequest(c, "参数错误：user_id")
// 			return
// 		}

// 		documentId := str.DefaultToInt(req.DocumentId, 0)
// 		if documentId <= 0 {
// 			response.BadRequest(c, "参数错误：document_id")
// 			return
// 		}

// 		result1 := copyDocument(userId, documentId, c, req.DocumentName)
// 		if result1 == nil {
// 			response.Fail(c, "创建文档失败")
// 			return
// 		}
// 		models.StructToMap(result1, result)
// 	} else {
// 		result["document"] = map[string]any{
// 			"id": req.DocumentId2,
// 		}
// 	}

// 	// 创建JWT
// 	token, err := jwt.CreateJwt(&jwt.Data{
// 		Id:       req.UserId,
// 		Nickname: "测试用户",
// 	})
// 	if err != nil {
// 		response.Fail(c, err.Error())
// 		return
// 	}
// 	result["token"] = token

// 	response.Success(c, result)
// }
