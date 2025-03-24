package communication

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/safereview"
	safereviewBase "kcaitech.com/kcserver/common/safereview/base"
	"kcaitech.com/kcserver/common/services"
	"kcaitech.com/kcserver/common/storage"
	myTime "kcaitech.com/kcserver/utils/time"
	"kcaitech.com/kcserver/utils/websocket"
)

type resourceServe struct {
	ws         *websocket.Ws
	userId     string
	documentId int64
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
	if _, err = storage.Bucket.PutObjectByte(path, *binaryData); err != nil {

		msgErr("上传失败", &serverData, &err)
		return
	}
	log.Println("上传成功", serv.documentId, path)

	_ = serv.ws.WriteJSON(&serverData)
	documentService.DB.Model(&document).Where("id = ?", serv.documentId).UpdateColumn("size", gorm.Expr("size + ?", len(*binaryData)))

	go func() {
		base64Str := base64.StdEncoding.EncodeToString(*binaryData)
		reviewResponse, err := safereview.Client.ReviewPictureFromBase64(base64Str)
		if err != nil {
			log.Println("图片审核失败", err)
			return
		} else if reviewResponse.Status != safereviewBase.ReviewImageResultPass {
			log.Println("图片审核不通过", err, reviewResponse)
			document.LockedAt = myTime.Time(time.Now())
			document.LockedReason += "[图片审核不通过：" + reviewResponse.Reason + "]"
			_, _ = documentService.UpdatesById(serv.documentId, &document)
		}
	}()

}
