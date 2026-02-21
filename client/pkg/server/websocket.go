package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 30 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type WSMessage struct {
	StreamID       string `json:"stream_id"`
	SequenceNumber int64  `json:"sequence_number"`
	Timestamp      int64  `json:"timestamp"`
	Message        string `json:"message"`
}

type WSResponse struct {
	StreamID       string `json:"stream_id"`
	SequenceNumber int64  `json:"sequence_number"`
	Timestamp      int64  `json:"timestamp"`
	Response       string `json:"response"`
	Success        bool   `json:"success"`
}

type WebSocketHandler struct {
	streamingClient pb.StreamingServerClient
}

func NewWebSocketHandler(client pb.StreamingServerClient) *WebSocketHandler {
	return &WebSocketHandler{
		streamingClient: client,
	}
}

func (h *WebSocketHandler) HandleBidirectional(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}
	defer conn.Close()

	stream, err := h.streamingClient.BidirectionalStream(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to create bidirectional stream")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(1011, "gRPC connection failed"))
		return
	}

	// Handle incoming responses from gRPC
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			resp, err := stream.Recv()
			if err != nil {
				return
			}

			wsResp := WSResponse{
				StreamID:       resp.StreamId,
				SequenceNumber: resp.SequenceNumber,
				Timestamp:      resp.Timestamp,
				Response:       resp.Response,
				Success:        resp.Success,
			}

			data, _ := json.Marshal(wsResp)
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}()

	// Handle incoming messages from WebSocket
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			errResp := WSResponse{Success: false, Response: "invalid message format"}
			data, _ := json.Marshal(errResp)
			conn.WriteMessage(websocket.TextMessage, data)
			continue
		}

		pbMsg := &pb.StreamMessage{
			StreamId:       wsMsg.StreamID,
			SequenceNumber: wsMsg.SequenceNumber,
			Timestamp:      wsMsg.Timestamp,
			Message:        wsMsg.Message,
		}

		if err := stream.Send(pbMsg); err != nil {
			break
		}
	}

	stream.CloseSend()
	<-done
}

func (h *WebSocketHandler) HandleServerStream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}
	defer conn.Close()

	// Read single message from WebSocket
	_, message, err := conn.ReadMessage()
	if err != nil {
		return
	}

	var wsMsg WSMessage
	if err := json.Unmarshal(message, &wsMsg); err != nil {
		errResp := WSResponse{Success: false, Response: "invalid message format"}
		data, _ := json.Marshal(errResp)
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	pbMsg := &pb.StreamMessage{
		StreamId:       wsMsg.StreamID,
		SequenceNumber: wsMsg.SequenceNumber,
		Timestamp:      wsMsg.Timestamp,
		Message:        wsMsg.Message,
	}

	stream, err := h.streamingClient.ServerStream(r.Context(), pbMsg)
	if err != nil {
		log.Error().Err(err).Msg("failed to create server stream")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(1011, "gRPC connection failed"))
		return
	}

	// Receive all responses and forward to WebSocket
	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}

		wsResp := WSResponse{
			StreamID:       resp.StreamId,
			SequenceNumber: resp.SequenceNumber,
			Timestamp:      resp.Timestamp,
			Response:       resp.Response,
			Success:        resp.Success,
		}

		data, _ := json.Marshal(wsResp)
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}

func (h *WebSocketHandler) HandleClientStream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}
	defer conn.Close()

	stream, err := h.streamingClient.ClientStream(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to create client stream")
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(1011, "gRPC connection failed"))
		return
	}

	// Read messages until client closes
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			// Client closed connection, close gRPC stream and get response
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			errResp := WSResponse{Success: false, Response: "invalid message format"}
			data, _ := json.Marshal(errResp)
			conn.WriteMessage(websocket.TextMessage, data)
			continue
		}

		pbMsg := &pb.StreamMessage{
			StreamId:       wsMsg.StreamID,
			SequenceNumber: wsMsg.SequenceNumber,
			Timestamp:      wsMsg.Timestamp,
			Message:        wsMsg.Message,
		}

		if err := stream.Send(pbMsg); err != nil {
			break
		}
	}

	// Close send and receive summary
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Error().Err(err).Msg("failed to close client stream")
		return
	}

	wsResp := WSResponse{
		StreamID:       resp.StreamId,
		SequenceNumber: resp.SequenceNumber,
		Timestamp:      resp.Timestamp,
		Response:       resp.Response,
		Success:        resp.Success,
	}

	data, _ := json.Marshal(wsResp)
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	conn.WriteMessage(websocket.TextMessage, data)
}
