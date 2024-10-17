FROM golang:1.23-alpine AS builder

COPY . /github.com/solumD/chat-server
WORKDIR /github.com/solumD/chat-server

RUN go mod download
RUN go build -o ./bin/chat_server cmd/grpc_server/main.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /github.com/solumD/chat-server/bin/chat_server .

ADD .env .

CMD ["./chat_server"]