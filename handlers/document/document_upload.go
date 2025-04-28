package document

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"kcaitech.com/kcserver/utils"
	"kcaitech.com/kcserver/utils/my_map"

	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/str"
)

// type Data map[string]any

type Header struct {
	UserId       string `json:"user_id"`
	DocumentId   string `json:"document_id"`
	ProjectId    string `json:"project_id"`
	LastCmdVerId string `json:"last_cmd_id"`
}

type ResponseStatusType string

const (
	ResponseStatusSuccess ResponseStatusType = "success"
	ResponseStatusFail    ResponseStatusType = "fail"
)

type Response struct {
	Status  ResponseStatusType `json:"status,omitempty"`
	Message string             `json:"message,omitempty"`
	Data    Data               `json:"data,omitempty"`
}

type PageSvg struct {
	Name    string
	Content string
}

type UploadData struct {
	DocumentMeta Data            `json:"document_meta"`
	Pages        json.RawMessage `json:"pages"`
	// FreeSymbols         json.RawMessage `json:"freesymbols"` // 这个现在在document meta里
	MediaNames   []string `json:"media_names"`
	MediasSize   uint64   `json:"medias_size"`
	DocumentText string   `json:"document_text"`
	// PageImageList *[][]byte `json:"page_image_list"`
	PageSvgs []string `json:"pageSvgs"`
}

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
	return _storage.Bucket.PutObjectByte(path, content)
}

