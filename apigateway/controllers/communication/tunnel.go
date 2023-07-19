package communication

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log"
	"protodesign.cn/kcserver/utils/websocket"
)

type TunnelServer interface {
	SetCloseHandler(handler func(code int, text string))
	WriteMessage(messageType websocket.MessageType, data []byte) error
	ReadMessage() (websocket.MessageType, []byte, error)
	Close()
}

type TunnelClient interface {
	WriteMessageLock(needLock bool, messageType websocket.MessageType, data []byte) error
	WriteJSONLock(needLock bool, v any) error
	Lock()
	Unlock()
}

type Tunnel struct {
	Id                string
	Server            TunnelServer
	Client            TunnelClient
	ReceiveFromClient func(tunnelDataType TunnelDataType, data []byte, serverCmd ServerCmd) error
}

func (tunnel *Tunnel) DefaultServerToClient() {
	needUnlock := false
	defer func() {
		tunnel.Server.Close()
		log.Println("document ws服务端关闭", tunnel.Id)
		if needUnlock {
			tunnel.Client.Unlock()
		}
	}()
	for {
		tunnelType, data, err := tunnel.Server.ReadMessage()
		if err != nil {
			log.Println("document ws服务端数据读取失败", err)
			return
		}
		var tunnelData any = nil
		if tunnelType == websocket.MessageTypeBinary {
			tunnel.Client.Lock()
			needUnlock = true
		} else {
			tunnelData = &CmdData{}
			if err := json.Unmarshal(data, tunnelData); err != nil {
				log.Println("document ws服务端数据解析失败", err)
				return
			}
		}
		if err := tunnel.Client.WriteJSONLock(tunnelType != websocket.MessageTypeBinary, &ServerCmd{
			CmdType: ServerCmdTypeTunnelData,
			CmdId:   uuid.New().String(),
			Data:    CmdData{"tunnel_id": tunnel.Id, "data_type": tunnelType, "data": tunnelData},
		}); err != nil {
			log.Println("document ws客户端数据写入失败", err)
			return
		}
		if tunnelType == websocket.MessageTypeBinary {
			if err := tunnel.Client.WriteMessageLock(false, websocket.MessageTypeBinary, data); err != nil {
				log.Println("document ws客户端数据写入失败", err)
				return
			}
			tunnel.Client.Unlock()
			needUnlock = false
		}
	}
}

func (tunnel *Tunnel) DefaultClientToServer(tunnelDataType TunnelDataType, data []byte, serverCmd ServerCmd) error {
	if tunnel.Server.WriteMessage(websocket.MessageType(tunnelDataType), data) != nil {
		return errors.New("数据发送失败")
	}
	return nil
}
