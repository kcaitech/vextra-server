package services

import (
	"protodesign.cn/kcserver/common/models"
)

type AppVersionService struct {
	*DefaultService
}

func NewAppVersionService() *AppVersionService {
	that := &AppVersionService{
		DefaultService: NewDefaultService(&models.AppVersion{}),
	}
	that.That = that
	return that
}

// FindAll 查询所有版本记录
func (s *AppVersionService) FindAll() *[]models.AppVersion {
	var result []models.AppVersion
	_ = s.Find(
		&result,
		&OrderLimitArgs{"code desc", 0},
	)
	return &result
}

// GetLatest 查询最新的版本信息
func (s *AppVersionService) GetLatest(userId int64) *models.AppVersion {
	userService := NewUserService()
	var user models.User
	userExists := userService.Get(&user, "id = ?", userId) == nil

	var result models.AppVersion
	if !userExists || user.WebAppChannel == "" {
		_ = s.Get(
			&result,
			&OrderLimitArgs{"code desc", 1},
		)
	} else {
		if s.Get(
			&result,
			&OrderLimitArgs{"code desc", 1},
			"web_app_channel = ?",
			user.WebAppChannel,
		) != nil {
			_ = s.Get(
				&result,
				&OrderLimitArgs{"code desc", 1},
			)
		}
	}
	return &result
}
