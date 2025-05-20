package ws

import (
	"encoding/json"
	"log"

	"gorm.io/gorm"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/websocket"
)

type ThumbnailServe struct {
	ws         *websocket.Ws
	userId     string
	documentId string
	storage    *storage.StorageClient
	dbModule   *models.DBModule
	review     safereview.Client
}

func NewThumbnailServe(ws *websocket.Ws, userId string, documentId string) *ThumbnailServe {
	serv := ThumbnailServe{
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

func (serv *ThumbnailServe) start(documentId string) {

}

func (serv *ThumbnailServe) close() {

}

func (serv *ThumbnailServe) handle(data *TransData, binaryData *([]byte)) {
	type ThumbnailHeader struct {
		Name string `json:"name"`
	}

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

	if binaryData == nil {
		msgErr("数据错误", &serverData, nil)
		return
	}

	thumbnailHeader := ThumbnailHeader{}
	err := json.Unmarshal([]byte(data.Data), &thumbnailHeader)
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

	// 删除旧的缩略图
	thumbnailDir := document.Path + "/thumbnail/"
	objects := serv.storage.Bucket.ListObjects(thumbnailDir)
	for object := range objects {
		if object.Err != nil {
			log.Println("列出缩略图异常：", object.Err)
			continue
		}
		if err := serv.storage.Bucket.DeleteObject(object.Key); err != nil {
			log.Println("删除旧缩略图异常：", err)
		}
	}

	path := document.Path + "/thumbnail/" + thumbnailHeader.Name
	log.Println("开始上传缩略图", serv.documentId, path)
	if _, err = serv.storage.Bucket.PutObjectByte(path, *binaryData); err != nil {
		msgErr("上传失败", &serverData, &err)
		return
	}
	log.Println("缩略图上传成功", serv.documentId, path)

	_ = serv.ws.WriteJSON(&serverData)
	serv.dbModule.DB.Model(&document).Where("id = ?", serv.documentId).UpdateColumn("size", gorm.Expr("size + ?", len(*binaryData)))

	if serv.review != nil {
		go reviewResGo(serv.documentId, thumbnailHeader.Name, *binaryData)
	}
}
