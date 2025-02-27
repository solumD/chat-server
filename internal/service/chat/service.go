package chat

import (
	"context"
	"fmt"

	"github.com/solumD/chat-server/internal/client/db"
	"github.com/solumD/chat-server/internal/model"
	"github.com/solumD/chat-server/internal/repository"
	"github.com/solumD/chat-server/internal/service"

	"google.golang.org/protobuf/types/known/emptypb"
)

// Структура сервисного слоя с объектами репо слоя
// и транзакционного менеджера
type srv struct {
	chatRepository repository.ChatRepository
	txManager      db.TxManager
}

// NewService возвращает объект сервисного слоя
func NewService(chatRepository repository.ChatRepository, txManager db.TxManager) service.ChatService {
	return &srv{
		chatRepository: chatRepository,
		txManager:      txManager,
	}
}

// NewMockService возвращает объект мока сервисного слоя
func NewMockService(deps ...interface{}) service.ChatService {
	serv := srv{}

	for _, v := range deps {
		switch s := v.(type) {
		case repository.ChatRepository:
			serv.chatRepository = s
		case db.TxManager:
			serv.txManager = s
		}
	}

	return &serv
}

// CreateChat отправляет запрос в репо слой на создание чата
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
