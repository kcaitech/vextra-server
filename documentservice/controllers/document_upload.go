package controllers

import (
	"encoding/json"
	"log"
	"protodesign.cn/kcserver/utils/my_map"
	"protodesign.cn/kcserver/utils/websocket"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/storage"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
)

func UploadDocument(c *gin.Context) {
	type Data map[string]any

	type Header struct {
		UserId     string `json:"user_id"`
		DocumentId string `json:"document_id"`
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

	ws, err := websocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Fail(c, "建立ws连接失败："+err.Error())
		return
	}
	defer ws.Close()

	resp := Response{
		Status: ResponseStatusFail,
	}

	header := Header{}
	if err := ws.ReadJSON(&header); err != nil {
		resp.Message = "Header结构错误"
		_ = ws.WriteJSON(&resp)
		log.Println("Header结构错误", err)
		return
	}
	userId := str.DefaultToInt(header.UserId, 0)
	documentId := str.DefaultToInt(header.DocumentId, 0)
	lastCmdId := header.LastCmdId
	if (userId <= 0 && documentId <= 0) || (userId > 0 && documentId > 0) || (documentId > 0 && lastCmdId == "") { // userId和documentId必须只传一个
		resp.Message = "参数错误"
		_ = ws.WriteJSON(&resp)
		log.Println("参数错误", userId, documentId)
		return
	}

	// 获取文档信息
	documentService := services.NewDocumentService()
	var document models.Document
	docPath := uuid.New().String()
	if documentId > 0 {
		if documentService.GetById(documentId, &document) != nil {
			resp.Message = "文档不存在"
			_ = ws.WriteJSON(&resp)
			return
		}
		docPath = document.Path
	}

	documentSize := uint64(0)

	type UploadData struct {
		DocumentMeta Data            `json:"document_meta"`
		Pages        json.RawMessage `json:"pages"`
		DocumentSyms json.RawMessage `json:"document_syms"`
		MediaNames   []string        `json:"media_names"`
	}
	uploadData := UploadData{}
	if err := ws.ReadJSON(&uploadData); err != nil {
		resp.Message = "UploadData结构错误"
		_ = ws.WriteJSON(&resp)
		log.Println("UploadData结构错误", err)
		return
	}

	uploadWaitGroup := sync.WaitGroup{}

	// pages部分
	var pages []struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(uploadData.Pages, &pages); err != nil {
		resp.Message = "Pages内有元素缺少Id " + err.Error()
		_ = ws.WriteJSON(&resp)
		ws.Close()
		return
	}
	var pagesRaw []json.RawMessage
	if err := json.Unmarshal(uploadData.Pages, &pagesRaw); err != nil {
		resp.Message = "Pages格式错误 " + err.Error()
		_ = ws.WriteJSON(&resp)
		ws.Close()
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
				_ = ws.WriteJSON(&resp)
				ws.Close()
				return
			} else {
				idToVersionId.Set(pageId, result.VersionID)
			}
		}(pagePath, pageContent)
	}

	// medias部分
	nextMedia := func() []byte {
		messageType, data, err := ws.ReadMessage()
		if err != nil {
			resp.Message = "ws连接异常"
			_ = ws.WriteJSON(&resp)
			ws.Close()
			return nil
		}
		if messageType != websocket.MessageTypeBinary {
			resp.Message = "media格式错误"
			_ = ws.WriteJSON(&resp)
			ws.Close()
			return nil
		}
		return data
	}
	for _, mediaName := range uploadData.MediaNames {
		path := docPath + "/medias/" + mediaName
		media := nextMedia()
		if media == nil {
			return
		}
		documentSize += uint64(len(media))
		uploadWaitGroup.Add(1)
		go func(path string, media []byte) {
			defer uploadWaitGroup.Done()
			if _, err := storage.Bucket.PutObjectByte(path, media); err != nil {
				resp.Message = "对象上传错误"
				log.Println("对象上传错误", err)
				_ = ws.WriteJSON(&resp)
				ws.Close()
				return
			}
		}(path, media)
	}

	uploadWaitGroup.Wait()
	if ws.IsClose() {
		return
	}

	// 设置versionId
	pagesList := uploadData.DocumentMeta["pagesList"].([]any)
	for _, page := range pagesList {
		pageItem, ok := page.(map[string]any)
		pageId, ok1 := pageItem["id"].(string)
		versionId, ok2 := idToVersionId.Get(pageId)
		if !ok || !ok1 || !ok2 {
			resp.Message = "pagesList格式错误"
			_ = ws.WriteJSON(&resp)
			ws.Close()
			return
		}
		pageItem["versionId"] = versionId
	}
	uploadData.DocumentMeta["lastCmdId"] = lastCmdId
	// 上传document-meta.json
	documentMetaStr, err := json.Marshal(uploadData.DocumentMeta)
	if err != nil {
		resp.Message = "document-meta.json格式错误"
		_ = ws.WriteJSON(&resp)
		ws.Close()
		return
	}
	path := docPath + "/document-meta.json"
	putObjectResult, err := storage.Bucket.PutObjectByte(path, documentMetaStr)
	if err != nil {
		resp.Message = "对象上传错误"
		log.Println("document-meta.json上传错误", err)
		_ = ws.WriteJSON(&resp)
		ws.Close()
		return
	}
	documentVersionId := putObjectResult.VersionID
	documentSize += uint64(len(uploadData.DocumentMeta))

	// 上传document-syms.json
	symsPath := docPath + "/document-syms.json"
	if _, err := storage.Bucket.PutObjectByte(symsPath, uploadData.DocumentSyms); err != nil {
		resp.Message = "对象上传错误"
		log.Println("document-syms.json上传错误", err)
		_ = ws.WriteJSON(&resp)
		ws.Close()
		return
	}
	documentSize += uint64(len(uploadData.DocumentSyms))

	// 获取文档名称
	documentName, _ := uploadData.DocumentMeta["name"].(string)
	documentAccessRecordService := services.NewDocumentAccessRecordService()
	// 创建文档记录和历史记录
	now := myTime.Time(time.Now())
	if documentId <= 0 {
		newDocument := models.Document{
			UserId:    userId,
			Path:      docPath,
			DocType:   models.DocTypeShareable,
			Name:      documentName,
			Size:      documentSize,
			VersionId: documentVersionId,
		}
		if err := documentService.Create(&newDocument); err != nil {
			resp.Message = "对象上传错误."
			_ = ws.WriteJSON(&resp)
			ws.Close()
			return
		}
		documentId = newDocument.Id
		_ = documentAccessRecordService.Create(&models.DocumentAccessRecord{
			UserId:         userId,
			DocumentId:     documentId,
			LastAccessTime: now,
		})
	} else {
		document.Name = documentName
		document.Size = documentSize
		document.VersionId = documentVersionId
		if err := documentService.UpdatesById(documentId, &document); err != nil {
			resp.Message = "对象上传错误.."
			_ = ws.WriteJSON(&resp)
			ws.Close()
			return
		}
		var documentAccessRecord models.DocumentAccessRecord
		err := documentAccessRecordService.Get(&documentAccessRecord, "user_id = ? and document_id = ?", userId, documentId)
		if err != nil && err != services.ErrRecordNotFound {
			resp.Message = "对象上传错误..."
			_ = ws.WriteJSON(&resp)
			ws.Close()
			return
		}
		if err == services.ErrRecordNotFound {
			_ = documentAccessRecordService.Create(&models.DocumentAccessRecord{
				UserId:         userId,
				DocumentId:     documentId,
				LastAccessTime: now,
			})
		} else {
			documentAccessRecord.LastAccessTime = now
			_ = documentAccessRecordService.UpdatesById(documentAccessRecord.Id, &documentAccessRecord)
		}
	}

	// 创建文档版本记录
	if err := documentService.DocumentVersionService.Create(&models.DocumentVersion{
		DocumentId: documentId,
		VersionId:  documentVersionId,
		LastCmdId:  str.DefaultToInt(lastCmdId, 0),
	}); err != nil {

	}

	resp.Status = ResponseStatusSuccess
	resp.Data = Data{
		"doc_id":     str.IntToString(documentId),
		"version_id": documentVersionId,
	}
	_ = ws.WriteJSON(&resp)
}
