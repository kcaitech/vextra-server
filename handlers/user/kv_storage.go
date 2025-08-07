/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package handlers

import (
	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
	"kcaitech.com/kcserver/utils/sliceutil"
)

var AllowedKeyList = []string{
	"FontList",
	"Preferences",
}

func GetUserKVStorage(c *gin.Context) {

	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}

	key := c.Query("key")
	if key == "" {
		common.BadRequest(c, "参数错误：key")
		return
	}

	if len(sliceutil.FilterT(func(code string) bool {
		return key == code
	}, AllowedKeyList...)) == 0 {
		common.BadRequest(c, "不允许的key: "+key)
		return
	}

	result := map[string]any{}
	userKVStorageService := services.NewUserKVStorageService()
	userKVStorage, err := userKVStorageService.GetOne(userId, key)

	if err == nil {
		result[key] = userKVStorage
	}
	common.Success(c, result)
}

func SetUserKVStorage(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}
	if req.Key == "" {
		common.BadRequest(c, "参数错误：key")
		return
	}
	if len(sliceutil.FilterT(func(code string) bool {
		return req.Key == code
	}, AllowedKeyList...)) == 0 {
		common.BadRequest(c, "不允许的key: "+req.Key)
		return
	}

	userKVStorageService := services.NewUserKVStorageService()
	if !userKVStorageService.SetOne(userId, req.Key, req.Value) {
		common.ServerError(c, "操作失败")
		return
	}

	common.Success(c, "")
}
