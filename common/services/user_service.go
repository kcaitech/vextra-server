package services

import (
	"protodesign.cn/kcserver/common/models"
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
