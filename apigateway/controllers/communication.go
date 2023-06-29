package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"protodesign.cn/kcserver/common/gin/response"
	. "protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/utils/str"
	"protodesign.cn/kcserver/utils/websocket"
)

type headerType struct {
	Token string `json:"token"`
}

type clientCmdTypeType uint8

const (
	clientCmdTypeReturn      clientCmdTypeType = iota // 返回cmd执行结果
	clientCmdTypeOpenTunnel                           // 打开一条虚拟通道
	clientCmdTypeCloseTunnel                          // 关闭一条虚拟通道
	clientCmdTypeTunnelData                           // 虚拟通道数据
)

// tunnelTypeType 虚拟通道类型
type tunnelTypeType uint8

const (
	tunnelTypeDocOp tunnelTypeType = iota // 文档操作
)

type cmdDataType map[string]any

type clientCmdType struct {
	CmdType    clientCmdTypeType `json:"cmd_type"`
	CmdId      string            `json:"cmd_id"`
	TunnelType tunnelTypeType    `json:"tunnel_type,omitempty"`
	Status     cmdStatusType     `json:"status,omitempty"`
	Message    string            `json:"message,omitempty"`
	Data       json.RawMessage   `json:"data,omitempty"`
}

type tunnelDataTypeType uint8

const (
	tunnelDataTypeNone   = tunnelDataTypeType(0)
	tunnelDataTypeText   = tunnelDataTypeType(websocket.MessageTypeText)
	tunnelDataTypeBinary = tunnelDataTypeType(websocket.MessageTypeBinary)
)

type tunnelCmdDataType struct {
	DataType tunnelDataTypeType `json:"data_type"`
	Data     json.RawMessage    `json:"data"`
}

type serverCmdTypeType uint8

const (
	serverCmdTypeInitResult  serverCmdTypeType = iota // 返回初始化结果
	serverCmdTypeReturn                               // 返回cmd执行结果
	serverCmdTypeCloseTunnel                          // 关闭一条虚拟通道
	serverCmdTypeTunnelData                           // 虚拟通道数据
)

type cmdStatusType string

const (
	cmdStatusTypeSuccess cmdStatusType = "success"
	cmdStatusTypeFail    cmdStatusType = "fail"
)

type serverCmdType struct {
	CmdType serverCmdTypeType `json:"cmd_type"`
	CmdId   string            `json:"cmd_id"`
	Status  cmdStatusType     `json:"status,omitempty"`
	Message string            `json:"message,omitempty"`
	Data    cmdDataType       `json:"data,omitempty"`
}

type tunnelType struct {
	Id       string
	ServerWs *websocket.Ws
}

