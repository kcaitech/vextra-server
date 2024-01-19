package common

const (
	ApiVersionPath = "/api/v1"

	// test环境
	//StorageHost     = "https://storage.protodesign.cn"
	//FileStorageHost = "https://storage.protodesign.cn/files"
	// 正式环境
	StorageHost     = "https://storage1.protodesign.cn"
	FileStorageHost = "https://storage2.protodesign.cn"

	ApiGatewayHost      = "apigateway.kc.svc.cluster.local:10000"
	AuthServiceHost     = "authservice.kc.svc.cluster.local:10001"
	UserServiceHost     = "userservice.kc.svc.cluster.local:10002"
	DocumentServiceHost = "documentservice.kc.svc.cluster.local:10003"
	DocOpHost           = "docop-server.kc.svc.cluster.local:10010"
)
