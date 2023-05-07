package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/common/storage"
	"protodesign.cn/kcserver/utils/storage/base"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
	"time"
)

// GetDocumentAccessKey 获取文档访问密钥
func GetDocumentAccessKey(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	documentService := services.NewDocumentService()

	docId := str.DefaultToInt(c.Query("doc_id"), 0)
	if docId == 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}

	document := models.Document{}
	err = documentService.GetById(docId, &document)
	if err != nil {
		response.BadRequest(c, "文档不存在")
		return
	}

	if document.UserId != userId {
		response.Forbidden(c, "")
		return
	}

	accessKey, err := storage.Bucket.GenerateAccessKey(
		document.Path+"/*",
		base.AuthOpGetObject|base.AuthOpListObject,
		3600,
	)
	if err != nil {
		log.Println("生成密钥失败" + err.Error())
		response.Fail(c, "生成密钥失败")
		return
	}

	now := myTime.Time(time.Now())
	documentAccessRecord := models.DocumentAccessRecord{}
	err = documentService.DocumentAccessRecordService.Get(&documentAccessRecord, "document_id = ? and user_id = ?", docId, userId)
	if err != nil {
		if err != services.ErrRecordNotFound {
			log.Println("documentAccessRecordService.Get错误：" + err.Error())
		} else {
			_ = documentService.DocumentAccessRecordService.Create(&models.DocumentAccessRecord{
				DocumentId:     docId,
				UserId:         userId,
				LastAccessTime: now,
			})
		}
	} else {
		documentAccessRecord.LastAccessTime = now
		_ = documentService.DocumentAccessRecordService.UpdatesById(documentAccessRecord.Id, &documentAccessRecord)
	}

	response.Success(c, accessKey)
}
