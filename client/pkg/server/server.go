package server

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/evio"
	"github.com/zufardhiyaulhaq/echo-grpc/client/pkg/settings"
	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
)

var ctx = context.Background()

type Server struct {
	settings settings.Settings
	client   pb.ServerClient
}

func NewServer(settings settings.Settings, client pb.ServerClient) Server {
	return Server{
		settings: settings,
		client:   client,
	}
}

func (e Server) ServeEcho() {
	var events evio.Events

	events.Data = func(c evio.Conn, in []byte) (out []byte, action evio.Action) {
		key := uuid.New().String()
		value := string(in)

		reply, err := e.client.GetReply(ctx, &pb.Message{
			Message: value,
		})
		if err != nil {
			out = []byte(err.Error())
			return
		}

		out = []byte(key + ":" + reply.Response)

		return
	}

	if err := evio.Serve(events, "tcp://0.0.0.0:"+e.settings.EchoPort); err != nil {
		log.Fatal().Err(err)
	}
}

func (e Server) ServeHTTP() {
	handler := NewHandler(e.settings, e.client)

	r := mux.NewRouter()

	r.HandleFunc("/grpc/{key}", handler.Handle)
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
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(key + ":" + reply.Response))
}
