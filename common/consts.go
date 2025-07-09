package common

const (
	RedisKeyDocumentVersioningLastUpdateTime = "server_version_last_update:"
	RedisKeyDocumentVersioningMutex          = "server_version_mutex:"
	RedisKeyDocumentVersion                  = "server_version:"
	RedisKeyDocumentLastCmdVerId             = "server_last_cmd_ver_id:"
	RedisKeyDocumentComment                  = "server_comment:"
	RedisKeyDocumentOpMutex                  = "server_op_mutex:"
	RedisKeyDocumentOp                       = "server_op:"
	RedisKeyDocumentSelection                = "server_selection:"
	RedisKeyDocumentSelectionData            = "server_selection_data:"
	RedisKeyRateLimit                        = "server_ratelimit:"
)
