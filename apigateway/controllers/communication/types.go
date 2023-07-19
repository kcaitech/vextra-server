package communication

import (
	"encoding/json"
	"protodesign.cn/kcserver/utils/websocket"
)

type Header struct {
	Token string `json:"token"`
}

type ClientCmdType uint8

const (
	ClientCmdTypeReturn            ClientCmdType = iota // 返回cmd执行结果
	ClientCmdTypeOpenTunnel                             // 打开一条虚拟通道
	ClientCmdTypeCloseTunnel                            // 关闭一条虚拟通道
	ClientCmdTypeTunnelData                             // 虚拟通道数据
	ClientCmdTypeHeartbeat         = 255                // 心跳包
	ClientCmdTypeHeartbeatResponse = 254                // 心跳包响应
)

// TunnelType 虚拟通道类型
type TunnelType uint8

const (
	TunnelTypeDocOp             TunnelType = iota // 文档操作
	TunnelTypeDocResourceUpload                   // 文档资源上传
	TunnelTypeDocCommentOp                        // 文档评论操作
)

type Data map[string]any

type CmdData Data

type ClientCmd struct {
	CmdType    ClientCmdType   `json:"cmd_type"`
	CmdId      string          `json:"cmd_id"`
	TunnelType TunnelType      `json:"tunnel_type,omitempty"`
	Status     CmdStatusType   `json:"status,omitempty"`
	Message    string          `json:"message,omitempty"`
	Data       json.RawMessage `json:"data,omitempty"`
}

type TunnelDataType uint8

const (
	TunnelDataTypeNone   = TunnelDataType(0)
	TunnelDataTypeText   = TunnelDataType(websocket.MessageTypeText)
	TunnelDataTypeBinary = TunnelDataType(websocket.MessageTypeBinary)
)

type TunnelCmdData struct {
	TunnelId string          `json:"tunnel_id"`
	DataType TunnelDataType  `json:"data_type"`
	Data     json.RawMessage `json:"data"`
}

type ServerCmdType uint8

const (
	ServerCmdTypeInitResult        ServerCmdType = iota // 返回初始化结果
	ServerCmdTypeReturn                                 // 返回cmd执行结果
	ServerCmdTypeCloseTunnel                            // 关闭一条虚拟通道
	ServerCmdTypeTunnelData                             // 虚拟通道数据
	ServerCmdTypeHeartbeat         = 255                // 心跳包
	ServerCmdTypeHeartbeatResponse = 254                // 心跳包响应
)

type CmdStatusType string

const (
	CmdStatusSuccess CmdStatusType = "success"
	CmdStatusFail    CmdStatusType = "fail"
)

type ServerCmd struct {
	CmdType ServerCmdType `json:"cmd_type"`
	CmdId   string        `json:"cmd_id"`
	Status  CmdStatusType `json:"status,omitempty"`
	Message string        `json:"message,omitempty"`
	Data    CmdData       `json:"data,omitempty"`
}
