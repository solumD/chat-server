package chat

import (
	"github.com/solumD/chat-server/internal/service"
	desc "github.com/solumD/chat-server/pkg/chat_v1"
)

// API сруктура с заглушками gRPC-методов (при их отсутствии) и
// объект сервисного слоя (его интерфейса)
type API struct {
	desc.UnimplementedChatV1Server
	chatService service.ChatService
}

// NewAPI возвращает новый объект имплементации API-слоя
func NewAPI(chatService service.ChatService) *API {
	return &API{
		chatService: chatService,
	}
}
