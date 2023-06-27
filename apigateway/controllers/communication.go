package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"protodesign.cn/kcserver/common/gin/response"
	. "protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/utils/str"
	"protodesign.cn/kcserver/utils/websocket"
)

type headerDataType struct {
	Token string `json:"token"`
}

type cmdType uint8

const (
	cmdTypeOpenTunnel  cmdType = iota // 打开一条虚拟通道
	cmdTypeCloseTunnel                // 关闭一条虚拟通道
)

// tunnelType 虚拟通道类型
type tunnelType uint8

const (
	tunnelTypeDocOp tunnelType = iota // 文档操作
)

type cmdDataType struct {
	CmdType    cmdType    `json:"cmd_type"`
	TunnelType tunnelType `json:"tunnel_type,omitempty"`
	Id         string     `json:"id,omitempty"`
}

type serverCmdType uint8

const (
	serverCmdTypeInitResult serverCmdType = iota // 返回隧道初始化结果
	serverCmdTypeReturn                          // 返回cmd执行结果
)

type serverCmdDataType struct {
	CmdType serverCmdType `json:"cmd_type"`
	Id      string        `json:"id,omitempty"`
	Data    any           `json:"data,omitempty"`
}

// Communication websocket连接
func Communication(c *gin.Context) {
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Fail(c, "建立ws连接失败："+err.Error())
		return
	}
	defer ws.Close()

	ws.SetCloseHandler(func(code int, text string) {
		log.Println("ws连接关闭", code, text)
	})

	headerData := &headerDataType{}
	if err := ws.ReadJSON(headerData); err != nil {
		_ = ws.WriteJSON(&serverCmdDataType{
			CmdType: serverCmdTypeInitResult,
			Data:    "Header结构错误",
		})
		log.Println("Header结构错误", err)
		return
	}
	parseData, err := ParseJwt(headerData.Token)
	if err != nil {
		_ = ws.WriteJSON(&serverCmdDataType{
			CmdType: serverCmdTypeInitResult,
			Data:    "Token错误",
		})
		log.Println("Token错误", err)
		return
	}
	userId := str.DefaultToInt(parseData.Id, 0)
	if userId <= 0 {
		_ = ws.WriteJSON(&serverCmdDataType{
			CmdType: serverCmdTypeInitResult,
			Data:    "UserId错误",
		})
		log.Println("UserId错误", userId)
		return
	}
	wsId := uuid.New().String()
	_ = ws.WriteJSON(&serverCmdDataType{
		CmdType: serverCmdTypeInitResult,
		Id:      wsId,
		Data:    "success",
	})
	if ws.IsClose() {
		return
	}
	log.Println("websocket连接成功，wsId：", wsId)

	//for {
	//
	//}

}
