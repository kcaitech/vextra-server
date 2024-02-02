package communication

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/mongo"
	"protodesign.cn/kcserver/common/redis"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/snowflake"
	"protodesign.cn/kcserver/utils/sliceutil"
	"protodesign.cn/kcserver/utils/str"
	"protodesign.cn/kcserver/utils/websocket"
	"sync"
)

type docOpTunnelServer struct {
	HandleClose func(code int, text string)
	IsClose     bool
	CloseLock   sync.Mutex
}

func (server *docOpTunnelServer) SetCloseHandler(handler func(code int, text string)) {
	server.HandleClose = handler
}

func (server *docOpTunnelServer) WriteMessage(messageType websocket.MessageType, data []byte) (err error) {
	return nil
}

func (server *docOpTunnelServer) ReadMessage() (websocket.MessageType, []byte, error) {
	return websocket.MessageTypeNone, nil, websocket.ErrClosed
}

func (server *docOpTunnelServer) Close() {
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

type ReceiveData struct {
	Type string `json:"type"` // commit pullCmds
	Cmds string `json:"cmds"` // commit
	From string `json:"from"` // pullCmds
	To   string `json:"to"`   // pullCmds
}

type SendData struct {
	Type string `json:"type"` // commitResult pullCmdsResult update
	Cmds string `json:"cmds"` // pullCmdsResult update
	From string `json:"from"` // pullCmdsResult
	To   string `json:"to"`   // pullCmdsResult
}

type CmdItem struct {
	Id           int64  `json:"id" bson:"_id"`
	BatchStartId int64  `json:"batch_start_id" bson:"batch_start_id"`
	BatchEndId   int64  `json:"batch_end_id" bson:"batch_end_id"`
	BatchLength  int    `json:"batch_length" bson:"batch_length"`
	DocumentId   int64  `json:"document_id" bson:"document_id"`
	UserId       int64  `json:"user_id" bson:"user_id"`
	Cmd          bson.M `json:"cmd" bson:"cmd"`
}

func (cmdItem CmdItem) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(cmdItem)
}

