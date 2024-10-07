package communication

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redsync/redsync/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/mongo"
	"kcaitech.com/kcserver/common/redis"
	"kcaitech.com/kcserver/common/services"
	"kcaitech.com/kcserver/common/snowflake"
	autosave "kcaitech.com/kcserver/controllers/document"
	"kcaitech.com/kcserver/utils/sliceutil"
	"kcaitech.com/kcserver/utils/str"
	"kcaitech.com/kcserver/utils/websocket"
)

type ReceiveData struct {
	Type string `json:"type"` // commit pullCmds
	Cmds string `json:"cmds"` // commit
	From string `json:"from"` // pullCmds
	To   string `json:"to"`   // pullCmds
}

type Cmd struct {
	BaseVer     string      `json:"baseVer" bson:"baseVer"`
	BatchId     string      `json:"batchId" bson:"batchId"`
	Ops         []bson.M    `json:"ops" bson:"ops"`
	IsRecovery  bool        `json:"isRecovery" bson:"isRecovery"`
	Description string      `json:"description" bson:"description"`
	Time        int64       `json:"time" bson:"time"`
	Posttime    int64       `json:"posttime" bson:"posttime"`
	DataFmtVer  interface{} `json:"dataFmtVer,omitempty" bson:"dataFmtVer,omitempty"` // int | string
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

type opServe struct {
	ws   *websocket.Ws
	quit chan struct{}
	// isready bool
	genSId     func() string
	mutex      *redsync.Mutex
	permType   models.PermType
	documentId int64
	userId     int64
}

// 从redis中获取最后一条cmd的id，若redis中没有则从mongodb中获取
func getPreviousId(documentId int64) (int64, error) {
	documentIdStr := str.IntToString(documentId)
	previousId, err := redis.Client.Get(context.Background(), "Document LastCmdId[DocumentId:"+documentIdStr+"]").Int64() // redis也是最终一致性
	if err == nil {
		return previousId, nil
	}

	if err != redis.Nil { // todo redis获取不到时，应该去mongodb获取，并更新到redis？
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

	cur, err := documentCollection.Find(nil, reqParams, findOptions) // todo 要去主节点找
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

	return previousId, nil
}

func NewOpServe(ws *websocket.Ws, userId int64, documentId int64, versionId string, genSId func() string) *opServe {

	documentService := services.NewDocumentService()
	var document models.Document
	if documentService.GetById(documentId, &document) != nil {
		// serverCmd.Message = "通道建立失败，文档不存在"
		// _ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，文档不存在", documentId)
		return nil
	}
	// 权限校验
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeReadOnly {
		// serverCmd.Message = "通道建立失败"
		// if err != nil {
		// 	serverCmd.Message += "，无权限"
		// }
		// _ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，权限校验错误", err, permType)
		return nil
	}
	// 验证文档版本信息
	var documentVersion models.DocumentVersion
	if versionId != "" { // todo 是否将获取的version直接返回前端。后端也准备将文档上传到cdn
		if err := documentService.DocumentVersionService.Get(&documentVersion, "document_id = ? and version_id = ?", documentId, versionId); err != nil {
			// serverCmd.Message = "通道建立失败，文档版本错误"
			// _ = clientWs.WriteJSON(&serverCmd)
			log.Println("document ws建立失败，文档版本不存在", documentId, versionId)
			return nil
		}
	}
	if !document.LockedAt.IsZero() && document.UserId != userId {
		// serverCmd.Message = "通道建立失败，审核不通过"
		// _ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，审核不通过", documentId)
		return nil
	}

	documentIdStr := str.IntToString(documentId)
	mutex := redis.RedSync.NewMutex("Document Op Mutex[DocumentId:"+documentIdStr+"]", redsync.WithExpiry(time.Second*10))

	serv := opServe{
		ws: ws,
		// isready: false,
		genSId:     genSId,
		mutex:      mutex,
		permType:   permType,
		documentId: documentId,
		userId:     userId,
	}

	serv.start(documentId, documentVersion)
	// serv.isready = true
	return &serv
}

func (serv *opServe) start(documentId int64, documentVersion models.DocumentVersion) {
	go func() {
		// defer tunnelServer.Close()

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
		cur, err := documentCollection.Find(nil, reqParams, findOptions) // 有没有可能一个batch的cmd没拉取完整
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

		serv.send(string(cmdItemListData))

		// if err := tunnel.Client.WriteJSONLock(true, &ServerCmd{
		// 	CmdType: ServerCmdTypeTunnelData,
		// 	CmdId:   uuid.New().String(),
		// 	Data: CmdData{
		// 		"tunnel_id": tunnel.Id,
		// 		"data_type": websocket.MessageTypeText,
		// 		"data": SendData{
		// 			Type:     "update",
		// 			CmdsData: string(cmdItemListData),
		// 		},
		// 	},
		// }); err != nil {
		// 	log.Println("初始数据发送失败", err)
		// 	return
		// }

		documentIdStr := str.IntToString(documentId)
		pubsub := redis.Client.Subscribe(context.Background(), "Document Op[DocumentId:"+documentIdStr+"]")
		defer pubsub.Close()
		channel := pubsub.Channel()
		for {
			select {
			case v, ok := <-channel:
				if !ok {
					break
				}
				// todo 需要判断下当前lastversion版本,当接不上时需要拉取cmds
				serv.send(v.Payload)
			case <-serv.quit:
				break
			}
		}
		// for {
		// 	msg, err := pubsub.ReceiveMessage(context.Background())
		// 	if err != nil {
		// 		log.Println("读取redis订阅消息失败", err)
		// 		return
		// 	}
		// 	if err := sendUpdate(msg.Payload); err != nil {
		// 		log.Println("数据发送失败", err)
		// 		return
		// 	}
		// }
	}()
}

func (serv *opServe) close() {
	close(serv.quit)
}

func (serv *opServe) send(data string) {
	// jsonData := &Data{}
	// if err := json.Unmarshal([]byte(data), jsonData); err != nil {
	// 	log.Println("comment, redis data wrong", err)
	// 	return
	// }

	sendData := SendData{
		Type:     "updata",
		CmdsData: data,
	}

	bytes, err := json.Marshal(&sendData)
	if err != nil {
		log.Println("op, marshal data fail", err)
		return
	}

	sid := serv.genSId()
	serverData := TransData{
		Type:   DataTypes_Op,
		DataId: sid,
		Data:   string(bytes),
	}
	if err := serv.ws.WriteJSONLock(true, &serverData); err != nil {
		log.Println("op, send data fail", err)
		return
	}
}

func (serv *opServe) handleCommit(data *TransData, receiveData *ReceiveData) {
	serverData := TransData{}
	serverData.Type = data.Type
	serverData.DataId = data.DataId

	msgErr := func(msg string, serverData *TransData, err *error) {
		serverData.Err = msg
		log.Println(msg)
		_ = serv.ws.WriteJSON(serverData)
	}

	// var receiveData = ReceiveData{}
	// if err := json.Unmarshal([]byte(data.Data), &receiveData); err != nil {
	// 	msgErr("数据解析失败", &serverData)
	// 	return
	// }

	if serv.permType < models.PermTypeEditable {
		msgErr("has no permision", &serverData, nil)
		return
	}
	var cmds []ReceiveCmd
	if err := json.Unmarshal([]byte(receiveData.Cmds), &cmds); err != nil {
		msgErr("数据解析失败1", &serverData, &err)
		return
	}
	if len(cmds) == 0 {
		msgErr("参数错误", &serverData, nil)
		return
	}

	// cmdIdList := sliceutil.MapT(func(cmd ReceiveCmd) string {
	// 	return cmd.Id
	// }, cmds...)

	// 上锁
	if err := serv.mutex.Lock(); err != nil {
		msgErr("获取锁失败", &serverData, &err)
		return
	}
	defer func() {
		if _, err := serv.mutex.Unlock(); err != nil {
			log.Println("释放锁失败 documentOpMutex.Unlock", err)
		}
	}()

	// 从redis中获取最后一条cmd的id，若redis中没有则从mongodb中获取
	previousId, err := getPreviousId(serv.documentId)
	if err != nil {
		msgErr("get previous id failed", &serverData, &err)
		return
	}

	cmdItemList := sliceutil.MapT(func(cmd ReceiveCmd) CmdItem {
		cmdItem := CmdItem{
			CmdItem0: CmdItem0{
				Id:           snowflake.NextId(),
				PreviousId:   previousId,
				BatchStartId: 0,
				BatchEndId:   0,
				BatchLength:  0,
				DocumentId:   serv.documentId,
				UserId:       serv.userId,
				CmdId:        cmd.Id,
			},
			Cmd: cmd.Cmd,
		}
		previousId = cmdItem.Id
		return cmdItem
	}, cmds...)

	// cmdItem0List := sliceutil.MapT(func(cmdItem CmdItem) CmdItem0 {
	// 	return cmdItem.CmdItem0
	// }, cmdItemList...)
	// cmdItem0ListString := ""
	// if cmdItem0ListStringByte, err := json.Marshal(cmdItem0List); err != nil {
	// 	cmdItem0ListString = string(cmdItem0ListStringByte)
	// }

	batchStartId := cmdItemList[0].Id
	batchEndId := cmdItemList[len(cmdItemList)-1].Id
	batchLength := len(cmdItemList)
	for i := range cmdItemList { // todo batchid?
		cmdItemList[i].BatchStartId = batchStartId
		cmdItemList[i].BatchEndId = batchEndId
		cmdItemList[i].BatchLength = batchLength
	}

	cmdItemListData, err := json.Marshal(cmdItemList)
	if err != nil {
		msgErr("json marshal failed", &serverData, &err)
		return
	}
	documentIdStr := str.IntToString(serv.documentId)
	documentCollection := mongo.DB.Collection("document1")

	// 先删除，防止插入失败时无法正确更新lastcmdid
	if _, err = redis.Client.Del(context.Background(), "Document LastCmdId[DocumentId:"+documentIdStr+"]").Result(); err != nil {
		msgErr("redis del fail", &serverData, &err)
		return
	}

	insertRes, err := documentCollection.InsertMany(context.Background(), sliceutil.ConvertToAnySlice(cmdItemList)) // 这是个原子性操作
	if err != nil && mongo.IsDuplicateKeyError(err) {
		index := len(insertRes.InsertedIDs) + 1
		duplicateCmdCmdId := cmdItemList[index].CmdId
		duplicateCmdDocumentId := cmdItemList[index].DocumentId
		duplicateCmd := &CmdItem{}
		if err := documentCollection.FindOne(context.Background(), bson.M{"document_id": duplicateCmdDocumentId, "cmd_id": duplicateCmdCmdId}).Decode(duplicateCmd); err != nil {
			log.Println("重复数据查询失败", duplicateCmdDocumentId, duplicateCmdCmdId, err)
			msgErr("数据插入失败", &serverData, &err)
			return
		}

		duplicate := map[string]any{}

		duplicate["type"] = "duplicate"
		duplicate["duplicateCmd"] = duplicateCmd

		duplicateStr, err := json.Marshal(duplicate)
		if err != nil {
			msgErr("数据插入失败", &serverData, &err)
			return
		}

		serverData.Data = string(duplicateStr)
		msgErr("duplicate", &serverData, &err)
		return

	} else if err != nil {
		// 插入失败
		// 删除已插入的数据（cmdItemList）
		// if res, err := documentCollection.DeleteMany(context.Background(), bson.M{
		// 	"document_id": serv.documentId,
		// 	"_id": bson.M{
		// 		"$in": sliceutil.MapT(func(cmdItem CmdItem) int64 {
		// 			return cmdItem.Id
		// 		}, cmdItemList...),
		// 	},
		// }); err != nil {
		// 	log.Println("数据删除失败", err, res)
		// 	// todo 删除失败的处理
		// }
		// redis.Client.Del(context.Background(), "Document LastCmdId[DocumentId:"+documentIdStr+"]")
		// insertFailedData := map[string]any{}
		// if duplicateCmd != nil {
		// 	insertFailedData["type"] = "duplicate"
		// 	insertFailedData["duplicateCmd"] = duplicateCmd
		// }
		// _ = sendErrorInsertFailed(cmdIdList, insertFailedData)
		// return errors.New("数据插入失败")

		msgErr("数据插入失败", &serverData, &err)
		return
	} else {
		if _, err = redis.Client.Set(context.Background(), "Document LastCmdId[DocumentId:"+documentIdStr+"]", previousId, time.Second*1).Result(); err != nil {
			log.Println("Document LastCmdId[DocumentId:"+documentIdStr+"]"+"设置失败", err)
			// return errors.New("数据插入失败")
		}
		redis.Client.Publish(context.Background(), "Document Op[DocumentId:"+documentIdStr+"]", cmdItemListData) // 通知客户端是通过redis订阅来触发的
		autosave.AutoSave(serv.documentId)
		// return nil

		_ = serv.ws.WriteJSON(serverData) // sucess
	}
}

func (serv *opServe) handlePullCmds(data *TransData, receiveData *ReceiveData) {
	serverData := TransData{}
	serverData.Type = data.Type
	serverData.DataId = data.DataId

	msgErr := func(msg string, serverData *TransData, err *error) {
		serverData.Err = msg
		log.Println(msg)
		_ = serv.ws.WriteJSON(serverData)
	}

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
		"document_id": serv.documentId,
		"_id":         idFilter,
	}, findOptions)
	if err != nil {
		msgErr("数据查询失败", &serverData, &err)
		return
	}
	if err := cur.All(nil, &cmdItemList); err != nil {
		msgErr("数据查询失败", &serverData, &err)
		return
	}
	cmdItemListData, err := json.Marshal(cmdItemList)
	if err != nil {
		msgErr("json编码错误", &serverData, &err)
		return
	}
	// previousId, err := getPreviousId(serv.documentId)
	// if err != nil {
	// 	log.Println(documentId, "lastCmdId查询失败", err)
	// 	_ = sendErrorPullCmdsFailed(receiveData.From, receiveData.To)
	// 	return errors.New("数据查询失败")
	// }

	sendData := SendData{
		Type:     "pullCmdsResult",
		CmdsData: string(cmdItemListData),
		From:     receiveData.From,
		To:       receiveData.To,
		// PreviousId: str.IntToString(previousId),
	}

	sendDataStr, err := json.Marshal(sendData)
	if err != nil {
		msgErr("json编码错误", &serverData, &err)
		return
	}

	sendData.Data = sendDataStr
	_ = serv.ws.WriteJSON(serverData)

}

func (serv *opServe) handle(data *TransData, binaryData *([]byte)) {
	serverData := TransData{}
	serverData.Type = data.Type
	serverData.DataId = data.DataId

	msgErr := func(msg string, serverData *TransData, err *error) {
		serverData.Err = msg
		log.Println(msg, err)
		_ = serv.ws.WriteJSON(serverData)
	}

	var receiveData = ReceiveData{}
	if err := json.Unmarshal([]byte(data.Data), &receiveData); err != nil {
		msgErr("数据解析失败", &serverData, &err)
		return
	}
	if receiveData.Type == "commit" {
		serv.handleCommit(data, &receiveData)
	} else if receiveData.Type == "pullCmds" {
		serv.handlePullCmds(data, &receiveData)
	} else {
		msgErr("unknow data type "+receiveData.Type, &serverData, nil)
	}
}
