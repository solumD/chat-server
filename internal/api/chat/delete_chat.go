package chat

import (
	"context"
	"fmt"
	"log"

	"github.com/solumD/chat-server/internal/logger"
	desc "github.com/solumD/chat-server/pkg/chat_v1"

	"go.uber.org/zap"
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

	logger.Info("deleted chat", zap.Int64("chatID", req.GetId()))

	return &emptypb.Empty{}, nil
}
