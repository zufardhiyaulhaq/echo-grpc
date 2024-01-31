#################
# Base image
#################
FROM alpine:3.19.1 as client-echo-grpc-base

USER root

RUN addgroup -g 10001 client-echo-grpc && \
    adduser --disabled-password --system --gecos "" --home "/home/client-echo-grpc" --shell "/sbin/nologin" --uid 10001 client-echo-grpc && \
    mkdir -p "/home/client-echo-grpc" && \
    chown client-echo-grpc:0 /home/client-echo-grpc && \
    chmod g=u /home/client-echo-grpc && \
    chmod g=u /etc/passwd
RUN apk add --update --no-cache alpine-sdk curl

ENV USER=client-echo-grpc
USER 10001
WORKDIR /home/client-echo-grpc

#################
# Builder image
#################
FROM golang:1.21-alpine AS client-echo-grpc-builder
RUN apk add --update --no-cache alpine-sdk
WORKDIR /app
COPY . .
RUN make client.build

#################
# Final image
#################
FROM client-echo-grpc-base

COPY --from=client-echo-grpc-builder /app/bin/client-echo-grpc /usr/local/bin

# Command to run the executable
ENTRYPOINT ["client-echo-grpc"]
