package communication

import (
	"context"
	"github.com/google/uuid"
	"log"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/redis"
	"kcaitech.com/kcserver/common/services"
	"kcaitech.com/kcserver/utils/str"
	"kcaitech.com/kcserver/utils/websocket"
	"sync"
)

type docCommentTunnelServer struct {
	HandleClose  func(code int, text string)
	IsClose      bool
	CloseLock    sync.Mutex
	ToClientChan chan []byte
}

func (server *docCommentTunnelServer) SetCloseHandler(handler func(code int, text string)) {
	server.HandleClose = handler
}

func (server *docCommentTunnelServer) WriteMessage(messageType websocket.MessageType, data []byte) error {
	return nil
}

func (server *docCommentTunnelServer) ReadMessage() (websocket.MessageType, []byte, error) {
	if server.IsClose {
		return websocket.MessageTypeNone, nil, websocket.ErrClosed
	}
	data, ok := <-server.ToClientChan
	if !ok {
		return websocket.MessageTypeNone, nil, websocket.ErrClosed
	}
	return websocket.MessageTypeText, data, nil
}

func (server *docCommentTunnelServer) Close() {
	server.CloseLock.Lock()
	defer server.CloseLock.Unlock()
	if server.IsClose {
		return
	}
	server.IsClose = true
	close(server.ToClientChan)
	if server.HandleClose != nil {
		server.HandleClose(0, "")
	}
}

func OpenDocCommentOpTunnel(clientWs *websocket.Ws, clientCmdData CmdData, serverCmd ServerCmd, data Data) *Tunnel {
	clientCmdDataData, ok := clientCmdData["data"].(map[string]any)
	documentIdStr, ok1 := clientCmdDataData["document_id"].(string)
	userId, ok2 := data["userId"].(int64)
	if !ok || !ok1 || documentIdStr == "" || !ok2 || userId <= 0 {
		serverCmd.Message = "通道建立失败，参数错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document comment ws建立失败，参数错误", ok, ok1, ok2, documentIdStr, userId)
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

	tunnelId := uuid.New().String()
	tunnelServer := &docCommentTunnelServer{
		ToClientChan: make(chan []byte),
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
		pubsub := redis.Client.Subscribe(context.Background(), "Document Comment[DocumentId:"+documentIdStr+"]")
		defer pubsub.Close()
		for v := range pubsub.Channel() {
			if tunnelServer.IsClose {
				break
			}
			tunnelServer.ToClientChan <- []byte(v.Payload)
		}
	}()

	return tunnel
}
