package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/zufardhiyaulhaq/echo-grpc/client/pkg/settings"
	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
)

var ctx = context.Background()

type Server struct {
	settings        settings.Settings
	client          pb.ServerClient
	streamingClient pb.StreamingServerClient
}

func NewServer(settings settings.Settings, client pb.ServerClient, streamingClient pb.StreamingServerClient) Server {
	return Server{
		settings:        settings,
		client:          client,
		streamingClient: streamingClient,
	}
}

func (e Server) ServeHTTP() {
	handler := NewHandler(e.settings, e.client)

	r := mux.NewRouter()

	r.HandleFunc("/grpc/{key}", handler.Handle)
	wsHandler := NewWebSocketHandler(e.streamingClient)
	r.HandleFunc("/ws/stream/bidirectional", wsHandler.HandleBidirectional)
	r.HandleFunc("/ws/stream/server", wsHandler.HandleServerStream)
	r.HandleFunc("/ws/stream/client", wsHandler.HandleClientStream)
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello!"))
	})
	r.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello!"))
	})

	err := http.ListenAndServe(":"+e.settings.HTTPPort, r)
	if err != nil {
		log.Fatal().Err(err)
	}
}

type Handler struct {
	settings settings.Settings
	client   pb.ServerClient
}

func NewHandler(settings settings.Settings, client pb.ServerClient) Handler {
	return Handler{
		settings: settings,
		client:   client,
	}
}

func (h Handler) Handle(w http.ResponseWriter, req *http.Request) {
	key := uuid.New().String()
	value := mux.Vars(req)["key"]

	reply, err := h.client.GetReply(ctx, &pb.Message{
		Message: value,
	})
	if err != nil {
		log.Info().Msg(err.Error())
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(key + ":" + reply.Response + fmt.Sprintf(":success:%v", reply.Success)))
}
