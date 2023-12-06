package services

import (
	"errors"
	"fmt"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/storage"
	"protodesign.cn/kcserver/utils/str"
)

type FeedbackService struct {
	*DefaultService
}

func NewFeedbackService() *FeedbackService {
	that := &FeedbackService{
		DefaultService: NewDefaultService(&models.Feedback{}),
	}
	that.That = that
	return that
}

func (s *FeedbackService) UploadImage(userId int64, fileBytes []byte, contentType string) (string, error) {
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
		return "", errors.New(fmt.Sprintf("不支持的文件类型：%s", contentType))
	}
	fileName := fmt.Sprintf("%s.%s", str.GetUid(), suffix)
	imagePath := fmt.Sprintf("/feedback/%s/%s", str.IntToString(userId), fileName)
	if _, err := storage.FilesBucket.PutObjectByte(imagePath, fileBytes); err != nil {
		return "", errors.New("上传文件失败")
	}
	return imagePath, nil
}
