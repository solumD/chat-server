package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	desc "github.com/solumD/chat-server/pkg/chat_v1"
)

const grpcPort = 50052

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterChatV1Server(s, &server{})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serv: %s", err)
	}
}

type server struct {
	desc.UnimplementedChatV1Server
}

// CreateChat creates new chat with given name and users
func (s *server) CreateChat(_ context.Context, req *desc.CreateChatRequest) (*desc.CreateChatResponse, error) {
	log.Printf("[Create] request data |\nchat's name: %v,usernames: %v",
		req.Name,
		req.Usernames,
	)

	return &desc.CreateChatResponse{
		Id: gofakeit.Int64(),
	}, nil
}

// DeleteChat deletes chat by id
func (s *server) DeleteChat(_ context.Context, req *desc.DeleteChatRequest) (*emptypb.Empty, error) {
	log.Printf("[Delete] request data |\nid: %v", req.Id)

	return nil, nil
}

// SendMessage sends message from on user to another
func (s *server) SendMessage(_ context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	log.Printf("[SendMessage] request data |\nfrom: %v, text: %v, timestamp: %v",
		req.Info.From,
		req.Info.Text,
		req.Info.Timestamp,
	)

	return nil, nil
}
