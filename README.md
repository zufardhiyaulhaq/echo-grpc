# Echo gRPC
A client and server gRPC echo.

### Usage
1. Run
```
source .env.example
make server.run
make client.run
```

2. Testing
Call HTTP server in client and client will make a call to gRPC server
```
curl http://localhost:8080/grpc/test
195d1194-5930-42c4-8280-53c3743e5b3b:from server:test

grpc-health-probe -addr=127.0.0.1:8081 -v
parsed options:
> addr=127.0.0.1:8081 conn_timeout=1s rpc_timeout=1s
> tls=false
> spiffe=false
establishing connection
connection established (took 4.2449ms)
time elapsed: connect=4.2449ms rpc=2.793724ms
status: SERVING
```

3. Streaming
WebSocket endpoints for gRPC streaming:
```
wscat -c ws://localhost:8080/ws/stream/server
> {"stream_id":"1","message":"hello"}
```

4. Testing gRPC Server with grpcurl

Unary RPC:
```bash
grpcurl -plaintext -d '{"message":"hello"}' \
  localhost:8081 com.gopay.echo.Server/GetReply
```

Server Streaming (receives 5 echoes with 1s intervals):
```bash
grpcurl -plaintext -d '{"stream_id":"1","message":"hello"}' \
  localhost:8081 com.gopay.echo.streaming.StreamingServer/ServerStream
```

Client Streaming (send multiple messages, receive summary):
```bash
grpcurl -plaintext -d @ localhost:8081 \
  com.gopay.echo.streaming.StreamingServer/ClientStream <<EOF
{"stream_id":"1","sequence_number":1,"message":"msg1"}
{"stream_id":"1","sequence_number":2,"message":"msg2"}
{"stream_id":"1","sequence_number":3,"message":"msg3"}
EOF
```

Bidirectional Streaming:
```bash
grpcurl -plaintext -d @ localhost:8081 \
  com.gopay.echo.streaming.StreamingServer/BidirectionalStream <<EOF
{"stream_id":"1","sequence_number":1,"message":"hello"}
{"stream_id":"1","sequence_number":2,"message":"world"}
EOF
```
