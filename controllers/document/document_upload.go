package document

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"kcaitech.com/kcserver/common/safereview"
	safereviewBase "kcaitech.com/kcserver/common/safereview/base"
	"kcaitech.com/kcserver/utils/my_map"

	"github.com/google/uuid"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/services"
	"kcaitech.com/kcserver/common/storage"
	storagebase "kcaitech.com/kcserver/utils/storage/base"
	"kcaitech.com/kcserver/utils/str"
	myTime "kcaitech.com/kcserver/utils/time"
)

// type Data map[string]any

type Header struct {
	UserId     string `json:"user_id"`
	DocumentId string `json:"document_id"`
	ProjectId  string `json:"project_id"`
	LastCmdId  string `json:"last_cmd_id"`
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

type UploadData struct {
	DocumentMeta Data            `json:"document_meta"`
	Pages        json.RawMessage `json:"pages"`
	// FreeSymbols         json.RawMessage `json:"freesymbols"` // 这个现在在document meta里
	MediaNames    []string  `json:"media_names"`
	MediasSize    uint64    `json:"medias_size"`
	DocumentText  string    `json:"document_text"`
	PageImageList *[][]byte `json:"page_image_list"`
}

type Media struct {
	Name    string
	Content *[]byte
}

// @deplecated
// func UploadDocument(c *gin.Context) {
// 	ws, err := websocket.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		response.Fail(c, "建立ws连接失败："+err.Error())
// 		return
// 	}
// 	resp := Response{
// 		Status: ResponseStatusFail,
// 	}
// 	defer func() {
// 		ws.WriteJSON(&resp)
// 		ws.Close()
// 	}()

// 	header := Header{}
// 	if err := ws.ReadJSON(&header); err != nil {
// 		resp.Message = "Header结构错误"
// 		log.Println("Header结构错误", err)
// 		return
// 	}

// 	uploadData := UploadData{}
// 	if err := ws.ReadJSON(&uploadData); err != nil {
// 		resp.Message = "UploadData结构错误"
// 		log.Println("UploadData结构错误", err)
// 		return
// 	}

// 	documentId := str.DefaultToInt(header.DocumentId, 0)
// 	isFirstUpload := documentId <= 0

// 	medias := []Media{}
// 	if isFirstUpload {
// 		uploadData.MediasSize = 0
// 		nextMedia := func() []byte {
// 			messageType, data, err := ws.ReadMessage()
// 			if err != nil {
// 				resp.Message = "ws连接异常"
// 				log.Println("ws连接异常", err)
// 				return nil
// 			}
// 			if messageType != websocket.MessageTypeBinary {
// 				resp.Message = "media格式错误"
// 				log.Println("media格式错误")
// 				return nil
// 			}
// 			return data
// 		}

// 		for _, name := range uploadData.MediaNames {
// 			m := nextMedia()
// 			if nil == m {
// 				return
// 			}
// 			medias = append(medias, Media{
// 				Name:    name,
// 				Content: &m,
// 			})
// 			uploadData.MediasSize += uint64(len(m))
// 		}
// 	}

// 	UploadDocumentData(&header, &uploadData, &medias, &resp)
// }

func reviewText(text string, newDocument *models.Document) {
	if text == "" {
		return
	}
	go func() {
		reviewResponse, err := safereview.Client.ReviewText(text)
		if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
			newDocument.LockedAt = myTime.Time(time.Now())
			newDocument.LockedReason = "文本审核不通过：" + reviewResponse.Reason
			if wordsBytes, err := json.Marshal(reviewResponse.Words); err == nil {
				newDocument.LockedWords += string(wordsBytes)
			}
		}
	}()
}

func reviewPages(PageImageList *[][]byte, newDocument *models.Document, docPath string, documentService *services.DocumentService) {
	if PageImageList == nil || len(*PageImageList) == 0 {
		return
	}
	go func() {
		needUpdateDocument := false
		for i, image := range *PageImageList {
			path := docPath + "/page_image/" + str.IntToString(int64(i)) + ".png"
			if _, err := storage.Bucket.PutObjectByte(path, image); err != nil {
				log.Println("图片上传错误", err)
			}
			if len(image) == 0 {
				continue
			}
			base64Str := base64.StdEncoding.EncodeToString(image)
			reviewResponse, err := safereview.Client.ReviewPictureFromBase64(base64Str)
			if err != nil {
				log.Println("图片审核失败", err)
				continue
			} else if reviewResponse.Status != safereviewBase.ReviewImageResultPass {
				newDocument.LockedAt = myTime.Time(time.Now())
				newDocument.LockedReason += "{图片审核不通过[page:" + str.IntToString(int64(i)) + "]：" + reviewResponse.Reason + "}"
				needUpdateDocument = true
			}
		}
		if needUpdateDocument {
			_, _ = documentService.UpdatesById(newDocument.Id, newDocument)
		}
	}()
}

