package document

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/models"
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
		common.Unauthorized(c)
		return
	}

	projectId := c.Query("project_id")
	cursor := c.Query("cursor")
	limit := utils.QueryInt(c, "limit", 20) // 默认每页20条

	var result *[]services.AccessRecordAndFavoritesQueryResItem
	var hasMore bool

	documentService := services.NewDocumentService()

	// 根据是否有项目ID，使用不同的分页查询函数
	if projectId != "" {
		result, hasMore = documentService.FindDocumentByProjectIdWithCursor(projectId, userId, cursor, limit)
	} else {
		result, hasMore = documentService.FindDocumentByUserIdWithCursor(userId, cursor, limit)
	}

	// 获取文档相关用户信息
	userIds := make([]string, 0)
	for _, item := range *result {
		userIds = append(userIds, item.Document.UserId)
	}

	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	for i := range *result {
		item := &(*result)[i]
		userInfo, ok := userMap[item.Document.UserId]
		if ok {
			item.User = &models.UserProfile{
				Id:       userInfo.UserID,
				Nickname: userInfo.Nickname,
				Avatar:   userInfo.Avatar,
			}
		}
	}

	// 构建包含下一页游标信息的响应
	var nextCursor string
	if hasMore && len(*result) > 0 {
		// 使用最后一条记录的访问时间作为下一页的游标
		lastItem := (*result)[len(*result)-1]
		nextCursor = lastItem.DocumentAccessRecord.LastAccessTime.Format(time.RFC3339)
	}

	common.SuccessWithCursor(c, result, hasMore, nextCursor)
}

// DeleteUserDocument 删除用户的某份文档
func DeleteUserDocument(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	documentId := (c.Query("doc_id"))
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}
	documentService := services.NewDocumentService()
	document := models.Document{}
	if documentService.GetById(documentId, &document) != nil {
		common.BadRequest(c, "文档不存在")
		return
	}
	if document.ProjectId == "" {
		if _, err := documentService.Delete(
			"user_id = ? and id = ?", userId, documentId,
		); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
			common.ServerError(c, "删除错误")
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
			common.Forbidden(c, "")
			return
		}
		if _, err := documentService.Delete(
			"id = ?", documentId,
		); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
			common.ServerError(c, "删除错误")
			return
		}
		_, err = documentService.UpdateColumns(map[string]any{"delete_by": userId}, "deleted_at is not null and id = ?", documentId, &services.Unscoped{})
		if err != nil {
			log.Println("更新文档删除者失败", err.Error())
		}
	}
	common.Success(c, "")
}

// GetUserDocumentInfo 获取用户某份文档的信息
func GetUserDocumentInfo(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		log.Println("获取userId失败", err.Error())
		common.Unauthorized(c)
		return
	}
	documentId := (c.Query("doc_id"))
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}
	permType := models.PermType(str.DefaultToInt(c.Query("perm_type"), 0))
	if permType < models.PermTypeReadOnly || permType > models.PermTypeEditable {
		permType = models.PermTypeNone
	}
	docService := services.NewDocumentService()
	result := docService.GetDocumentInfoByDocumentAndUserId(documentId, userId, permType)
	if result == nil {
		common.Resp(c, http.StatusNotFound, "文档不存在", nil)
		return
	}
	locked, _ := docService.GetLocked(documentId)

	// 处理锁定信息中的LockedWords
	if len(locked) > 0 {
		_storage := services.GetStorageClient()
		for i := range locked {
			item := &locked[i]
			// 如果是缩略图类型的媒体锁定，设置LockedWords为GetDocumentThumbnail的返回值
			if item.LockedType == models.LockedTypeMedia && item.LockedTarget == "thumbnail" {
				item.LockedWords = GetDocumentThumbnail(nil, result.Document.Path, _storage)
			}
		}
	}

	if len(locked) > 0 && result.Document.UserId != userId {
		common.Resp(c, common.StatusDocumentNotFound, "审核不通过", nil)
		return
	}

	// 获取文档对应的user信息
	docUserId := result.Document.UserId
	authClient := services.GetKCAuthClient()
	token, err := utils.GetAccessToken(c)
	if err != nil {
		log.Println("获取token失败")
		common.Unauthorized(c)
		return
	}
	userInfo, err := authClient.GetUserInfoById(token, docUserId)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}

	common.Success(c, gin.H{
		"user": models.UserProfile{
			Id:       userInfo.UserID,
			Nickname: userInfo.Nickname,
			Avatar:   userInfo.Avatar,
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
		"locked_info":                  locked,
	})
}

type SetDocumentNameReq struct {
	DocId string `json:"doc_id" binding:"required"`
	Name  string `json:"name" binding:"required"`
}

