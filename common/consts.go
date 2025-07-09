package common

const (
	RedisKeyDocumentVersioningLastUpdateTime = "Document Versioning LastUpdateTime[DocumentId:%s]"
	RedisKeyDocumentVersioningMutex          = "Document Versioning Mutex[DocumentId::%s]"
	RedisKeyDocumentComment                  = "Document Comment[DocumentId:%s]"
	RedisKeyDocumentLastCmdVerId             = "Document LastCmdVerId[DocumentId:%s]"
	RedisKeyDocumentOpMutex                  = "Document Op Mutex[DocumentId:%s]"
	RedisKeyDocumentOp                       = "Document Op[DocumentId:%s]"
	RedisKeyDocumentSelection                = "Document Selection[DocumentId:%s]"
	RedisKeyDocumentSelectionData            = "Document Selection Data[DocumentId:%s]"
	RedisKeyDocumentVersion                  = "Document Version[DocumentId:%s]"
	RedisKeyRateLimit                        = "ratelimit:"
)
