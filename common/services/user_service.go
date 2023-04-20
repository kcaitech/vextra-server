package services

import (
	"gorm.io/gorm"
	"protodesign.cn/kcserver/common/models"
)

type UserService struct {
	DB *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{
		DB: models.DB,
	}
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.DB.Create(user).Error
}

func (s *UserService) GetUser(id uint) (*models.User, error) {
	var user models.User
	err := s.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetUserByNickname(nickname string) (*models.User, error) {
	var user models.User
	err := s.DB.Where("nickname = ?", nickname).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
