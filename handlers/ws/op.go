package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redsync/redsync/v4"
	autoupdate "kcaitech.com/kcserver/handlers/document"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/mongo"
	"kcaitech.com/kcserver/providers/redis"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/sliceutil"
	"kcaitech.com/kcserver/utils/websocket"
)

type ReceiveData struct {
	Type string `json:"type"` // commit pullCmds
	Cmds string `json:"cmds"` // commit
	From int    `json:"from"` // pullCmds
	To   int    `json:"to"`   // pullCmds
}
type Cmd = models.Cmd
type CmdItem = models.CmdItem
type ReceiveCmd = models.Cmd

type SendData struct {
	Type       string   `json:"type"`                  // pullCmdsResult update errorInvalidParams errorNoPermission errorInsertFailed errorPullCmdsFailed
	CmdsData   string   `json:"cmds_data,omitempty"`   // pullCmdsResult update
	From       int      `json:"from,omitempty"`        // pullCmdsResult errorPullCmdsFailed
	To         int      `json:"to,omitempty"`          // pullCmdsResult errorPullCmdsFailed
	PreviousId string   `json:"previous_id,omitempty"` // pullCmdsResult
	CmdIdList  []string `json:"cmd_id_list,omitempty"` // errorInsertFailed
	// Data       any      `json:"data,omitempty"`        // errorInsertFailed
}

type opServe struct {
	ws   *websocket.Ws
	quit chan struct{}
	// isready bool
	genSId     func() string
	mutex      *redsync.Mutex
	permType   models.PermType
	documentId string
	userId     string
	// dbModule   *models.DBModule
	redis *redis.RedisDB
	// mongo      *mongo.MongoDB
}

// 从redis中获取最后一条cmd的id，若redis中没有则从mongodb中获取
func (serv *opServe) getPreviousId(documentId string) (uint, error) {
	// documentIdStr := str.IntToString(documentId)
	previousId, err := serv.redis.Client.Get(context.Background(), "Document lastCmdVerId[DocumentId:"+documentId+"]").Int() // redis也是最终一致性
	if err == nil {
		return uint(previousId), nil
	}

	if err != redis.Nil { // todo redis获取不到时，应该去mongodb获取，并更新到redis？
		log.Println("Document lastCmdVerId[DocumentId:"+documentId+"]"+"获取失败", err)
		return 0, err
	}

	cmdService := services.GetCmdService()
	cmdItem, err := cmdService.GetLastCmdItem(documentId)
	if err != nil {
		return 0, err
	}
	if cmdItem != nil {
		return cmdItem.VerId, nil
	}
	return 0, nil
}

