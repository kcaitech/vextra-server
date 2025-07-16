package ws

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync/atomic"

	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/str"
	"kcaitech.com/kcserver/utils/websocket"
)

func decodeBinaryMessage(data []byte) (string, []byte, error) {
	// 假设长度前缀为4字节
	lengthPrefixSize := uint32(4)

	if len(data) < int(lengthPrefixSize) {

		return "", nil, fmt.Errorf("data is too short to contain the length prefix")
	}

	// 读取字符串长度
	strLength := binary.LittleEndian.Uint32(data[:lengthPrefixSize])

	// 检查数据是否足够长以包含整个字符串
	if len(data) < int(lengthPrefixSize+strLength) {
		return "", nil, fmt.Errorf("data is too short to contain the string")
	}

	// 提取字符串部分
	strBytes := data[lengthPrefixSize : lengthPrefixSize+strLength]
	str := string(strBytes)

	// 提取 ArrayBuffer 部分
	bufferStart := lengthPrefixSize + strLength
	bufferData := data[bufferStart:]

	return str, bufferData, nil
}

type BindData struct {
	DocumentId string `json:"document_id"`
	// VersionId  string `json:"version_id"`
	Perm string `json:"perm_type,omitempty"`
}

type StartData struct {
	LastCmdVersion uint `json:"last_cmd_version,omitempty"`
}

type ServeFace interface {
	handle(data *TransData, binaryData *([]byte))
	close()
}

// 检查接口是否为nil（包括底层值为nil的情况）
func isInterfaceNil(i interface{}) bool {
	if i == nil {
		return true
	}
	return reflect.ValueOf(i).IsNil()
}

type WSClient struct {
	serveMap   map[string]ServeFace
	ws         *websocket.Ws
	token      string
	genSId     func() string
	userId     string
	documentId string
	versionId  string
}

func NewWSClient(ws *websocket.Ws, token string, userId string) *WSClient {
	var _sid atomic.Int64
	genSId := func() string {
		sid := _sid.Add(1)
		return "s" + str.IntToString(sid)
	}
	return &WSClient{
		ws:       ws,
		token:    token,
		userId:   userId,
		genSId:   genSId,
		serveMap: map[string]ServeFace{},
	}
}

func (c *WSClient) msgErr(msg string, serverData *TransData, err *error) {
	serverData.Msg = msg
	if err != nil {
		log.Println(msg, *err)
	} else {
		log.Println(msg)
	}
	_ = c.ws.WriteJSON(serverData)
}

func (c *WSClient) msgErrWithCode(msg string, serverData *TransData, err *error, errCode int) {
	serverData.Msg = msg
	serverData.Code = int32(errCode)
	if err != nil {
		log.Println(msg, *err)
	} else {
		log.Println(msg)
	}
	_ = c.ws.WriteJSON(serverData)
}

func (c *WSClient) bindServe(t string, s ServeFace) {
	old := c.serveMap[t]
	if old != nil {
		(old).close()
	}
	if !isInterfaceNil(s) {
		c.serveMap[t] = s
	} else {
		delete(c.serveMap, t)
	}
}

func (c *WSClient) handleBind(clientData *TransData) {
	serverData := TransData{}
	serverData.Type = clientData.Type
	serverData.DataId = clientData.DataId

	bindData := BindData{}
	err := json.Unmarshal([]byte(clientData.Data), &bindData)
	if err != nil {
		c.msgErr("wrong bind struct", &serverData, nil)
		return
	}

	documentId := (bindData.DocumentId)
	if documentId == "" {
		c.msgErr("wrong document id", &serverData, nil)
		return
	}

	permType := models.PermType(str.DefaultToInt(bindData.Perm, 0))
	if permType < models.PermTypeReadOnly || permType > models.PermTypeEditable {
		permType = models.PermTypeNone
	}

	docService := services.NewDocumentService()
	docInfo := docService.GetDocumentInfoByDocumentAndUserId(documentId, c.userId, permType)
	if docInfo == nil {
		c.msgErrWithCode("文档不存在", &serverData, nil, http.StatusNotFound)
		return
	}
	locked, _ := docService.GetLocked(documentId)
	if len(locked) > 0 && docInfo.Document.UserId != c.userId {
		c.msgErrWithCode("审核不通过", &serverData, nil, common.StatusDocumentNotFound)
		return
	}

	accessKey, err_code, err := common.GetDocumentAccessKey(c.userId, documentId)
	if err != nil {
		c.msgErrWithCode(err.Error(), &serverData, nil, err_code)
		return
	}

	retstr, err := json.Marshal(&map[string]any{
		"doc_info":   docInfo,
		"access_key": accessKey,
	})
	if err != nil {
		c.msgErr("unknow", &serverData, nil)
		return
	}

	c.documentId = documentId
	c.versionId = docInfo.Document.VersionId

	serverData.Data = string(retstr)
	// send back message
	c.ws.WriteJSON(serverData)
}

