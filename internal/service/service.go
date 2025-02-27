package service

import (
	"context"

	"github.com/solumD/chat-server/internal/model"

	"google.golang.org/protobuf/types/known/emptypb"
)

// ChatService - интерфейс сервисного слоя
type ChatService interface {
	CreateChat(ctx context.Context, chat *model.Chat) (int64, error)
	DeleteChat(ctx context.Context, chatID int64) (*emptypb.Empty, error)
	GetUserChats(ctx context.Context, username string) ([]*model.Chat, error)
	SendMessage(ctx context.Context, message *model.Message) (*emptypb.Empty, error)
}
