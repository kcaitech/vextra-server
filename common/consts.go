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
)
