package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"protodesign.cn/kcserver/common/gin/response"
	. "protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
)

// IncrementalUpload 增量上传
func IncrementalUpload(c *gin.Context) {
	type headerData struct {
		Token string `json:"token"`
		DocId string `json:"doc_id"`
	}

	type opData struct {
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
		data := map[string]any{
			"code":    "error",
			"message": msg,
		}
		message, _ := json.Marshal(data)
		conn.WriteMessage(websocket.TextMessage, message)
		log.Println("数据格式错误：" + msg)
		closeConn("数据格式错误")
	}
	paramsError := func(msg string) {
		if hasClose {
			return
		}
		if msg == "" {
			msg = "参数错误"
		}
		data := map[string]any{
			"code":    "error",
			"message": msg,
		}
		message, _ := json.Marshal(data)
		conn.WriteMessage(websocket.TextMessage, message)
		log.Println("参数错误：" + msg)
		closeConn(msg)
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
		paramsError("headerData")
		return
	}
	// 获取用户信息
	parseData, err := ParseJwt(headerDataVal.Token)
	if err != nil {
		paramsError("Token")
		return
	}
	userId, err := str.ToInt(parseData.Id)
	if err != nil {
		paramsError("Id")
		return
	}
	if userId <= 0 {
		paramsError("Id")
		return
	}
	// 获取文档信息
	docId := str.DefaultToInt(headerDataVal.DocId, 0)
	if docId <= 0 {
		paramsError("docId")
		return
	}
	documentService := services.NewDocumentService()
	var document models.Document
	if documentService.GetById(docId, &document) != nil {
		paramsError("文档不存在")
		return
	}

}
