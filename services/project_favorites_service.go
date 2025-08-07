/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

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
