package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	autoupdate "kcaitech.com/kcserver/handlers/document"
	"kcaitech.com/kcserver/scripts/migrate-20250702/config"
	"kcaitech.com/kcserver/utils/str"
)

// getOldMedias 获取旧的媒体文件
func getOldMedias(path string, mediaNames []string, sourceConf config.Config) ([]autoupdate.Media, error) {
	if len(mediaNames) == 0 {
		log.Printf("No media files to migrate for document %s", path)
		return []autoupdate.Media{}, nil
	}

	storage, err := CreateStorageClient(sourceConf)
	if err != nil {
		log.Printf("Failed to create storage client: %v", err)
		return nil, err
	}

	var bucketName string
	switch sourceConf.Source.Storage.Provider {
	case "minio":
		bucketName = sourceConf.Source.Storage.Minio.Bucket
	case "oss":
		bucketName = sourceConf.Source.Storage.OSS.Bucket
	case "s3":
		bucketName = sourceConf.Source.Storage.S3.Bucket
	}

	var oldMedias []autoupdate.Media

	for _, mediaName := range mediaNames {
		// 构建媒体文件路径
		mediaPath := fmt.Sprintf("%s/medias/%s", path, mediaName)
		log.Printf("Looking for media file: %s", mediaPath)

		// 从源 bucket 获取媒体文件
		mediaContent, err := storage.GetObject(bucketName, mediaPath)
		if err != nil {
			log.Printf("Failed to get media object %s: %v", mediaPath, err)
			continue // 继续处理其他文件，不中断整个流程
		}

		// 创建 Media 结构
		media := autoupdate.Media{
			Name:    mediaName,
			Content: &mediaContent,
		}

		oldMedias = append(oldMedias, media)
		log.Printf("Retrieved media file: %s, size: %d bytes", mediaName, len(mediaContent))
	}

	log.Printf("Successfully retrieved %d/%d media files for document %s", len(oldMedias), len(mediaNames), path)
	return oldMedias, nil
}

func MigrateDocumentStorage(documentId int64, generateApiUrl string, config config.Config, path string) error {
	const maxRetries = 3
	const retryDelay = 3 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("documentId: %d, attempt: %d/%d", documentId, attempt, maxRetries)

		err := migrateDocumentStorageOnce(documentId, generateApiUrl, config, path)
		if err == nil {
			return nil
		}

		log.Printf("attempt %d failed: %v", attempt, err)
		if attempt < maxRetries {
			log.Printf("retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	log.Printf("all %d attempts failed for document %d", maxRetries, documentId)
	return fmt.Errorf("failed after %d attempts", maxRetries)
}

// MigrateVersionResp 迁移专用的版本响应结构体，用于兼容API的实际返回格式
type MigrateVersionResp struct {
	LastCmdId    string                `json:"lastCmdId"` // API实际返回的字段名
	DocumentData autoupdate.ExFromJson `json:"documentData"`
	DocumentText string                `json:"documentText"`
	MediasSize   uint64                `json:"mediasSize"`
	PagePngs     []string              `json:"pages_png_generated"`
}

func migrateDocumentStorageOnce(documentId int64, generateApiUrl string, sourceMinioConf config.Config, path string) error {
	// var generateApiUrl = "http://192.168.0.131:8088/generate" // 旧版本更新服务地址
	documentIdStr := str.IntToString(documentId)
	resp, err := http.Get(generateApiUrl + "?documentId=" + documentIdStr)
	if err != nil {
		log.Println(generateApiUrl, "http.NewRequest err", err)
		return err
	}

	defer resp.Body.Close()
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(generateApiUrl, "io.ReadAll err", err)
		return err
	}
	if resp.StatusCode != 200 {
		log.Println(generateApiUrl, "请求失败", resp.StatusCode, string(body))
		return errors.New("请求失败")
	}

	// 先用迁移专用的结构体解析API响应
	migrateVersion := MigrateVersionResp{}
	err = json.Unmarshal(body, &migrateVersion)
	if err != nil {
		log.Println(generateApiUrl, "resp", err)
		return err
	}

	// 转换为标准的VersionResp结构体
	version := autoupdate.VersionResp{
		LastCmdVerId: migrateVersion.LastCmdId, // 字段名转换
		DocumentData: migrateVersion.DocumentData,
		DocumentText: migrateVersion.DocumentText,
		MediasSize:   migrateVersion.MediasSize,
		PagePngs:     migrateVersion.PagePngs,
	}

	// 获取旧的媒体文件
	oldMedias, err := getOldMedias(path, version.DocumentData.MediaNames, sourceMinioConf)
	if err != nil {
		log.Printf("Failed to get old medias for document %d: %v", documentId, err)
		// 不中断流程，继续处理其他数据
	}

	log.Println("auto update document, start upload data", documentId)
	// upload document data
	// header := autoupdate.Header{
	// 	DocumentId:   documentIdStr,
	// 	LastCmdVerId: version.LastCmdVerId,
	// }
	response := autoupdate.Response{}

	// 传递媒体文件到 UploadDocumentData
	var mediasPtr *[]autoupdate.Media
	if len(oldMedias) > 0 {
		mediasPtr = &oldMedias
	}

	autoupdate.UpdateDocumentData(documentIdStr, version.LastCmdVerId, &version, mediasPtr, &response)
	if response.Status != autoupdate.ResponseStatusSuccess {
		log.Println("auto update failed", response.Message)
		return errors.New("auto update failed")
	}
	log.Println("auto update successed")
	return nil
}
