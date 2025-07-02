package storage

import (
	"log"

	"kcaitech.com/kcserver/services"
)

func TidyDocumentStorage(removedDocumentPaths []string) {
	storageClient := services.GetStorageClient()

	for _, document := range removedDocumentPaths {
		// 遍历对象列表并删除
		for objectInfo := range storageClient.Bucket.ListObjects(document) {
			if objectInfo.Err != nil {
				log.Printf("列出对象时发生错误: %v", objectInfo.Err)
				continue
			}

			// 删除对象
			if err := storageClient.Bucket.DeleteObject(objectInfo.Key); err != nil {
				log.Printf("删除对象 %s 时发生错误: %v", objectInfo.Key, err)
			} else {
				log.Printf("成功删除对象: %s", objectInfo.Key)
			}
		}
	}
}
