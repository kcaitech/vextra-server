package communication

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"kcaitech.com/kcserver/common"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/redis"
	"kcaitech.com/kcserver/common/services"
	"kcaitech.com/kcserver/utils/str"
	"kcaitech.com/kcserver/utils/websocket"
	"sync"
	"time"
)

type docSelectionTunnelServer struct {
	HandleClose  func(code int, text string)
	IsClose      bool
	CloseLock    sync.Mutex
	ToClientChan chan []byte
	ToServerChan chan []byte
}

func (server *docSelectionTunnelServer) SetCloseHandler(handler func(code int, text string)) {
	server.HandleClose = handler
}

func (server *docSelectionTunnelServer) WriteMessage(messageType websocket.MessageType, data []byte) (err error) {
	if server.IsClose {
		return websocket.ErrClosed
	}
	defer func() {
		if err1 := recover(); err1 != nil {
			log.Println("通道写入时panic", err1)
			err = websocket.ErrClosed
		}
	}()
	server.ToServerChan <- data
	return nil
}

func (server *docSelectionTunnelServer) ReadMessage() (websocket.MessageType, []byte, error) {
	if server.IsClose {
		return websocket.MessageTypeNone, nil, websocket.ErrClosed
	}
	data, ok := <-server.ToClientChan
	if !ok {
		return websocket.MessageTypeNone, nil, websocket.ErrClosed
	}
	return websocket.MessageTypeText, data, nil
}

func (server *docSelectionTunnelServer) Close() {
	server.CloseLock.Lock()
	defer server.CloseLock.Unlock()
	if server.IsClose {
		return
	}
	server.IsClose = true
	close(server.ToClientChan)
	close(server.ToServerChan)
	if server.HandleClose != nil {
		server.HandleClose(0, "")
	}
}

type DocSelectionData struct {
	SelectPageId      string          `json:"select_page_id,omitempty"`
	SelectShapeIdList []string        `json:"select_shape_id_list"`
	HoverShapeId      string          `json:"hover_shape_id,omitempty"`
	CursorStart       int             `json:"cursor_start,omitempty"`
	CursorEnd         int             `json:"cursor_end,omitempty"`
	CursorAtBefore    bool            `json:"cursor_at_before,omitempty"`
	UserId            string          `json:"user_id,omitempty"`
	Permission        models.PermType `json:"permission,omitempty"`
	Avatar            string          `json:"avatar,omitempty"`
	Nickname          string          `json:"nickname,omitempty"`
	EnterTime         int64           `json:"enter_time,omitempty"`
}

type DocSelectionOpType uint8

const (
	DocSelectionOpTypeUpdate DocSelectionOpType = iota
	DocSelectionOpTypeExit
)

type DocSelectionOpData struct {
	Type   DocSelectionOpType `json:"type"`
	UserId string             `json:"user_id"`
	Data   *DocSelectionData  `json:"data,omitempty"`
}

