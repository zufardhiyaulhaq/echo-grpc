#################
# Base image
#################
FROM alpine:3.20.0 as server-echo-grpc-base

USER root

RUN addgroup -g 10001 server-echo-grpc && \
    adduser --disabled-password --system --gecos "" --home "/home/server-echo-grpc" --shell "/sbin/nologin" --uid 10001 server-echo-grpc && \
    mkdir -p "/home/server-echo-grpc" && \
    chown server-echo-grpc:0 /home/server-echo-grpc && \
    chmod g=u /home/server-echo-grpc && \
    chmod g=u /etc/passwd
RUN apk add --update --no-cache alpine-sdk curl wget
RUN wget -O /bin/grpc_health_probe-linux-amd64 https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.6/grpc_health_probe-linux-amd64
RUN chmod 755 /bin/grpc_health_probe-linux-amd64

ENV USER=server-echo-grpc
USER 10001
WORKDIR /home/server-echo-grpc

#################
# Builder image
#################
FROM golang:1.22-alpine AS server-echo-grpc-builder
RUN apk add --update --no-cache alpine-sdk
WORKDIR /app
COPY . .
RUN make server.build

#################
# Final image
#################
FROM server-echo-grpc-base

COPY --from=server-echo-grpc-builder /app/bin/server-echo-grpc /usr/local/bin

# Command to run the executable
ENTRYPOINT ["server-echo-grpc"]
