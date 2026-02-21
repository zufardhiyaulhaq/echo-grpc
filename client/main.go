package main

import (
	"crypto/tls"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/zufardhiyaulhaq/echo-grpc/client/pkg/server"
	"github.com/zufardhiyaulhaq/echo-grpc/client/pkg/settings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
)

func main() {
	settings, err := settings.NewSettings()
	if err != nil {
		log.Fatal().AnErr("failed to get settings", err)
	}

	log.Info().Msg("creating grpc connection")

	var opts []grpc.DialOption

	if settings.GRPCServerTLS {
		log.Info().Msg("setting gRPC to call with TLS")
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		creds := credentials.NewTLS(config)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if settings.GRPCKeepalive {
		log.Info().Msg("setting gRPC to enable keepalive")
		keepaliveParams := keepalive.ClientParameters{
			Time:    settings.GRPCKeepaliveTime,
			Timeout: settings.GRPCKeepaliveTimeout,
		}
		opts = append(opts, grpc.WithKeepaliveParams(keepaliveParams))
	}

	conn, err := grpc.Dial(settings.GRPCServerHost+":"+settings.GRPCServerPort, opts...)
	if err != nil {
		log.Fatal().AnErr("failed to start connection", err)
	}
	defer conn.Close()

	client := pb.NewServerClient(conn)
	streamingClient := pb.NewStreamingServerClient(conn)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	log.Info().Msg("starting server")
	server := server.NewServer(settings, client, streamingClient)

	go func() {
		log.Info().Msg("starting HTTP server")
		server.ServeHTTP()
		wg.Done()
	}()

	wg.Wait()
}
