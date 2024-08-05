package communication

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	doc_versioning_service "protodesign.cn/kcserver/apigateway/common/doc-versioning-service"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/mongo"
	"protodesign.cn/kcserver/common/redis"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/snowflake"
	"protodesign.cn/kcserver/utils/sliceutil"
	"protodesign.cn/kcserver/utils/str"
	"protodesign.cn/kcserver/utils/websocket"
	"sync"
	"time"
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

type Cmd struct {
	BaseVer     string   `json:"baseVer" bson:"baseVer"`
	BatchId     string   `json:"batchId" bson:"batchId"`
	Ops         []bson.M `json:"ops" bson:"ops"`
	IsRecovery  bool     `json:"isRecovery" bson:"isRecovery"`
	Description string   `json:"description" bson:"description"`
	Time        int64    `json:"time" bson:"time"`
	Posttime    int64    `json:"posttime" bson:"posttime"`
	DataFmtVer  int64    `json:"dataFmtVer,omitempty" bson:"dataFmtVer,omitempty"`
}

type ReceiveCmd struct {
	Cmd `json:",inline" bson:",inline"`
	Id  string `json:"id" bson:"id"`
}

type CmdItem0 struct {
	Id           int64  `json:"id" bson:"_id"`
	PreviousId   int64  `json:"previous_id" bson:"previous_id"`
	BatchStartId int64  `json:"batch_start_id" bson:"batch_start_id"`
	BatchEndId   int64  `json:"batch_end_id" bson:"batch_end_id"`
	BatchLength  int    `json:"batch_length" bson:"batch_length"`
	DocumentId   int64  `json:"document_id" bson:"document_id"`
	UserId       int64  `json:"user_id" bson:"user_id"`
	CmdId        string `json:"cmd_id" bson:"cmd_id"`
}

type CmdItem struct {
	CmdItem0 `json:",inline" bson:",inline"`
	Cmd      Cmd `json:"cmd" bson:"cmd"`
}

func (cmdItem CmdItem) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(cmdItem)
}

