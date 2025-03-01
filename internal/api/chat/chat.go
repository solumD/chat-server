package chat

import (
	"context"
	"fmt"
	"log"

	"github.com/solumD/chat-server/internal/api/chat/errors"
	"github.com/solumD/chat-server/internal/converter"
	"github.com/solumD/chat-server/internal/logger"
	"github.com/solumD/chat-server/internal/service"
	desc "github.com/solumD/chat-server/pkg/chat_v1"
	"google.golang.org/protobuf/types/known/emptypb"

	"go.uber.org/zap"
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

// CreateChat отправляет запрос в сервисный слой на создание чата
func (i *API) CreateChat(ctx context.Context, req *desc.CreateChatRequest) (*desc.CreateChatResponse, error) {
	convertedChat := converter.ToChatFromDesc(req)
	if convertedChat == nil {
		return nil, errors.ErrDescChatIsNil
	}

	chatID, err := i.chatService.CreateChat(ctx, convertedChat)
	if err != nil {
		return nil, err
	}

	logger.Info("inserted chat", zap.Int64("chatID", chatID))

	return &desc.CreateChatResponse{
		Id: chatID,
	}, nil
}

// DeleteChat отправляет запрос на удаление чата в сервисный слой
func (i *API) DeleteChat(ctx context.Context, req *desc.DeleteChatRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("req is nil")
	}
	_, err := i.chatService.DeleteChat(ctx, req.GetId())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	logger.Info("deleted chat", zap.Int64("chatID", req.GetId()))

	return &emptypb.Empty{}, nil
}

func (i *API) GetUserChats(ctx context.Context, req *desc.GetUserChatsRequest) (*desc.GetUserChatsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("req is nil")
	}

	chatsInfo, err := i.chatService.GetUserChats(ctx, req.GetUsername())
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	logger.Info("got user's chats", zap.String("username", req.GetUsername()), zap.Any("chatsInfo", chatsInfo))

	return &desc.GetUserChatsResponse{
		Chats: converter.ToDescChatInfoFromService(chatsInfo),
	}, nil
}

// ConnectChat подключает юзера к чату и возвращает stream сообщений
func (i *API) ConnectChat(req *desc.ConnectChatRequest,
	stream desc.ChatV1_ConnectChatServer) error {

	err := i.chatService.ConnectChat(stream.Context(), req.GetId(), req.GetUsername(), stream)
	if err != nil {
		return err
	}

	return nil
}

// SendMessage отправляет запрос в сервисный слой на отправку (сохранение) сообщения
func (i *API) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	convertedMessage := converter.ToMessageFromDesc(req)
	if convertedMessage == nil {
		return nil, errors.ErrDescMessageIsNil
	}
	_, err := i.chatService.SendMessage(ctx, convertedMessage)
	if err != nil {
		return nil, err
	}

	logger.Info("sent message in chat", zap.Int64("chatID", req.GetId()))

	return &emptypb.Empty{}, nil
}
