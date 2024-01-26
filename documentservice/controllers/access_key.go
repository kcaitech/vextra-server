package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log"
	"protodesign.cn/kcserver/common"
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
	var documentPermission *models.DocumentPermission
	var isPublicPerm bool
	if documentPermission, isPublicPerm, err = documentService.GetDocumentPermissionByDocumentAndUserId(&permType, documentId, userId); err != nil {
		response.Fail(c, "")
		return
	}
	if permType <= models.PermTypeNone {
		response.Forbidden(c, "")
		return
	}
	if !document.LockedAt.IsZero() && document.UserId != userId {
		response.Forbidden(c, "审核不通过")
		return
	}
	if documentPermission == nil && isPublicPerm {
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

	accessKeyValue, err := storage.Bucket.GenerateAccessKey(
		document.Path+"/*",
		base.AuthOpGetObject|base.AuthOpListObject,
		3600,
		"U"+str.IntToString(userId)+"D"+str.IntToString(documentId),
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
		if !errors.Is(err, services.ErrRecordNotFound) {
			log.Println("documentAccessRecordService.Get错误：" + err.Error())
		} else {
			_ = documentService.DocumentAccessRecordService.Create(&models.DocumentAccessRecord{
				DocumentId:     documentId,
				UserId:         userId,
				LastAccessTime: now,
			})
		}
	} else {
		_, _ = documentService.DocumentAccessRecordService.UpdateColumns(map[string]any{
			"last_access_time": now,
			"deleted_at":       nil,
		}, "id = ?", documentAccessRecord.Id, &services.Unscoped{})
	}

	storageConfig := storage.Bucket.GetConfig()

	response.Success(c, map[string]any{
		"access_key":        accessKeyValue.AccessKey,
		"secret_access_key": accessKeyValue.SecretAccessKey,
		"session_token":     accessKeyValue.SessionToken,
		"signer_type":       accessKeyValue.SignerType,
		"provider":          storageConfig.Provider,
		"region":            storageConfig.Region,
		"bucket_name":       storageConfig.BucketName,
		"endpoint":          common.StorageHost,
	})
}
