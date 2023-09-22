package common

const (
	ApiVersionPath      = "/api/v1"
	StorageHost         = "https://storage.protodesign.cn/files"
	BaseServiceHost     = "192.168.0.18"
	ApiGatewayHost      = "apigateway.kc.svc.cluster.local:10000"
	AuthServiceHost     = "authservice.kc.svc.cluster.local:10001"
	UserServiceHost     = "userservice.kc.svc.cluster.local:10002"
	DocumentServiceHost = "documentservice.kc.svc.cluster.local:10003"
	DocOpHost           = BaseServiceHost + ":10010"
)
