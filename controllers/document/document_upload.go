package document

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"kcaitech.com/kcserver/common/safereview"
	safereviewBase "kcaitech.com/kcserver/common/safereview/base"
	"kcaitech.com/kcserver/utils/my_map"
	"kcaitech.com/kcserver/utils/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"kcaitech.com/kcserver/common/gin/response"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/services"
	"kcaitech.com/kcserver/common/storage"
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
	DocumentMeta        Data            `json:"document_meta"`
	Pages               json.RawMessage `json:"pages"`
	FreeSymbols         json.RawMessage `json:"freesymbols"`
	MediaNames          []string        `json:"media_names"`
	MediasSize          uint64          `json:"medias_size"`
	DocumentText        string          `json:"document_text"`
	PageImageBase64List []string        `json:"page_image_base64_list"`
}

func UploadDocument(c *gin.Context) {
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Fail(c, "建立ws连接失败："+err.Error())
		return
	}
	resp := Response{
		Status: ResponseStatusFail,
	}
	defer func() {
		ws.WriteJSON(&resp)
		ws.Close()
	}()

	header := Header{}
	if err := ws.ReadJSON(&header); err != nil {
		resp.Message = "Header结构错误"
		log.Println("Header结构错误", err)
		return
	}

	uploadData := UploadData{}
	if err := ws.ReadJSON(&uploadData); err != nil {
		resp.Message = "UploadData结构错误"
		log.Println("UploadData结构错误", err)
		return
	}

	documentId := str.DefaultToInt(header.DocumentId, 0)
	isFirstUpload := documentId <= 0

	medias := []([]byte){}
	if isFirstUpload {
		nextMedia := func() []byte {
			messageType, data, err := ws.ReadMessage()
			if err != nil {
				resp.Message = "ws连接异常"
				log.Println("ws连接异常", err)
				return nil
			}
			if messageType != websocket.MessageTypeBinary {
				resp.Message = "media格式错误"
				log.Println("media格式错误")
				return nil
			}
			return data
		}

		for range uploadData.MediaNames {
			m := nextMedia()
			if nil == m {
				return
			}
			medias = append(medias, m)
		}
	}

	UploadDocumentData(&header, &uploadData, &medias, &resp)
}

