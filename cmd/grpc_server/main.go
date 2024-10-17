package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	sq "github.com/Masterminds/squirrel"
	"github.com/solumD/chat-server/internal/config"
	desc "github.com/solumD/chat-server/pkg/chat_v1"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", ".env", "path to config file")
}

func main() {
	flag.Parse()

	err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	grpcConfig, err := config.NewGRPCConfig()
	if err != nil {
		log.Fatalf("failed to get grpc config: %v", err)
	}

	lis, err := net.Listen("tcp", grpcConfig.Address())
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	pgConfig, err := config.NewPGConfig()
	if err != nil {
		log.Fatalf("failed to get pg config: %v", err)
	}

	ctx := context.Background()
	pgPool, err := pgxpool.Connect(ctx, pgConfig.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pgPool.Close()

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterChatV1Server(s, &server{pool: pgPool})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serv: %s", err)
	}
}

type server struct {
	desc.UnimplementedChatV1Server
	pool *pgxpool.Pool
}

// CreateChat creates new chat with given name and users
func (s *server) CreateChat(ctx context.Context, req *desc.CreateChatRequest) (*desc.CreateChatResponse, error) {
	fn := "CreateChat"
	log.Printf("[%s] request data | chat's name: %v, usernames: %v", fn, req.Name, req.Usernames)

	// добавляем новый чат
	query, args, err := sq.Insert("chats").
		PlaceholderFormat(sq.Dollar).
		Columns("chat_name").
		Values(req.Name).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		log.Printf("%s: failed to create builder: %v", fn, err)
		return nil, err
	}

	var chatID int64
	err = s.pool.QueryRow(ctx, query, args...).Scan(&chatID)
	if err != nil {
		log.Printf("%s: failed to insert chat: %v", fn, err)
		return nil, err
	}

	userIDs := []int64{}   // существующие пользователи
	newUsers := []string{} // новые пользователи (у них пока что нет id)

	for _, user := range req.Usernames {
		query, args, err := sq.Select("id").
			From("users").
			PlaceholderFormat(sq.Dollar).
			Where(sq.Eq{"username": user}).ToSql()

		if err != nil {
			log.Printf("%s: failed to create query: %v", fn, err)
			return nil, err
		}

		row := s.pool.QueryRow(ctx, query, args...)
		var userID int64
		err = row.Scan(&userID)
		if err == pgx.ErrNoRows {
			newUsers = append(newUsers, user) // сохраняем имя нового пользователя
		} else if err != nil {
			log.Printf("%s: failed to select user: %v", fn, err)
			return nil, err
		} else {
			userIDs = append(userIDs, userID) // сохраняем id уже существующего пользователя
		}
	}

	// добавляем новых пользователей и сохраняем их id
	for _, name := range newUsers {
		query, args, err := sq.Insert("users").
			PlaceholderFormat(sq.Dollar).
			Columns("username").
			Values(name).
			Suffix("RETURNING id").
			ToSql()
		if err != nil {
			log.Printf("%s: failed to create query: %v", fn, err)
			return nil, err
		}

		var newUserID int64
		if err = s.pool.QueryRow(ctx, query, args...).Scan(&newUserID); err != nil {
			log.Printf("%s: failed to insert user: %v", fn, err)
			return nil, err
		}

		userIDs = append(userIDs, newUserID) // сохраняем id
	}

	// добавляем id чата и id всех его пользователей в таблицу users_in_chats
	for _, userID := range userIDs {
		query, args, err := sq.Insert("users_in_chats").
			PlaceholderFormat(sq.Dollar).
			Columns("chat_id", "user_id").
			Values(chatID, userID).
			ToSql()

		if err != nil {
			log.Printf("%s: failed to create query: %v", fn, err)
			return nil, err
		}

		if _, err = s.pool.Exec(ctx, query, args...); err != nil {
			log.Printf("%s: failed to insert user: %v", fn, err)
			return nil, err
		}
	}

	log.Printf("%s: created chat with name %s", fn, req.Name)
	return &desc.CreateChatResponse{
		Id: chatID,
	}, nil
}

