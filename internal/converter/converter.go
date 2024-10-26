package converter

import (
	"github.com/solumD/chat-server/internal/model"
	desc "github.com/solumD/chat-server/pkg/chat_v1"
)

// ToUserFromDesc конвертирует модель для создания чата API слоя в
// модель сервисного слоя
func ToChatFromDesc(chat *desc.CreateChatRequest) *model.Chat {
	return &model.Chat{
		Name:      chat.Name,
		Usernames: chat.Usernames,
	}
}

// ToUserFromDesc конвертирует модель сообщения API слоя в
// модель сервисного слоя
func ToMessageFromDesc(message *desc.SendMessageRequest) *model.Message {
	return &model.Message{
		ChatID: message.Id,
		From:   message.From,
		Text:   message.Text,
	}
}