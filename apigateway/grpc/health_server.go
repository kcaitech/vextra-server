package grpc

import (
	"context"
	health_server "protodesign.cn/kcserver/apigateway/grpc/generated/health_server"
)

type healthServer struct {
	health_server.UnimplementedHealthServer
}

func NewHealthServer() health_server.HealthServer {
	return &healthServer{}
}

func (s *healthServer) HealthCheck(ctx context.Context, req *health_server.HealthCheckRequest) (*health_server.HealthCheckResponse, error) {
	return &health_server.HealthCheckResponse{
		Status: "healthy",
	}, nil
}
