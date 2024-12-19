package chat

import (
	"github.com/solumD/chat-server/pkg/chat_v1"
)

// ConnectChat connects user to chat
func (i *API) ConnectChat(req *chat_v1.ConnectChatRequest,
	stream chat_v1.ChatV1_ConnectChatServer) error {

	err := i.chatService.ConnectChat(stream.Context(), req.GetId(), req.GetUsername(), stream)
	if err != nil {
		return err
	}

	return nil
}