type SendData struct {
	Type       string   `json:"type"`                  // pullCmdsResult update errorInvalidParams errorNoPermission errorInsertFailed errorPullCmdsFailed
	CmdsData   string   `json:"cmds_data,omitempty"`   // pullCmdsResult update
	From       string   `json:"from,omitempty"`        // pullCmdsResult errorPullCmdsFailed
	To         string   `json:"to,omitempty"`          // pullCmdsResult errorPullCmdsFailed
	PreviousId string   `json:"previous_id,omitempty"` // pullCmdsResult
	CmdIdList  []string `json:"cmd_id_list,omitempty"` // errorInsertFailed
	Data       any      `json:"data,omitempty"`        // errorInsertFailed
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
	var documentVersion models.DocumentVersion
	if versionId != "" {
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

	_send := func(data SendData) error {
		return tunnel.Client.WriteJSONLock(true, &ServerCmd{
			CmdType: ServerCmdTypeTunnelData,
			CmdId:   uuid.New().String(),
			Data: CmdData{
				"tunnel_id": tunnel.Id,
				"data_type": websocket.MessageTypeText,
				"data":      data,
			},
		})
	}
	// 参数格式错误
	sendErrorInvalidParams := func() error {
		return _send(SendData{
			Type: "errorInvalidParams",
		})
	}
	// 无权限
	sendErrorNoPermission := func() error {
		return _send(SendData{
			Type: "errorNoPermission",
		})
	}
	// 数据插入失败
	sendErrorInsertFailed := func(cmdIdList []string, data any) error {
		return _send(SendData{
			Type:      "errorInsertFailed",
			CmdIdList: cmdIdList,
			Data:      data,
		})
	}
	// 发送更新数据
	sendUpdate := func(cmdsData string) error {
		return _send(SendData{
			Type:     "update",
			CmdsData: cmdsData,
		})
	}
	// 拉取数据失败
	sendErrorPullCmdsFailed := func(from string, to string) error {
		return _send(SendData{
			Type: "errorPullCmdsFailed",
			From: from,
			To:   to,
		})
	}

	// 从redis中获取最后一条cmd的id，若redis中没有则从mongodb中获取
	getPreviousId := func() (int64, error) {
		previousId, err := redis.Client.Get(context.Background(), "Document LastCmdId[DocumentId:"+documentIdStr+"]").Int64()
		if err != nil {
			if err != redis.Nil {
				log.Println("Document LastCmdId[DocumentId:"+documentIdStr+"]"+"获取失败", err)
				return 0, err
			}
			cmdItemList := make([]CmdItem, 0)
			documentCollection := mongo.DB.Collection("document1")
			reqParams := bson.M{
				"document_id": documentId,
			}
			findOptions := options.Find()
			findOptions.SetSort(bson.D{{"_id", -1}})
			findOptions.SetLimit(1)
			cur, err := documentCollection.Find(nil, reqParams, findOptions)
			if err != nil {
				return 0, err
			}
			if err := cur.All(nil, &cmdItemList); err != nil {
				return 0, err
			}
			if len(cmdItemList) > 0 {
				previousId = cmdItemList[0].Id
			} else {
				previousId = 0
			}
		}
		return previousId, nil
	}

	documentOpMutex := redis.RedSync.NewMutex("Document Op Mutex[DocumentId:"+documentIdStr+"]", redsync.WithExpiry(time.Second*10))

	tunnel.ReceiveFromClient = func(tunnelDataType TunnelDataType, data []byte, serverCmd ServerCmd) error {
		var receiveData ReceiveData
		if err := json.Unmarshal(data, &receiveData); err != nil {
			log.Println("数据解析失败", err)
			_ = sendErrorInvalidParams()
			return errors.New("数据解析失败")
		}
		if receiveData.Type == "commit" {
			if permType < models.PermTypeEditable {
				_ = sendErrorNoPermission()
				return errors.New("无权限")

			}
			var cmds []ReceiveCmd
			if err := json.Unmarshal([]byte(receiveData.Cmds), &cmds); err != nil {
				log.Println("数据解析失败", err)
				_ = sendErrorInvalidParams()
				return errors.New("数据解析失败")
			}
			if len(cmds) == 0 {
				_ = sendErrorInvalidParams()
				return nil
			}
			cmdIdList := sliceutil.MapT(func(cmd ReceiveCmd) string {
				return cmd.Id
			}, cmds...)

			// 上锁
			if err := documentOpMutex.Lock(); err != nil {
				log.Println(documentId, "获取锁失败 documentOpMutex.Lock", err)
				_ = sendErrorInsertFailed(cmdIdList, nil)
				return errors.New("数据插入失败")
			}
			defer func() {
				if _, err := documentOpMutex.Unlock(); err != nil {
					log.Println(documentId, "释放锁失败 documentOpMutex.Unlock", err)
				}
			}()

			// 从redis中获取最后一条cmd的id，若redis中没有则从mongodb中获取
			previousId, err := getPreviousId()
			if err != nil {
				log.Println(documentId, "lastCmdId查询失败", err)
				_ = sendErrorInsertFailed(cmdIdList, nil)
				return errors.New("数据插入失败")
			}

			cmdItemList := sliceutil.MapT(func(cmd ReceiveCmd) CmdItem {
				cmdItem := CmdItem{
					CmdItem0: CmdItem0{
						Id:           snowflake.NextId(),
						PreviousId:   previousId,
						BatchStartId: 0,
						BatchEndId:   0,
						BatchLength:  0,
						DocumentId:   documentId,
						UserId:       userId,
						CmdId:        cmd.Id,
					},
					Cmd: cmd.Cmd,
				}
				previousId = cmdItem.Id
				return cmdItem
			}, cmds...)

			cmdItem0List := sliceutil.MapT(func(cmdItem CmdItem) CmdItem0 {
				return cmdItem.CmdItem0
			}, cmdItemList...)
			cmdItem0ListString := ""
			if cmdItem0ListStringByte, err := json.Marshal(cmdItem0List); err != nil {
				cmdItem0ListString = string(cmdItem0ListStringByte)
			}

			batchStartId := cmdItemList[0].Id
			batchEndId := cmdItemList[len(cmdItemList)-1].Id
			batchLength := len(cmdItemList)
			for i := range cmdItemList {
				cmdItemList[i].BatchStartId = batchStartId
				cmdItemList[i].BatchEndId = batchEndId
				cmdItemList[i].BatchLength = batchLength
			}

			cmdItemListData, err := json.Marshal(cmdItemList)
			if err != nil {
				log.Println("json编码错误 cmdItemsData", err)
				_ = sendErrorInsertFailed(cmdIdList, nil)
				return errors.New("数据插入失败")
			}

			documentCollection := mongo.DB.Collection("document1")
			var duplicateCmd *CmdItem
			if func() error {
				if insertRes, err := documentCollection.InsertMany(context.Background(), sliceutil.ConvertToAnySlice(cmdItemList)); err != nil {
					log.Println("数据插入失败："+cmdItem0ListString, err)

					// duplicate key error
					if mongo.IsDuplicateKeyError(err) {
						index := 0
						if insertRes != nil && len(insertRes.InsertedIDs) > 0 {
							lastInsertedID := insertRes.InsertedIDs[len(insertRes.InsertedIDs)-1]
							for i, cmdItem := range cmdItemList {
								if cmdItem.Id == lastInsertedID {
									index = i + 1
									break
								}
							}
							if index == 0 {
								index = -1
							}
						}
						if index >= 0 {
							//duplicateCmd = &cmdItemList[index]
							// 从mongodb中获取重复的cmd
							duplicateCmdCmdId := cmdItemList[index].CmdId
							duplicateCmdDocumentId := cmdItemList[index].DocumentId
							duplicateCmd = &CmdItem{}
							if err := documentCollection.FindOne(context.Background(), bson.M{"document_id": duplicateCmdDocumentId, "cmd_id": duplicateCmdCmdId}).Decode(duplicateCmd); err != nil {
								log.Println("重复数据查询失败", duplicateCmdDocumentId, duplicateCmdCmdId, err)
							}
						}
					}

					return errors.New("数据插入失败")
				}
				if _, err = redis.Client.Set(context.Background(), "Document LastCmdId[DocumentId:"+documentIdStr+"]", previousId, time.Second*1).Result(); err != nil {
					log.Println("Document LastCmdId[DocumentId:"+documentIdStr+"]"+"设置失败", err)
					return errors.New("数据插入失败")
				}
				redis.Client.Publish(context.Background(), "Document Op[DocumentId:"+documentIdStr+"]", cmdItemListData)
				doc_versioning_service.Trigger(documentId)
				return nil
			}() != nil {
				// 删除已插入的数据（cmdItemList）
				if res, err := documentCollection.DeleteMany(context.Background(), bson.M{
					"document_id": documentId,
					"_id": bson.M{
						"$in": sliceutil.MapT(func(cmdItem CmdItem) int64 {
							return cmdItem.Id
						}, cmdItemList...),
					},
				}); err != nil {
					log.Println("数据删除失败", err, res)
					// todo 删除失败的处理
				}
				redis.Client.Del(context.Background(), "Document LastCmdId[DocumentId:"+documentIdStr+"]")
				insertFailedData := map[string]any{}
				if duplicateCmd != nil {
					insertFailedData["type"] = "duplicate"
					insertFailedData["duplicateCmd"] = duplicateCmd
				}
				_ = sendErrorInsertFailed(cmdIdList, insertFailedData)
				return errors.New("数据插入失败")
			}

		} else if receiveData.Type == "pullCmds" {
			var cmdItemList []CmdItem
			fromId := str.DefaultToInt(receiveData.From, 0)
			toId := str.DefaultToInt(receiveData.To, 0)
			idFilter := bson.M{
				"$gte": fromId,
			}
			findOptions := options.Find()
			findOptions.SetSort(bson.D{{"_id", 1}})
			if toId > 0 {
				idFilter["$lte"] = toId
			} else {
				//findOptions.SetLimit(100)
			}
			cur, err := mongo.DB.Collection("document1").Find(nil, bson.M{
				"document_id": documentId,
				"_id":         idFilter,
			}, findOptions)
			if err != nil {
				log.Println("数据查询失败", err)
				_ = sendErrorPullCmdsFailed(receiveData.From, receiveData.To)
				return errors.New("数据查询失败")
			}
			if err := cur.All(nil, &cmdItemList); err != nil {
				log.Println("数据查询失败", err)
				_ = sendErrorPullCmdsFailed(receiveData.From, receiveData.To)
				return errors.New("数据查询失败")
			}
			cmdItemListData, err := json.Marshal(cmdItemList)
			if err != nil {
				log.Println("json编码错误 cmdItemsData", err)
				_ = sendErrorPullCmdsFailed(receiveData.From, receiveData.To)
				return errors.New("数据查询失败")
			}
			previousId, err := getPreviousId()
			if err != nil {
				log.Println(documentId, "lastCmdId查询失败", err)
				_ = sendErrorPullCmdsFailed(receiveData.From, receiveData.To)
				return errors.New("数据查询失败")
			}
			if err := tunnel.Client.WriteJSONLock(true, &ServerCmd{
				CmdType: ServerCmdTypeTunnelData,
				CmdId:   uuid.New().String(),
				Data: CmdData{
					"tunnel_id": tunnel.Id,
					"data_type": websocket.MessageTypeText,
					"data": SendData{
						Type:       "pullCmdsResult",
						CmdsData:   string(cmdItemListData),
						From:       receiveData.From,
						To:         receiveData.To,
						PreviousId: str.IntToString(previousId),
					},
				},
			}); err != nil {
				log.Println("数据发送失败", err)
				_ = sendErrorPullCmdsFailed(receiveData.From, receiveData.To)
				return errors.New("数据查询失败")
			}
		}

		return nil
	}

	go func() {
		defer tunnelServer.Close()

		// 根据版本号发送初始cmdList数据
		cmdItemList := make([]CmdItem, 0)
		documentCollection := mongo.DB.Collection("document1")
		reqParams := bson.M{
			"document_id": documentId,
			"_id": bson.M{
				"$gt": documentVersion.LastCmdId,
			},
		}
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{"_id", 1}})
		cur, err := documentCollection.Find(nil, reqParams, findOptions)
		if err != nil {
			log.Println("cmdList查询失败", err)
			return
		}
		if err := cur.All(nil, &cmdItemList); err != nil {
			log.Println("cmdList查询失败", err)
			return
		}
		cmdItemListData, err := json.Marshal(cmdItemList)
		if err != nil {
			log.Println("json编码错误 cmdItemsData", err)
			return
		}
		if err := tunnel.Client.WriteJSONLock(true, &ServerCmd{
			CmdType: ServerCmdTypeTunnelData,
			CmdId:   uuid.New().String(),
			Data: CmdData{
				"tunnel_id": tunnel.Id,
				"data_type": websocket.MessageTypeText,
				"data": SendData{
					Type:     "update",
					CmdsData: string(cmdItemListData),
				},
			},
		}); err != nil {
			log.Println("初始数据发送失败", err)
			return
		}

		pubsub := redis.Client.Subscribe(context.Background(), "Document Op[DocumentId:"+documentIdStr+"]")
		defer pubsub.Close()
		for {
			msg, err := pubsub.ReceiveMessage(context.Background())
			if err != nil {
				log.Println("读取redis订阅消息失败", err)
				return
			}
			if err := sendUpdate(msg.Payload); err != nil {
				log.Println("数据发送失败", err)
				return
			}
		}
	}()

	return tunnel
}