func OpenDocSelectionOpTunnel(clientWs *websocket.Ws, clientCmdData CmdData, serverCmd ServerCmd, data Data) *Tunnel {
	clientCmdDataData, ok := clientCmdData["data"].(map[string]any)
	documentIdStr, ok1 := clientCmdDataData["document_id"].(string)
	userId, ok2 := data["userId"].(int64)
	if !ok || !ok1 || documentIdStr == "" || !ok2 || userId <= 0 {
		serverCmd.Message = "通道建立失败，参数错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document comment ws建立失败，参数错误", ok, ok1, ok2, documentIdStr, userId)
		return nil
	}
	userIdStr := str.IntToString(userId)
	userService := services.NewUserService()
	user := models.User{}
	if err := userService.GetById(userId, &user); err != nil {
		serverCmd.Message = "通道建立失败，用户信息错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document comment ws建立失败，用户不存在", err, userId)
		return nil
	}
	documentId := str.DefaultToInt(documentIdStr, 0)
	if documentId <= 0 {
		serverCmd.Message = "通道建立失败，documentId错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document comment ws建立失败，documentId错误", documentId)
		return nil
	}
	// 权限校验
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeReadOnly {
		serverCmd.Message = "通道建立失败"
		if err != nil {
			serverCmd.Message += "，无权限"
		}
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，权限校验错误", err, permType)
		return nil
	}

	enterTime := time.Now().UnixNano() / 1000000

	tunnelId := uuid.New().String()
	tunnelServer := &docSelectionTunnelServer{
		ToClientChan: make(chan []byte),
		ToServerChan: make(chan []byte),
	}
	tunnel := &Tunnel{
		Id:     tunnelId,
		Server: tunnelServer,
		Client: clientWs,
	}
	// 转发客户端数据到服务端
	tunnel.ReceiveFromClient = tunnel.DefaultClientToServer
	// 转发服务端数据到客户端
	go tunnel.DefaultServerToClient()

	go func() {
		defer tunnelServer.Close()
		// 获取文档当前所有用户的选区数据
		result, err := redis.Client.HGetAll(context.Background(), "Document Selection Data[DocumentId:"+documentIdStr+"]").Result()
		if err != nil {
			log.Println("获取文档选区数据失败", err)
			return
		}
		for userIdStr, data := range result {
			selectionData := &DocSelectionData{}
			if err := json.Unmarshal([]byte(data), selectionData); err != nil {
				log.Println("获取文档选区数据失败：数据解码错误", err)
				return
			}
			docSelectionOpData := &DocSelectionOpData{
				Type:   DocSelectionOpTypeUpdate,
				UserId: userIdStr,
				Data:   selectionData,
			}
			if data, err := json.Marshal(docSelectionOpData); err == nil {
				if tunnelServer.IsClose {
					return
				}
				tunnelServer.ToClientChan <- data
			}
		}
		// 首次发送当前用户的选区数据
		selectionData := &DocSelectionData{
			SelectShapeIdList: []string{},
			UserId:            userIdStr,
			Permission:        permType,
			Avatar:            common.FileStorageHost + user.Avatar,
			Nickname:          user.Nickname,
			EnterTime:         enterTime,
		}
		selectionDataJson, _ := json.Marshal(selectionData)
		docSelectionOpData := &DocSelectionOpData{
			Type:   DocSelectionOpTypeUpdate,
			UserId: userIdStr,
			Data:   selectionData,
		}
		// todo 使用Document Selection Data[DocumentId:{documentId}][UserId:{UserId}]单独存储用户的选区数据，并设置有效期，配合Keys、MGet即可获取文档中所有用户的选区数据
		if docSelectionOpDataJson, err := json.Marshal(docSelectionOpData); err == nil {
			redis.Client.HSet(context.Background(), "Document Selection Data[DocumentId:"+documentIdStr+"]", userIdStr, string(selectionDataJson))
			redis.Client.Expire(context.Background(), "Document Selection Data[DocumentId:"+documentIdStr+"]", time.Hour*1)
			redis.Client.Publish(context.Background(), "Document Selection[DocumentId:"+documentIdStr+"]", string(docSelectionOpDataJson))
		}
		// 开始转发
		subscribe := redis.Client.Subscribe(context.Background(), "Document Selection[DocumentId:"+documentIdStr+"]")
		defer subscribe.Close()
		subscribeChan := subscribe.Channel()
		for {
			select {
			case data, ok := <-tunnelServer.ToServerChan:
				if !ok { // 通道已关闭
					docSelectionOpData := &DocSelectionOpData{
						Type:   DocSelectionOpTypeExit,
						UserId: userIdStr,
					}
					if data, err := json.Marshal(docSelectionOpData); err == nil {
						redis.Client.HDel(context.Background(), "Document Selection Data[DocumentId:"+documentIdStr+"]", userIdStr)
						redis.Client.Publish(context.Background(), "Document Selection[DocumentId:"+documentIdStr+"]", string(data))
					}
					return
				}
				selectionData := &DocSelectionData{}
				if err := json.Unmarshal(data, selectionData); err != nil {
					log.Println("document selection数据解码错误", err)
					continue
				}
				selectionData.UserId = userIdStr
				selectionData.Permission = permType
				selectionData.Avatar = common.FileStorageHost + user.Avatar
				selectionData.Nickname = user.Nickname
				selectionData.EnterTime = enterTime
				selectionDataJson, _ := json.Marshal(selectionData)
				docSelectionOpData := &DocSelectionOpData{
					Type:   DocSelectionOpTypeUpdate,
					UserId: userIdStr,
					Data:   selectionData,
				}
				if docSelectionOpDataJson, err := json.Marshal(docSelectionOpData); err != nil {
					log.Println("document selection数据编码错误", err)
					continue
				} else {
					redis.Client.HSet(context.Background(), "Document Selection Data[DocumentId:"+documentIdStr+"]", userIdStr, string(selectionDataJson))
					redis.Client.Expire(context.Background(), "Document Selection Data[DocumentId:"+documentIdStr+"]", time.Hour*1)
					redis.Client.Publish(context.Background(), "Document Selection[DocumentId:"+documentIdStr+"]", string(docSelectionOpDataJson))
				}
			case data, ok := <-subscribeChan:
				if !ok { // 通道已关闭
					docSelectionOpData := &DocSelectionOpData{
						Type:   DocSelectionOpTypeExit,
						UserId: userIdStr,
					}
					if data, err := json.Marshal(docSelectionOpData); err == nil {
						redis.Client.Publish(context.Background(), "Document Selection[DocumentId:"+documentIdStr+"]", string(data))
					}
					return
				}
				if tunnelServer.IsClose {
					return
				}
				tunnelServer.ToClientChan <- []byte(data.Payload)
			}
		}
	}()

	return tunnel
}
