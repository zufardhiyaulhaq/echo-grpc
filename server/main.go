package main

import (
	"fmt"
	"log"
	"net"

	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
	"google.golang.org/grpc"
)

func main() {
	settings, err := NewSettings()
	if err != nil {
		log.Fatalf("failed to get settings: %v", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", settings.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterServerServer(grpcServer, NewServer())
	grpcServer.Serve(listener)
}
