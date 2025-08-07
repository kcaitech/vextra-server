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
	"fmt"

	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/utils"
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
	id, err := utils.GenerateBase62ID()
	if err != nil {
		return "", err
	}
	fileName := fmt.Sprintf("%s.%s", id, suffix)
	imagePath := fmt.Sprintf("/feedback/%s/%s", (userId), fileName)
	if _, err := s.storage.AttatchBucket.PutObjectByte(imagePath, fileBytes, ""); err != nil {
		return "", errors.New("上传文件失败")
	}
	return imagePath, nil
}
