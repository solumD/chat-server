package chat

import (
	"github.com/solumD/chat-server/internal/client/db"
	"github.com/solumD/chat-server/internal/repository"
	"github.com/solumD/chat-server/internal/service"
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
