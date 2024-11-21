package auth

import (
	"context"
	"fmt"

	"github.com/solumD/auth/pkg/access_v1"
)

// Client интерфейс клиента auth
type Client interface {
	Check(ctx context.Context, endpoint string) error
}

type client struct {
	accessClient access_v1.AccessV1Client
}

// New возвращает новый объект клиента auth
func New(accessClient access_v1.AccessV1Client) Client {
	return &client{
		accessClient: accessClient,
	}
}

// Check отправляет запрос в сервис auth на проверкку доступа
func (c *client) Check(ctx context.Context, endpoint string) error {
	req := &access_v1.CheckRequest{
		EndpointAddress: endpoint,
	}

	if _, err := c.accessClient.Check(ctx, req); err != nil {
		return fmt.Errorf("access check error: %v", err)
	}

	return nil
}
