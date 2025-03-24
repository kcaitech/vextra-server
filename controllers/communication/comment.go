package communication

import (
	"context"
	"log"

	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/redis"
	"kcaitech.com/kcserver/common/services"
	"kcaitech.com/kcserver/utils/str"
	"kcaitech.com/kcserver/utils/websocket"
)

type commnetServe struct {
	ws   *websocket.Ws
	quit chan struct{}
	// isready bool
	genSId func() string
}

func NewCommentServe(ws *websocket.Ws, userId string, documentId int64, genSId func() string) *commnetServe {

	// 权限校验
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeReadOnly {
		log.Println("NO comment perm", err, permType)
		return nil
	}

	serv := commnetServe{
		ws: ws,
		// isready: false,
		genSId: genSId,
		quit:   make(chan struct{}),
	}
	serv.start(documentId)
	// serv.isready = true
	return &serv
}

func (serv *commnetServe) start(documentId int64) {

	documentIdStr := str.IntToString(documentId)
	// 监控评论变化
	go func() {
		// defer tunnelServer.Close()
		pubsub := redis.Client.Subscribe(context.Background(), "Document Comment[DocumentId:"+documentIdStr+"]")
		defer pubsub.Close()
		channel := pubsub.Channel()
		for {
			select {
			case v, ok := <-channel:
				if !ok {
					break
				}
				serv.send(v.Payload)
			case <-serv.quit:
				return
			}
		}
	}()
}

func (serv *commnetServe) close() {
	close(serv.quit)
}

func (serv *commnetServe) handle(data *TransData, binaryData *([]byte)) {
	// nothing
}

func (serv *commnetServe) send(data string) {
	// jsonData := &Data{}
	// if err := json.Unmarshal([]byte(data), jsonData); err != nil {
	// 	log.Println("comment, redis data wrong", err)
	// 	return
	// }
	sid := serv.genSId()
	serverData := TransData{
		Type:   DataTypes_Comment,
		DataId: sid,
		Data:   data,
	}
	if err := serv.ws.WriteJSONLock(true, &serverData); err != nil {
		log.Println("comment, send data fail", err)
		return
	}
}
