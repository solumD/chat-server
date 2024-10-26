package chat

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

// DeleteChat отправляет запрос в репо слой на удаление чата
func (s *srv) DeleteChat(ctx context.Context, chatID int64) (*emptypb.Empty, error) {
	err := s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		var errTx error
		_, errTx = s.chatRepository.DeleteChat(ctx, chatID)
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
