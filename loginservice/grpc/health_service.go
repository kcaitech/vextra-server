package grpc

import (
	"context"
	health_server "protodesign.cn/kcserver/loginservice/grpc/generated/health_service"
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
