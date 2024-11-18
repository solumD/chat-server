package auth

import "context"

type AuthClient interface {
	Check(ctx context.Context, endpoint string) error
}

type client struct {
}
