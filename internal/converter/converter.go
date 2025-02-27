package converter

import (
	"github.com/solumD/chat-server/internal/model"

	desc "github.com/solumD/chat-server/pkg/chat_v1"
)

// ToChatFromDesc конвертирует модель для создания чата API слоя в
// модель сервисного слоя
func ToChatFromDesc(chat *desc.CreateChatRequest) *model.Chat {
	if chat == nil {
		return nil
	}

	return &model.Chat{
		Name:      chat.Name,
		Usernames: chat.Usernames,
	}
}

// ToMessageFromDesc конвертирует модель сообщения API слоя в
// модель сервисного слоя
func ToMessageFromDesc(message *desc.SendMessageRequest) *model.Message {
	if message == nil {
		return nil
	}

	return &model.Message{
		ChatID: message.Id,
		From:   message.From,
		Text:   message.Text,
	}
}

// ToDescChatInfoFromService конвертирует модель списка информации о чатах из
// сервисного слоя в модель API слоя
func ToDescChatInfoFromService(chatsInfo []*model.Chat) []*desc.ChatInfo {
	if chatsInfo == nil {
		return nil
	}

	descChatsInfo := []*desc.ChatInfo{}
	for _, c := range chatsInfo {
		ci := &desc.ChatInfo{}
		ci.Id = c.ID
		ci.Name = c.Name
		ci.Usernames = c.Usernames

		descChatsInfo = append(descChatsInfo, ci)
	}

	return descChatsInfo
}
