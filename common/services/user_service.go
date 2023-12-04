package services

import (
	"errors"
	"fmt"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/storage"
	"protodesign.cn/kcserver/utils/str"
)

type UserService struct {
	*DefaultService
}

func NewUserService() *UserService {
	that := &UserService{
		DefaultService: NewDefaultService(&models.User{}),
	}
	that.That = that
	return that
}

func (s *UserService) GetByNickname(nickname string) (*models.User, error) {
	var modelData models.User
	err := s.DB.Where("nickname = ?", nickname).First(&modelData).Error
	if err != nil {
		return nil, err
	}
	return &modelData, nil
}

func (s *UserService) UploadUserAvatar(user *models.User, fileBytes []byte, contentType string) (string, error) {
	fileSize := len(fileBytes)
	if fileSize > 1024*1024*5 {
		return "", errors.New("文件大小不能超过5MB")
	}
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
	if user.Uid == "" {
		user.Uid = str.GetUid()
	}
	fileName := fmt.Sprintf("%s.%s", str.GetUid(), suffix)
	avatarPath := fmt.Sprintf("/users/%s/avatar/%s", user.Uid, fileName)
	if _, err := storage.FilesBucket.PutObjectByte(avatarPath, fileBytes); err != nil {
		return "", errors.New("上传文件失败")
	}
	user.Avatar = avatarPath
	if _, err := s.UpdateColumnsById(user.Id, map[string]any{
		"avatar": avatarPath,
	}); err != nil {
		return "", errors.New("更新错误")
	}
	return avatarPath, nil
}
