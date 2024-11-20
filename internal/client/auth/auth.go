package auth

import "context"

// Client интерфейс клиента auth
type Client interface {
	Check(ctx context.Context, endpoint string) error
}

type client struct {
}