func NewOpServe(ws *websocket.Ws, userId string, documentId string, versionId string, lastCmdVerId uint, genSId func() string) *opServe {

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
	if err := documentService.GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeReadOnly {
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
	lockedInfo, err := documentService.GetLocked(documentId)
	if err != nil {
		log.Println("document ws建立失败，获取审核信息失败", documentId)
		return nil
	}
	if len(lockedInfo) > 0 && document.UserId != userId {
		// serverCmd.Message = "通道建立失败，审核不通过"
		// _ = clientWs.WriteJSON(&serverCmd)
		log.Println("document ws建立失败，审核不通过", documentId)
		return nil
	}

	// documentIdStr := str.IntToString(documentId)
	redis := services.GetRedisDB()
	mutex := redis.RedSync.NewMutex("Document Op Mutex[DocumentId:"+documentId+"]", redsync.WithExpiry(time.Second*10))

	serv := opServe{
		ws: ws,
		// isready: false,
		genSId:     genSId,
		mutex:      mutex,
		permType:   permType,
		documentId: documentId,
		userId:     userId,
		quit:       make(chan struct{}),
		// dbModule:   dbModule,
		redis: redis,
		// mongo:      mongo,
	}

	if lastCmdVerId == 0 {
		lastCmdVerId = (documentVersion.LastCmdVerId)
	}

	serv.start(documentId, lastCmdVerId)
	// serv.isready = true
	return &serv
}

func (serv *opServe) start(documentId string, lastCmdVersion uint) {
	go func() {
		cmdService := services.GetCmdService()
		cmdItemList, err := cmdService.GetCmdItemsFromStart(documentId, lastCmdVersion)
		if err != nil {
			log.Println("cmdList查询失败", err)
			return
		}
		cmdItemListData, err := json.Marshal(cmdItemList)
		if err != nil {
			log.Println("json编码错误 cmdItemsData", err)
			return
		}
		log.Println("Document Op[DocumentId:"+documentId+"]"+"获取成功", len(cmdItemList), lastCmdVersion)
		serv.send(string(cmdItemListData))

		// documentIdStr := str.IntToString(documentId)
		pubsub := serv.redis.Client.Subscribe(context.Background(), "Document Op[DocumentId:"+documentId+"]")
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
				return
			}
		}
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
		Type:     "update",
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
		serverData.Msg = msg
		if err != nil {
			log.Println(msg, *err)
		} else {
			log.Println(msg)
		}
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
	previousId, err := serv.getPreviousId(serv.documentId)
	if err != nil {
		msgErr("get previous id failed", &serverData, &err)
		return
	}

	batchStartId := previousId + 1
	batchLength := len(cmds)

	cmdItemList := sliceutil.MapT(func(cmd ReceiveCmd) CmdItem {
		previousId += 1
		return CmdItem{
			VerId:        previousId, // 有redis锁 todo 不对，不能保证唯一
			BatchStartId: batchStartId,
			BatchLength:  uint(batchLength),
			DocumentId:   serv.documentId,
			UserId:       serv.userId,
			Cmd:          cmd,
		}
	}, cmds...)

	cmdItemListData, err := json.Marshal(cmdItemList)
	if err != nil {
		msgErr("json marshal failed", &serverData, &err)
		return
	}
	documentId := (serv.documentId)

	// 先删除，防止插入失败时无法正确更新lastCmdVerId
	if _, err = serv.redis.Client.Del(context.Background(), "Document lastCmdVerId[DocumentId:"+documentId+"]").Result(); err != nil {
		msgErr("redis del fail", &serverData, &err)
		return
	}

	cmdServices := services.GetCmdService()
	insertRes, err := cmdServices.SaveCmdItems(cmdItemList)

	if err != nil && mongo.IsDuplicateKeyError(err) {
		log.Println("重复数据插入失败1", err)

		index := len(insertRes.InsertedIDs)
		duplicateCmdCmdId := cmdItemList[index].Cmd.Id
		duplicateCmdDocumentId := cmdItemList[index].DocumentId
		duplicateCmd, err := cmdServices.GetCmd(duplicateCmdDocumentId, duplicateCmdCmdId)
		if err != nil {
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
		msgErr("数据插入失败", &serverData, &err)
		return
	} else {
		if _, err = serv.redis.Client.Set(context.Background(), "Document lastCmdVerId[DocumentId:"+documentId+"]", previousId, time.Hour*1).Result(); err != nil {
			log.Println("Document lastCmdVerId[DocumentId:"+documentId+"]"+"设置失败", err)
			// return errors.New("数据插入失败")
		}
		log.Println("广播", string(cmdItemListData))
		serv.redis.Client.Publish(context.Background(), "Document Op[DocumentId:"+documentId+"]", cmdItemListData) // 通知客户端是通过redis订阅来触发的
		// return nil
		// debug
		// log.Panic()
		_ = serv.ws.WriteJSON(serverData) // sucess
		go autoupdate.AutoUpdate(serv.documentId, services.GetConfig())
	}
}

func (serv *opServe) handlePullCmds(data *TransData, receiveData *ReceiveData) {
	serverData := TransData{}
	serverData.Type = data.Type
	serverData.DataId = data.DataId
	msgErr := func(msg string, serverData *TransData, err *error) {
		serverData.Msg = msg
		log.Println(msg, err)
		_ = serv.ws.WriteJSON(serverData)
	}

	var cmdItemList []CmdItem
	var err error
	fromId := (receiveData.From)
	toId := (receiveData.To)
	log.Println("handlePullCmds", receiveData)

	cmdsService := services.GetCmdService()
	if toId == 0 {
		cmdItemList, err = cmdsService.GetCmdItemsFromStart(serv.documentId, uint(fromId))
	} else {
		cmdItemList, err = cmdsService.GetCmdItems(serv.documentId, uint(fromId), uint(toId))
	}
	if err != nil {
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
	// 	log.Println(documentId, "lastCmdVerId查询失败", err)
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
	log.Println("pullCmdsEnd", string(sendDataStr))
	serverData.Data = string(sendDataStr)
	err = serv.ws.WriteJSON(serverData)
	if err != nil {
		log.Println("数据发送失败", err)
		// msgErr("数据发送失败", &serverData, &err)
		return
	} else {
		log.Println("pullCmdsEnd 数据发送成功", sendData)
	}

}

func (serv *opServe) handle(data *TransData, binaryData *([]byte)) {
	serverData := TransData{}
	serverData.Type = data.Type
	serverData.DataId = data.DataId

	msgErr := func(msg string, serverData *TransData, err *error) {
		serverData.Msg = msg
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
