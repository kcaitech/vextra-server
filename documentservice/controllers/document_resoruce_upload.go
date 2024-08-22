package controllers

import (
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/safereview"
	safereviewBase "protodesign.cn/kcserver/common/safereview/base"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/storage"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
	"protodesign.cn/kcserver/utils/websocket"
	"time"
)

func UploadDocumentResource(c *gin.Context) {
	type Data map[string]any

	type Header struct {
		UserId     string `json:"user_id"`
		DocumentId string `json:"document_id"`
	}

	type ResponseStatusType string

	const (
		ResponseStatusSuccess ResponseStatusType = "success"
		ResponseStatusFail    ResponseStatusType = "fail"
	)

	type Response struct {
		Status  ResponseStatusType `json:"status,omitempty"`
		Message string             `json:"message,omitempty"`
		Data    Data               `json:"data,omitempty"`
	}

	type ResourceHeader struct {
		Name string `json:"name"`
	}

	ws, err := websocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		response.Fail(c, "建立ws连接失败："+err.Error())
		return
	}
	defer ws.Close()

	resp := Response{
		Status: ResponseStatusFail,
	}

	header := Header{}
	if err := ws.ReadJSON(&header); err != nil {
		resp.Message = "Header结构错误"
		_ = ws.WriteJSON(&resp)
		log.Println("Header结构错误", err)
		return
	}
	userId := str.DefaultToInt(header.UserId, 0)
	documentId := str.DefaultToInt(header.DocumentId, 0)
	if userId <= 0 || documentId <= 0 {
		resp.Message = "参数错误"
		_ = ws.WriteJSON(&resp)
		log.Println("参数错误", userId, documentId)
		return
	}

	// 权限校验
	var permType models.PermType
	if err := services.NewDocumentService().GetPermTypeByDocumentAndUserId(&permType, documentId, userId); err != nil || permType < models.PermTypeEditable {
		resp.Message = "权限校验失败"
		if err != nil {
			resp.Message += "，无权限"
		}
		_ = ws.WriteJSON(&resp)
		return
	}

	// 获取文档信息
	documentService := services.NewDocumentService()
	var document models.Document
	if documentService.GetById(documentId, &document) != nil {
		resp.Message = "文档不存在"
		_ = ws.WriteJSON(&resp)
		log.Println("文档不存在", documentId)
		return
	}

	for {
		resp = Response{
			Status: ResponseStatusFail,
		}
		resourceHeader := ResourceHeader{}
		if err := ws.ReadJSON(&resourceHeader); err != nil || resourceHeader.Name == "" {
			if errors.Is(err, websocket.ErrClosed) {
				return
			}
			resp.Message = "ResourceHeader结构错误"
			_ = ws.WriteJSON(&resp)
			log.Println("ResourceHeader结构错误", err)
			return
		}
		messageType, resourceData, err := ws.ReadMessage()
		if messageType != websocket.MessageTypeBinary || err != nil || len(resourceData) == 0 {
			if errors.Is(err, websocket.ErrClosed) {
				return
			}
			resp.Message = "ResourceData结构错误"
			_ = ws.WriteJSON(&resp)
			log.Println("ResourceData结构错误", err, len(resourceData))
			return
		}
		path := document.Path + "/medias/" + resourceHeader.Name
		log.Println("开始上传", documentId, path)
		if _, err = storage.Bucket.PutObjectByte(path, resourceData); err != nil {
			resp.Message = "上传失败"
			_ = ws.WriteJSON(&resp)
			log.Println("上传失败", documentId, path, err)
			return
		}
		log.Println("上传成功", documentId, path)
		resp.Status = ResponseStatusSuccess
		_ = ws.WriteJSON(&resp)
		documentService.DB.Model(&document).Where("id = ?", documentId).UpdateColumn("size", gorm.Expr("size + ?", len(resourceData)))

		go func() {
			base64Str := base64.StdEncoding.EncodeToString(resourceData)
			reviewResponse, err := safereview.Client.ReviewPictureFromBase64(base64Str)
			if err != nil {
				log.Println("图片审核失败", err)
				return
			} else if reviewResponse.Status != safereviewBase.ReviewImageResultPass {
				log.Println("图片审核不通过", err, reviewResponse)
				document.LockedAt = myTime.Time(time.Now())
				document.LockedReason += "[图片审核不通过：" + reviewResponse.Reason + "]"
				_, _ = documentService.UpdatesById(documentId, &document)
			}
		}()
	}
}