func UploadDocumentData(header *Header, uploadData *UploadData, medias *[]([]byte), resp *Response) {

	// resp := Response{
	// 	Status: ResponseStatusFail,
	// }

	// header := Header{}
	// if err := ws.ReadJSON(&header); err != nil {
	// 	resp.Message = "Header结构错误"
	// 	_ = ws.WriteJSON(&resp)
	// 	log.Println("Header结构错误", err)
	// 	return
	// }

	userId := str.DefaultToInt(header.UserId, 0)
	documentId := str.DefaultToInt(header.DocumentId, 0)
	projectId := str.DefaultToInt(header.ProjectId, 0)
	lastCmdId := header.LastCmdId
	if (userId <= 0 && documentId <= 0) || (userId > 0 && documentId > 0) || (documentId > 0 && lastCmdId == "") { // userId和documentId必须只传一个
		resp.Message = "参数错误"
		// _ = ws.WriteJSON(&resp)
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
			// _ = ws.WriteJSON(&resp)
			return
		}
		teamId = project.TeamId
		permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
		if err != nil || permType == nil {
			log.Println("获取项目权限失败", err)
		}
		if err != nil || permType == nil || *permType < models.ProjectPermTypeEditable {
			resp.Message = "无权限"
			// _ = ws.WriteJSON(&resp)
			return
		}
	}

	// 获取文档信息
	documentService := services.NewDocumentService()
	var document models.Document
	docPath := uuid.New().String()
	if !isFirstUpload {
		if documentService.GetById(documentId, &document) != nil {
			resp.Message = "文档不存在"
			// _ = ws.WriteJSON(&resp)
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

	if uploadData.DocumentText != "" {
		go func() {
			reviewResponse, err := safereview.Client.ReviewText(uploadData.DocumentText)
			if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
				newDocument.LockedAt = myTime.Time(time.Now())
				newDocument.LockedReason = "文本审核不通过：" + reviewResponse.Reason
				if wordsBytes, err := json.Marshal(reviewResponse.Words); err == nil {
					newDocument.LockedWords += string(wordsBytes)
				}
			}
		}()
	}
	if len(uploadData.PageImageBase64List) > 0 {
		go func() {
			needUpdateDocument := false
			for i, base64Str := range uploadData.PageImageBase64List {
				path := docPath + "/page_image/" + str.IntToString(int64(i)) + ".png"
				image, err := base64.StdEncoding.DecodeString(base64Str)
				if err != nil {
					log.Println("图片base64解码错误", err)
				} else if _, err := storage.Bucket.PutObjectByte(path, image); err != nil {
					log.Println("图片上传错误", err)
				}
				if len(base64Str) == 0 {
					continue
				}
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
				_, _ = documentService.UpdatesById(newDocument.Id, &newDocument)
			}
		}()
	}
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
		// _ = ws.WriteJSON(&resp)
		// ws.Close()
		return
	}
	var pagesRaw []json.RawMessage
	if err := json.Unmarshal(uploadData.Pages, &pagesRaw); err != nil {
		resp.Message = "Pages格式错误 " + err.Error()
		log.Println("Pages格式错误", err)
		// _ = ws.WriteJSON(&resp)
		// ws.Close()
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
			if result, err := storage.Bucket.PutObjectByte(pagePath, pageContent); err != nil {
				resp.Message = "对象上传错误"
				log.Println("对象上传错误", err)
				// _ = ws.WriteJSON(&resp)
				// ws.Close()
				return
			} else {
				idToVersionId.Set(pageId, result.VersionID)
			}
		}(pagePath, pageContent)
	}

	// medias部分
	if isFirstUpload {
		mediaInfoList := make([]struct {
			Name    string
			Content []byte
		}, len(uploadData.MediaNames))
		for idx, mediaName := range uploadData.MediaNames {
			path := docPath + "/medias/" + mediaName
			media := (*medias)[idx]
			if media == nil {
				continue // return?
			}
			mediaInfoList = append(mediaInfoList, struct {
				Name    string
				Content []byte
			}{
				Name:    mediaName,
				Content: media,
			})
			documentSize += uint64(len(media))
			uploadWaitGroup.Add(1)
			go func(path string, media []byte) {
				defer uploadWaitGroup.Done()
				if _, err := storage.Bucket.PutObjectByte(path, media); err != nil {
					resp.Message = "对象上传错误"
					log.Println("对象上传错误", err)
					// _ = ws.WriteJSON(&resp)
					// ws.Close()
					return
				}
			}(path, media)
		}
		if len(mediaInfoList) > 0 {
			go func() {
				needUpdateDocument := false
				for _, mediaInfo := range mediaInfoList {
					base64Str := base64.StdEncoding.EncodeToString(mediaInfo.Content)
					if len(mediaInfo.Content) == 0 || len(base64Str) == 0 {
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
					_, _ = documentService.UpdatesById(newDocument.Id, &newDocument)
				}
			}()
		}
	} else {
		documentSize += uploadData.MediasSize
	}

	uploadWaitGroup.Wait()
	// if ws.IsClose() {
	// 	log.Println("ws已关闭")
	// }

	freesymbolsVersionId := ""
	// 上传freesymbols.json
	freeSymbolsPath := docPath + "/freesymbols.json"
	if result, err := storage.Bucket.PutObjectByte(freeSymbolsPath, uploadData.FreeSymbols); err != nil {
		resp.Message = "对象上传错误"
		log.Println("freesymbols.json上传错误", err)
		// _ = ws.WriteJSON(&resp)
		// ws.Close()
		return
	} else {
		freesymbolsVersionId = result.VersionID
	}
	documentSize += uint64(len(uploadData.FreeSymbols))

	// 设置versionId
	pagesList := uploadData.DocumentMeta["pagesList"].([]any)
	for _, page := range pagesList {
		pageItem, ok := page.(map[string]any)
		pageId, ok1 := pageItem["id"].(string)
		versionId, ok2 := idToVersionId.Get(pageId)
		if !ok || !ok1 || !ok2 {
			resp.Message = "pagesList格式错误"
			log.Println("pagesList格式错误")
			// _ = ws.WriteJSON(&resp)
			// ws.Close()
			return
		}
		pageItem["versionId"] = versionId
	}
	uploadData.DocumentMeta["lastCmdId"] = lastCmdId
	uploadData.DocumentMeta["freesymbolsVersionId"] = freesymbolsVersionId
	// 上传document-meta.json
	documentMetaStr, err := json.Marshal(uploadData.DocumentMeta)
	if err != nil {
		resp.Message = "document-meta.json格式错误"
		log.Println("document-meta.json格式错误", err)
		// _ = ws.WriteJSON(&resp)
		// ws.Close()
		return
	}
	path := docPath + "/document-meta.json"
	putObjectResult, err := storage.Bucket.PutObjectByte(path, documentMetaStr)
	if err != nil {
		resp.Message = "对象上传错误"
		log.Println("document-meta.json上传错误", err)
		// _ = ws.WriteJSON(&resp)
		// ws.Close()
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
			log.Println("对象上传错误", err)
			// _ = ws.WriteJSON(&resp)
			// ws.Close()
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
			resp.Message = "对象上传错误.."
			log.Println("对象上传错误..", err)
			// _ = ws.WriteJSON(&resp)
			// ws.Close()
			return
		}
		var documentAccessRecord models.DocumentAccessRecord
		err := documentAccessRecordService.Get(&documentAccessRecord, "user_id = ? and document_id = ?", userId, documentId)
		if err != nil && !errors.Is(err, services.ErrRecordNotFound) {
			resp.Message = "对象上传错误..."
			log.Println("对象上传错误...", err)
			// _ = ws.WriteJSON(&resp)
			// ws.Close()
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
		resp.Message = "对象上传错误...."
		log.Println("对象上传错误....", err)
		// _ = ws.WriteJSON(&resp)
		// ws.Close()
		return
	}

	resp.Status = ResponseStatusSuccess
	resp.Data = Data{
		"doc_id":     str.IntToString(documentId),
		"version_id": documentVersionId,
	}
	// _ = ws.WriteJSON(&resp)
}