func UploadDocumentData(header *Header, uploadData *UploadData, medias *[]Media, resp *Response) {

	userId := header.UserId
	documentId := (header.DocumentId)
	projectId := (header.ProjectId)
	lastCmdVerId := header.LastCmdVerId
	if (userId == "" && documentId == "") || (userId != "" && documentId != "") || (documentId != "" && lastCmdVerId == "") { // userId和documentId必须只传一个 // todo这不对吧，要鉴权
		resp.Message = "参数错误"
		log.Println("参数错误", userId, documentId)
		return
	}
	isFirstUpload := documentId == ""
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

	docPath := ""
	docId := ""
	var document = models.Document{}
	if !isFirstUpload {
		if documentService.GetById(documentId, &document) != nil {
			resp.Message = "文档不存在"
			return
		}
		docPath = document.Path
		docId = document.Id
	} else {
		id, err := utils.GenerateBase62ID()
		if err != nil {
			resp.Message = err.Error()
			return
		}
		docPath = id
		docId = id
	}

	newDocument := models.Document{
		UserId:    userId,
		Path:      docPath,
		DocType:   models.DocTypeShareable,
		TeamId:    teamId,
		ProjectId: projectId,
	}

	documentSize := uint64(0)

	// reviewText(uploadData.DocumentText, &newDocument)
	// reviewPages(uploadData.PageImageList, &newDocument, docPath, documentService)

	// if !newDocument.LockedAt.IsZero() && !isFirstUpload {
	// 	_, _ = documentService.UpdateColumnsById(documentId, map[string]any{
	// 		"locked_at": nil,
	// 	})
	// }

	uploadWaitGroup := sync.WaitGroup{}

	// pages部分
	var pages []struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(uploadData.Pages, &pages); err != nil {
		resp.Message = "Pages内有元素缺少Id " + err.Error()
		log.Println("Pages内有元素缺少Id", err)
		return
	}
	var pagesRaw = []json.RawMessage{}
	if err := json.Unmarshal(uploadData.Pages, &pagesRaw); err != nil {
		resp.Message = "Pages格式错误 " + err.Error()
		log.Println("Pages格式错误", err)
		return
	}
	idToVersionId := my_map.NewSyncMap[string, string]()
	// 上传pages目录
	for i := 0; i < len(pages); i++ {
		pageId := pages[i].Id
		pagePath := docPath + "/pages/" + pageId + ".json"
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
		// reviewMedias(medias, &newDocument, documentService, services.GetSafereviewClient())

		// upload medias
		for _, media := range *medias {
			path := docPath + "/medias/" + media.Name
			if media.Content == nil {
				continue
			}
			uploadWaitGroup.Add(1)
			go func(path string, media []byte) {
				defer uploadWaitGroup.Done()
				_storage := services.GetStorageClient()
				if _, err := _storage.Bucket.PutObjectByte(path, media); err != nil {
					resp.Message = "对象上传错误"
					log.Println("对象上传错误", err)
					return
				}
			}(path, *media.Content)
		}
	}
	documentSize += uploadData.MediasSize

	review(&newDocument, uploadData, docPath, pages, medias)

	uploadWaitGroup.Wait()

	// 设置versionId
	pagesList := uploadData.DocumentMeta["pagesList"].([]any)
	for _, page := range pagesList {
		pageItem, ok := page.(map[string]any)
		pageId, ok1 := pageItem["id"].(string)
		versionId, ok2 := idToVersionId.Get(pageId)
		if !ok || !ok1 || !ok2 {
			resp.Message = "对象上传错误"
			log.Println("对象上传错误1")
			return
		}
		pageItem["versionId"] = versionId
	}
	uploadData.DocumentMeta["lastCmdVerId"] = lastCmdVerId

	documentMetaStr, err := json.Marshal(uploadData.DocumentMeta)
	if err != nil {
		resp.Message = "document-meta.json格式错误"
		log.Println("document-meta.json格式错误", err)
		return
	}
	path := docPath + "/document-meta.json"
	_storage := services.GetStorageClient()
	putObjectResult, err := compressPutObjectByte(path, documentMetaStr, _storage)
	if err != nil {
		resp.Message = "对象上传错误"
		log.Println("document-meta.json上传错误", err)
		return
	}
	documentVersionId := putObjectResult.VersionID
	documentSize += uint64(len(uploadData.DocumentMeta))

	// 获取文档名称
	documentName, _ := uploadData.DocumentMeta["name"].(string)
	documentAccessRecordService := services.NewDocumentAccessRecordService()
	// 创建文档记录和历史记录
	now := (time.Now())
	if isFirstUpload {
		newDocument.Name = documentName
		newDocument.Size = documentSize
		newDocument.VersionId = documentVersionId
		newDocument.Id = docId
		if err := documentService.Create(&newDocument); err != nil {
			resp.Message = "对象上传错误."
			log.Println("对象上传错误2", err)
			return
		}
		documentId = newDocument.Id
		_ = documentAccessRecordService.Create(&models.DocumentAccessRecord{
			UserId:         userId,
			DocumentId:     documentId,
			LastAccessTime: now,
		})
	} else {
		document.Size = documentSize
		document.VersionId = documentVersionId
		if _, err := documentService.UpdatesById(documentId, &document); err != nil {
			resp.Message = "对象上传错误"
			log.Println("对象上传错误3", err)
			return
		}
		var documentAccessRecord = models.DocumentAccessRecord{}
		err := documentAccessRecordService.Get(&documentAccessRecord, "user_id = ? and document_id = ?", userId, documentId)
		if err != nil && !errors.Is(err, services.ErrRecordNotFound) {
			resp.Message = "对象上传错误"
			log.Println("对象上传错误4", err)
			return
		}
		if errors.Is(err, services.ErrRecordNotFound) {
			_ = documentAccessRecordService.Create(&models.DocumentAccessRecord{
				UserId:         userId,
				DocumentId:     documentId,
				LastAccessTime: now,
			})
		} else {
			documentAccessRecord.LastAccessTime = now
			_, _ = documentAccessRecordService.UpdatesById(documentAccessRecord.Id, &documentAccessRecord)
		}
	}

	// 创建文档版本记录
	if err := documentService.DocumentVersionService.Create(&models.DocumentVersion{
		DocumentId:   documentId,
		VersionId:    documentVersionId,
		LastCmdVerId: uint(str.DefaultToInt(lastCmdVerId, 0)),
	}); err != nil {
		resp.Message = "对象上传错误"
		log.Println("对象上传错误5", err)
		return
	}

	resp.Status = ResponseStatusSuccess
	resp.Data = Data{
		"document_id": (documentId),
		"version_id":  documentVersionId,
	}
}
