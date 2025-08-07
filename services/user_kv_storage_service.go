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
	"errors"

	"kcaitech.com/kcserver/models"
)

type UserKVStorageService struct {
	*DefaultService
}

func NewUserKVStorageService() *UserKVStorageService {
	that := &UserKVStorageService{
		DefaultService: NewDefaultService(&models.UserKVStorage{}),
	}
	that.That = that
	return that
}

func (s *UserKVStorageService) GetOne(userId string, key string) (string, error) {
	var result models.UserKVStorage
	if s.Get(
		&result,
		"user_id = ? and `key` = ?",
		userId,
		key,
	) == nil {
		return result.Value, nil
	}
	return "", errors.New("not found")
}

func (s *UserKVStorageService) SetOne(userId string, key string, value string) bool {
	var result models.UserKVStorage
	if s.Get(
		&result,
		"user_id = ? and `key` = ?",
		userId,
		key,
	) == nil {
		result.Value = value
		count, _ := s.Updates(&result)
		return count > 0
	}

	result.UserId = userId
	result.Key = key
	result.Value = value
	return s.Create(&result) == nil
}