// SetDocumentName 设置文档名称
func SetDocumentName(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req SetDocumentNameReq
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
	document := models.Document{}
	if documentService.Get(&document, "id = ?", documentId) != nil {
		common.BadRequest(c, "文档不存在")
		return
	}
	if document.ProjectId == "" {
		if document.UserId != userId {
			common.Forbidden(c, "")
			return
		}
	} else {
		projectService := services.NewProjectService()
		projectPermType, err := projectService.GetProjectPermTypeByForUser(document.ProjectId, userId)
		// 管理员以上权限或文档创建者且可编辑权限
		if !(err == nil && projectPermType != nil && ((*projectPermType) >= models.ProjectPermTypeAdmin || ((*projectPermType) == models.ProjectPermTypeEditable && document.UserId == userId))) {
			common.Forbidden(c, "")
			return
		}
	}

	reviewClient := services.GetSafereviewClient()
	if reviewClient != nil {
		reviewResponse, err := (reviewClient).ReviewText(req.Name)
		if err != nil {
			log.Println("名称审核失败", req.Name, err)
			common.ReviewFail(c, "审核失败")
			return
		} else if reviewResponse != nil && reviewResponse.Status != safereviewBase.ReviewTextResultPass {
			log.Println("名称审核不通过", req.Name, reviewResponse)
			common.ReviewFail(c, "审核不通过")
			return
		}
	}

	if _, err = documentService.UpdateColumns(
		map[string]any{"name": req.Name},
		"id = ?", documentId,
	); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		common.ServerError(c, "更新错误")
		return
	}
	if errors.Is(err, services.ErrRecordNotFound) {
		common.Forbidden(c, "")
		return
	}
	common.Success(c, "")
}

type CopyDocumentReq struct {
	DocId string `json:"doc_id" binding:"required"`
}

type CopyDocumentRes struct {
	CopyId string `json:"copy_id"`
}