func OpenDocOpTunnel(clientWs *websocket.Ws, clientCmdData CmdData, serverCmd ServerCmd, data Data) *Tunnel {
	clientCmdDataData, ok := clientCmdData["data"].(map[string]any)
	documentIdStr, ok1 := clientCmdDataData["document_id"].(string)
	userId, ok2 := data["userId"].(int64)
	if !ok || !ok1 || documentIdStr == "" || !ok2 || userId <= 0 {
		serverCmd.Message = "通道建立失败，参数错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，参数错误", ok, ok1, ok2, documentIdStr, userId)
		return nil
	}
	versionId, _ := clientCmdDataData["version_id"].(string)

	// 获取文档信息
	documentId := str.DefaultToInt(documentIdStr, 0)
	if documentId <= 0 {
		serverCmd.Message = "通道建立失败，documentId错误"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，documentId错误", documentId)
		return nil
	}
	documentService := services.NewDocumentService()
	var document models.Document
	if documentService.GetById(documentId, &document) != nil {
		serverCmd.Message = "通道建立失败，文档不存在"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，文档不存在", documentId)
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
	// 验证文档版本信息
	if versionId != "" {
		var documentVersion models.DocumentVersion
		if err := documentService.DocumentVersionService.Get(&documentVersion, "document_id = ? and version_id = ?", documentId, versionId); err != nil {
			serverCmd.Message = "通道建立失败，文档版本错误"
			_ = clientWs.WriteJSON(&serverCmd)
			log.Println("document ws建立失败，文档版本不存在", documentId, versionId)
			return nil
		}
	}
	if !document.LockedAt.IsZero() && document.UserId != userId {
		serverCmd.Message = "通道建立失败，审核不通过"
		_ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，审核不通过", documentId)
		return nil
	}

	tunnelServer := &docOpTunnelServer{}

	tunnelId := uuid.New().String()
	tunnel := &Tunnel{
		Id:     tunnelId,
		Server: tunnelServer,
		Client: clientWs,
	}

	tunnel.ReceiveFromClient = func(tunnelDataType TunnelDataType, data []byte, serverCmd ServerCmd) error {
		var receiveData ReceiveData
		if err := json.Unmarshal(data, &receiveData); err != nil {
			log.Println("数据解析失败", err)
			return errors.New("数据解析失败")
		}
		if receiveData.Type == "commit" {
			if permType < models.PermTypeEditable {
				return errors.New("无权限")

			}
			var cmds []map[string]any
			if err := json.Unmarshal([]byte(receiveData.Cmds), &cmds); err != nil {
				log.Println("数据解析失败", err)
				return errors.New("数据解析失败")
			}
			if len(cmds) == 0 {
				return nil
			}
			cmdItemList := sliceutil.MapT(func(cmd map[string]any) CmdItem {
				return CmdItem{
					Id:           snowflake.NextId(),
					BatchStartId: 0,
					BatchEndId:   0,
					BatchLength:  0,
					DocumentId:   documentId,
					UserId:       userId,
					Cmd:          cmd,
				}
			}, cmds...)
			batchStartId := cmdItemList[0].Id
			batchEndId := cmdItemList[len(cmdItemList)-1].Id
			batchLength := len(cmdItemList)
			for i := range cmdItemList {
				cmdItemList[i].BatchStartId = batchStartId
				cmdItemList[i].BatchEndId = batchEndId
				cmdItemList[i].BatchLength = batchLength
			}

			cmdItemsData, err := json.Marshal(cmdItemList)
			if err != nil {
				log.Println("json编码错误 cmdItemsData", err)
				return errors.New("数据插入失败")
			}

			session, err := mongo.Client.StartSession()
			if err != nil {
				log.Println("事务开启失败 StartSession", err)
				return errors.New("数据插入失败")
			}
			err = session.StartTransaction()
			if err != nil {
				log.Println("事务开启失败 StartTransaction", err)
				return errors.New("数据插入失败")
			}

			ctx := context.Background()
			err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
				_, err := mongo.DB.Collection("document1").InsertMany(nil, sliceutil.ConvertToAnySlice(cmdItemList))
				if err != nil {
					log.Println("数据插入失败", err)
					return errors.New("数据插入失败")
				}
				return nil
			})

			if err != nil {
				_ = session.AbortTransaction(ctx)
			} else {
				_ = session.CommitTransaction(ctx)
			}
			session.EndSession(ctx)

			redis.Client.Publish(context.Background(), "Document Op[DocumentId:"+documentIdStr+"]", cmdItemsData)

		} else if receiveData.Type == "pullCmds" {
			var cmdItems []CmdItem
			fromId := str.DefaultToInt(receiveData.From, 0)
			toId := str.DefaultToInt(receiveData.To, 0)
			if fromId <= 0 || toId <= 0 {
				return errors.New("参数错误")
			}
			cur, err := mongo.DB.Collection("document1").Find(nil, bson.M{
				"document_id": documentId,
				"_id":         bson.M{"$gte": fromId, "$lte": toId},
			})
			if err != nil {
				log.Println("数据查询失败", err)
				return errors.New("数据查询失败")
			}
			if err := cur.All(nil, &cmdItems); err != nil {
				log.Println("数据查询失败", err)
				return errors.New("数据查询失败")
			}
			cmdItemsData, err := json.Marshal(cmdItems)
			if err != nil {
				log.Println("json编码错误 cmdItemsData", err)
				return errors.New("数据查询失败")
			}
			if err := tunnel.Client.WriteJSONLock(false, &ServerCmd{
				CmdType: ServerCmdTypeTunnelData,
				CmdId:   uuid.New().String(),
				Data: CmdData{"tunnel_id": tunnel.Id, "data_type": websocket.MessageTypeText, "data": SendData{
					Type: "pullCmdsResult",
					Cmds: string(cmdItemsData),
					From: receiveData.From,
					To:   receiveData.To,
				}},
			}); err != nil {
				log.Println("数据发送失败", err)
				return errors.New("数据查询失败")
			}
		}
		return nil
	}

	// 监听redis 返回结果到前端tunnel.client
	go func() {
		pubsub := redis.Client.Subscribe(context.Background(), "Document Op[DocumentId:"+documentIdStr+"]")
		defer pubsub.Close()
		defer tunnelServer.Close()
		for {
			msg, err := pubsub.ReceiveMessage(context.Background())
			if err != nil {
				log.Println("读取redis订阅消息失败", err)
				return
			}
			if err := tunnel.Client.WriteJSONLock(false, &ServerCmd{
				CmdType: ServerCmdTypeTunnelData,
				CmdId:   uuid.New().String(),
				Data: CmdData{"tunnel_id": tunnel.Id, "data_type": websocket.MessageTypeText, "data": &SendData{
					Type: "update",
					Cmds: msg.Payload,
				}},
			}); err != nil {
				log.Println("数据发送失败", err)
				return
			}
		}
	}()

	return tunnel
}
