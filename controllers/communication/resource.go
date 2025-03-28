package communication

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/services"
	myTime "kcaitech.com/kcserver/utils/time"
	"kcaitech.com/kcserver/utils/websocket"
)

type resourceServe struct {
	ws         *websocket.Ws
	userId     string
	documentId int64
	storage    *storage.StorageClient
	dbModule   *models.DBModule
	review     safereview.Client
}

func NewResourceServe(ws *websocket.Ws, userId string, documentId int64) *resourceServe {

	// 权限校验
	// var permType models.PermType
	// if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeReadOnly {
	// 	log.Println("NO comment perm", err, permType)
	// 	return nil
	// }

	serv := resourceServe{
		ws:         ws,
		userId:     userId,
		documentId: documentId,
		storage:    services.GetStorageClient(),
		dbModule:   services.GetDBModule(),
		review:     services.GetSafereviewClient(),
	}
	serv.start(documentId)
	return &serv
}

func (serv *resourceServe) start(documentId int64) {

}

func (serv *resourceServe) close() {

}

func (serv *resourceServe) handle(data *TransData, binaryData *([]byte)) {

	type ResourceHeader struct {
		Name string `json:"name"`
	}

	serverData := TransData{}
	serverData.Type = data.Type
	serverData.DataId = data.DataId
	msgErr := func(msg string, serverData *TransData, err *error) {
		serverData.Err = msg
		if err != nil {
			log.Println(msg, *err)
		} else {
			log.Println(msg)
		}
		_ = serv.ws.WriteJSON(serverData)
	}

	if binaryData == nil {
		msgErr("数据错误", &serverData, nil)
		return
	}

	resourceHeader := ResourceHeader{}
	err := json.Unmarshal([]byte(data.Data), &resourceHeader)
	if err != nil {
		msgErr("数据错误", &serverData, nil)
		return
	}

	// 权限校验
	documentService := services.NewDocumentService()
	var permType models.PermType
	if err := documentService.GetPermTypeByDocumentAndUserId(&permType, serv.documentId, serv.userId); err != nil || permType < models.PermTypeEditable {
		msgErr("无权限", &serverData, &err)
		return
	}

	// 获取文档信息
	var document models.Document
	if documentService.GetById(serv.documentId, &document) != nil {
		msgErr("文档不存在", &serverData, nil)
		return
	}

	path := document.Path + "/medias/" + resourceHeader.Name
	log.Println("开始上传", serv.documentId, path)
	if _, err = serv.storage.Bucket.PutObjectByte(path, *binaryData); err != nil {

		msgErr("上传失败", &serverData, &err)
		return
	}
	log.Println("上传成功", serv.documentId, path)

	_ = serv.ws.WriteJSON(&serverData)
	serv.dbModule.DB.Model(&document).Where("id = ?", serv.documentId).UpdateColumn("size", gorm.Expr("size + ?", len(*binaryData)))

	if serv.review != nil {
		go reviewResGo(serv.documentId, *binaryData)
	}

}

func reviewResGo(documentId int64, resourceData []byte) {
	reviewClient := services.GetSafereviewClient()
	if reviewClient == nil {
		return
	}
	base64Str := base64.StdEncoding.EncodeToString(resourceData)
	reviewResponse, err := (reviewClient).ReviewPictureFromBase64(base64Str)
	if err != nil {
		log.Println("图片审核失败", err)
		return
	} else if reviewResponse.Status != safereview.ReviewImageResultPass {
		log.Println("图片审核不通过", err, reviewResponse)
		documentService := services.NewDocumentService()
		// 审核不通过，锁定文档
		lockedInfo, err := documentService.GetLocked(documentId)
		if err != nil {
			log.Println("获取锁定信息失败", err)
			return
		}
		if lockedInfo == nil {
			lockedInfo = &models.DocumentLock{
				DocumentId: documentId,
				// LockedAt:   myTime.Time(time.Now()),
				LockedReason: "[图片审核不通过：" + reviewResponse.Reason + "]",
			}
		} else {
			lockedInfo.LockedAt = myTime.Time(time.Now())
			lockedInfo.LockedReason += "[图片审核不通过：" + reviewResponse.Reason + "]"
		}
		documentService.UpdateLocked(documentId, time.Now(), lockedInfo.LockedReason, lockedInfo.LockedWords)
	}
}
