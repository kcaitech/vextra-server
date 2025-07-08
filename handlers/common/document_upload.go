package common

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"kcaitech.com/kcserver/utils/my_map"

	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/str"
)

type Header struct {
	UserId       string `json:"user_id"`
	DocumentId   string `json:"document_id"`
	ProjectId    string `json:"project_id"`
	LastCmdVerId string `json:"last_cmd_id"`
}

type ResponseStatusType string

// const (
// 	ResponseStatusSuccess ResponseStatusType = "success"
// 	ResponseStatusFail    ResponseStatusType = "fail"
// )

// type Response struct {
// 	Status  ResponseStatusType `json:"status,omitempty"`
// 	Message string             `json:"message,omitempty"`
// 	Data    Data               `json:"data,omitempty"`
// }

// type PageSvg struct {
// 	Name    string
// 	Content string
// }

type Media struct {
	Name    string
	Content *[]byte
}

func compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	_, err := gw.Write(data)
	if err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func compressPutObjectByte(path string, content []byte, _storage *storage.StorageClient) (*storage.UploadInfo, error) {
	if len(content) > 128 {
		presed, _ := compress(content)
		if presed != nil {
			content = presed
		}
	}
	return _storage.Bucket.PutObjectByte(path, content, "")
}

func uploadDocumentData(document *models.Document, lastCmdVerId string, uploadData *VersionResp, medias *[]Media, resp *Response) {

	documentSize := uint64(0)

	uploadWaitGroup := sync.WaitGroup{}

	// pages部分
	var pages []struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(uploadData.DocumentData.Pages, &pages); err != nil {
		resp.Message = "Pages内有元素缺少Id " + err.Error()
		log.Println("Pages内有元素缺少Id", err)
		return
	}
	var pagesRaw = []json.RawMessage{}
	if err := json.Unmarshal(uploadData.DocumentData.Pages, &pagesRaw); err != nil {
		resp.Message = "Pages格式错误 " + err.Error()
		log.Println("Pages格式错误", err)
		return
	}
	idToVersionId := my_map.NewSyncMap[string, string]()
	// 上传pages目录
	for i := 0; i < len(pages); i++ {
		pageId := pages[i].Id
		pagePath := document.Path + "/pages/" + pageId + ".json"
		pageContent := pagesRaw[i]
		documentSize += uint64(len(pageContent))
		uploadWaitGroup.Add(1)
		go func(pagePath string, pageContent json.RawMessage) {
			defer uploadWaitGroup.Done()
			if result, err := compressPutObjectByte(pagePath, pageContent, services.GetStorageClient()); err != nil {
				resp.Message = "对象上传错误"
				log.Println("对象上传错误", err)
				return
			} else {
				idToVersionId.Set(pageId, result.VersionID)
			}
		}(pagePath, pageContent)
	}

	// medias部分
	if medias != nil { // 非首次上传即版本更新，不会有medias
		// upload medias
		for _, media := range *medias {
			if media.Content == nil {
				continue
			}
			path := document.Path + "/medias/" + media.Name
			uploadWaitGroup.Add(1)
			go func(path string, media []byte) {
				defer uploadWaitGroup.Done()
				_storage := services.GetStorageClient()
				if _, err := _storage.Bucket.PutObjectByte(path, media, ""); err != nil {
					resp.Message = "对象上传错误"
					log.Println("对象上传错误", err)
					return
				}
			}(path, *media.Content)
		}
	}
	documentSize += uploadData.MediasSize

	var _medias []Media // 审核时需要
	if medias != nil {
		_medias = *medias
	}
	for _, mediaName := range uploadData.DocumentData.MediaNames {
		mediaPath := document.Path + "/medias/" + mediaName
		mediaData, err := services.GetStorageClient().Bucket.GetObject(mediaPath)
		if err != nil {
			log.Printf("获取媒体文件 %s 失败: %v", mediaName, err)
			continue
		}
		_medias = append(_medias, Media{
			Name:    mediaName,
			Content: &mediaData,
		})
	}

	review(document, uploadData, document.Path, pages, &_medias)

	uploadWaitGroup.Wait()

	// 设置versionId
	pagesList := uploadData.DocumentData.DocumentMeta["pagesList"].([]any)
	for _, page := range pagesList {
		pageItem, ok := page.(map[string]any)
		pageId, ok1 := pageItem["id"].(string)
		versionId, ok2 := idToVersionId.Get(pageId)
		if !ok || !ok1 || !ok2 {
			resp.Message = "对象上传错误" + fmt.Sprintf("ok: %v, ok1: %v, ok2: %v", ok, ok1, ok2)
			log.Println("对象上传错误1", ok, ok1, ok2, len(pages))
			return
		}
		pageItem["versionId"] = versionId
	}
	uploadData.DocumentData.DocumentMeta["lastCmdVer"] = str.DefaultToInt(lastCmdVerId, 0)

	documentMetaStr, err := json.Marshal(uploadData.DocumentData.DocumentMeta)
	if err != nil {
		resp.Message = "document-meta.json格式错误"
		log.Println("document-meta.json格式错误", err)
		return
	}
	path := document.Path + "/document-meta.json"
	_storage := services.GetStorageClient()
	putObjectResult, err := compressPutObjectByte(path, documentMetaStr, _storage)
	if err != nil {
		resp.Message = "对象上传错误"
		log.Println("document-meta.json上传错误", err)
		return
	}
	// documentVersionId := putObjectResult.VersionID
	documentSize += uint64(len(uploadData.DocumentData.DocumentMeta))

	document.Size = documentSize
	document.VersionId = putObjectResult.VersionID
}

