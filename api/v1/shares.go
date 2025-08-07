/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package v1

import (
	"github.com/gin-gonic/gin"
	handlers "kcaitech.com/kcserver/handlers/document"
)

func loadShareRoutes(api *gin.RouterGroup) {
	router := api.Group("/share")
	// 分享
	router.GET("/receives", handlers.GetUserReceiveSharesList)            // 查询用户加入的文档分享列表
	router.DELETE("/", handlers.DeleteUserShare)                          // 退出共享
	router.PUT("/set", handlers.SetDocumentShareType)                     // 设置分享类型
	router.GET("/grantees", handlers.GetDocumentSharesList)               // 查询某个文档对所有用户的分享列表
	router.PUT("/", handlers.SetDocumentSharePermission)                  // 修改分享权限
	router.DELETE("/perm", handlers.DeleteDocumentSharePermission)        // 移除分享权限
	router.POST("/apply", handlers.ApplyDocumentPermission)               // 申请文档权限
	router.GET("/apply", handlers.GetDocumentPermissionRequestsList)      // 获取申请列表
	router.POST("/apply/audit", handlers.ReviewDocumentPermissionRequest) // 权限申请审核
}
