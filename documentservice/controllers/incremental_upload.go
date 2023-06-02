package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"log"
	"net/http"
	"protodesign.cn/kcserver/common/gin/response"
	. "protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/mongo"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/snowflake"
	"protodesign.cn/kcserver/utils/sliceutil"
	"protodesign.cn/kcserver/utils/str"
)

// IncrementalUpload 增量上传
func IncrementalUpload(c *gin.Context) {
	type headerDataType struct {
		Token     string `json:"token"`
		DocId     string `json:"doc_id"`
		VersionId string `json:"version_id"`
	}

	type CmdType map[string]any

	type commitDataType struct {
		Type string    `json:"type"`
		Cmds []CmdType `json:"cmds"`
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
	var headerDataVal headerDataType
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
	userId := str.DefaultToInt(parseData.Id, 0)
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

	_ = conn.WriteMessage(websocket.TextMessage, []byte("success"))

	type documentDataType struct {
		Id     int64   `bson:"_id"`
		UserId string  `bson:"user_id"`
		Cmd    CmdType `bson:"cmd"`
	}

	collection := mongo.DB.Collection("document")

	documentUpdateList := []documentDataType{}
	if cur, err := collection.Find(nil, bson.M{"document_id": headerDataVal.DocId}); err == nil {
		if err := cur.All(nil, &documentUpdateList); err == nil && len(documentUpdateList) > 0 {
			updateDataList := sliceutil.MapT(func(item documentDataType) CmdType {
				item.Cmd["_serverId"] = str.IntToString(item.Id)
				item.Cmd["userId"] = item.UserId
				return item.Cmd
			}, documentUpdateList...)
			_ = conn.WriteJSON(&commitDataType{
				Type: "update",
				Cmds: updateDataList,
			})
		}
	}

	for {
		commitDataVal := commitDataType{}
		err := conn.ReadJSON(&commitDataVal)
		if err != nil {
			connError()
			return
		}
		if commitDataVal.Type != "commit" {
			continue
		}
		cmds := sliceutil.MapT(func(cmd CmdType) any {
			return CmdType{
				"_id":         snowflake.NextId(),
				"document_id": headerDataVal.DocId,
				"user_id":     parseData.Id,
				"unit_id":     cmd["_unitId"],
				"cmd":         cmd,
			}
		}, commitDataVal.Cmds...)
		if _, err := collection.InsertMany(nil, cmds); err != nil {
			log.Println("mongo插入失败", err)
		}
	}

}
