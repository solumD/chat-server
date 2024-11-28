package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/solumD/chat-server/internal/client/db"
	"github.com/solumD/chat-server/internal/client/db/mocks"
	"github.com/solumD/chat-server/internal/logger"
	"github.com/solumD/chat-server/internal/model"
	"github.com/solumD/chat-server/internal/repository"
	repoMocks "github.com/solumD/chat-server/internal/repository/mocks"
	"github.com/solumD/chat-server/internal/service/chat"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestSendMessage(t *testing.T) {
	t.Parallel()
	type chatRepositoryMockFunc func(mc *minimock.Controller) repository.ChatRepository
	type txManagerMockFunc func(mc *minimock.Controller) db.TxManager

	type args struct {
		ctx context.Context
		req *model.Message
	}

	var (
		ctx = context.Background()
		mc  = minimock.NewController(t)

		id   = gofakeit.Int64()
		from = gofakeit.Username()
		text = gofakeit.Fruit()

		repoErr      = fmt.Errorf("repo error")
		emptyFromErr = fmt.Errorf("from can't be empty")
		emptyTextErr = fmt.Errorf("message's text can't be empty")

		req = &model.Message{
			ChatID: id,
			From:   from,
			Text:   text,
		}

		emptyFromReq = &model.Message{
			ChatID: id,
			From:   "",
			Text:   text,
		}

		emptyTextReq = &model.Message{
			ChatID: id,
			From:   from,
			Text:   "",
		}
		res = &emptypb.Empty{}
	)
	defer t.Cleanup(mc.Finish)

	tests := []struct {
		name               string
		args               args
		want               *emptypb.Empty
		err                error
		chatRepositoryMock chatRepositoryMockFunc
		txManagerMock      txManagerMockFunc
	}{
		{
			name: "success from repo",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: res,
			err:  nil,
			chatRepositoryMock: func(mc *minimock.Controller) repository.ChatRepository {
				mock := repoMocks.NewChatRepositoryMock(mc)
				mock.SendMessageMock.Expect(ctx, req).Return(res, nil)
				return mock
			},
			txManagerMock: func(mc *minimock.Controller) db.TxManager {
				mock := mocks.NewTxManagerMock(mc)
				mock.ReadCommittedMock.Set(func(ctx context.Context, f db.Handler) (err error) {
					return f(ctx)
				})
				return mock
			},
		},
		{
			name: "error from repo",
			args: args{
				ctx: ctx,
				req: req,
			},
			want: nil,
			err:  repoErr,
			chatRepositoryMock: func(mc *minimock.Controller) repository.ChatRepository {
				mock := repoMocks.NewChatRepositoryMock(mc)
				mock.SendMessageMock.Expect(ctx, req).Return(nil, repoErr)
				return mock
			},
			txManagerMock: func(mc *minimock.Controller) db.TxManager {
				mock := mocks.NewTxManagerMock(mc)
				mock.ReadCommittedMock.Set(func(ctx context.Context, f db.Handler) (err error) {
					return f(ctx)
				})
				return mock
			},
		},
		{
			name: "error empty from",
			args: args{
				ctx: ctx,
				req: emptyFromReq,
			},
			want: nil,
			err:  emptyFromErr,
			chatRepositoryMock: func(mc *minimock.Controller) repository.ChatRepository {
				mock := repoMocks.NewChatRepositoryMock(mc)
				return mock
			},
			txManagerMock: func(mc *minimock.Controller) db.TxManager {
				mock := mocks.NewTxManagerMock(mc)
				return mock
			},
		},
		{
			name: "error empty text",
			args: args{
				ctx: ctx,
				req: emptyTextReq,
			},
			want: nil,
			err:  emptyTextErr,
			chatRepositoryMock: func(mc *minimock.Controller) repository.ChatRepository {
				mock := repoMocks.NewChatRepositoryMock(mc)
				return mock
			},
			txManagerMock: func(mc *minimock.Controller) db.TxManager {
				mock := mocks.NewTxManagerMock(mc)
				return mock
			},
		},
	}

	logger.MockInit()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			authRepoMock := tt.chatRepositoryMock(mc)
			txManagerMock := tt.txManagerMock(mc)

			service := chat.NewMockService(authRepoMock, txManagerMock)

			newID, err := service.SendMessage(tt.args.ctx, tt.args.req)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, newID)
		})
	}
}
