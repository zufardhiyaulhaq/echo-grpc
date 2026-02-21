# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Set environment variables first
source .env.example

# Run locally
make server.run      # Start gRPC server
make client.run      # Start HTTP client (requires server running)

# Build binaries
make server.build    # Output: ./bin/server-echo-grpc
make client.build    # Output: ./bin/client-echo-grpc

# Docker
make server.image.build
make client.image.build
make grpc.up         # docker-compose up
make grpc.down       # docker-compose down
```

## Testing the Service

```bash
# HTTP endpoint (client forwards to gRPC server)
curl http://localhost:8080/grpc/test

# gRPC health check
grpc-health-probe -addr=127.0.0.1:8081
```

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

## Architecture

**Two-binary system**: Client (HTTP gateway) → Server (gRPC)

- **Server** (`server/`): Pure gRPC server implementing `Server.GetReply`, `Health`, and `StreamingServer` services
- **Client** (`client/`): HTTP server using gorilla/mux that proxies requests to the gRPC server, plus WebSocket handlers for streaming
- **Proto** (`proto/`): Service definitions - `server.proto` (echo), `healthcheck.proto` (gRPC health), `streaming.proto` (streaming RPCs)

**Configuration**: Both binaries use `envconfig` for settings via environment variables. See `.env.example` for all options including keepalive and TLS settings.

**Key endpoints**:
- `GET /grpc/{key}` - Echo request through gRPC server (unary)
- `GET /healthz`, `GET /readyz` - Client health checks
- `WS /ws/stream/bidirectional` - Bidirectional streaming
- `WS /ws/stream/server` - Server streaming (5 echoes)
- `WS /ws/stream/client` - Client streaming (summary on close)
- gRPC `Health.Check` - Server health check

## Proto Generation

Proto files are in `proto/`. Generated Go files (`*.pb.go`, `*_grpc.pb.go`) are committed. Use `protoc` with `protoc-gen-go` and `protoc-gen-go-grpc` plugins to regenerate.
