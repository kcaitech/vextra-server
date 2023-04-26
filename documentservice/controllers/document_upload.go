package controllers

import (
	"encoding/json"
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
)

type tokenData struct {
	Token string `json:"token"`
}

type uploadData struct {
	Code string `json:"code"`
	Data struct {
		DocumentMeta     json.RawMessage          `json:"document_meta"`
		Pages            json.RawMessage          `json:"pages"`
		PageRefartboards []json.RawMessage        `json:"page_refartboards"`
		PageRefsyms      []json.RawMessage        `json:"page_refsyms"`
		Artboards        []map[string]interface{} `json:"artboards"`
		ArtboardsRefsyms [][]string               `json:"artboard_refsyms"`
		Symbols          []map[string]interface{} `json:"symbols"`
		Medias           []string                 `json:"medias"`
	} `json:"data"`
}

// UploadHandler 上传文档
func UploadHandler(c *gin.Context) {
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

	closeConn := func(msg string) {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
			websocket.CloseNormalClosure,
			msg,
		))
	}

	connError := func() {
		log.Println("ws连接异常")
		closeConn("ws连接异常")
	}

	dataFormatError := func(msg string) {
		log.Println("数据格式错误：" + msg)
		closeConn("数据格式错误")
	}

	uploadError := func(err error) {
		log.Println("对象上传错误：" + err.Error())
		closeConn("对象上传错误")
	}

	paramsError := func(msg string) {
		if msg == "" {
			msg = "参数错误"
		}
		log.Println("参数错误" + err.Error())
		closeConn(msg)
	}

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
	var tokenDataVal tokenData
	if err := json.Unmarshal(content, &tokenDataVal); err != nil {
		paramsError("")
		return
	}
	// 获取用户信息
	parseData, err := ParseJwt(tokenDataVal.Token)
	if err != nil {
		paramsError("")
		return
	}
	userId := parseData.Id
	if userId <= 0 {
		paramsError("")
		return
	}

	documentService := services.NewDocumentService()
	docId := uuid.New().String()

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

	// 解析uploadData
	var data uploadData
	if err := json.Unmarshal(content, &data); err != nil {
		dataFormatError("uploadData格式错误 " + err.Error())
		return
	}
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
	if len(pages) != len(data.Data.PageRefartboards) || len(pages) != len(data.Data.PageRefsyms) {
		dataFormatError("uploadData.Data.Pages和PageRefartboards、PageRefsyms长度不对应")
		return
	}

	// 上传pages、page-symrefs、page-artboardrefs目录
	for i := 0; i < len(pages); i++ {
		pageId := pages[i].Id
		path := docId + "/pages/" + pageId + ".json"
		if _, err = storage.Bucket.PubObjectByte(path, pagesRaw[i]); err != nil {
			uploadError(err)
			return
		}
		path = docId + "/page-symrefs/" + pageId + ".json"
		if _, err = storage.Bucket.PubObjectByte(path, data.Data.PageRefsyms[i]); err != nil {
			uploadError(err)
			return
		}
		path = docId + "/page-artboardrefs/" + pageId + ".json"
		if _, err = storage.Bucket.PubObjectByte(path, data.Data.PageRefartboards[i]); err != nil {
			uploadError(err)
			return
		}
	}

	// 上传document-meta.json
	path := docId + "/document-meta.json"
	if _, err = storage.Bucket.PubObjectByte(path, data.Data.DocumentMeta); err != nil {
		uploadError(err)
		return
	}

	// 获取文档名称
	var documentMeta struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data.Data.DocumentMeta, &documentMeta); err != nil {
		documentMeta.Name = docId
	}

	// 创建文档记录
	if err = documentService.Create(&models.Document{
		UserId:  userId,
		Path:    "/" + docId,
		DocType: models.DocTypePrivate,
		Name:    documentMeta.Name,
		Size:    uint(len(content)),
	}); err != nil {
		uploadError(err)
	}

	closeConn("")
}
