package main

import (
	"net"

	"github.com/rs/zerolog/log"
	"github.com/zufardhiyaulhaq/echo-grpc/server/pkg/settings"

	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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

	if settings.GRPCKeepalive {
		log.Info().Msg("setting gRPC to enable keepalive")
		keepaliveParams := keepalive.ServerParameters{
			Time:    settings.GRPCKeepaliveTime,
			Timeout: settings.GRPCKeepaliveTimeout,
		}
		opts = append(opts, grpc.KeepaliveParams(keepaliveParams))
	}

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterServerServer(grpcServer, NewServer())
	pb.RegisterHealthServer(grpcServer, NewServer())
	grpcServer.Serve(listener)
}
