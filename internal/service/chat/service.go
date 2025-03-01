package chat

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/solumD/chat-server/internal/client/db"
	"github.com/solumD/chat-server/internal/logger"
	"github.com/solumD/chat-server/internal/model"
	"github.com/solumD/chat-server/internal/repository"
	"github.com/solumD/chat-server/internal/service"
	"github.com/solumD/chat-server/pkg/chat_v1"
	"go.uber.org/zap"

	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *srv) GetUserChats(ctx context.Context, username string) ([]*model.Chat, error) {
	var chatsInfo []*model.Chat
	err := s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		var errTx error
		chatsInfo, errTx = s.chatRepository.GetUserChats(ctx, username)
		if errTx != nil {
			return errTx
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return chatsInfo, nil
}

// ConnectChat подключает пользователя к чату по id
func (s *srv) ConnectChat(ctx context.Context, chatID int64, username string,
	stream chat_v1.ChatV1_ConnectChatServer,
) error {
	logger.Info("connecting user to chat", zap.Int64("chatID", chatID), zap.String("username", username))

	username = strings.TrimSpace(username)

	// проверка, что чат есть в базе, а пользователь в нем состоит
	err := s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		errTx := s.chatRepository.CheckChat(ctx, chatID, username)
		if errTx != nil {
			return errTx
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to connect user to chat", zap.Error(err))
		return err
	}

	s.mu.Lock()
	// проверяем, соединялся ли уже кто-то с чатом, если нет, то создаем соединение и канал для сообщений
	if _, exist := s.chatStreams[chatID]; !exist {
		s.chatStreams[chatID] = make(map[string]chat_v1.ChatV1_ConnectChatServer)
		s.msgChans[chatID] = make(chan *chat_v1.Message, 100)
	}

	// добавляем или заменяем стрим пользователя и получаем канал с сообщениями
	s.chatStreams[chatID][username] = stream
	chatMsgChan := s.msgChans[chatID]
	s.mu.Unlock()

	logger.Info("connected user to chat", zap.Int64("chatID", chatID), zap.String("username", username))

	for {
		select {
		// если от кого-то пришло сообщение, то отправляем его всем подлюченным пользователям
		case msg, ok := <-chatMsgChan:
			if !ok {
				return nil
			}

			for _, st := range s.chatStreams[chatID] {
				if err := st.Send(msg); err != nil {
					return err
				}
			}

		case <-stream.Context().Done():

			s.mu.Lock()
			// удаляем подключение пользователя
			delete(s.chatStreams[chatID], username)

			// если подключений не осталось, то удаляем из мапы чат и канал с его сообщениями
			if len(s.chatStreams[chatID]) == 0 {
				delete(s.chatStreams, chatID)
				delete(s.msgChans, chatID)
			}
			s.mu.Unlock()

			return nil
		}
	}

}

// SendMessage сохраняет сообщение в репо и отправляет его в чат через stream
func (s *srv) SendMessage(ctx context.Context, message *model.Message) (*emptypb.Empty, error) {
	// проверяем, существует ли канал для сообщений чата
	s.mu.RLock()
	chatMsgChan, exist := s.msgChans[message.ChatID]
	s.mu.RUnlock()

	if !exist {
		return nil, fmt.Errorf("chat's %d connection not exist. connect to create it", message.ChatID)
	}

	// проверяем, подключен ли пользователь к чату
	s.mu.RLock()
	_, exist = s.chatStreams[message.ChatID][message.From]
	s.mu.RUnlock()

	if !exist {
		return nil, fmt.Errorf("user %s is not connected to chat. connect to chat", message.From)
	}

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

	chatMsgChan <- &chat_v1.Message{
		From: message.From,
		Text: message.Text,
	}

	return &emptypb.Empty{}, nil
}
