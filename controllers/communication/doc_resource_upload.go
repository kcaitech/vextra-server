package communication

import (
	"log"
	"sync"

	"github.com/google/uuid"
	"kcaitech.com/kcserver/utils/str"
	"kcaitech.com/kcserver/utils/websocket"
)

type docResTunnelServer struct {
	HandleClose func(code int, text string)
	IsClose     bool
	CloseLock   sync.Mutex
}

func (server *docResTunnelServer) SetCloseHandler(handler func(code int, text string)) {
	server.HandleClose = handler
}

func (server *docResTunnelServer) WriteMessage(messageType websocket.MessageType, data []byte) (err error) {
	return nil
}

func (server *docResTunnelServer) ReadMessage() (websocket.MessageType, []byte, error) {
	return websocket.MessageTypeNone, nil, websocket.ErrClosed
}

func (server *docResTunnelServer) Close() {
	server.CloseLock.Lock()
	defer server.CloseLock.Unlock()
	if server.IsClose {
		return
	}
	server.IsClose = true
	if server.HandleClose != nil {
		server.HandleClose(0, "")
	}
}

func OpenDocResourceUploadTunnel(clientWs *websocket.Ws, clientCmdData CmdData, serverCmd ServerCmd, data Data) *Tunnel {
	clientCmdDataData, ok := clientCmdData["data"].(map[string]any)
	documentIdStr, ok1 := clientCmdDataData["document_id"].(string)
	userId, ok2 := data["userId"].(int64)
	if !ok || !ok1 || documentIdStr == "" || !ok2 || userId <= 0 {
		serverCmd.Message = "通道建立失败，参数错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document resource upload ws建立失败，参数错误", ok, ok1, ok2, documentIdStr, userId)
		return nil
	}
	documentId := str.DefaultToInt(documentIdStr, 0)
	if documentId <= 0 {
		serverCmd.Message = "通道建立失败，documentId错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document resource upload ws建立失败，documentId错误", documentId)
		return nil
	}

	// serverWs, err := websocket.NewClient("ws://"+common.DocumentServiceHost+common.ApiVersionPath+"/documents/resource_upload", nil)
	// if err != nil {
	// 	serverCmd.Message = "通道建立失败"
	// 	_ = clientWs.WriteJSON(&serverCmd)
	// 	log.Println("document resource upload ws建立失败", err)
	// 	return nil
	// }
	// if err := serverWs.WriteJSON(Data{
	// 	"document_id": documentIdStr,
	// 	"user_id":     str.IntToString(userId),
	// }); err != nil {
	// 	serverCmd.Message = "通道建立失败"
	// 	_ = clientWs.WriteJSON(&serverCmd)
	// 	log.Println("document resource upload ws建立失败（鉴权）", err)
	// 	return nil
	// }

	tunnelServer := &docResTunnelServer{}

	tunnelId := uuid.New().String()
	tunnel := &Tunnel{
		Id:     tunnelId,
		Server: tunnelServer,
		Client: clientWs,
	}
	// 转发客户端数据到服务端
	tunnel.ReceiveFromClient = tunnel.DefaultClientToServer
	// 转发服务端数据到客户端
	go tunnel.DefaultServerToClient()

	return tunnel
}
