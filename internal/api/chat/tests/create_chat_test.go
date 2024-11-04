package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/solumD/chat-server/internal/api/chat"
	"github.com/solumD/chat-server/internal/api/chat/errors"
	"github.com/solumD/chat-server/internal/model"
	"github.com/solumD/chat-server/internal/service"
	serviceMocks "github.com/solumD/chat-server/internal/service/mocks"
	desc "github.com/solumD/chat-server/pkg/chat_v1"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
)

func TestCreateChat(t *testing.T) {
	t.Parallel()

	type chatServiceMockFunc func(mc *minimock.Controller) service.ChatService

	type args struct {
		ctx context.Context
		req *desc.CreateChatRequest
	}

	var (
		ctx = context.Background()
		mc  = minimock.NewController(t)

		name      = gofakeit.Name()
		usernames = []string{gofakeit.Name(), gofakeit.Name(), gofakeit.Name(), gofakeit.Name()}
		id        = gofakeit.Int64()

		serviceErr = fmt.Errorf("service err")

		req = &desc.CreateChatRequest{
			Name:      name,
			Usernames: usernames,
		}

		info = &model.Chat{
			Name:      name,
			Usernames: usernames,
		}

		res = &desc.CreateChatResponse{
			Id: id,
		}
	)
	defer t.Cleanup(mc.Finish)

	tests := []struct {
		name            string
		args            args
		want            *desc.CreateChatResponse
		err             error
		chatServiceMock chatServiceMockFunc
	}{
		{
			name: "success case",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: res,
			err:  nil,
			chatServiceMock: func(mc *minimock.Controller) service.ChatService {
				mock := serviceMocks.NewChatServiceMock(mc)
				mock.CreateChatMock.Expect(ctx, info).Return(id, nil)
				return mock
			},
		},
		{
			name: "error case",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: nil,
			err:  serviceErr,
			chatServiceMock: func(mc *minimock.Controller) service.ChatService {
				mock := serviceMocks.NewChatServiceMock(mc)
				mock.CreateChatMock.Expect(ctx, info).Return(0, serviceErr)
				return mock
			},
		},
		{
			name: "error req is nil",
			args: args{
				ctx: ctx,
				req: nil,
			},
			want: nil,
			err:  errors.ErrDescChatIsNil,
			chatServiceMock: func(mc *minimock.Controller) service.ChatService {
				mock := serviceMocks.NewChatServiceMock(mc)
				return mock
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			chatServiceMock := tt.chatServiceMock(mc)
			api := chat.NewAPI(chatServiceMock)

			res, err := api.CreateChat(tt.args.ctx, tt.args.req)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, res)
		})
	}
}
