/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package document

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/utils"
)

// GetDocumentAccessKey 获取文档访问密钥
func GetDocumentAccessKey(c *gin.Context) {
	userId, msg := utils.GetUserId(c)
	if msg != nil {
		common.Unauthorized(c)
		return
	}

	documentId := (c.Query("doc_id"))
	if documentId == "" {
		common.BadRequest(c, "参数错误：doc_id")
		return
	}

	key, code, err := common.GetDocumentAccessKey(userId, documentId, true)
	if err == nil {
		common.Success(c, key)
	} else if code == http.StatusUnauthorized {
		common.Unauthorized(c)
	} else {
		common.ServerError(c, err.Error())
	}
}
