package main

import (
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog/log"
	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StreamingServer struct {
	pb.UnimplementedStreamingServerServer
}

func NewStreamingServer() *StreamingServer {
	return &StreamingServer{}
}

func (s *StreamingServer) BidirectionalStream(stream pb.StreamingServer_BidirectionalStreamServer) error {
	var serverSeq int64 = 0

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if msg.StreamId == "" {
			return status.Error(codes.InvalidArgument, "stream_id is required")
		}

		serverSeq++
		response := &pb.StreamResponse{
			StreamId:       msg.StreamId,
			SequenceNumber: serverSeq,
			Timestamp:      time.Now().UnixNano(),
			Response:       "from server: " + msg.Message,
			Success:        true,
		}

		if err := stream.Send(response); err != nil {
			return err
		}

		log.Info().
			Str("stream_id", msg.StreamId).
			Int64("seq", serverSeq).
			Msg("bidirectional: echoed message")
	}
}

func (s *StreamingServer) ServerStream(msg *pb.StreamMessage, stream pb.StreamingServer_ServerStreamServer) error {
	if msg.StreamId == "" {
		return status.Error(codes.InvalidArgument, "stream_id is required")
	}

	log.Info().
		Str("stream_id", msg.StreamId).
		Msg("server stream: starting 5 echoes")

	for i := 1; i <= 5; i++ {
		response := &pb.StreamResponse{
			StreamId:       msg.StreamId,
			SequenceNumber: int64(i),
			Timestamp:      time.Now().UnixNano(),
			Response:       "from server: " + msg.Message + " (echo " + fmt.Sprintf("%d/5", i) + ")",
			Success:        true,
		}

		if err := stream.Send(response); err != nil {
			return err
		}

		log.Info().
			Str("stream_id", msg.StreamId).
			Int("echo", i).
			Msg("server stream: sent echo")

		if i < 5 {
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func (s *StreamingServer) ClientStream(stream pb.StreamingServer_ClientStreamServer) error {
	var count int64 = 0
	var streamId string

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			response := &pb.StreamResponse{
				StreamId:       streamId,
				SequenceNumber: count,
				Timestamp:      time.Now().UnixNano(),
				Response:       fmt.Sprintf("from server: received %d messages", count),
				Success:        true,
			}

			log.Info().
				Str("stream_id", streamId).
				Int64("count", count).
				Msg("client stream: completed")

			return stream.SendAndClose(response)
		}
		if err != nil {
			return err
		}

		if msg.StreamId == "" {
			return status.Error(codes.InvalidArgument, "stream_id is required")
		}

		if streamId == "" {
			streamId = msg.StreamId
		}

		count++
		log.Info().
			Str("stream_id", msg.StreamId).
			Int64("count", count).
			Msg("client stream: received message")
	}
}
