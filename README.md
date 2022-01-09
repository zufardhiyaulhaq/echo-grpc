# Echo gRPC
A client and server gRPC echo.

### Usage
1. Run
```
source .env.example
make server.run
make clinet.run
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
