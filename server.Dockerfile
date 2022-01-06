#################
# Base image
#################
FROM alpine:3.12 as server-echo-grpc-base

USER root

RUN addgroup -g 10001 server-echo-grpc && \
    adduser --disabled-password --system --gecos "" --home "/home/server-echo-grpc" --shell "/sbin/nologin" --uid 10001 server-echo-grpc && \
    mkdir -p "/home/server-echo-grpc" && \
    chown server-echo-grpc:0 /home/server-echo-grpc && \
    chmod g=u /home/server-echo-grpc && \
    chmod g=u /etc/passwd
RUN apk add --update --no-cache alpine-sdk curl

ENV USER=server-echo-grpc
USER 10001
WORKDIR /home/server-echo-grpc

#################
# Builder image
#################
FROM golang:1.16-alpine AS server-echo-grpc-builder
RUN apk add --update --no-cache alpine-sdk
WORKDIR /app
COPY . .
RUN make build

#################
# Final image
#################
FROM server-echo-grpc-base

COPY --from=server-echo-grpc-builder /app/bin/server-echo-grpc /usr/local/bin

# Command to run the executable
ENTRYPOINT ["server-echo-grpc"]
