package common

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/services"
)

// GetDocumentAccessKey 获取文档访问密钥
func GetDocumentAccessKey(userId string, documentId string) (*map[string]any, int, error) {
	documentService := services.NewDocumentService()

	document := models.Document{}
	err := documentService.GetById(documentId, &document)
	if err != nil {
		// response.BadRequest(c, "文档不存在")
		return nil, StatusDocumentNotFound, fmt.Errorf("文档不存在")
	}

	var permType models.PermType
	var documentPermission *models.DocumentPermission
	var isPublicPerm bool
	if documentPermission, isPublicPerm, err = documentService.GetDocumentPermissionByDocumentAndUserId(&permType, documentId, userId); err != nil {
		// response.Fail(c, "")
		return nil, http.StatusOK, err
	}
	if permType <= models.PermTypeNone {
		// no permission
		return nil, http.StatusForbidden, fmt.Errorf("Forbidden")
	}
	locked, _ := documentService.GetLocked(documentId)
	if len(locked) > 0 && document.UserId != userId {
		// response.Forbidden(c, "审核不通过")
		return nil, StatusContentReviewFail, fmt.Errorf("审核不通过")
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
			return nil, 0, fmt.Errorf("权限创建失败")
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
		return nil, 0, fmt.Errorf("生成密钥失败")
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
		"bucket_name":       storageConfig.DocumentBucket,
		"endpoint":          documentStorageUrl,
	}, http.StatusOK, nil
}
