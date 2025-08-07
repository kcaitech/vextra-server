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
	handlers "kcaitech.com/kcserver/handlers"
	"kcaitech.com/kcserver/services"
)

func loadAccessRoutes(api *gin.RouterGroup) {
	router := api.Group("/access")
	router.POST("/token", handlers.AccessToken)
	router.GET("/ws", handlers.AccessWs)
	// 下面的需要用户登录
	router.Use(services.GetKCAuthClient().AuthRequired())
	router.POST("/create", handlers.AccessCreate)
	router.GET("/list", handlers.AccessList)
	router.POST("/update", handlers.AccessUpdate)
	router.POST("/delete", handlers.AccessDelete)
}
