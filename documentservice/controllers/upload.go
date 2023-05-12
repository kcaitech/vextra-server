package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"protodesign.cn/kcserver/common/gin/response"
	. "protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/storage"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
	"sync"
	"time"
)

type headerData struct {
	Token string `json:"token"`
	DocId string `json:"doc_id"`
}

type uploadData struct {
	Code string `json:"code"`
	Data struct {
		DocumentMeta     json.RawMessage   `json:"document_meta"`
		Pages            json.RawMessage   `json:"pages"`
		PageRefartboards []json.RawMessage `json:"page_refartboards"`
		PageRefsyms      []json.RawMessage `json:"page_refsyms"`
		Artboards        json.RawMessage   `json:"artboards"`
		ArtboardsRefsyms []json.RawMessage `json:"artboard_refsyms"`
		Symbols          json.RawMessage   `json:"symbols"`
		MediaNames       []string          `json:"media_names"`
	} `json:"data"`
}

// UploadDocumentByUser 用户上传文档
func UploadDocumentByUser(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Fail(c, "建立ws连接失败："+err.Error())
		return
	}
	defer conn.Close()

	// 连接已经关闭的标识
	hasClose := false
	closeConn := func(msg string) {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
			websocket.CloseNormalClosure,
			msg,
		))
		hasClose = true
	}
	connError := func() {
		if hasClose {
			return
		}
		log.Println("ws连接异常")
		closeConn("ws连接异常")
	}
	dataFormatError := func(msg string) {
		if hasClose {
			return
		}
		log.Println("数据格式错误：" + msg)
		closeConn("数据格式错误")
	}
	uploadError := func(err error) {
		if hasClose {
			return
		}
		log.Println("对象上传错误：" + err.Error())
		closeConn("对象上传错误")
	}
	paramsError := func(msg string) {
		if hasClose {
			return
		}
		if msg == "" {
			msg = "参数错误"
		}
		log.Println("参数错误：" + msg)
		closeConn(msg)
	}
	done := func(data map[string]interface{}) {
		if hasClose {
			return
		}
		data["code"] = "done"
		message, _ := json.Marshal(data)
		conn.WriteMessage(websocket.TextMessage, message)
		closeConn("上传成功")
	}

	var size uint64

	// 获取Token
	messageType, reader, err := conn.NextReader()
	if err != nil {
		connError()
		return
	}
	if messageType != websocket.TextMessage {
		dataFormatError("未传Token")
		return
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		connError()
		return
	}
	size += uint64(len(content))
	var headerDataVal headerData
	if err := json.Unmarshal(content, &headerDataVal); err != nil {
		paramsError("")
		return
	}
	// 获取用户信息
	parseData, err := ParseJwt(headerDataVal.Token)
	if err != nil {
		paramsError("")
		return
	}
	userId, err := str.ToInt(parseData.Id)
	if err != nil {
		paramsError("")
		return
	}
	if userId <= 0 {
		paramsError("")
		return
	}
	// 获取文档信息
	documentService := services.NewDocumentService()
	var document models.Document
	docPath := uuid.New().String()
	docId := str.DefaultToInt(headerDataVal.DocId, 0)
	if docId > 0 {
		if documentService.GetById(docId, &document) != nil {
			paramsError("文档不存在")
			return
		}
		if document.UserId != userId {
			paramsError("无编辑权限")
			return
		}
		docPath = document.Path
	}

	// 获取maindoc
	messageType, reader, err = conn.NextReader()
	if err != nil {
		connError()
		return
	}
	if messageType != websocket.TextMessage {
		dataFormatError("未传maindoc")
		return
	}
	content, err = io.ReadAll(reader)
	if err != nil {
		connError()
		return
	}
	size += uint64(len(content))
	// 解析uploadData
	var data uploadData
	if err := json.Unmarshal(content, &data); err != nil {
		dataFormatError("uploadData格式错误 " + err.Error())
		return
	}

	uploadWaitGroup := sync.WaitGroup{}

	// pages部分
	var pages []struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(data.Data.Pages, &pages); err != nil {
		dataFormatError("uploadData.Data.Pages内有元素缺少Id " + err.Error())
		return
	}
	var pagesRaw []json.RawMessage
	if err := json.Unmarshal(data.Data.Pages, &pagesRaw); err != nil {
		dataFormatError("uploadData.Data.Pages格式错误 " + err.Error())
		return
	}
	pageRefsyms := data.Data.PageRefsyms
	pageRefartboards := data.Data.PageRefartboards
	if len(pages) != len(pageRefsyms) || len(pages) != len(pageRefartboards) {
		dataFormatError("uploadData.Data.Pages和PageRefartboards、PageRefsyms长度不对应")
		return
	}
	// 上传pages、page-symrefs、page-artboardrefs目录
	for i := 0; i < len(pages); i++ {
		pageId := pages[i].Id
		pagePath := docPath + "/pages/" + pageId + ".json"
		pageContent := pagesRaw[i]
		pageRefsymsPath := docPath + "/page-symrefs/" + pageId + ".json"
		pageRefsymsContent := pageRefsyms[i]
		pageRefartboardPath := docPath + "/page-artboardrefs/" + pageId + ".json"
		pageRefartboardContent := pageRefartboards[i]

		uploadWaitGroup.Add(1)
		go func() {
			defer uploadWaitGroup.Done()
			if _, err = storage.Bucket.PubObjectByte(pagePath, pageContent); err != nil {
				uploadError(err)
				return
			}
			if _, err = storage.Bucket.PubObjectByte(pageRefsymsPath, pageRefsymsContent); err != nil {
				uploadError(err)
				return
			}
			if _, err = storage.Bucket.PubObjectByte(pageRefartboardPath, pageRefartboardContent); err != nil {
				uploadError(err)
				return
			}
		}()
	}

	// artboards部分
	var artboards []struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(data.Data.Artboards, &artboards); err != nil {
		dataFormatError("uploadData.Data.Artboards内有元素缺少Id " + err.Error())
		return
	}
	var artboardsRaw []json.RawMessage
	if err := json.Unmarshal(data.Data.Artboards, &artboardsRaw); err != nil {
		dataFormatError("uploadData.Data.Artboards格式错误 " + err.Error())
		return
	}
	artboardsRefsyms := data.Data.ArtboardsRefsyms
	if len(artboards) != len(artboardsRefsyms) {
		dataFormatError("uploadData.Data.Artboards和ArtboardsRefsyms长度不对应")
		return
	}
	// 上传artboards、artboards-refsyms目录
	for i := 0; i < len(artboards); i++ {
		artboardId := artboards[i].Id
		artboardPath := docPath + "/artboards/" + artboardId + ".json"
		artboardContent := artboardsRaw[i]
		artboardRefsymsPath := docPath + "/artboards-refsyms/" + artboardId + ".json"
		artboardRefsymsContent := artboardsRefsyms[i]
		uploadWaitGroup.Add(1)
		go func() {
			defer uploadWaitGroup.Done()
			if _, err = storage.Bucket.PubObjectByte(artboardPath, artboardContent); err != nil {
				uploadError(err)
				return
			}
			if _, err = storage.Bucket.PubObjectByte(artboardRefsymsPath, artboardRefsymsContent); err != nil {
				uploadError(err)
				return
			}
		}()
	}

	// symbols部分
	var symbols []struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(data.Data.Symbols, &symbols); err != nil {
		dataFormatError("uploadData.Data.Symbols内有元素缺少Id " + err.Error())
		return
	}
	var symbolsRaw []json.RawMessage
	if err := json.Unmarshal(data.Data.Symbols, &symbolsRaw); err != nil {
		dataFormatError("uploadData.Data.Symbols格式错误 " + err.Error())
		return
	}
	// 上传symbols目录
	for i := 0; i < len(symbols); i++ {
		symbolId := symbols[i].Id
		symbolContent := symbolsRaw[i]
		path := docPath + "/symbols/" + symbolId + ".json"
		uploadWaitGroup.Add(1)
		go func() {
			defer uploadWaitGroup.Done()
			if _, err = storage.Bucket.PubObjectByte(path, symbolContent); err != nil {
				uploadError(err)
				return
			}
		}()
	}

	// medias部分
	nextMedia := func() []byte {
		messageType, reader, err = conn.NextReader()
		if err != nil {
			connError()
			return nil
		}
		if messageType != websocket.BinaryMessage {
			dataFormatError("media格式错误")
			return nil
		}
		content, err = io.ReadAll(reader)
		if err != nil {
			connError()
			return nil
		}
		size += uint64(len(content))
		return content
	}
	for _, mediaName := range data.Data.MediaNames {
		path := docPath + "/medias/" + mediaName
		media := nextMedia()
		if media == nil {
			return
		}
		uploadWaitGroup.Add(1)
		go func() {
			defer uploadWaitGroup.Done()
			if _, err = storage.Bucket.PubObjectByte(path, media); err != nil {
				uploadError(err)
				return
			}
		}()
	}

	uploadWaitGroup.Wait()
	if hasClose {
		return
	}

	// 上传document-meta.json
	path := docPath + "/document-meta.json"
	if _, err = storage.Bucket.PubObjectByte(path, data.Data.DocumentMeta); err != nil {
		uploadError(err)
		return
	}

	// 获取文档名称
	var documentMeta struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data.Data.DocumentMeta, &documentMeta); err != nil {
		documentMeta.Name = docPath
	}
	documentAccessRecordService := services.NewDocumentAccessRecordService()
	// 创建文档记录和历史记录
	now := myTime.Time(time.Now())
	if docId <= 0 {
		newDocument := models.Document{
			UserId:  userId,
			Path:    docPath,
			DocType: models.DocTypePrivate,
			Name:    documentMeta.Name,
			Size:    size,
		}
		if err = documentService.Create(&newDocument); err != nil {
			uploadError(err)
			return
		}
		docId = newDocument.Id
		_ = documentAccessRecordService.Create(&models.DocumentAccessRecord{
			UserId:         userId,
			DocumentId:     docId,
			LastAccessTime: now,
		})
	} else {
		if err := documentService.UpdatesById(docId, &document); err != nil {
			uploadError(err)
			return
		}
		var documentAccessRecord models.DocumentAccessRecord
		err := documentAccessRecordService.Get(&documentAccessRecord, "user_id = ? and document_id = ?", userId, docId)
		if err != nil && err != services.ErrRecordNotFound {
			uploadError(err)
			return
		}
		if err == services.ErrRecordNotFound {
			_ = documentAccessRecordService.Create(&models.DocumentAccessRecord{
				UserId:         userId,
				DocumentId:     docId,
				LastAccessTime: now,
			})
		} else {
			documentAccessRecord.LastAccessTime = now
			_ = documentAccessRecordService.UpdatesById(documentAccessRecord.Id, &documentAccessRecord)
		}
	}

	done(map[string]interface{}{
		"doc_id": fmt.Sprintf("%d", docId),
	})
}
