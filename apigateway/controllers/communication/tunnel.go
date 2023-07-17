package communication

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"protodesign.cn/kcserver/utils/websocket"
)

type Tunnel struct {
	Id                string
	ServerWs          *websocket.Ws
	ClientWs          *websocket.Ws
	ReceiveFromClient func(tunnelDataType TunnelDataType, data []byte, serverCmd ServerCmd)
}

func (tunnel *Tunnel) DefaultServerToClient() {
	needUnlock := false
	defer func() {
		tunnel.ServerWs.Close()
		log.Println("document ws服务端关闭", tunnel.Id)
		if needUnlock {
			tunnel.ClientWs.Unlock()
		}
	}()
	for {
		tunnelType, data, err := tunnel.ServerWs.ReadMessage()
		if err != nil {
			log.Println("document ws服务端数据读取失败", err)
			return
		}
		var tunnelData any = nil
		if tunnelType == websocket.MessageTypeBinary {
			tunnel.ClientWs.Lock()
			needUnlock = true
		} else {
			tunnelData = &CmdData{}
			if err := json.Unmarshal(data, tunnelData); err != nil {
				log.Println("document ws服务端数据解析失败", err)
				return
			}
		}
		if err := tunnel.ClientWs.WriteJSONLock(tunnelType != websocket.MessageTypeBinary, &ServerCmd{
			CmdType: ServerCmdTypeTunnelData,
			CmdId:   uuid.New().String(),
			Data:    CmdData{"tunnel_id": tunnel.Id, "data_type": tunnelType, "data": tunnelData},
		}); err != nil {
			log.Println("document ws客户端数据写入失败", err)
			return
		}
		if tunnelType == websocket.MessageTypeBinary {
			if err := tunnel.ClientWs.WriteMessageLock(false, websocket.MessageTypeBinary, data); err != nil {
				log.Println("document ws客户端数据写入失败", err)
				return
			}
			tunnel.ClientWs.Unlock()
			needUnlock = false
		}
	}
}

func (tunnel *Tunnel) DefaultClientToServer(tunnelDataType TunnelDataType, data []byte, serverCmd ServerCmd) {
	err := tunnel.ServerWs.WriteMessage(websocket.MessageType(tunnelDataType), data)
	if err != nil {
		serverCmd.Message = "数据发送失败"
		_ = tunnel.ClientWs.WriteJSON(&serverCmd)
		log.Println("数据发送失败", err)
		return
	}
	serverCmd.Status = CmdStatusSuccess
	_ = tunnel.ClientWs.WriteJSON(&serverCmd)
}
