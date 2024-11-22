package interceptor

import (
	"context"

	"github.com/solumD/chat-server/internal/client/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type authInterceptor struct {
	authClient auth.Client
}

// NewAuthInterceptor возвращает структуру интерцептора auth
func NewAuthInterceptor(authClient auth.Client) *authInterceptor {
	return &authInterceptor{
		authClient: authClient,
	}
}

// Get возвращает интерцептор, который делает запрос к сервису auth
func (i *authInterceptor) Get() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		if err = i.authClient.Check(ctx, info.FullMethod); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}