func getTunnelDataByCmdData(cmdData tunnelCmdDataType, ws *websocket.Ws) (tunnelDataTypeType, []byte) {
	dataType := cmdData.DataType
	if dataType < tunnelDataTypeText || dataType > tunnelDataTypeBinary {
		log.Println("dataType错误", dataType)
		return tunnelDataTypeNone, nil
	}
	if dataType != tunnelDataTypeBinary {
		if len(cmdData.Data) == 0 {
			log.Println("data错误")
			return tunnelDataTypeNone, nil
		}
		return dataType, cmdData.Data
	} else {
		messageType, data, err := ws.ReadMessage()
		if err != nil || messageType != websocket.MessageTypeBinary || len(data) == 0 {
			log.Println("binary data错误", err)
			return tunnelDataTypeNone, nil
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

	tunnelMap := make(map[string]*tunnelType)
	getTunnelByCmdData := func(cmdData cmdDataType) (string, *tunnelType) {
		v, ok := cmdData["tunnel_id"]
		tunnelId, ok1 := v.(string)
		if !ok || !ok1 || tunnelId == "" {
			log.Println("tunnelId错误", ok, ok1, tunnelId)
			return "", nil
		}
		tunnel, ok := tunnelMap[tunnelId]
		if !ok {
			log.Println("tunnel不存在", tunnelId)
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

	header := headerType{}
	cmdId := uuid.New().String()
	serverCmd := serverCmdType{
		CmdType: serverCmdTypeInitResult,
		Status:  cmdStatusTypeFail,
		CmdId:   cmdId,
	}
	if err := ws.ReadJSON(&header); err != nil {
		serverCmd.Message = "Header结构错误"
		_ = ws.WriteJSON(&serverCmd)
		log.Println("Header结构错误", err)
		return
	}
	jwtParseData, err := ParseJwt(header.Token)
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
	serverCmd.Status = cmdStatusTypeSuccess
	serverCmd.Data = cmdDataType{"communication_id": communicationId}
	_ = ws.WriteJSON(&serverCmd)
	if ws.IsClose() {
		return
	}
	log.Println("websocket连接成功，communicationId：", communicationId)

	for {
		cmdData := clientCmdType{}
		cmdId := uuid.New().String()
		serverCmd := serverCmdType{
			CmdType: serverCmdTypeReturn,
			Status:  cmdStatusTypeFail,
			CmdId:   cmdId,
		}
		if err := ws.ReadJSON(&cmdData); err != nil {
			serverCmd.Message = "cmdData结构错误"
			_ = ws.WriteJSON(&serverCmd)
			log.Println("cmdData结构错误", err)
			continue
		}
		serverCmd.Data = cmdDataType{"cmd_id": cmdData.CmdId}
		cmdDataData := cmdDataType{}
		tunnelCmdDataData := tunnelCmdDataType{}
		var v any
		if cmdData.CmdType != clientCmdTypeTunnelData {
			v = &cmdDataData
		} else {
			v = &tunnelCmdDataData
		}
		if err := json.Unmarshal(cmdData.Data, v); err != nil {
			serverCmd.Message = "cmdData结构错误"
			_ = ws.WriteJSON(&serverCmd)
			log.Println("cmdData.Data结构错误", err)
			continue
		}
		switch cmdData.CmdType {
		case clientCmdTypeReturn:
			v, ok := cmdDataData["cmd_id"]
			cmdId, ok1 := v.(string)
			if !ok || !ok1 || cmdId == "" {
				log.Println("cmdId错误", ok, ok1, cmdId)
				continue
			}
		case clientCmdTypeOpenTunnel:
			switch cmdData.TunnelType {
			case tunnelTypeDocOp:
				serverWs, err := websocket.NewClient("ws://192.168.0.18:10010", nil)
				if err != nil {
					serverCmd.Message = "通道建立失败"
					_ = ws.WriteJSON(&serverCmd)
					log.Println("document ws建立失败", err)
				}
				tunnelId := uuid.New().String()
				tunnelMap[tunnelId] = &tunnelType{
					Id:       tunnelId,
					ServerWs: serverWs,
				}
				// todo 关闭前的处理
				serverWs.SetCloseHandler(func(code int, text string) {
					delete(tunnelMap, tunnelId)
					_ = ws.WriteJSON(&serverCmdType{
						CmdType: serverCmdTypeCloseTunnel,
						CmdId:   uuid.New().String(),
						Data:    cmdDataType{"tunnel_id": tunnelId},
					})
					log.Println("document ws连接关闭", tunnelId)
				})
				go func(tunnelId string) {
					needUnlock := false
					defer func() {
						serverWs.Close()
						log.Println("document ws服务端关闭", tunnelId)
						if needUnlock {
							ws.Unlock()
						}
					}()
					for {
						tunnelType, data, err := serverWs.ReadMessage()
						if err != nil {
							log.Println("document ws服务端数据读取失败", err)
							return
						}
						var tunnelData any = nil
						if tunnelType == websocket.MessageTypeBinary {
							ws.Lock()
							needUnlock = true
						} else {
							tunnelData = &cmdDataType{}
							if err := json.Unmarshal(data, tunnelData); err != nil {
								log.Println("document ws服务端数据解析失败", err)
								return
							}
						}
						if err := ws.WriteJSONLock(tunnelType != websocket.MessageTypeBinary, &serverCmdType{
							CmdType: serverCmdTypeTunnelData,
							CmdId:   uuid.New().String(),
							Data:    cmdDataType{"tunnel_id": tunnelId, "data_type": tunnelType, "data": tunnelData},
						}); err != nil {
							log.Println("document ws客户端数据写入失败", err)
							return
						}
						if tunnelType == websocket.MessageTypeBinary {
							if err := ws.WriteMessageLock(false, websocket.MessageTypeBinary, data); err != nil {
								log.Println("document ws客户端数据写入失败", err)
								return
							}
							ws.Unlock()
							needUnlock = false
						}
					}
				}(tunnelId)
				serverCmd.Status = cmdStatusTypeSuccess
				serverCmd.Data["tunnel_id"] = tunnelId
				_ = ws.WriteJSON(&serverCmd)
			default:
				serverCmd.Message = "tunnel_type错误"
				_ = ws.WriteJSON(&serverCmd)
				log.Println("cmdData.TunnelType错误", cmdData.TunnelType)
			}
		case clientCmdTypeCloseTunnel:
			tunnelId, tunnel := getTunnelByCmdData(cmdDataData)
			if tunnel == nil {
				serverCmd.Message = "tunnel_id错误"
				_ = ws.WriteJSON(&serverCmd)
				continue
			}
			tunnel.ServerWs.Close()
			delete(tunnelMap, tunnelId)
			serverCmd.Status = cmdStatusTypeSuccess
			_ = ws.WriteJSON(&serverCmd)
		case clientCmdTypeTunnelData:
			tunnelId, tunnel := getTunnelByCmdData(cmdDataData)
			if cmdData.Status != cmdStatusTypeSuccess {
				// todo
				log.Println("tunnel通讯异常，关闭tunnel", tunnelId, cmdData.Message)
				tunnel.ServerWs.Close()
				delete(tunnelMap, tunnelId)
				continue
			}
			serverCmd.Data["tunnel_id"] = tunnelId
			if tunnel == nil {
				serverCmd.Message = "tunnel_id错误"
				_ = ws.WriteJSON(&serverCmd)
				continue
			}
			tunnelDataType, data := getTunnelDataByCmdData(tunnelCmdDataData, ws)
			if len(data) == 0 {
				serverCmd.Message = "数据错误"
				_ = ws.WriteJSON(&serverCmd)
				continue
			}
			err := tunnel.ServerWs.WriteMessage(websocket.MessageType(tunnelDataType), data)
			if err != nil {
				serverCmd.Message = "数据发送失败"
				_ = ws.WriteJSON(&serverCmd)
				log.Println("数据发送失败", err)
				continue
			}
			serverCmd.Status = cmdStatusTypeSuccess
			_ = ws.WriteJSON(&serverCmd)
		default:
			serverCmd.Message = "cmd_type错误"
			_ = ws.WriteJSON(&serverCmd)
			log.Println("cmdData.CmdType错误", cmdData.CmdType)
		}
	}
}
