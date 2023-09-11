package communication

import (
	"github.com/google/uuid"
	"log"
	"protodesign.cn/kcserver/common"
	"protodesign.cn/kcserver/utils/str"
	"protodesign.cn/kcserver/utils/websocket"
)

func OpenDocUploadTunnel(clientWs *websocket.Ws, clientCmdData CmdData, serverCmd ServerCmd, data Data) *Tunnel {
	clientCmdDataData, _ := clientCmdData["data"].(map[string]any)
	projectIdStr, _ := clientCmdDataData["project_id"].(string)
	userId, ok := data["userId"].(int64)
	if !ok || userId <= 0 {
		serverCmd.Message = "通道建立失败，参数错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document upload ws建立失败，参数错误", ok, userId)
		return nil
	}

	serverWs, err := websocket.NewClient("ws://"+common.DocumentServiceHost+common.ApiVersionPath+"/documents/document_upload", nil)
	if err != nil {
		serverCmd.Message = "通道建立失败"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document upload ws建立失败", err)
		return nil
	}
	sendToServerData := Data{
		"user_id": str.IntToString(userId),
	}
	if projectIdStr != "" {
		sendToServerData["project_id"] = projectIdStr
	}
	if err := serverWs.WriteJSON(sendToServerData); err != nil {
		serverCmd.Message = "通道建立失败"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document upload ws建立失败（鉴权）", err)
		return nil
	}

	tunnelId := uuid.New().String()
	tunnel := &Tunnel{
		Id:     tunnelId,
		Server: serverWs,
		Client: clientWs,
	}
	// 转发客户端数据到服务端
	tunnel.ReceiveFromClient = tunnel.DefaultClientToServer
	// 转发服务端数据到客户端
	go tunnel.DefaultServerToClient()

	return tunnel
}
