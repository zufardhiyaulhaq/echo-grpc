# gRPC Streaming Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add three gRPC streaming RPCs (bidirectional, client, server) with WebSocket HTTP endpoints.

**Architecture:** New `StreamingServer` gRPC service in separate proto file. Server implements streaming handlers. Client exposes WebSocket endpoints that proxy to gRPC streams.

**Tech Stack:** Go 1.22, gRPC, Protocol Buffers, gorilla/websocket, gorilla/mux

---

### Task 1: Create Streaming Proto Definition

**Files:**
- Create: `proto/streaming.proto`

**Step 1: Create the proto file**

```protobuf
syntax = "proto3";

package com.gopay.echo.streaming;

option go_package = "github.com/zufardhiyaulhaq/echo-grpc/proto";

service StreamingServer {
    rpc ClientStream(stream StreamMessage) returns (StreamResponse);
    rpc ServerStream(StreamMessage) returns (stream StreamResponse);
    rpc BidirectionalStream(stream StreamMessage) returns (stream StreamResponse);
}

message StreamMessage {
    string stream_id = 1;
    int64 sequence_number = 2;
    int64 timestamp = 3;
    string message = 4;
}

message StreamResponse {
    string stream_id = 1;
    int64 sequence_number = 2;
    int64 timestamp = 3;
    string response = 4;
    bool success = 5;
}
```

**Step 2: Commit**

```bash
git add proto/streaming.proto
git commit -m "proto: add streaming service definition"
```

---

### Task 2: Generate Go Code from Proto

**Files:**
- Generate: `proto/streaming.pb.go`
- Generate: `proto/streaming_grpc.pb.go`

**Step 1: Generate the Go code**

Run:
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/streaming.proto
```

Expected: Two new files created in `proto/` directory.

**Step 2: Verify generation**

Run: `ls proto/streaming*.go`
Expected: `proto/streaming.pb.go  proto/streaming_grpc.pb.go`

**Step 3: Commit**

```bash
git add proto/streaming.pb.go proto/streaming_grpc.pb.go
git commit -m "proto: generate Go code for streaming service"
```

---

### Task 3: Implement StreamingServer - BidirectionalStream

**Files:**
- Create: `server/streaming.go`

**Step 1: Create streaming.go with BidirectionalStream**

```go
package main