// DeleteChat deletes chat by id
func (s *server) DeleteChat(ctx context.Context, req *desc.DeleteChatRequest) (*emptypb.Empty, error) {
	fn := "DeleteChat"
	log.Printf("[%s] request data |\nid: %v", fn, req.Id)

	builderDeleteChat := sq.Update("chats").
		PlaceholderFormat(sq.Dollar).
		Set("is_deleted", 1).
		Where(sq.Eq{"id": req.Id})

	query, args, err := builderDeleteChat.ToSql()
	if err != nil {
		log.Printf("%s: failed to create builder: %v", fn, err)
		return nil, err
	}

	res, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("%s: failed to delete chat: %v", fn, err)
		return nil, err
	}

	log.Printf("%s updated %d rows", fn, res.RowsAffected())

	return &emptypb.Empty{}, nil
}

// SendMessage sends message from on user to another
func (s *server) SendMessage(ctx context.Context, req *desc.SendMessageRequest) (*emptypb.Empty, error) {
	fn := "SendMessage"
	log.Printf("[%s] request data |\nchat: %v from: %v, text: %v",
		fn,
		req.Id,
		req.From,
		req.Text,
	)

	// выбираем чат с указанным id
	query, args, err := sq.Select("is_deleted").
		From("chats").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"id": req.Id}).
		GroupBy("id").
		ToSql()

	if err != nil {
		log.Printf("%s: failed to create query: %v", fn, err)
		return nil, err
	}

	row := s.pool.QueryRow(ctx, query, args...)
	var isDeleted int
	err = row.Scan(&isDeleted)
	if err == pgx.ErrNoRows {
		log.Printf("%s: chat %d not found", fn, req.Id) // чата с указанными id не найдено
		return nil, fmt.Errorf("chat %d was not found ", req.Id)
	} else if err != nil {
		log.Printf("%s: failed to select count of chat: %v", fn, err)
		return nil, err
	}

	// проверяем удален ли чат
	if isDeleted == 1 {
		log.Printf("%s: chat %d was deleted", fn, req.Id)
		return nil, fmt.Errorf("chat %d was deleted", req.Id)
	}

	// выбираем id юзера с указанными именем
	query, args, err = sq.Select("id").
		From("users").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"username": req.From}).
		ToSql()

	if err != nil {
		log.Printf("%s: failed to create query: %v", fn, err)
		return nil, err
	}

	var userID int
	row = s.pool.QueryRow(ctx, query, args...)
	err = row.Scan(&userID)
	if err == pgx.ErrNoRows {
		log.Printf("%s: user %s doesn't exist", fn, req.From) // юзер не найден
		return nil, fmt.Errorf("user %s doesn't exist", req.From)
	} else if err != nil {
		log.Printf("%s: failed to select user: %v", fn, err)
		return nil, err
	}

	// проверяем, состоит ли юзер в указанном чате
	query, args, err = sq.Select("").
		From("users_in_chats").
		PlaceholderFormat(sq.Dollar).
		Where(sq.And{sq.Eq{"chat_id": req.Id}, sq.Eq{"user_id": userID}}).
		ToSql()

	if err != nil {
		log.Printf("%s: failed to create query: %v", fn, err)
		return nil, err
	}

	row = s.pool.QueryRow(ctx, query, args...)
	err = row.Scan()
	if err == pgx.ErrNoRows {
		log.Printf("%s: user %s doesn't exist in chat %d", fn, req.From, req.Id) // юзер не состоит в указанном чате
		return nil, fmt.Errorf("user %s doesn't exist in chat %d", req.From, req.Id)
	} else if err != nil {
		log.Printf("%s: failed to select count of users_and_chats: %v", fn, err)
		return nil, err
	}

	// после всех проверок сохраняем сообщение юзера
	query, args, err = sq.Insert("messages").
		PlaceholderFormat(sq.Dollar).
		Columns("chat_id", "user_id", "message_text").
		Values(req.Id, userID, req.Text).
		ToSql()

	if err != nil {
		log.Printf("%s: failed to create query: %v", fn, err)
		return nil, err
	}

	res, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("%s: failed to insert message: %v", fn, err)
		return nil, err
	}

	log.Printf("%s: inserted %d message", fn, res.RowsAffected())

	return &emptypb.Empty{}, nil
}
