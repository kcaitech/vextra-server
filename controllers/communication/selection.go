package communication

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/redis"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/str"
	"kcaitech.com/kcserver/utils/websocket"
)

type DocSelectionData struct {
	SelectPageId      string          `json:"select_page_id,omitempty"`
	SelectShapeIdList []string        `json:"select_shape_id_list"`
	HoverShapeId      string          `json:"hover_shape_id,omitempty"`
	CursorStart       int             `json:"cursor_start,omitempty"`
	CursorEnd         int             `json:"cursor_end,omitempty"`
	CursorAtBefore    bool            `json:"cursor_at_before,omitempty"`
	UserId            string          `json:"user_id,omitempty"`
	Permission        models.PermType `json:"permission,omitempty"`
	Avatar            string          `json:"avatar,omitempty"`
	Nickname          string          `json:"nickname,omitempty"`
	EnterTime         int64           `json:"enter_time,omitempty"`
}

type DocSelectionOpType uint8

const (
	DocSelectionOpTypeUpdate DocSelectionOpType = iota
	DocSelectionOpTypeExit
)

type DocSelectionOpData struct {
	Type   DocSelectionOpType `json:"type"`
	UserId string             `json:"user_id"`
	Data   *DocSelectionData  `json:"data,omitempty"`
}

type selectionServe struct {
	ws         *websocket.Ws
	quit       chan struct{}
	genSId     func() string
	documentId int64
	userId     string
	user       *models.UserProfile
	enterTime  int64
	permType   models.PermType
	redis      *redis.RedisDB
}

func NewSelectionServe(ws *websocket.Ws, token, userId string, documentId int64, genSId func() string) *selectionServe {

	// 权限校验
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeReadOnly {
		log.Println("NO comment perm", err, permType)
		return nil
	}

	jwtClient := services.GetJWTClient()
	userInfo, err := jwtClient.GetUserInfo(token)
	if err != nil {
		log.Println("document comment ws建立失败，用户不存在", err, userId)
		return nil
	}
	// userId := userInfo.UserID
	// userIdStr := str.IntToString(userId)
	// userService := services.NewUserService()
	// user := models.User{}
	// if err := userService.GetById(userId, &user); err != nil {
	// 	// serverCmd.Message = "通道建立失败，用户信息错误"
	// 	// _ = clientWs.WriteJSON(&serverCmd)
	// 	log.Println("document comment ws建立失败，用户不存在", err, userId)
	// 	return nil
	// }
	userProfile := models.UserProfile{
		UserId:   userInfo.UserID,
		Nickname: userInfo.Profile.Nickname,
		Avatar:   userInfo.Profile.Avatar,
	}

	serv := selectionServe{
		ws:         ws,
		genSId:     genSId,
		documentId: documentId,
		userId:     userInfo.UserID,
		user:       &userProfile,
		enterTime:  time.Now().UnixNano() / 1000000,
		permType:   permType,
		quit:       make(chan struct{}),
		redis:      services.GetRedisDB(),
	}
	serv.start(documentId)
	// serv.isready = true
	return &serv
}

func (serv *selectionServe) start(documentId int64) {

	documentIdStr := str.IntToString(documentId)
	// 监控选区变化
	go func() {
		// defer tunnelServer.Close()
		subscribe := serv.redis.Client.Subscribe(context.Background(), "Document Selection[DocumentId:"+documentIdStr+"]")
		defer subscribe.Close()
		channel := subscribe.Channel()
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

func (serv *selectionServe) close() {
	userIdStr := (serv.userId)
	documentIdStr := str.IntToString(serv.documentId)
	docSelectionOpData := &DocSelectionOpData{
		Type:   DocSelectionOpTypeExit,
		UserId: userIdStr,
	}
	if data, err := json.Marshal(docSelectionOpData); err == nil {
		serv.redis.Client.HDel(context.Background(), "Document Selection Data[DocumentId:"+documentIdStr+"]", userIdStr)
		serv.redis.Client.Publish(context.Background(), "Document Selection[DocumentId:"+documentIdStr+"]", string(data))
	}
	close(serv.quit)
}

func (serv *selectionServe) handle(data *TransData, binaryData *([]byte)) {
	serverData := TransData{}
	serverData.Type = data.Type
	serverData.DataId = data.DataId

	userIdStr := (serv.userId)
	documentIdStr := str.IntToString(serv.documentId)
	msgErr := func(msg string, serverData *TransData, err *error) {
		serverData.Err = msg
		if err != nil {
			log.Println(msg, *err)
		} else {
			log.Println(msg)
		}
		_ = serv.ws.WriteJSON(serverData)
	}
	selectionData := &DocSelectionData{}
	if err := json.Unmarshal([]byte(data.Data), selectionData); err != nil {
		msgErr("document selection数据解码错误", &serverData, &err)
		return
	}

	selectionData.UserId = userIdStr
	selectionData.Permission = serv.permType
	// todo
	// selectionData.Avatar = config.Config.StorageUrl.Attatch + serv.user.Avatar
	selectionData.Avatar = serv.user.Avatar
	selectionData.Nickname = serv.user.Nickname
	selectionData.EnterTime = serv.enterTime
	selectionDataJson, _ := json.Marshal(selectionData)
	docSelectionOpData := &DocSelectionOpData{
		Type:   DocSelectionOpTypeUpdate,
		UserId: userIdStr,
		Data:   selectionData,
	}
	if docSelectionOpDataJson, err := json.Marshal(docSelectionOpData); err != nil {
		msgErr("document selection数据解码错误", &serverData, &err)
		return
	} else {
		serv.redis.Client.HSet(context.Background(), "Document Selection Data[DocumentId:"+documentIdStr+"]", userIdStr, string(selectionDataJson))
		serv.redis.Client.Expire(context.Background(), "Document Selection Data[DocumentId:"+documentIdStr+"]", time.Hour*1)
		serv.redis.Client.Publish(context.Background(), "Document Selection[DocumentId:"+documentIdStr+"]", string(docSelectionOpDataJson))
		serv.ws.WriteJSON(serverData)
	}
}

func (serv *selectionServe) send(data string) {
	// jsonData := &Data{}
	// if err := json.Unmarshal([]byte(data), jsonData); err != nil {
	// 	log.Println("comment, redis data wrong", err)
	// 	return
	// }
	sid := serv.genSId()
	serverData := TransData{
		Type:   DataTypes_Selection,
		DataId: sid,
		Data:   data,
	}
	if err := serv.ws.WriteJSONLock(true, &serverData); err != nil {
		log.Println("selection, send data fail", err)
	}
}
