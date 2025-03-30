package services

import (
	"errors"
	"fmt"

	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/utils/str"
)

type FeedbackService struct {
	*DefaultService
	storage *storage.StorageClient
}

func NewFeedbackService() *FeedbackService {
	that := &FeedbackService{
		DefaultService: NewDefaultService(&models.Feedback{}),
		storage:        storageClient,
	}
	that.That = that
	return that
}

func (s *FeedbackService) UploadImage(userId string, fileBytes []byte, contentType string) (string, error) {
	var suffix string
	switch contentType {
	case "image/jpeg":
		suffix = "jpg"
	case "image/png":
		suffix = "png"
	case "image/gif":
		suffix = "gif"
	case "image/bmp":
		suffix = "bmp"
	case "image/tiff":
		suffix = "tif"
	case "image/webp":
		suffix = "webp"
	default:
		return "", (fmt.Errorf("不支持的文件类型：%s", contentType))
	}
	fileName := fmt.Sprintf("%s.%s", str.GetUid(), suffix)
	imagePath := fmt.Sprintf("/feedback/%s/%s", (userId), fileName)
	if _, err := s.storage.AttatchBucket.PutObjectByte(imagePath, fileBytes); err != nil {
		return "", errors.New("上传文件失败")
	}
	return imagePath, nil
}