func copyDocument(userId string, documentId string, c *gin.Context, documentName string, _storage *storage.StorageClient, insert bool) (result *CopyDocumentRes) {
	documentService := services.NewDocumentService()
	sourceDocument := models.Document{}
	if err := documentService.Get(&sourceDocument, "id = ?", documentId); err != nil {
		common.Forbidden(c, "")
		return
	}
	lockedInfo, err := documentService.GetLocked(documentId)
	if err != nil {
		common.Forbidden(c, "")
		return
	}
	if len(lockedInfo) > 0 {
		common.Forbidden(c, "审核不通过")
		return
	}
	if sourceDocument.ProjectId == "" {
		var permType models.PermType
		if err := documentService.GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeEditable {
			common.Forbidden(c, "")
			return
		}
	} else {
		projectService := services.NewProjectService()
		projectPermType, err := projectService.GetProjectPermTypeByForUser(sourceDocument.ProjectId, userId)
		if !(err == nil && projectPermType != nil && (*projectPermType) >= models.ProjectPermTypeEditable) { // 需有可编辑权限
			common.Forbidden(c, "")
			return
		}
	}

	documentName = strings.ReplaceAll(documentName, "%s", sourceDocument.Name)

	path := uuid.NewString()
	targetDocumentId := path

	documentVersion := models.DocumentVersion{}
	if err := documentService.DocumentVersionService.Get(&documentVersion, "document_id = ? and version_id = ?", documentId, sourceDocument.VersionId); err != nil {
		common.ServerError(c, "获取文档版本失败")
		return
	}
	// 获取文档cmd
	cmdServices := services.GetCmdService()
	documentCmdList, err := cmdServices.GetCmdItemsFromStart(documentId, documentVersion.LastCmdVerId)
	if err != nil {
		log.Println("获取文档cmd失败:", err, "documentId:", documentId, "LastCmdVerId:", documentVersion.LastCmdVerId)
		common.ServerError(c, "获取文档cmd失败")
		return
	}
	var minBaseVer uint
	if len(documentCmdList) == 0 {
		minBaseVer = 0
	} else {
		minBaseVer = documentCmdList[0].Cmd.BaseVer
	}
	for _, item := range documentCmdList {
		if item.Cmd.BaseVer < minBaseVer {
			minBaseVer = item.Cmd.BaseVer
		}
	}

	// 获取minBaseVer到documentVersion.lastCmdVerId之间的cmd
	baseVerCmdList, err := cmdServices.GetCmdItems(documentId, minBaseVer, documentVersion.LastCmdVerId)
	if err != nil {
		log.Println("获取基础版本cmd失败:", err, "documentId:", documentId, "minBaseVer:", minBaseVer, "LastCmdVerId:", documentVersion.LastCmdVerId)
		common.ServerError(c, "获取基础版本cmd失败")
		return
	}

	lastCmdVerId := uint(0)
	if len(baseVerCmdList) > 0 {
		lastCmdVerId = baseVerCmdList[len(baseVerCmdList)-1].VerId
	}

	documentCmdList1 := append(baseVerCmdList, documentCmdList...)

	offset := uint(0)
	for _, item := range documentCmdList1 {
		offset = min(offset, item.VerId, item.BatchStartId, item.Cmd.BaseVer)
	}

	for i := range documentCmdList1 {
		cmd := &documentCmdList1[i]
		cmd.DocumentId = targetDocumentId
		cmd.VerId -= offset
		cmd.BatchStartId -= offset
		cmd.Cmd.BaseVer -= offset
	}
	cmdServices.SaveCmdItems(documentCmdList1)

	lastCmdVerId -= offset
	// 复制目录
	if _, err := _storage.Bucket.CopyDirectory(sourceDocument.Path, targetDocumentId); err != nil {
		log.Println("复制目录失败：", err)
		common.ServerError(c, "复制失败")
		return
	}

	documentMetaBytes, err := _storage.Bucket.GetObject(targetDocumentId + "/document-meta.json")
	if err != nil {
		log.Println("获取document-meta.json失败：", targetDocumentId+"/document-meta.json", err)
		common.ServerError(c, "复制失败")
		return
	}

	// 检查数据是否被压缩
	if len(documentMetaBytes) > 2 &&
		documentMetaBytes[0] == 0x1f &&
		documentMetaBytes[1] == 0x8b {
		// 尝试解压缩 gzip 数据
		reader, err := gzip.NewReader(bytes.NewReader(documentMetaBytes))
		if err != nil {
			common.ServerError(c, "复制失败：文档元数据格式错误")
			return
		}
		// 读取解压缩后的数据
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			common.ServerError(c, "复制失败：文档元数据解压缩错误")
			return
		}
		reader.Close()
		documentMetaBytes = decompressed
	}

	documentMeta := map[string]any{}
	if err := json.Unmarshal(documentMetaBytes, &documentMeta); err != nil {
		common.ServerError(c, "复制失败：文档元数据解析错误")
		return
	}

	for _, page := range documentMeta["pagesList"].([]any) {
		pageItem, ok := page.(map[string]any)
		if !ok {
			common.ServerError(c, "pageItem转map失败：复制失败")
			return
		}
		pageObjectInfo, err := _storage.Bucket.GetObjectInfo(targetDocumentId + "/pages/" + pageItem["id"].(string) + ".json")
		if err != nil {
			log.Println("获取pageObjectInfo失败：", err)
			common.ServerError(c, "获取pageObjectInfo失败：复制失败")
			return
		}
		pageItem["versionId"] = pageObjectInfo.VersionID
	}

	documentMeta["lastCmdVerId"] = (lastCmdVerId)
	documentMetaBytes, err = json.Marshal(documentMeta)
	if err != nil {
		log.Println("documentMeta转json失败：", err)
		common.ServerError(c, "复制失败")
		return
	}

	documentMetaUploadInfo, err := _storage.Bucket.PutObjectByte(targetDocumentId+"/document-meta.json", documentMetaBytes, "")
	if err != nil {
		log.Println("documentMeta上传失败：", err)
		common.ServerError(c, "复制失败")
		return
	}

	// 创建文档版本
	if err := documentService.DocumentVersionService.Create(&models.DocumentVersion{
		DocumentId:   targetDocumentId,
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

	// commentIdMap := map[string]string{}
	for i := range documentCommentList {
		item := &documentCommentList[i]
		item.DocumentId = (targetDocumentId)
	}
	_, err = commentService.SaveCommentItems(documentCommentList)
	if err != nil {
		log.Println("评论复制失败：", err)
	}
	// 复制文档
	targetDocument := models.Document{
		Id:        targetDocumentId,
		UserId:    userId,
		Path:      targetDocumentId,
		DocType:   sourceDocument.DocType,
		Name:      documentName,
		Size:      sourceDocument.Size,
		TeamId:    sourceDocument.TeamId,
		ProjectId: sourceDocument.ProjectId,
		VersionId: documentMetaUploadInfo.VersionID,
	}

	if insert {
		// 确保设置文档ID
		if err := documentService.Create(&targetDocument); err != nil {
			log.Println("创建文档记录失败:", err, "userId:", userId, "documentName:", documentName)
			common.ServerError(c, "创建失败: "+err.Error())
			return
		}
		// 添加最近访问
		documentAccessRecord := models.DocumentAccessRecord{
			UserId:     userId,
			DocumentId: targetDocument.Id,
		}
		documentService.DocumentAccessRecordService.Create(&documentAccessRecord)
	}

	result = &CopyDocumentRes{
		CopyId: targetDocumentId,
	}

	return
}

// CopyDocument 复制文档
func CopyDocument(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req CopyDocumentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "")
		return
	}
	documentId := (req.DocId)
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}
	result := copyDocument(userId, documentId, c, "%s_副本", services.GetStorageClient(), true)
	if result != nil {
		common.Success(c, result)
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
// 		common.BadRequest(c, "")
// 		return
// 	}

// 	if req.VerifyCode != "123456" {
// 		common.Forbidden(c, "")
// 		return
// 	}

// 	result := map[string]any{}

// 	if req.DocumentId2 == "" {
// 		userId := req.UserId
// 		if userId == "" {
// 			common.BadRequest(c, "参数错误：user_id")
// 			return
// 		}

// 		documentId := str.DefaultToInt(req.DocumentId, 0)
// 		if documentId <= 0 {
// 			common.BadRequest(c, "参数错误：document_id")
// 			return
// 		}

// 		result1 := copyDocument(userId, documentId, c, req.DocumentName)
// 		if result1 == nil {
// 			common.Fail(c, "创建文档失败")
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
// 		common.Fail(c, err.Error())
// 		return
// 	}
// 	result["token"] = token

// 	common.Success(c, result)
// }
