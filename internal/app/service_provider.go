package app

import (
	"context"
	"log"

	api "github.com/solumD/chat-server/internal/api/chat"
	"github.com/solumD/chat-server/internal/client/db"
	"github.com/solumD/chat-server/internal/client/db/pg"
	"github.com/solumD/chat-server/internal/client/db/transaction"
	"github.com/solumD/chat-server/internal/closer"
	"github.com/solumD/chat-server/internal/config"
	"github.com/solumD/chat-server/internal/repository"
	chatRepo "github.com/solumD/chat-server/internal/repository/chat"
	"github.com/solumD/chat-server/internal/service"
	chatSrv "github.com/solumD/chat-server/internal/service/chat"
)

// Структура приложения со всеми зависимостями
type serviceProvider struct {
	pgConfig   config.PGConfig
	grpcConfig config.GRPCConfig

	dbClient  db.Client
	txManager db.TxManager

	chatRepository repository.ChatRepository
	chatService    service.ChatService
	chatImpl       *api.API
}

// NewServiceProvider возвращает новый объект API слоя
func NewServiceProvider() *serviceProvider {
	return &serviceProvider{}
}

// PGConfig инициализирует конфиг postgres
func (s *serviceProvider) PGConfig() config.PGConfig {
	if s.pgConfig == nil {
		cfg, err := config.NewPGConfig()
		if err != nil {
			log.Fatalf("failed to get pg config: %v", err)
		}

		s.pgConfig = cfg
	}

	return s.pgConfig
}

// GRPCConfig инициализирует конфиг grpc
func (s *serviceProvider) GRPCConfig() config.GRPCConfig {
	if s.grpcConfig == nil {
		cfg, err := config.NewGRPCConfig()
		if err != nil {
			log.Fatalf("failed to get grpc config: %v", err)
		}

		s.grpcConfig = cfg
	}

	return s.grpcConfig
}

// DBClient инициализирует клиент базы данных
func (s *serviceProvider) DBClient(ctx context.Context) db.Client {
	if s.dbClient == nil {
		cl, err := pg.New(ctx, s.PGConfig().DSN())
		if err != nil {
			log.Fatalf("failed to create a db client: %v", err)
		}

		err = cl.DB().Ping(ctx)
		if err != nil {
			log.Fatalf("ping error: %v", err)
		}
		closer.Add(cl.Close)

		s.dbClient = cl
	}

	return s.dbClient
}

// TxManager инициализирует менеджер транзакций
func (s *serviceProvider) TxManager(ctx context.Context) db.TxManager {
	if s.txManager == nil {
		s.txManager = transaction.NewTransactionManager(s.DBClient(ctx).DB())
	}

	return s.txManager
}

// ChatRepository инициализирует репо слой
func (s *serviceProvider) ChatReposistory(ctx context.Context) repository.ChatRepository {
	if s.chatRepository == nil {
		s.chatRepository = chatRepo.NewRepository(s.DBClient(ctx))
	}

	return s.chatRepository
}

// ChatService иницилизирует сервисный слой
func (s *serviceProvider) ChatService(ctx context.Context) service.ChatService {
	if s.chatService == nil {
		s.chatService = chatSrv.NewService(s.ChatReposistory(ctx), s.TxManager(ctx))
	}

	return s.chatService
}

// ChatAPI инициализирует api слой
func (s *serviceProvider) ChatAPI(ctx context.Context) *api.API {
	if s.chatImpl == nil {
		s.chatImpl = api.NewAPI(s.ChatService(ctx))
	}

	return s.chatImpl
}
