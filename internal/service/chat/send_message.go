package chat

import (
	"context"
	"fmt"

	"github.com/solumD/chat-server/internal/model"

	"google.golang.org/protobuf/types/known/emptypb"
)

// SendMessage отправляет запрос в репо слой на отправку (сохранение) сообщения
func (s *srv) SendMessage(ctx context.Context, message *model.Message) (*emptypb.Empty, error) {
	if len(message.From) == 0 {
		return nil, fmt.Errorf("from can't be empty")
	}
	if len(message.Text) == 0 {
		return nil, fmt.Errorf("message's text can't be empty")
	}

	err := s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		var errTx error
		_, errTx = s.chatRepository.SendMessage(ctx, message)
		if errTx != nil {
			return errTx
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
