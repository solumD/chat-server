package chat

import (
	"context"

	"github.com/solumD/chat-server/internal/model"

	"google.golang.org/protobuf/types/known/emptypb"
)

// SendMessage отправляет запрос в репо слой на отправку (сохранение) сообщения
func (s *srv) SendMessage(ctx context.Context, message *model.Message) (*emptypb.Empty, error) {
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
