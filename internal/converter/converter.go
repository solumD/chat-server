package converter

import (
	"fmt"

	"github.com/solumD/chat-server/internal/model"

	desc "github.com/solumD/chat-server/pkg/chat_v1"
)

var (
	ErrDescChatIsNil    = fmt.Errorf("desc chat is nil")
	ErrDescMessageIsNil = fmt.Errorf("desc message is nil")
)

// ToChatFromDesc конвертирует модель для создания чата API слоя в
// модель сервисного слоя
func ToChatFromDesc(chat *desc.CreateChatRequest) (*model.Chat, error) {
	if chat == nil {
		return nil, ErrDescChatIsNil
	}

	return &model.Chat{
		Name:      chat.Name,
		Usernames: chat.Usernames,
	}, nil
}

// ToMessageFromDesc конвертирует модель сообщения API слоя в
// модель сервисного слоя
func ToMessageFromDesc(message *desc.SendMessageRequest) (*model.Message, error) {
	if message == nil {
		return nil, ErrDescMessageIsNil
	}

	return &model.Message{
		ChatID: message.Id,
		From:   message.From,
		Text:   message.Text,
	}, nil
}
