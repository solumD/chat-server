package chat

import (
	"context"

	"github.com/solumD/chat-server/internal/api/chat/errors"
	"github.com/solumD/chat-server/internal/converter"
	"github.com/solumD/chat-server/internal/logger"
	desc "github.com/solumD/chat-server/pkg/chat_v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SendMessage отправляет запрос в сервисный слой на отправку (сохранение) сообщения
func (i *API) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	convertedMessage := converter.ToMessageFromDesc(req)
	if convertedMessage == nil {
		return nil, errors.ErrDescMessageIsNil
	}
	_, err := i.chatService.SendMessage(ctx, convertedMessage)
	if err != nil {
		return nil, err
	}

	logger.Info("sent message in chat", zap.Int64("chatID", req.GetId()))

	return &emptypb.Empty{}, nil
}
