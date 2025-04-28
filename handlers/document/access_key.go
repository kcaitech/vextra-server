package document

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
)

// GetDocumentAccessKey 获取文档访问密钥
func GetDocumentAccessKey(c *gin.Context) {
	userId, msg := utils.GetUserId(c)
	if msg != nil {
		response.Unauthorized(c)
		return
	}

	documentId := (c.Query("doc_id"))
	if documentId == "" {
		response.BadRequest(c, "参数错误：doc_id")
		return
	}

	key, msg, code := GetDocumentAccessKey1(userId, documentId)
	if msg == nil {
		response.Success(c, key)
	} else if code == http.StatusUnauthorized {
		response.Unauthorized(c)
	} else {
		response.ServerError(c, msg.Error())
	}
}

// GetDocumentAccessKey 获取文档访问密钥
func GetDocumentAccessKey1(userId string, documentId string) (*map[string]any, error, int) {
	documentService := services.NewDocumentService()

	document := models.Document{}
	err := documentService.GetById(documentId, &document)
	if err != nil {
		// response.BadRequest(c, "文档不存在")
		return nil, fmt.Errorf("文档不存在"), response.StatusDocumentNotFound
	}

	var permType models.PermType
	var documentPermission *models.DocumentPermission
	var isPublicPerm bool
	if documentPermission, isPublicPerm, err = documentService.GetDocumentPermissionByDocumentAndUserId(&permType, documentId, userId); err != nil {
		// response.Fail(c, "")
		return nil, err, http.StatusOK
	}
	if permType <= models.PermTypeNone {
		// no permission
		return nil, fmt.Errorf("Forbidden"), http.StatusForbidden
	}
	locked, _ := documentService.GetLocked(documentId)
	if len(locked) > 0 && document.UserId != userId {
		// response.Forbidden(c, "审核不通过")
		return nil, fmt.Errorf("审核不通过"), response.StatusContentReviewFail
	}
	log.Println("documentPermission", documentPermission, isPublicPerm)
	if documentPermission == nil && isPublicPerm {
		if err := documentService.DocumentPermissionService.Create(&models.DocumentPermission{
			ResourceType: models.ResourceTypeDoc,
			ResourceId:   documentId,
			GranteeId:    userId,
			PermType:     permType,
		}); err != nil {
			// response.Fail(c, "权限创建失败")
			return nil, fmt.Errorf("权限创建失败"), 0
		}
	}

	_storage := services.GetStorageClient()
	accessKeyValue, err := _storage.Bucket.GenerateAccessKey(
		document.Path+"/*",
		storage.AuthOpGetObject|storage.AuthOpListObject,
		3600,
		"U"+(userId)+"D"+(documentId),
	)
	if err != nil {
		log.Println("生成密钥失败", err)
		// response.Fail(c, "生成密钥失败")
		return nil, fmt.Errorf("生成密钥失败"), 0
	}

	// 插入/更新访问记录
	now := (time.Now())
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

	storageConfig := _storage.Bucket.GetConfig()
	documentStorageUrl := services.GetConfig().StorageUrl.Document
	return &map[string]any{
		"access_key":        accessKeyValue.AccessKey,
		"secret_access_key": accessKeyValue.SecretAccessKey,
		"session_token":     accessKeyValue.SessionToken,
		"signer_type":       accessKeyValue.SignerType,
		"provider":          services.GetConfig().Storage.Provider,
		"region":            storageConfig.Region,
		"bucket_name":       storageConfig.BucketName,
		"endpoint":          documentStorageUrl,
	}, nil, http.StatusOK
}
