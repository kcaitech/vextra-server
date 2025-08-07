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
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	// todo 需要检查数据库连接是否正常
	common.Success(c, map[string]string{
		"status": "healthy",
	})
}
