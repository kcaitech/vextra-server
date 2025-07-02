package storage

import (
	"fmt"
	"log"
	"strings"

	"kcaitech.com/kcserver/scripts/migrate-20250702/config"
	"kcaitech.com/kcserver/services"
)

func MigrateTeamAvatars(sourceConf config.Config) {
	storage, err := CreateStorageClient(sourceConf)
	if err != nil {
		log.Printf("Failed to create storage client: %v", err)
		return
	}

	var filesBucket string
	switch sourceConf.Source.Storage.Provider {
	case "minio":
		filesBucket = sourceConf.Source.Storage.Minio.FilesBucket
	case "oss":
		filesBucket = sourceConf.Source.Storage.OSS.FilesBucket
	case "s3":
		filesBucket = sourceConf.Source.Storage.S3.FilesBucket
	}

	objectCh := storage.ListObjects(filesBucket, "teams/")

	fileCount := 0
	successCount := 0
	failCount := 0
	totalSize := int64(0)

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Error listing objects: %v", object.Err)
			break
		}

		fileCount++
		totalSize += object.Size
		log.Printf("处理文件 %d: %s (size: %d bytes)", fileCount, object.Key, object.Size)

		// 下载源文件
		fileContent, err := storage.GetObject(filesBucket, object.Key)
		if err != nil {
			log.Printf("Failed to get source object %s: %v", object.Key, err)
			failCount++
			continue
		}

		// 解析路径获取ID和文件类型，只处理teams文件
		if strings.HasPrefix(object.Key, "teams/") {
			err = uploadTeamFile(object.Key, fileContent)
		}

		if err != nil {
			log.Printf("Failed to upload %s: %v", object.Key, err)
			failCount++
		} else {
			log.Printf("✅ Successfully uploaded: %s", object.Key)
			successCount++
		}
	}

	log.Printf("=== teams文件夹迁移完成 ===")
	log.Printf("teams总文件数: %d, teams失败数量: %d, teams成功上传: %d, teams总大小: %.2f MB", fileCount, failCount, successCount, float64(totalSize)/(1024*1024))
	log.Printf("=== END teams MIGRATION ===")
}

// uploadTeamFile 上传团队文件
func uploadTeamFile(filePath string, fileContent []byte) error {
	// 解析路径: teams/{teamId}/avatar/{filename}
	parts := strings.Split(filePath, "/")
	teamId := parts[1]

	// 使用存储服务上传文件到attach bucket
	_storage := services.GetStorageClient()
	if _storage == nil {
		log.Printf("存储客户端为空")
		return fmt.Errorf("存储客户端未初始化")
	}

	if _storage.AttatchBucket == nil {
		log.Printf("AttatchBucket为空")
		return fmt.Errorf("AttatchBucket未初始化")
	}

	path := fmt.Sprintf("/teams/%s/avatar/%s", teamId, parts[3])
	if _, err := _storage.AttatchBucket.PutObjectByte(path, fileContent, ""); err != nil {
		log.Printf("上传文件失败: %v", err)
		return fmt.Errorf("上传文件失败: %v", err)
	}
	log.Printf("Successfully updated team %s avatar: %s", teamId, path)
	return nil
}