func reviewMedias(medias *[]Media, newDocument *models.Document, documentService *services.DocumentService) {
	if medias == nil || len(*medias) == 0 {
		return
	}
	go func() {
		needUpdateDocument := false
		for _, mediaInfo := range *medias {
			base64Str := base64.StdEncoding.EncodeToString(*mediaInfo.Content)
			if len(*mediaInfo.Content) == 0 || len(base64Str) == 0 {
				continue
			}
			reviewResponse, err := safereview.Client.ReviewPictureFromBase64(base64Str)
			if err != nil {
				log.Println("图片审核失败", err)
				continue
			} else if reviewResponse.Status != safereviewBase.ReviewImageResultPass {
				log.Println("图片审核不通过", err, reviewResponse)
				newDocument.LockedAt = myTime.Time(time.Now())
				newDocument.LockedReason += "{图片审核不通过[media:" + mediaInfo.Name + "]：" + reviewResponse.Reason + "}"
				needUpdateDocument = true
			}
		}
		if needUpdateDocument {
			_, _ = documentService.UpdatesById(newDocument.Id, newDocument)
		}
	}()
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

func compressPutObjectByte(path string, content []byte) (*storagebase.UploadInfo, error) {
	if len(content) > 128 {
		presed, _ := compress(content)
		if presed != nil {
			content = presed
		}
	}
	return storage.Bucket.PutObjectByte(path, content)
}

func UploadDocumentData(header *Header, uploadData *UploadData, medias *[]Media, resp *Response) {

	userId := header.UserId
	documentId := str.DefaultToInt(header.DocumentId, 0)
	projectId := str.DefaultToInt(header.ProjectId, 0)
	lastCmdId := header.LastCmdId
	if (userId == "" && documentId <= 0) || (userId != "" && documentId > 0) || (documentId > 0 && lastCmdId == "") { // userId和documentId必须只传一个 // todo这不对吧，要鉴权
		resp.Message = "参数错误"
		log.Println("参数错误", userId, documentId)
		return
	}
	isFirstUpload := documentId <= 0
	var teamId int64
	if projectId < 0 {
		projectId = 0
	} else if projectId > 0 {
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
	docPath := uuid.New().String()
	var document = models.Document{}
	if !isFirstUpload {
		if documentService.GetById(documentId, &document) != nil {
			resp.Message = "文档不存在"
			return
		}
		docPath = document.Path
	}

	newDocument := models.Document{
		UserId:    userId,
		Path:      docPath,
		DocType:   models.DocTypeShareable,
		TeamId:    teamId,
		ProjectId: projectId,
	}

	documentSize := uint64(0)

	reviewText(uploadData.DocumentText, &newDocument)
	reviewPages(uploadData.PageImageList, &newDocument, docPath, documentService)

	if !newDocument.LockedAt.IsZero() && !isFirstUpload {
		_, _ = documentService.UpdateColumnsById(documentId, map[string]any{
			"locked_at": nil,
		})
	}

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
			if result, err := compressPutObjectByte(pagePath, pageContent); err != nil {
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
		reviewMedias(medias, &newDocument, documentService)

		// upload medias
		for _, media := range *medias {
			path := docPath + "/medias/" + media.Name
			if media.Content == nil {
				continue
			}
			uploadWaitGroup.Add(1)
			go func(path string, media []byte) {
				defer uploadWaitGroup.Done()
				if _, err := storage.Bucket.PutObjectByte(path, media); err != nil {
					resp.Message = "对象上传错误"
					log.Println("对象上传错误", err)
					return
				}
			}(path, *media.Content)
		}
	}
	documentSize += uploadData.MediasSize

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
	uploadData.DocumentMeta["lastCmdId"] = lastCmdId

	documentMetaStr, err := json.Marshal(uploadData.DocumentMeta)
	if err != nil {
		resp.Message = "document-meta.json格式错误"
		log.Println("document-meta.json格式错误", err)
		return
	}
	path := docPath + "/document-meta.json"
	putObjectResult, err := compressPutObjectByte(path, documentMetaStr)
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
	now := myTime.Time(time.Now())
	if isFirstUpload {
		newDocument.Name = documentName
		newDocument.Size = documentSize
		newDocument.VersionId = documentVersionId
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
		DocumentId: documentId,
		VersionId:  documentVersionId,
		LastCmdId:  str.DefaultToInt(lastCmdId, 0),
	}); err != nil {
		resp.Message = "对象上传错误"
		log.Println("对象上传错误5", err)
		return
	}

	resp.Status = ResponseStatusSuccess
	resp.Data = Data{
		"document_id": str.IntToString(documentId),
		"version_id":  documentVersionId,
	}
}
