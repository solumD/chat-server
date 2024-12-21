package chat

import (
	"context"

	"github.com/solumD/chat-server/internal/logger"
	"github.com/solumD/chat-server/pkg/chat_v1"

	"go.uber.org/zap"
)

// ConnectChat подключает пользователя к чату по id
func (s *srv) ConnectChat(ctx context.Context, chatID int64, username string,
	stream chat_v1.ChatV1_ConnectChatServer,
) error {
	logger.Info("connecting user to chat", zap.Int64("chatID", chatID), zap.String("username", username))

	// проверка, что чат есть в базе, а пользователь в нем состоит
	err := s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		errTx := s.chatRepository.CheckChat(ctx, chatID, username)
		if errTx != nil {
			return errTx
		}

		return nil
	})

	if err != nil {
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
