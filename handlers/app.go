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
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// GetAppVersionList 获取APP版本列表
// func GetAppVersionList(c *gin.Context) {
// 	appVersionService := services.NewAppVersionService()
// 	result := appVersionService.FindAll()
// 	response.Success(c, result)
// }

// GetLatestAppVersion 获取最新的版本信息
// func GetLatestAppVersion(c *gin.Context) {
// 	userId, _ := auth.GetUserId(c)

// 	appVersionService := services.NewAppVersionService()
// 	result := appVersionService.GetLatest(userId)
// 	response.Success(c, result)
// }

type Package struct {
	Version string `yaml:"version"`
}

func LoadPackageVersion() *string {
	var def = ""
	content, err := os.ReadFile("package.yaml")
	if err != nil {
		log.Printf("load package.yaml fail %v", err)
		return &def
	}
	config := &Package{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		log.Printf("unmarshal package.yaml fail %v", err)
		return &def
	}
	return &config.Version
}

var version *string

// func GetAppVersion(c *gin.Context) {
// 	if version == nil {
// 		version = LoadPackageVersion()
// 	}
// 	response.Success(c, version)
// }
