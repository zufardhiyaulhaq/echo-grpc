package main

import (
	"context"

	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
)

type Server struct {
	pb.UnimplementedServerServer
	pb.UnimplementedHealthServer
}

func (s *Server) GetReply(ctx context.Context, msg *pb.Message) (*pb.Response, error) {
	return &pb.Response{
		Response: "from server:" + msg.Message,
	}, nil
}

func (s *Server) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}, nil
}
func (s *Server) Watch(req *pb.HealthCheckRequest, watch pb.Health_WatchServer) error {
	return nil
}

func NewServer() *Server {
	return &Server{}
}