func (c *WSClient) handleStart(clientData *TransData) {
	serverData := TransData{}
	serverData.Type = clientData.Type
	serverData.DataId = clientData.DataId

	if c.documentId == "" {
		c.msgErr("not bind document", &serverData, nil)
		return
	}

	startdata := StartData{}
	err := json.Unmarshal([]byte(clientData.Data), &startdata)
	if err != nil {
		c.msgErr("start args error", &serverData, &err)
		return
	}

	lastCmdVersion := startdata.LastCmdVersion

	log.Println("LastCmdVersion", startdata.LastCmdVersion, lastCmdVersion)
	// bind comment
	commentServe := NewCommentServe(c.ws, c.userId, c.documentId, c.genSId)
	c.bindServe(DataTypes_Comment, commentServe)
	opServe := NewOpServe(c.ws, c.userId, c.documentId, c.versionId, lastCmdVersion, c.genSId) // todo VersionId
	c.bindServe(DataTypes_Op, opServe)
	resourceServe := NewResourceServe(c.ws, c.userId, c.documentId)
	c.bindServe(DataTypes_Resource, resourceServe)
	thumbnailServe := NewThumbnailServe(c.ws, c.userId, c.documentId)
	c.bindServe(DataTypes_Thumbnail, thumbnailServe)
	selectionServe := NewSelectionServe(c.ws, c.token, c.userId, c.documentId, c.genSId)
	c.bindServe(DataTypes_Selection, selectionServe)
	versionServe := NewVersionServe(c.ws, c.userId, c.documentId, c.genSId)
	c.bindServe(DataTypes_GenerateVersion, versionServe)

	c.ws.WriteJSON(serverData)
}

func (c *WSClient) Serve() {

	// doc upload
	docUploadServe := NewDocUploadServe(c.ws, c.userId)
	c.bindServe(DataTypes_DocUpload, docUploadServe)

	// close handlers
	defer (func() {
		for _, h := range c.serveMap {
			if nil != h {
				h.close()
			}
		}
	})()

	for {

		clientData := TransData{}
		serverData := TransData{}
		serverData.Type = clientData.Type
		serverData.DataId = clientData.DataId

		mt, bytes, err := c.ws.ReadMessage()
		// todo
		if err != nil {
			if errors.Is(err, websocket.ErrClosed) {
				log.Println("ws is closed", err)
				return
			}

			c.msgErr("read message err", &serverData, nil)
			continue
		}

		var binaryData *([]byte) = nil
		if mt == websocket.MessageTypeBinary {
			// get header
			h, binary, err := decodeBinaryMessage(bytes)
			if err != nil {
				c.msgErr("decode binary fail", &serverData, nil)
				continue
			}

			err1 := json.Unmarshal([]byte(h), &clientData)
			if err1 != nil {
				c.msgErr("wrong binary header", &serverData, nil)
				continue
			}

			binaryData = &binary
		} else if mt == websocket.MessageTypeText {
			err := json.Unmarshal(bytes, &clientData)
			if err != nil {
				c.msgErr("wrong data struct", &serverData, nil)
				continue
			}
		} else {
			c.msgErr("unknow message type", &serverData, nil)
			continue
		}

		log.Println("ws receive msg:", clientData.Type, mt)

		if clientData.Type == DataTypes_Heartbeat {
			_ = c.ws.WriteJSON(&serverData)
			continue
		}
		if clientData.Type == DataTypes_Bind {
			// permType := models.PermType(str.DefaultToInt(c.Query("perm_type"), 0))
			c.handleBind(&clientData)
			continue
		}
		if clientData.Type == DataTypes_Start {
			c.handleStart(&clientData)
			continue
		}
		serve := c.serveMap[clientData.Type]
		if serve != nil {
			serve.handle(&clientData, binaryData)
		} else {
			c.msgErr("no bind handler", &serverData, nil)
		}
	}
}
