package chat

import (
	"github.com/solumD/chat-server/internal/service"
	desc "github.com/solumD/chat-server/pkg/chat_v1"
)

// Implementation сруктура с заглушками gRPC-методов (при их отсутствии) и
// объект сервисного слоя (его интерфейса)
type Implementation struct {
	desc.UnimplementedChatV1Server
	chatService service.ChatService
}

// NewImplementation возвращает новый объект имплементации API-слоя
func NewImplementation(chatService service.ChatService) *Implementation {
	return &Implementation{
		chatService: chatService,
	}
}
