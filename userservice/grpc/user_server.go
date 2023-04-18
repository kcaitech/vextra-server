package grpc

import (
	"context"
	"protodesign.cn/kcserver/userservice/grpc/generated/user_server"
)

type userServer struct {
	user_server.UnimplementedUserServer
}

func NewUserServer() user_server.UserServer {
	return &userServer{}
}

func (s *healthServer) GetUser(ctx context.Context, req *user_server.GetUserRequest) (*user_server.GetUserResponse, error) {
	return &user_server.GetUserResponse{
		UserId: 1,
		Name:   "1",
		Email:  "1",
	}, nil
}
