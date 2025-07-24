package common

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
)

type AccessKeyInfo struct {
	AccessKey       string `json:"access_key"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
	SignerType      int    `json:"signer_type"`
	Provider        string `json:"provider"`
	Region          string `json:"region"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
}

// GetDocumentAccessKey 获取文档访问密钥
func GetDocumentAccessKey(userId string, documentId string) (*AccessKeyInfo, int, error) {
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
	return &AccessKeyInfo{
		AccessKey:       accessKeyValue.AccessKey,
		SecretAccessKey: accessKeyValue.SecretAccessKey,
		SessionToken:    accessKeyValue.SessionToken,
		SignerType:      accessKeyValue.SignerType,
		Provider:        string(services.GetConfig().Storage.Provider),
		Region:          storageConfig.Region,
		BucketName:      storageConfig.DocumentBucket,
		Endpoint:        documentStorageUrl,
	}, http.StatusOK, nil
}

type ThumbnailResponse struct {
	AccessKeyInfo
	ObjectKey string `json:"object_key"`
}

func GetDocumentThumbnailAccessKey(c *gin.Context, documentId string, _storage *storage.StorageClient) (*ThumbnailResponse, int, error) {
	documentService := services.NewDocumentService()
	userId, err := utils.GetUserId(c)
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}
	var permType models.PermType
	if _, _, err = documentService.GetDocumentPermissionByDocumentAndUserId(&permType, documentId, userId); err != nil {
		return nil, http.StatusOK, err
	}
	if permType <= models.PermTypeNone {
		return nil, http.StatusForbidden, fmt.Errorf("Forbidden")
	}
	accessKeyInfo, err := GetDocumentThumbnailAccessKeyNoCheckAuth(c, documentId, _storage)
	if err != nil {
		return nil, http.StatusOK, err
	}
	return accessKeyInfo, http.StatusOK, nil
}

func GetDocumentThumbnailAccessKeyNoCheckAuth(c *gin.Context, documentId string, _storage *storage.StorageClient) (*ThumbnailResponse, error) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		return nil, err
	}
	objects := _storage.Bucket.ListObjects(documentId + "/thumbnail/")
	for object := range objects {
		if object.Err != nil {
			continue
		}

		// 生成预签名URL，有效期1小时
		reqParams := make(url.Values)
		reqParams.Set("response-content-disposition", "inline")
		accessKeyValue, err := _storage.Bucket.GenerateAccessKey(object.Key, storage.AuthOpGetObject, 3600, "U"+(userId)+"D"+(documentId))
		if err != nil {
			return nil, err
		}

		storageConfig := _storage.Bucket.GetConfig()
		documentStorageUrl := services.GetConfig().StorageUrl.Document
		return &ThumbnailResponse{
			AccessKeyInfo: AccessKeyInfo{
				AccessKey:       accessKeyValue.AccessKey,
				SecretAccessKey: accessKeyValue.SecretAccessKey,
				SessionToken:    accessKeyValue.SessionToken,
				SignerType:      accessKeyValue.SignerType,
				Provider:        string(services.GetConfig().Storage.Provider),
				Region:          storageConfig.Region,
				BucketName:      storageConfig.DocumentBucket,
				Endpoint:        documentStorageUrl,
			},
			ObjectKey: object.Key,
		}, nil
	}
	return nil, errors.New("thumbnail not found")
}
