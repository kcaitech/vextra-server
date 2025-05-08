package services

import (
	"kcaitech.com/kcserver/models"
)

// ProjectFavoritesService 项目收藏服务
type ProjectFavoritesService struct {
	*DefaultService
}

// NewProjectFavoritesService 创建项目收藏服务
func NewProjectFavoritesService() *ProjectFavoritesService {
	that := &ProjectFavoritesService{
		DefaultService: NewDefaultService(&models.ProjectFavorite{}),
	}
	that.That = that
	return that
}

// GetUserFavoriteProjects 获取用户收藏的项目列表
func (s *ProjectFavoritesService) GetUserFavoriteProjects(userId string) ([]models.ProjectFavorite, error) {
	var projectFavorites []models.ProjectFavorite
	err := s.Find(
		&projectFavorites,
		&WhereArgs{"user_id = ?", []any{userId}},
		&OrderLimitArgs{"id desc", 0},
	)
	if err != nil {
		return nil, err
	}
	return projectFavorites, nil
}
