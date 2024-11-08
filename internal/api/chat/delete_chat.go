package chat

import (
	"context"
	"fmt"
	"log"

	desc "github.com/solumD/chat-server/pkg/chat_v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

// DeleteChat отправляет запрос на удаление чата в сервисный слой
func (i *API) DeleteChat(ctx context.Context, req *desc.DeleteChatRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("req is nil")
	}
	_, err := i.chatService.DeleteChat(ctx, req.GetId())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Printf("deleted chat with id %d", req.GetId())

	return &emptypb.Empty{}, nil
}
