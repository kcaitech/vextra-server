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
	userConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Fail(c, "建立ws连接失败："+err.Error())
		return
	}
	defer userConn.Close()

	// 连接已经关闭的标识
	hasClose := false
	closeConn := func(msg string) {
		userConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
			websocket.CloseNormalClosure,
			msg,
		))
		hasClose = true
		userConn.Close()
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
		userConn.WriteMessage(websocket.TextMessage, message)
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
		userConn.WriteMessage(websocket.TextMessage, message)
		log.Println("参数错误：" + msg)
		closeConn(msg)
	}

	var size uint64

	// 获取Token
	messageType, reader, err := userConn.NextReader()
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
	var headerData headerDataType
	if err := json.Unmarshal(content, &headerData); err != nil {
		paramsError("headerData")
		return
	}
	// 获取用户信息
	parseData, err := ParseJwt(headerData.Token)
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
	docId := str.DefaultToInt(headerData.DocId, 0)
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

	documentServerConn, _, err := websocket.DefaultDialer.Dial("ws://192.168.0.10:10010", nil)
	if err != nil {
		log.Println("文档服务器连接失败", err)
		connError()
		return
	}
	defer documentServerConn.Close()

	data, err := json.Marshal(map[string]any{
		"documentId": document.Id,
		"userId":     userId,
	})
	if err != nil {
		log.Println("数据格式化失败", err)
		connError()
		documentServerConn.Close()
		return
	}
	_ = documentServerConn.WriteMessage(websocket.TextMessage, data)

	_ = userConn.WriteMessage(websocket.TextMessage, []byte("success"))

	// 创建一个从客户端读取并写入到后端的管道
	go func() {
		for {
			msgType, message, err := userConn.ReadMessage()
			if err != nil {
				log.Println("客户端连接读取异常", err)
				connError()
				documentServerConn.Close()
				break
			}
			err = documentServerConn.WriteMessage(msgType, message)
			if err != nil {
				log.Println("文件服务器连接写入异常", err)
				connError()
				documentServerConn.Close()
				break
			}
		}
	}()

	// 创建一个从后端读取并写入到客户端的管道
	for {
		msgType, message, err := documentServerConn.ReadMessage()
		if err != nil {
			log.Println("文件服务器连接读取异常", err)
			connError()
			documentServerConn.Close()
			break
		}
		err = userConn.WriteMessage(msgType, message)
		if err != nil {
			log.Println("客户端连接写入异常", err)
			connError()
			documentServerConn.Close()
			break
		}
	}

	//type documentDataType struct {
	//	Id     int64   `bson:"_id"`
	//	UserId string  `bson:"user_id"`
	//	Cmd    CmdType `bson:"cmd"`
	//}
	//
	//documentCollection := mongo.DB.Collection("document")
	//
	//var lastId int64
	//pullUpdate := func() bool {
	//	documentUpdateList := []documentDataType{}
	//	findOptions := options.Find()
	//	findOptions.SetSort(bson.D{{"_id", 1}})
	//	filter := bson.M{"document_id": headerData.DocId}
	//	if lastId > 0 {
	//		filter["_id"] = bson.M{"$gt": lastId}
	//	}
	//	if cur, err := documentCollection.Find(nil, filter, findOptions); err == nil {
	//		if err := cur.All(nil, &documentUpdateList); err == nil {
	//			if len(documentUpdateList) == 0 {
	//				return true
	//			}
	//			updateDataList := sliceutil.MapT(func(item documentDataType) CmdType {
	//				item.Cmd["serverId"] = str.IntToString(item.Id)
	//				item.Cmd["userId"] = item.UserId
	//				return item.Cmd
	//			}, documentUpdateList...)
	//			_ = userConn.WriteJSON(&commitDataType{
	//				Type: "update",
	//				Cmds: updateDataList,
	//			})
	//			lastId = documentUpdateList[len(documentUpdateList)-1].Id
	//			return true
	//		}
	//	}
	//	return false
	//}
	//pullUpdate()
	//go func() {
	//	for {
	//		time.Sleep(time.Second * 1)
	//		if !pullUpdate() {
	//			connError()
	//			return
	//		}
	//	}
	//}()
	//
	//for {
	//	commitData := commitDataType{}
	//	err := userConn.ReadJSON(&commitData)
	//	if err != nil {
	//		connError()
	//		return
	//	}
	//	if commitData.Type != "commit" {
	//		continue
	//	}
	//	cmds := sliceutil.MapT(func(cmd CmdType) any {
	//		return CmdType{
	//			"_id":         snowflake.NextId(),
	//			"document_id": headerData.DocId,
	//			"user_id":     parseData.Id,
	//			"unit_id":     cmd["unitId"],
	//			"cmd":         cmd,
	//		}
	//	}, commitData.Cmds...)
	//	if _, err := documentCollection.InsertMany(nil, cmds); err != nil {
	//		log.Println("mongo插入失败", err)
	//	}
	//}

}
