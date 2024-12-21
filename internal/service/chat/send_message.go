package chat

import (
	"context"
	"fmt"

	"github.com/solumD/chat-server/internal/model"
	"github.com/solumD/chat-server/pkg/chat_v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

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
