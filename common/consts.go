/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package common

const (
	RedisKeyDocumentVersioningLastUpdateTime = "server_document_version_last_update:"
	RedisKeyDocumentVersioningMutex          = "server_document_version_mutex:"
	RedisKeyDocumentVersion                  = "server_document_version:"
	RedisKeyDocumentLastCmdVerId             = "server_document_last_cmd_ver_id:"
	RedisKeyDocumentComment                  = "server_document_comment:"
	RedisKeyDocumentOpMutex                  = "server_document_op_mutex:"
	RedisKeyDocumentOp                       = "server_document_op:"
	RedisKeyDocumentSelection                = "server_document_selection:"
	RedisKeyDocumentSelectionData            = "server_document_selection_data:"
	RedisKeyRateLimit                        = "server_ratelimit:"
	RedisKeyDocumentConcurrentLimit          = "server_document_concurrent_limit:"
)