// 更新文档数据
func UpdateDocumentData(documentId string, lastCmdVerId string, uploadData *VersionResp, medias *[]Media, resp *Response) {

	// 获取文档信息
	documentService := services.NewDocumentService()

	var document = models.Document{}
	if documentService.GetById(documentId, &document) != nil {
		resp.Message = "文档不存在"
		return
	}

	// uploadData.DocumentData.DocumentMeta["lastCmdVer"] = str.DefaultToInt(lastCmdVerId, 0)

	uploadDocumentData(&document, lastCmdVerId, uploadData, medias, resp)
	if resp.Message != "" {
		return
	}

	// 获取文档名称
	// documentName, _ := uploadData.DocumentData.DocumentMeta["name"].(string)
	documentAccessRecordService := services.NewDocumentAccessRecordService()
	// 创建文档记录和历史记录
	now := (time.Now())

	// document.Size = documentSize
	// document.VersionId = documentVersionId
	if _, err := documentService.UpdatesById(documentId, &document); err != nil {
		resp.Message = "对象上传错误"
		log.Println("对象上传错误3", err)
		return
	}
	var documentAccessRecord = models.DocumentAccessRecord{}
	err := documentAccessRecordService.Get(&documentAccessRecord, "user_id = ? and document_id = ?", document.UserId, documentId)
	if err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		resp.Message = "对象上传错误"
		log.Println("对象上传错误4", err)
		return
	}
	if errors.Is(err, services.ErrRecordNotFound) {
		_ = documentAccessRecordService.Create(&models.DocumentAccessRecord{
			UserId:         document.UserId,
			DocumentId:     documentId,
			LastAccessTime: now,
		})
	} else {
		documentAccessRecord.LastAccessTime = now
		_, _ = documentAccessRecordService.UpdatesById(documentAccessRecord.Id, &documentAccessRecord)
	}

	// 创建文档版本记录
	if err := documentService.DocumentVersionService.Create(&models.DocumentVersion{
		DocumentId:   documentId,
		VersionId:    document.VersionId,
		LastCmdVerId: uint(str.DefaultToInt(lastCmdVerId, 0)),
	}); err != nil {
		resp.Message = "对象上传错误"
		log.Println("对象上传错误5", err)
		return
	}

	resp.Code = http.StatusOK
	resp.Data = Data{
		"document_id": (documentId),
		"version_id":  document.VersionId,
	}
}

// 上传新文档数据
func UploadNewDocumentData(userId string, projectId string, uploadData *VersionResp, medias *[]Media, resp *Response) {

	if userId == "" {
		resp.Message = "userId不能为空"
		return
	}
	var teamId string
	if projectId != "" {
		projectService := services.NewProjectService()
		project := models.Project{}
		if err := projectService.GetById(projectId, &project); err != nil {
			resp.Message = "项目不存在"
			return
		}
		teamId = project.TeamId
		permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
		if err != nil || permType == nil {
			log.Println("获取项目权限失败", err)
		}
		if err != nil || permType == nil || *permType < models.ProjectPermTypeEditable {
			resp.Message = "无权限"
			return
		}
	}

	// 获取文档信息
	documentService := services.NewDocumentService()

	// // 还是换成base62,与用户id、团队id等保持一致
	// document_id, err := utils.GenerateBase62ID()
	// if err != nil {
	// 	resp.Message = "生成文档id失败"
	// 	log.Println("生成文档id失败", err)
	// 	return
	// }
	// 还是换成uuid,与用户pageid,shapeid等保持一致
	document_id := uuid.New().String()

	newDocument := models.Document{
		Id:        document_id,
		UserId:    userId,
		Path:      document_id,
		DocType:   models.DocTypeShareable,
		TeamId:    teamId,
		ProjectId: projectId,
	}

	// uploadData.DocumentData.DocumentMeta["lastCmdVer"] = 0

	uploadDocumentData(&newDocument, "0", uploadData, medias, resp)
	if resp.Message != "" {
		return
	}

	// 获取文档名称
	documentName, _ := uploadData.DocumentData.DocumentMeta["name"].(string)
	documentAccessRecordService := services.NewDocumentAccessRecordService()
	// 创建文档记录和历史记录
	now := (time.Now())

	newDocument.Name = documentName

	if err := documentService.Create(&newDocument); err != nil {
		resp.Message = "对象上传错误."
		log.Println("对象上传错误2", err)
		return
	}

	_ = documentAccessRecordService.Create(&models.DocumentAccessRecord{
		UserId:         userId,
		DocumentId:     newDocument.Id,
		LastAccessTime: now,
	})

	// 创建文档版本记录
	if err := documentService.DocumentVersionService.Create(&models.DocumentVersion{
		DocumentId:   newDocument.Id,
		VersionId:    newDocument.VersionId,
		LastCmdVerId: uint(0),
	}); err != nil {
		resp.Message = "对象上传错误"
		log.Println("对象上传错误5", err)
		return
	}

	resp.Code = http.StatusOK
	resp.Data = Data{
		"document_id": (newDocument.Id),
		"version_id":  newDocument.VersionId,
	}
}
