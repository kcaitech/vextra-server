package communication

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"kcaitech.com/kcserver/common/gin/response"
	"kcaitech.com/kcserver/common/jwt"
	"kcaitech.com/kcserver/utils/my_map"
	"kcaitech.com/kcserver/utils/str"
	"kcaitech.com/kcserver/utils/websocket"
	"time"
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
		if err != nil || messageType != websocket.MessageTypeBinary {
			log.Println("binary data错误", err, messageType)
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

	tunnelMap := my_map.NewSyncMap[string, *Tunnel]()
	getTunnelIdByCmdData := func(cmdData CmdData) (string, bool) {
		v, ok := cmdData["tunnel_id"]
		tunnelId, ok1 := v.(string)
		if !ok || !ok1 || tunnelId == "" {
			return "", false
		}
		return tunnelId, true
	}
	getTunnelById := func(tunnelId string) *Tunnel {
		tunnel, ok := tunnelMap.Get(tunnelId)
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
			return tunnelId, nil
		}
		return tunnelId, tunnel
	}

	ws.SetCloseHandler(func(code int, text string) {
		log.Println("ws连接关闭", code, text)
		tunnelMap.Range(func(_ string, tunnelSession *Tunnel) bool {
			tunnelSession.Server.Close()
			return true
		})
	})

	header := Header{}
	cmdId := uuid.New().String()
	serverCmd := ServerCmd{
		CmdType: ServerCmdTypeInitResult,
		Status:  CmdStatusFail,
		CmdId:   cmdId,
	}
	if err := ws.ReadJSON(&header); err != nil {
		log.Println("Header结构错误", err)
		serverCmd.Message = "Header结构错误"
		_ = ws.WriteJSON(&serverCmd)
		return
	}
	jwtParseData, err := jwt.ParseJwt(header.Token)
	if err != nil {
		log.Println("Token错误", err)
		serverCmd.Message = "Token错误"
		_ = ws.WriteJSON(&serverCmd)
		return
	}
	userId := str.DefaultToInt(jwtParseData.Id, 0)
	if userId <= 0 {
		log.Println("UserId错误", userId)
		serverCmd.Message = "UserId错误"
		_ = ws.WriteJSON(&serverCmd)
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

	lastReceiveHeartbeatTime := time.Now()
	go func() {
		for {
			time.Sleep(time.Second * 1)
			if ws.IsClose() {
				return
			}
			if time.Now().Sub(lastReceiveHeartbeatTime).Seconds() > 60 {
				log.Println("心跳超时，断开连接")
				ws.Close()
				return
			}
			// 1秒发送一次心跳
			_ = ws.WriteJSON(&ServerCmd{
				CmdType: ServerCmdTypeHeartbeat,
				Status:  CmdStatusSuccess,
				CmdId:   uuid.New().String(),
			})
		}
	}()

	for {
		clientCmd := ClientCmd{}
		serverCmdId := uuid.New().String()
		serverCmd := ServerCmd{
			CmdType: ServerCmdTypeReturn,
			Status:  CmdStatusFail,
			CmdId:   serverCmdId,
		}
		if err := ws.ReadJSON(&clientCmd); err != nil {
			if errors.Is(err, websocket.ErrClosed) {
				log.Println("ws连接关闭", err)
				tunnelMap.Range(func(_ string, tunnelSession *Tunnel) bool {
					tunnelSession.Server.Close()
					return true
				})
				return
			}
			serverCmd.Message = "cmdData结构错误"
			log.Println("cmdData结构错误", err)
			_ = ws.WriteJSON(&serverCmd)
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
			log.Println("clientCmd.Data结构错误", err)
			_ = ws.WriteJSON(&serverCmd)
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
			if clientCmd.Status == CmdStatusFail && clientCmd.Message == CmdMessageTunnelIdError {
				tunnelId, tunnel := getTunnelByCmdData(clientCmdData)
				if tunnel != nil {
					tunnel.Server.Close()
				}
				tunnelMap.Delete(tunnelId)
			}
		case ClientCmdTypeOpenTunnel:
			var tunnel *Tunnel
			switch clientCmd.TunnelType {
			case TunnelTypeDocOp:
				tunnel = OpenDocOpTunnel(ws, clientCmdData, serverCmd, Data{"userId": userId})
			// case TunnelTypeDocResourceUpload:
			// 	tunnel = OpenDocResourceUploadTunnel(ws, clientCmdData, serverCmd, Data{"userId": userId})
			case TunnelTypeDocCommentOp:
				tunnel = OpenDocCommentOpTunnel(ws, clientCmdData, serverCmd, Data{"userId": userId})
			// case TunnelTypeDocUpload:
			// 	tunnel = OpenDocUploadTunnel(ws, clientCmdData, serverCmd, Data{"userId": userId})
			case TunnelTypeDocSelectionOp:
				tunnel = OpenDocSelectionOpTunnel(ws, clientCmdData, serverCmd, Data{"userId": userId})
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
			tunnelMap.Set(tunnelId, tunnel)
			// todo 关闭前的处理
			tunnel.Server.SetCloseHandler(func(code int, text string) {
				log.Println("document ws连接关闭", tunnelId)
				tunnelMap.Delete(tunnelId)
				_ = ws.WriteJSON(&ServerCmd{
					CmdType: ServerCmdTypeCloseTunnel,
					CmdId:   uuid.New().String(),
					Data:    CmdData{"tunnel_id": tunnelId},
				})
			})
		case ClientCmdTypeCloseTunnel:
			tunnelId, tunnel := getTunnelByCmdData(clientCmdData)
			if tunnel == nil {
				serverCmd.Message = CmdMessageTunnelIdError
				_ = ws.WriteJSON(&serverCmd)
				continue
			}
			tunnel.Server.Close()
			tunnelMap.Delete(tunnelId)
			serverCmd.Status = CmdStatusSuccess
			_ = ws.WriteJSON(&serverCmd)
		case ClientCmdTypeTunnelData:
			tunnelId, tunnel := getTunnelByTunnelCmdData(clientTunnelCmdData)
			serverCmd.Data["tunnel_id"] = tunnelId
			if tunnel == nil {
				serverCmd.Message = CmdMessageTunnelIdError
				_ = ws.WriteJSON(&serverCmd)
				continue
			}
			tunnelDataType, data := getTunnelDataByCmdData(clientTunnelCmdData, ws)
			if tunnelDataType == TunnelDataTypeNone {
				serverCmd.Message = "数据错误"
				_ = ws.WriteJSON(&serverCmd)
				continue
			}
			if err := tunnel.ReceiveFromClient(tunnelDataType, data, serverCmd); err != nil {
				serverCmd.Message = err.Error()
				log.Println("ReceiveFromClient error:", err)
				_ = ws.WriteJSON(&serverCmd)
				continue
			} else {
				serverCmd.Status = CmdStatusSuccess
				_ = ws.WriteJSON(&serverCmd)
			}
		case ClientCmdTypeHeartbeat:
			serverCmd.Status = CmdStatusSuccess
			serverCmd.CmdType = ServerCmdTypeHeartbeatResponse
			serverCmd.Data = CmdData{"cmd_id": clientCmd.CmdId}
			_ = ws.WriteJSON(&serverCmd)
		case ClientCmdTypeHeartbeatResponse:
			lastReceiveHeartbeatTime = time.Now()
		default:
			serverCmd.Message = "cmd_type错误"
			log.Println("clientCmd.CmdType错误", clientCmd.CmdType)
			_ = ws.WriteJSON(&serverCmd)
		}
	}
}
