package chat

import (
	"context"
	"log"

	"github.com/solumD/chat-server/internal/converter"
	desc "github.com/solumD/chat-server/pkg/chat_v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

// SendMessage отправляет запрос в сервисный слой на отправку (сохранение) сообщения
func (i *API) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	convertedMessage, err := converter.ToMessageFromDesc(req)
	if err != nil {
		return nil, err
	}
	_, err = i.chatService.SendMessage(ctx, convertedMessage)
	if err != nil {
		return nil, err
	}

	log.Printf("sent message in chat %d", req.GetId())

	return &emptypb.Empty{}, nil
}
