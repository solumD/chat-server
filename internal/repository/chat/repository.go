package chat

import (
	"context"
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/solumD/chat-server/internal/client/db"
	"github.com/solumD/chat-server/internal/model"
	"github.com/solumD/chat-server/internal/repository"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// названия таблиц
	usersTable        = "users"
	chatsTable        = "chats"
	usersInChatsTable = "users_in_chats"
	messagesTable     = "messages"

	// названия колонок (некоторые участвуют в нескольких таблицах)
	idColumn          = "id"
	usernameColumn    = "username"
	chatNameColumn    = "chat_name"
	chatIDColumn      = "chat_id"
	userIDColumn      = "user_id"
	messageTextColumn = "message_text"
	createdAtColumn   = "created_at"
	isDeletedColumn   = "is_deleted"
)

// Структура репо с клиентом базы данных (интерфейсом)
type repo struct {
	db db.Client
}

// NewRepository возвращает новый объект репо слоя
func NewRepository(db db.Client) repository.ChatRepository {
	return &repo{
		db: db,
	}
}

// CreateChat создает чат
func (r *repo) CreateChat(ctx context.Context, chat model.Chat) (int64, error) {
	// добавляем новый чат
	query, args, err := sq.Insert(chatsTable).
		PlaceholderFormat(sq.Dollar).
		Columns(chatNameColumn).
		Values(chat.Name).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return 0, err
	}

	q := db.Query{
		Name:     "chat_repository.CreateChat",
		QueryRaw: query,
	}

	var chatID int64
	err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&chatID)
	if err != nil {
		return 0, err
	}

	userIDs := []int64{}   // существующие пользователи
	newUsers := []string{} // новые пользователи (у них пока что нет id)

	for _, user := range chat.Usernames {
		query, args, err := sq.Select(idColumn).
			From(usersTable).
			PlaceholderFormat(sq.Dollar).
			Where(sq.Eq{usernameColumn: user}).ToSql()

		if err != nil {
			return 0, err
		}

		q := db.Query{
			Name:     "chat_repository.CreateChat",
			QueryRaw: query,
		}

		var userID int64
		err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&userID)
		if err == pgx.ErrNoRows {
			newUsers = append(newUsers, user) // сохраняем имя нового пользователя
		} else if err != nil {
			return 0, err
		} else {
			userIDs = append(userIDs, userID) // сохраняем id уже существующего пользователя
		}
	}

	// добавляем новых пользователей и сохраняем их id
	for _, name := range newUsers {
		query, args, err := sq.Insert(usersTable).
			PlaceholderFormat(sq.Dollar).
			Columns(usernameColumn).
			Values(name).
			Suffix("RETURNING id").
			ToSql()

		if err != nil {
			return 0, err
		}

		q := db.Query{
			Name:     "chat_repository.CreateChat",
			QueryRaw: query,
		}

		var newUserID int64
		err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&newUserID)
		if err != nil {
			return 0, err
		}

		userIDs = append(userIDs, newUserID) // сохраняем id
	}

	// добавляем id чата и id всех его пользователей в таблицу users_in_chats
	for _, userID := range userIDs {
		query, args, err := sq.Insert(usersInChatsTable).
			PlaceholderFormat(sq.Dollar).
			Columns(chatIDColumn, userIDColumn).
			Values(chatID, userID).
			ToSql()

		if err != nil {
			return 0, err
		}

		q := db.Query{
			Name:     "chat_repository.CreateChat",
			QueryRaw: query,
		}

		_, err = r.db.DB().ExecContext(ctx, q, args...)
		if err != nil {
			return 0, err
		}
	}

	log.Printf("created chat with name %s", chat.Name)
	return chatID, nil
}

// DeleteChat удаляет чат по id
func (r *repo) DeleteChat(ctx context.Context, chatID int64) (*emptypb.Empty, error) {
	query, args, err := sq.Update(chatsTable).
		PlaceholderFormat(sq.Dollar).
		Set(isDeletedColumn, 1).
		Where(sq.Eq{idColumn: chatID}).ToSql()

	if err != nil {
		return nil, err
	}

	q := db.Query{
		Name:     "chat_repository.DeleteChat",
		QueryRaw: query,
	}

	res, err := r.db.DB().ExecContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	log.Printf("updated %d rows", res.RowsAffected())

	return &emptypb.Empty{}, nil
}

// SendMessage отправляет (сохраняет) сообщение пользователя в чат
func (r *repo) SendMessage(ctx context.Context, message model.Message) (*emptypb.Empty, error) {
	// выбираем чат с указанным id
	query, args, err := sq.Select(isDeletedColumn).
		From(chatsTable).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{idColumn: message.ChatID}).
		GroupBy(idColumn).
		ToSql()

	if err != nil {
		return nil, err
	}

	q := db.Query{
		Name:     "chat_repository.SendMessage",
		QueryRaw: query,
	}

	var isDeleted int
	err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&isDeleted)
	if err == pgx.ErrNoRows {
		return nil, errors.Errorf("chat %d was not found ", message.ChatID) // чата с указанными id не найдено
	} else if err != nil {
		return nil, err
	}

	// проверяем удален ли чат
	if isDeleted == 1 {
		return nil, fmt.Errorf("chat %d was deleted", message.ChatID)
	}

	// выбираем id юзера с указанными именем
	query, args, err = sq.Select(idColumn).
		From(usersTable).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{usernameColumn: message.From}).
		ToSql()

	if err != nil {
		return nil, err
	}

	q = db.Query{
		Name:     "chat_repository.SendMessage",
		QueryRaw: query,
	}

	var userID int
	err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&userID)
	if err == pgx.ErrNoRows {
		return nil, errors.Errorf("user %s doesn't exist", message.From) // юзер не найден
	} else if err != nil {
		return nil, err
	}

	// проверяем, состоит ли юзер в указанном чате
	query, args, err = sq.Select("").
		From(usersInChatsTable).
		PlaceholderFormat(sq.Dollar).
		Where(sq.And{sq.Eq{chatIDColumn: message.ChatID}, sq.Eq{userIDColumn: userID}}).
		ToSql()

	if err != nil {
		return nil, err
	}

	q = db.Query{
		Name:     "chat_repository.SendMessage",
		QueryRaw: query,
	}

	err = r.db.DB().QueryRowContext(ctx, q, args...).Scan()
	if err == pgx.ErrNoRows {
		return nil, errors.Errorf("user %s doesn't exist in chat %d", message.From, message.ChatID) // юзер не состоит в указанном чате
	} else if err != nil {
		return nil, err
	}

	// после всех проверок сохраняем сообщение юзера
	query, args, err = sq.Insert(messagesTable).
		PlaceholderFormat(sq.Dollar).
		Columns(chatIDColumn, userIDColumn, messageTextColumn).
		Values(message.ChatID, userID, message.Text).
		ToSql()

	if err != nil {
		return nil, err
	}

	q = db.Query{
		Name:     "chat_repository.SendMessage",
		QueryRaw: query,
	}

	res, err := r.db.DB().ExecContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	log.Printf("inserted %d message", res.RowsAffected())

	return &emptypb.Empty{}, nil
}
