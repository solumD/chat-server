package chat

import (
	"context"
	"log"

	desc "github.com/solumD/chat-server/pkg/chat_v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// DeleteChat отправляет запрос на удаление чата в сервисный слой
func (i *Implementation) DeleteChat(ctx context.Context, req *desc.DeleteChatRequest) (*emptypb.Empty, error) {
	_, err := i.chatService.DeleteChat(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	log.Printf("deleted chat with id %d", req.GetId())

	return &emptypb.Empty{}, nil
}
