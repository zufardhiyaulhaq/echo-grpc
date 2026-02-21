# gRPC Streaming Design

## Overview

Add three gRPC streaming RPCs to the echo-grpc project: bidirectional streaming, client streaming, and server streaming. Expose all three via WebSocket endpoints in the client HTTP server.

## Requirements

- Three separate streaming RPCs in a new `StreamingServer` service
- New message types with `stream_id`, `sequence_number`, `timestamp` fields
- Server echoes with transformation: "from server:" prefix + server-generated metadata
- WebSocket HTTP endpoints for all streaming types

## Proto Definition

New file `proto/streaming.proto`:

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

## Server Implementation

New file `server/streaming.go` implementing `StreamingServer`:

| RPC | Behavior |
|-----|----------|
| ClientStream | Receive multiple messages, respond with summary at end: "from server: received N messages" |
| ServerStream | Receive one message, echo back 5 times with 1-second intervals |
| BidirectionalStream | Echo each message immediately with transformation |

Register `StreamingServer` alongside existing `Server` and `Health` services in `server/main.go`.

## Client Implementation

New file `client/pkg/server/websocket.go` with WebSocket handlers.

**New endpoints:**
- `GET /ws/stream/bidirectional` - Bidirectional streaming
- `GET /ws/stream/client` - Client streaming
- `GET /ws/stream/server` - Server streaming

**WebSocket message format (JSON):**
```json
// Client sends:
{"stream_id": "abc", "sequence_number": 1, "message": "hello"}

// Server responds:
{"stream_id": "abc", "sequence_number": 1, "timestamp": 1234567890, "response": "from server: hello", "success": true}
```

**Behavior:**

| Endpoint | Client Action | Server Action |
|----------|---------------|---------------|
| `/ws/stream/bidirectional` | Send messages anytime | Echo each immediately |
| `/ws/stream/client` | Send multiple, then close | Summary after close |
| `/ws/stream/server` | Send one message | 5 echoes with 1s intervals |

**New dependency:** `github.com/gorilla/websocket`

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                           CLIENT                                │
│                                                                 │
│  ┌─────────────┐     ┌──────────────────┐     ┌─────────────┐  │
│  │ HTTP/WS     │     │ WebSocket        │     │ gRPC        │  │
│  │ Handlers    │────▶│ Handlers         │────▶│ Streaming   │  │
│  │ (mux)       │     │ (gorilla/ws)     │     │ Client      │  │
│  └─────────────┘     └──────────────────┘     └──────┬──────┘  │
│        │                                              │         │
│        │ Unary (/grpc/{key})                         │         │
│        ▼                                              │         │
│  ┌─────────────┐                                     │         │
│  │ gRPC Unary  │                                     │         │
│  │ Client      │─────────────────────────────────────┤         │
│  └─────────────┘                                     │         │
└──────────────────────────────────────────────────────┼─────────┘
                                                       │
                                              gRPC (port 8081)
                                                       │
┌──────────────────────────────────────────────────────┼─────────┐
│                           SERVER                     │         │
│                                                      ▼         │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    gRPC Server                          │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │   │
│  │  │ Server      │  │ Streaming   │  │ Health          │  │   │
│  │  │ (unary)     │  │ Server      │  │ (health check)  │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────────┘  │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Error Handling

**WebSocket errors (client):**
- Invalid JSON → Error response: `{"success": false, "response": "invalid message format"}`
- gRPC failure → Close WebSocket with code 1011 + reason
- Client disconnect → Clean up gRPC stream gracefully

**gRPC errors (server):**
- Empty stream_id → `InvalidArgument` status
- Stream cancelled → Log and clean up
- Internal error → `Internal` status with description

**Timeouts:**
- WebSocket ping/pong every 30 seconds
- gRPC streams use existing keepalive settings
