package chat

import (
	"sync"

	"github.com/solumD/chat-server/internal/client/db"
	"github.com/solumD/chat-server/internal/repository"
	"github.com/solumD/chat-server/internal/service"
	"github.com/solumD/chat-server/pkg/chat_v1"
)

// Структура сервисного слоя с объектами репо слоя
// и транзакционного менеджера
type srv struct {
	chatRepository repository.ChatRepository
	txManager      db.TxManager

	chatStreams map[int64]map[string]chat_v1.ChatV1_ConnectChatServer
	msgChans    map[int64]chan *chat_v1.Message
	mu          *sync.RWMutex
}

// NewService возвращает объект сервисного слоя
func NewService(chatRepository repository.ChatRepository, txManager db.TxManager) service.ChatService {
	return &srv{
		chatRepository: chatRepository,
		txManager:      txManager,
		chatStreams:    make(map[int64]map[string]chat_v1.ChatV1_ConnectChatServer),
		msgChans:       make(map[int64]chan *chat_v1.Message),
		mu:             &sync.RWMutex{},
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
