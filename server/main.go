package main

import (
	"net"

	"github.com/rs/zerolog/log"
	"github.com/zufardhiyaulhaq/echo-grpc/server/pkg/settings"

	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
	"google.golang.org/grpc"
)

func main() {
	settings, err := settings.NewSettings()
	if err != nil {
		log.Fatal().AnErr("failed to get settings", err)
	}

	log.Info().Msg("starting grpc server")

	listener, err := net.Listen("tcp", "0.0.0.0:"+settings.Port)
	if err != nil {
		log.Fatal().AnErr("failed to listen connection", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterServerServer(grpcServer, NewServer())
	pb.RegisterHealthServer(grpcServer, NewServer())
	grpcServer.Serve(listener)
}
