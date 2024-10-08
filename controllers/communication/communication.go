package communication

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/gin/response"
	"kcaitech.com/kcserver/common/jwt"
	"kcaitech.com/kcserver/utils/str"
	"kcaitech.com/kcserver/utils/websocket"
)

// func getTunnelDataByCmdData(tunnelCmdData TunnelCmdData, ws *websocket.Ws) (TunnelDataType, []byte) {
// 	dataType := tunnelCmdData.DataType
// 	if dataType < TunnelDataTypeText || dataType > TunnelDataTypeBinary {
// 		log.Println("dataType错误", dataType)
// 		return TunnelDataTypeNone, nil
// 	}
// 	if dataType != TunnelDataTypeBinary {
// 		if len(tunnelCmdData.Data) == 0 {
// 			log.Println("data错误")
// 			return TunnelDataTypeNone, nil
// 		}
// 		return dataType, tunnelCmdData.Data
// 	} else {
// 		messageType, data, err := ws.ReadMessage()
// 		if err != nil || messageType != websocket.MessageTypeBinary {
// 			log.Println("binary data错误", err, messageType)
// 			return TunnelDataTypeNone, nil
// 		}
// 		return dataType, data
// 	}
// }

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
	VersionId  string `json:"version_id"`
}

// Communication websocket连接
func Communication(c *gin.Context) {

	// get token
	token := jwt.GetJwtFromAuthorization(c.GetHeader("Authorization"))
	if token == "" {
		response.Abort(c, http.StatusUnauthorized, "未登录", nil)
		return
	}

	jwtParseData, err := jwt.ParseJwt(token)
	if err != nil {
		log.Println("Token错误", err)
		response.Abort(c, http.StatusForbidden, "Token错误", nil)
		return
	}
	userId := str.DefaultToInt(jwtParseData.Id, 0)
	if userId <= 0 {
		log.Println("UserId错误", userId)
		response.Abort(c, http.StatusForbidden, "UserId错误", nil)
		return
	}

	// 建立ws连接
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Fail(c, "建立ws连接失败："+err.Error())
		return
	}
	defer ws.Close()

	log.Println("websocket连接成功")

	var _sid atomic.Int64
	genSId := func() string {
		sid := _sid.Add(1)
		return "s" + str.IntToString(sid)
	}

	type serveFace interface {
		handle(data *TransData, binaryData *([]byte))
		close()
	}

	serveMap := map[string]serveFace{}

	msgErr := func(err string, serverData *TransData) {
		serverData.Err = err
		log.Println(err)
		_ = ws.WriteJSON(serverData)
	}

	bindServe := func(t string, s serveFace) {
		old := serveMap[t]
		if old != nil {
			(old).close()
		}
		serveMap[t] = s
	}

	// doc upload
	docUploadServe := NewDocUploadServe(ws, userId)
	bindServe(DataTypes_DocUpload, docUploadServe)

	for {

		clientData := TransData{}
		serverData := TransData{}
		serverData.Type = clientData.Type
		serverData.DataId = clientData.DataId

		mt, bytes, err := ws.ReadMessage()
		// todo
		if err != nil {
			if errors.Is(err, websocket.ErrClosed) {
				log.Println("ws is closed", err)
				return
			}

			msgErr("read message err", &serverData)
			continue
		}

		var binaryData *([]byte) = nil
		if mt == websocket.MessageTypeBinary {
			// get header
			h, binary, err := decodeBinaryMessage(bytes)
			if err != nil {
				msgErr("decode binary fail", &serverData)
				continue
			}

			err1 := json.Unmarshal([]byte(h), &clientData)
			if err1 != nil {
				msgErr("wrong binary header", &serverData)
				continue
			}

			binaryData = &binary
		} else if mt == websocket.MessageTypeText {
			err := json.Unmarshal(bytes, &clientData)
			if err != nil {
				msgErr("wrong data struct", &serverData)
				continue
			}
		} else {
			msgErr("unknow message type", &serverData)
			continue
		}

		if clientData.Type == DataTypes_Bind {
			bindData := BindData{}
			err := json.Unmarshal([]byte(clientData.Data), &bindData)
			if err != nil {
				msgErr("wrong bind struct", &serverData)
				continue
			}

			documentId := str.DefaultToInt(bindData.DocumentId, 0)
			if documentId <= 0 {
				msgErr("wrong document id", &serverData)
				continue
			}

			// bind comment
			commentServe := NewCommentServe(ws, userId, documentId, genSId)
			bindServe(DataTypes_Comment, commentServe)

			opServe := NewOpServe(ws, userId, documentId, bindData.VersionId, genSId) // todo VersionId
			bindServe(DataTypes_Op, opServe)

			resourceServe := NewResourceServe(ws, userId, documentId)
			bindServe(DataTypes_Resource, resourceServe)

			selectionServe := NewSelectionServe(ws, userId, documentId, genSId)
			bindServe(DataTypes_Selection, selectionServe)
			// send back message
			ws.WriteJSON(serverData)
			return

		} else {
			serve := serveMap[clientData.Type]
			if serve != nil {
				serve.handle(&clientData, binaryData)
			} else {
				msgErr("no bind handler", &serverData)
			}
		}

	}
}
