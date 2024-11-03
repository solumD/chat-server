package app

import (
	"context"
	"log"
	"net"

	"github.com/solumD/chat-server/internal/closer"
	"github.com/solumD/chat-server/internal/config"
	desc "github.com/solumD/chat-server/pkg/chat_v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const configPath = ".env"

// App структура приложения
type App struct {
	serviceProvider *serviceProvider
	grpcServer      *grpc.Server
}

// NewApp возвращает объект приложения
func NewApp(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Run запускает приложение
func (a *App) Run() error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	return a.runGRPCServer()
}

// initDeps инициализирует все зависимости
func (a *App) initDeps(ctx context.Context) error {
	err := a.initConfig()
	if err != nil {
		return err
	}

	a.initServiceProvider()
	a.initGRPCServer(ctx)

	return nil
}

func (a *App) initConfig() error {
	err := config.Load(configPath)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) initServiceProvider() {
	a.serviceProvider = NewServiceProvider()
}

func (a *App) initGRPCServer(ctx context.Context) {
	a.grpcServer = grpc.NewServer(grpc.Creds(insecure.NewCredentials()))

	reflection.Register(a.grpcServer)

	desc.RegisterChatV1Server(a.grpcServer, a.serviceProvider.ChatAPI(ctx))

}

func (a *App) runGRPCServer() error {
	lis, err := net.Listen("tcp", a.serviceProvider.GRPCConfig().Address())
	if err != nil {
		return err
	}

	log.Printf("GRPC server is running on %s", a.serviceProvider.GRPCConfig().Address())
	err = a.grpcServer.Serve(lis)
	if err != nil {
		return err
	}

	return nil
}
