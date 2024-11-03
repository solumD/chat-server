package chat

import (
	"context"
	"log"

	"github.com/solumD/chat-server/internal/api/chat/errors"
	"github.com/solumD/chat-server/internal/converter"
	desc "github.com/solumD/chat-server/pkg/chat_v1"
)

// CreateChat отправляет запрос в сервисный слой на создание чата
func (i *API) CreateChat(ctx context.Context, req *desc.CreateChatRequest) (*desc.CreateChatResponse, error) {
	convertedChat := converter.ToChatFromDesc(req)
	if convertedChat == nil {
		return nil, errors.ErrDescChatIsNil
	}

	chatID, err := i.chatService.CreateChat(ctx, convertedChat)
	if err != nil {
		return nil, err
	}

	log.Printf("inserted chat with id %d", chatID)

	return &desc.CreateChatResponse{
		Id: chatID,
	}, nil
}
