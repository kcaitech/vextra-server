package common

const (
	ApiVersionPath = "/api/v1"

	// test环境
	//StorageHost     = "https://storage.moss.design"
	//FileStorageHost = "https://storage.moss.design/files"

	// 正式环境
	StorageHost     = "https://storage1.moss.design"
	FileStorageHost = "https://storage2.moss.design"

	ApiGatewayHost      = "apigateway.kc.svc.cluster.local:10000"
	AuthServiceHost     = "authservice.kc.svc.cluster.local:10001"
	UserServiceHost     = "userservice.kc.svc.cluster.local:10002"
	DocumentServiceHost = "documentservice.kc.svc.cluster.local:10003"
	DocOpHost           = "docop-server.kc.svc.cluster.local:10010"
)
