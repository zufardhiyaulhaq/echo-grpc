package main

import (
	"context"

	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
)

type Server struct {
	pb.UnimplementedServerServer
}

func (s *Server) GetStatus(ctx context.Context, point *pb.Empty) (*pb.Status, error) {
	return &pb.Status{
		Status: true,
	}, nil
}

func NewServer() *Server {
	return &Server{}
}
