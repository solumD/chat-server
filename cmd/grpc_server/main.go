package main

import (
	"context"
	"log"

	"github.com/solumD/chat-server/internal/app"
)

// TODO: перенести проверку access токена отсюда в клиент,
// чтобы в клиенте проверять токен и сразу извлекать из него имя
func main() {
	ctx := context.Background()

	a, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	err = a.Run()
	if err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}
