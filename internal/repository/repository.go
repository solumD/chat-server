package repository

import (
	"context"

	"github.com/solumD/chat-server/internal/model"

	"google.golang.org/protobuf/types/known/emptypb"
)

// ChatRepository - интерфейс репо слоя
type ChatRepository interface {
	CreateChat(ctx context.Context, chat *model.Chat) (int64, error)
	DeleteChat(ctx context.Context, chatID int64) (*emptypb.Empty, error)
	SendMessage(ctx context.Context, message *model.Message) (*emptypb.Empty, error)
}
