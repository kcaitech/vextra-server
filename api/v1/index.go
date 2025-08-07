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
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/services"
)

func LoadRoutes(apiGroup *gin.RouterGroup) {
	loadWsRoutes(apiGroup)    // 单独鉴权
	loadLoginRoutes(apiGroup) // 从refreshToken获取信息
	loadAccessRoutes(apiGroup)
	apiGroup.Use(services.GetKCAuthClient().AuthRequired())
	apiGroup.Use(common.Sha1SaveData)
	loadUserRoutes(apiGroup)
	loadDocumentRoutes(apiGroup)
	loadShareRoutes(apiGroup)
	loadTeamRoutes(apiGroup)
	loadFeedbackRoutes(apiGroup)
}
