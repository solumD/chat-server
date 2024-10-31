package chat

import (
	"context"
	"fmt"

	"github.com/solumD/chat-server/internal/model"
)

func (s *srv) CreateChat(ctx context.Context, chat *model.Chat) (int64, error) {
	if len(chat.Name) == 0 {
		return 0, fmt.Errorf("chat's name can't be empty")
	}

	var chatID int64
	err := s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		var errTx error
		chatID, errTx = s.chatRepository.CreateChat(ctx, chat)
		if errTx != nil {
			return errTx
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return chatID, nil
}
