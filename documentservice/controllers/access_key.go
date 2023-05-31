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

	documentId := str.DefaultToInt(c.Query("doc_id"), 0)
	if documentId == 0 {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}

	document := models.Document{}
	err = documentService.GetById(documentId, &document)
	if err != nil {
		response.BadRequest(c, "文档不存在")
		return
	}

	var permType models.PermType
	var isCreator bool
	var documentPermission *models.DocumentPermission
	if documentPermission, isCreator, err = documentService.GetDocumentPermissionByDocumentAndUserId(&permType, documentId, userId); err != nil {
		response.Fail(c, "")
		return
	}
	if permType <= models.PermTypeNone {
		response.Forbidden(c, "")
		return
	}
	if documentPermission == nil && !isCreator {
		if err := documentService.DocumentPermissionService.Create(&models.DocumentPermission{
			ResourceType: models.ResourceTypeDoc,
			ResourceId:   documentId,
			GranteeType:  models.GranteeTypeExternal,
			GranteeId:    userId,
			PermType:     permType,
		}); err != nil {
			response.Fail(c, "权限创建失败")
			return
		}
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

	// 插入/更新访问记录
	now := myTime.Time(time.Now())
	documentAccessRecord := models.DocumentAccessRecord{}
	err = documentService.DocumentAccessRecordService.Get(&documentAccessRecord, "document_id = ? and user_id = ?", documentId, userId, &services.Unscoped{})
	if err != nil {
		if err != services.ErrRecordNotFound {
			log.Println("documentAccessRecordService.Get错误：" + err.Error())
		} else {
			_ = documentService.DocumentAccessRecordService.Create(&models.DocumentAccessRecord{
				DocumentId:     documentId,
				UserId:         userId,
				LastAccessTime: now,
			})
		}
	} else {
		_ = documentService.DocumentAccessRecordService.UpdateColumns(map[string]any{
			"last_access_time": now,
			"deleted_at":       nil,
		}, "id = ?", documentAccessRecord.Id, &services.Unscoped{})
	}

	response.Success(c, accessKey)
}
