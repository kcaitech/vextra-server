package communication

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/utils/str"
	"protodesign.cn/kcserver/utils/websocket"
)

func getTunnelDataByCmdData(tunnelCmdData TunnelCmdData, ws *websocket.Ws) (TunnelDataType, []byte) {
	dataType := tunnelCmdData.DataType
	if dataType < TunnelDataTypeText || dataType > TunnelDataTypeBinary {
		log.Println("dataType错误", dataType)
		return TunnelDataTypeNone, nil
	}
	if dataType != TunnelDataTypeBinary {
		if len(tunnelCmdData.Data) == 0 {
			log.Println("data错误")
			return TunnelDataTypeNone, nil
		}
		return dataType, tunnelCmdData.Data
	} else {
		messageType, data, err := ws.ReadMessage()
		if err != nil || messageType != websocket.MessageTypeBinary || len(data) == 0 {
			log.Println("binary data错误", err)
			return TunnelDataTypeNone, nil
		}
		return dataType, data
	}
}

// Communication websocket连接
func Communication(c *gin.Context) {
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Fail(c, "建立ws连接失败："+err.Error())
		return
	}
	defer ws.Close()

	tunnelMap := make(map[string]*Tunnel)
	getTunnelIdByCmdData := func(cmdData CmdData) (string, bool) {
		v, ok := cmdData["tunnel_id"]
		tunnelId, ok1 := v.(string)
		if !ok || !ok1 || tunnelId == "" {
			return "", false
		}
		return tunnelId, true
	}
	getTunnelById := func(tunnelId string) *Tunnel {
		tunnel, ok := tunnelMap[tunnelId]
		if tunnelId == "" || !ok {
			log.Println("tunnel不存在", tunnelId)
			return nil
		}
		return tunnel
	}
	getTunnelByCmdData := func(cmdData CmdData) (string, *Tunnel) {
		tunnelId, ok := getTunnelIdByCmdData(cmdData)
		if !ok {
			log.Println("tunnelId错误")
			return "", nil
		}
		tunnel := getTunnelById(tunnelId)
		if tunnel == nil {
			return "", nil
		}
		return tunnelId, tunnel
	}
	getTunnelByTunnelCmdData := func(tunnelCmdData TunnelCmdData) (string, *Tunnel) {
		tunnelId := tunnelCmdData.TunnelId
		tunnel := getTunnelById(tunnelId)
		if tunnel == nil {
			return "", nil
		}
		return tunnelId, tunnel
	}

	ws.SetCloseHandler(func(code int, text string) {
		log.Println("ws连接关闭", code, text)
		for _, tunnelSession := range tunnelMap {
			tunnelSession.ServerWs.Close()
		}
	})

	header := Header{}
	cmdId := uuid.New().String()
	serverCmd := ServerCmd{
		CmdType: ServerCmdTypeInitResult,
		Status:  CmdStatusFail,
		CmdId:   cmdId,
	}
	if err := ws.ReadJSON(&header); err != nil {
		serverCmd.Message = "Header结构错误"
		_ = ws.WriteJSON(&serverCmd)
		log.Println("Header结构错误", err)
		return
	}
	jwtParseData, err := jwt.ParseJwt(header.Token)
	if err != nil {
		serverCmd.Message = "Token错误"
		_ = ws.WriteJSON(&serverCmd)
		log.Println("Token错误", err)
		return
	}
	userId := str.DefaultToInt(jwtParseData.Id, 0)
	if userId <= 0 {
		serverCmd.Message = "UserId错误"
		_ = ws.WriteJSON(&serverCmd)
		log.Println("UserId错误", userId)
		return
	}
	communicationId := uuid.New().String()
	serverCmd.Status = CmdStatusSuccess
	serverCmd.Data = CmdData{"communication_id": communicationId}
	_ = ws.WriteJSON(&serverCmd)
	if ws.IsClose() {
		return
	}
	log.Println("websocket连接成功，communicationId：", communicationId)

	for {
		clientCmd := ClientCmd{}
		serverCmdId := uuid.New().String()
		serverCmd := ServerCmd{
			CmdType: ServerCmdTypeReturn,
			Status:  CmdStatusFail,
			CmdId:   serverCmdId,
		}
		if err := ws.ReadJSON(&clientCmd); err != nil {
			if err == websocket.ErrClosed {
				for _, tunnelSession := range tunnelMap {
					tunnelSession.ServerWs.Close()
				}
				log.Println("ws连接关闭", err)
				return
			}
			serverCmd.Message = "cmdData结构错误"
			_ = ws.WriteJSON(&serverCmd)
			log.Println("cmdData结构错误", err)
			continue
		}
		serverCmd.Data = CmdData{"cmd_id": clientCmd.CmdId}
		clientCmdData := CmdData{}
		clientTunnelCmdData := TunnelCmdData{}
		var v any
		if clientCmd.CmdType != ClientCmdTypeTunnelData {
			v = &clientCmdData
		} else {
			v = &clientTunnelCmdData
		}
		if len(clientCmd.Data) == 0 {
			clientCmd.Data = []byte("{}")
		}
		if err := json.Unmarshal(clientCmd.Data, v); err != nil {
			serverCmd.Message = "clientCmd.Data结构错误"
			_ = ws.WriteJSON(&serverCmd)
			log.Println("clientCmd.Data结构错误", err)
			continue
		}
		switch clientCmd.CmdType {
		case ClientCmdTypeReturn:
			v, ok := clientCmdData["cmd_id"]
			cmdId, ok1 := v.(string)
			if !ok || !ok1 || cmdId == "" {
				log.Println("cmdId错误", ok, ok1, cmdId)
				continue
			}
		case ClientCmdTypeOpenTunnel:
			var tunnel *Tunnel
			switch clientCmd.TunnelType {
			case TunnelTypeDocOp:
				tunnel = OpenDocOpTunnel(ws, clientCmdData, serverCmd, Data{"userId": userId})
			case TunnelTypeDocResourceUpload:
				tunnel = OpenDocResourceUploadTunnel(ws, clientCmdData, serverCmd, Data{"userId": userId})
			default:
				serverCmd.Message = "tunnel_type错误"
				log.Println("clientCmd.TunnelType错误", clientCmd.TunnelType)
			}
			if tunnel != nil {
				serverCmd.Status = CmdStatusSuccess
				serverCmd.Data["tunnel_id"] = tunnel.Id
			}
			_ = ws.WriteJSON(&serverCmd)
			if tunnel == nil {
				continue
			}
			tunnelId := tunnel.Id
			tunnelMap[tunnelId] = tunnel
			// todo 关闭前的处理
			tunnel.ServerWs.SetCloseHandler(func(code int, text string) {
				delete(tunnelMap, tunnelId)
				_ = ws.WriteJSON(&ServerCmd{
					CmdType: ServerCmdTypeCloseTunnel,
					CmdId:   uuid.New().String(),
					Data:    CmdData{"tunnel_id": tunnelId},
				})
				log.Println("document ws连接关闭", tunnelId)
			})
		case ClientCmdTypeCloseTunnel:
			tunnelId, tunnel := getTunnelByCmdData(clientCmdData)
			if tunnel == nil {
				serverCmd.Message = "tunnel_id错误"
				_ = ws.WriteJSON(&serverCmd)
				continue
			}
			tunnel.ServerWs.Close()
			delete(tunnelMap, tunnelId)
			serverCmd.Status = CmdStatusSuccess
			_ = ws.WriteJSON(&serverCmd)
		case ClientCmdTypeTunnelData:
			tunnelId, tunnel := getTunnelByTunnelCmdData(clientTunnelCmdData)
			serverCmd.Data["tunnel_id"] = tunnelId
			if tunnel == nil {
				serverCmd.Message = "tunnel_id错误"
				_ = ws.WriteJSON(&serverCmd)
				continue
			}
			tunnelDataType, data := getTunnelDataByCmdData(clientTunnelCmdData, ws)
			if len(data) == 0 {
				serverCmd.Message = "数据错误"
				_ = ws.WriteJSON(&serverCmd)
				continue
			}
			tunnel.ReceiveFromClient(tunnelDataType, data, serverCmd)
		default:
			serverCmd.Message = "cmd_type错误"
			_ = ws.WriteJSON(&serverCmd)
			log.Println("clientCmd.CmdType错误", clientCmd.CmdType)
		}
	}
}
