FROM golang:1.12-alpine AS builder

WORKDIR /grpc-practice-server

COPY . .

ENV GOPROXY=https://goproxy.cn

RUN GO111MODULE=on CGO_ENABLED=0 go build -o server-bin ./server


FROM alpine

WORKDIR /grpc-practice-server

COPY --from=builder /grpc-practice-server/server-bin   /grpc-practice-server/server
COPY --from=builder /grpc-practice-server/crt          /grpc-practice-server/crt

#ENV GRPC_GO_LOG_SEVERITY_LEVEL=INFO
EXPOSE 10000

CMD ["./server", "-tls", "-cert_file", "crt/cert.pem", "-key_file", "crt/key.pem", "-sleep", "0"]
