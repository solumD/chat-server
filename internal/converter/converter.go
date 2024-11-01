package converter

import (
	"fmt"

	"github.com/solumD/chat-server/internal/model"

	desc "github.com/solumD/chat-server/pkg/chat_v1"
)

var (
	errDescChatIsNil    = fmt.Errorf("desc chat is nil")
	errDescMessageIsNil = fmt.Errorf("desc message is nil")
)

// ToChatFromDesc конвертирует модель для создания чата API слоя в
// модель сервисного слоя
func ToChatFromDesc(chat *desc.CreateChatRequest) (*model.Chat, error) {
	if chat == nil {
		return nil, fmt.Errorf("convertion failed: %v", errDescChatIsNil)
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
		return nil, fmt.Errorf("convertion failed: %v", errDescMessageIsNil)
	}

	return &model.Message{
		ChatID: message.Id,
		From:   message.From,
		Text:   message.Text,
	}, nil
}
