package chat

import (
	"context"

	"github.com/solumD/chat-server/internal/api/chat/errors"
	"github.com/solumD/chat-server/internal/converter"
	"github.com/solumD/chat-server/internal/logger"
	desc "github.com/solumD/chat-server/pkg/chat_v1"

	"go.uber.org/zap"
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

	logger.Info("inserted chat", zap.Int64("chatID", chatID))

	return &desc.CreateChatResponse{
		Id: chatID,
	}, nil
}