import (
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
```

**Step 2: Verify it compiles**

Run: `go build ./server/`
Expected: No errors

**Step 3: Commit**

```bash
git add server/streaming.go
git commit -m "server: implement BidirectionalStream RPC"
```

---

### Task 4: Implement StreamingServer - ServerStream

**Files:**
- Modify: `server/streaming.go`

**Step 1: Add ServerStream method**

Add after `BidirectionalStream`:

```go
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
```

**Step 2: Add "fmt" import**

Update imports at top of file:

```go
import (
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog/log"
	pb "github.com/zufardhiyaulhaq/echo-grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
```

**Step 3: Verify it compiles**

Run: `go build ./server/`
Expected: No errors

**Step 4: Commit**

```bash
git add server/streaming.go
git commit -m "server: implement ServerStream RPC"
```

---

### Task 5: Implement StreamingServer - ClientStream

**Files:**
- Modify: `server/streaming.go`

**Step 1: Add ClientStream method**

Add after `ServerStream`:

```go
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
```

**Step 2: Verify it compiles**

Run: `go build ./server/`
Expected: No errors

**Step 3: Commit**

```bash
git add server/streaming.go
git commit -m "server: implement ClientStream RPC"
```

---

### Task 6: Register StreamingServer in main.go

**Files:**
- Modify: `server/main.go`

**Step 1: Register StreamingServer**

Add after line 41 (`pb.RegisterHealthServer(grpcServer, NewServer())`):

```go
	pb.RegisterStreamingServerServer(grpcServer, NewStreamingServer())
```

**Step 2: Verify it compiles**

Run: `go build ./server/`
Expected: No errors

**Step 3: Commit**

```bash
git add server/main.go
git commit -m "server: register StreamingServer service"
```

---

### Task 7: Add gorilla/websocket Dependency

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

**Step 1: Add dependency**

Run: `go get github.com/gorilla/websocket`
Expected: go.mod updated with websocket dependency

**Step 2: Verify**

Run: `grep websocket go.mod`
Expected: `github.com/gorilla/websocket v1.x.x`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add gorilla/websocket"
```

---

### Task 8: Create WebSocket Handler - Base Structure

**Files:**
- Create: `client/pkg/server/websocket.go`

**Step 1: Create base WebSocket handler structure**

```go
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
	writeWait      = 10 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = (pongWait * 9) / 10
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
```

**Step 2: Verify it compiles**

Run: `go build ./client/...`
Expected: No errors

**Step 3: Commit**

```bash
git add client/pkg/server/websocket.go
git commit -m "client: add WebSocket handler base structure"
```

---

### Task 9: Implement BidirectionalStream WebSocket Handler

**Files:**
- Modify: `client/pkg/server/websocket.go`

**Step 1: Add BidirectionalStream handler**

Add at the end of the file:

```go
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
```

**Step 2: Verify it compiles**

Run: `go build ./client/...`
Expected: No errors

**Step 3: Commit**

```bash
git add client/pkg/server/websocket.go
git commit -m "client: implement bidirectional WebSocket handler"
```

---

### Task 10: Implement ServerStream WebSocket Handler

**Files:**
- Modify: `client/pkg/server/websocket.go`

**Step 1: Add ServerStream handler**

Add at the end of the file:

```go
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
```

**Step 2: Verify it compiles**

Run: `go build ./client/...`
Expected: No errors

**Step 3: Commit**

```bash
git add client/pkg/server/websocket.go
git commit -m "client: implement server stream WebSocket handler"
```

---

### Task 11: Implement ClientStream WebSocket Handler

**Files:**
- Modify: `client/pkg/server/websocket.go`

**Step 1: Add ClientStream handler**

Add at the end of the file:

```go
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
```

**Step 2: Verify it compiles**

Run: `go build ./client/...`
Expected: No errors

**Step 3: Commit**

```bash
git add client/pkg/server/websocket.go
git commit -m "client: implement client stream WebSocket handler"
```

---

### Task 12: Create StreamingServerClient in Client Main

**Files:**
- Modify: `client/main.go`

**Step 1: Create StreamingServerClient**

After line 53 (`client := pb.NewServerClient(conn)`), add:

```go
	streamingClient := pb.NewStreamingServerClient(conn)
```

**Step 2: Update server.NewServer call**

Change line 59 from:
```go
	server := server.NewServer(settings, client)
```

To:
```go
	server := server.NewServer(settings, client, streamingClient)
```

**Step 3: Verify it compiles (will fail - need to update Server struct next)**

Run: `go build ./client/...`
Expected: Compile error (too many arguments to NewServer) - this is expected, fixed in next task.

**Step 4: Commit (partial - will complete in next task)**

Do not commit yet - continue to next task.

---

### Task 13: Update Server Struct for Streaming Client

**Files:**
- Modify: `client/pkg/server/server.go`

**Step 1: Add streamingClient field to Server struct**

Update Server struct (around line 17):

```go
type Server struct {
	settings        settings.Settings
	client          pb.ServerClient
	streamingClient pb.StreamingServerClient
}
```

**Step 2: Update NewServer function**

Update NewServer (around line 22):

```go
func NewServer(settings settings.Settings, client pb.ServerClient, streamingClient pb.StreamingServerClient) Server {
	return Server{
		settings:        settings,
		client:          client,
		streamingClient: streamingClient,
	}
}
```

**Step 3: Verify it compiles**

Run: `go build ./client/...`
Expected: No errors

**Step 4: Commit**

```bash
git add client/main.go client/pkg/server/server.go
git commit -m "client: add streaming client to Server struct"
```

---

### Task 14: Register WebSocket Routes

**Files:**
- Modify: `client/pkg/server/server.go`

**Step 1: Add WebSocket routes in ServeHTTP**

In `ServeHTTP` function, after line 34 (`r.HandleFunc("/grpc/{key}", handler.Handle)`), add:

```go
	wsHandler := NewWebSocketHandler(e.streamingClient)
	r.HandleFunc("/ws/stream/bidirectional", wsHandler.HandleBidirectional)
	r.HandleFunc("/ws/stream/server", wsHandler.HandleServerStream)
	r.HandleFunc("/ws/stream/client", wsHandler.HandleClientStream)
```

**Step 2: Verify it compiles**

Run: `go build ./client/...`
Expected: No errors

**Step 3: Commit**

```bash
git add client/pkg/server/server.go
git commit -m "client: register WebSocket routes for streaming"
```

---

### Task 15: Manual Integration Test

**Step 1: Start server**

Run in terminal 1:
```bash
source .env.example
make server.run
```
Expected: "starting grpc server" log message

**Step 2: Start client**

Run in terminal 2:
```bash
source .env.example
make client.run
```
Expected: "starting HTTP server" log message

**Step 3: Test unary (existing)**

Run:
```bash
curl http://localhost:8080/grpc/test
```
Expected: UUID + ":from server:test:success:true"

**Step 4: Test server streaming with wscat**

Run:
```bash
wscat -c ws://localhost:8080/ws/stream/server
```

Then send:
```json
{"stream_id":"test1","sequence_number":1,"message":"hello"}
```

Expected: 5 responses with 1-second intervals, each containing "from server: hello (echo N/5)"

**Step 5: Test bidirectional with wscat**

Run:
```bash
wscat -c ws://localhost:8080/ws/stream/bidirectional
```

Then send multiple messages:
```json
{"stream_id":"test2","sequence_number":1,"message":"msg1"}
{"stream_id":"test2","sequence_number":2,"message":"msg2"}
```

Expected: Immediate response for each message

**Step 6: Test client streaming with wscat**

Run:
```bash
wscat -c ws://localhost:8080/ws/stream/client
```

Send multiple messages then close (Ctrl+C):
```json
{"stream_id":"test3","sequence_number":1,"message":"a"}
{"stream_id":"test3","sequence_number":2,"message":"b"}
```

Expected: Summary response after close: "from server: received 2 messages"

---

### Task 16: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Add streaming documentation**

Add after "Testing the Service" section:

```markdown
## Streaming Endpoints

WebSocket endpoints for gRPC streaming:
```bash
# Server streaming (5 echoes with 1s intervals)
wscat -c ws://localhost:8080/ws/stream/server
> {"stream_id":"1","message":"hello"}

# Bidirectional streaming (immediate echo)
wscat -c ws://localhost:8080/ws/stream/bidirectional

# Client streaming (summary after close)
wscat -c ws://localhost:8080/ws/stream/client
```
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add streaming endpoints to CLAUDE.md"
```

---

### Task 17: Final Commit - Update README

**Files:**
- Modify: `README.md`

**Step 1: Add streaming section to README**

Add after existing "Testing" section:

```markdown
3. Streaming
WebSocket endpoints for gRPC streaming:
```
wscat -c ws://localhost:8080/ws/stream/server
> {"stream_id":"1","message":"hello"}
```
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add streaming to README"
```
