package chat

import (
	"context"
	"fmt"
	"log"

	"github.com/solumD/chat-server/internal/client/db"
	"github.com/solumD/chat-server/internal/model"
	"github.com/solumD/chat-server/internal/repository"

	sq "github.com/Masterminds/squirrel"
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
func (r *repo) CreateChat(ctx context.Context, chat *model.Chat) (int64, error) {
	// добавляем новый чат
	chatID, err := r.createChat(ctx, chat.Name)
	if err != nil {
		return 0, err
	}

	// разделяем полученных юзеров на существующих и несуществующих
	userIDs, newUsers, err := r.divideUsers(ctx, chat.Usernames)
	if err != nil {
		return 0, err
	}

	// добавляем новых пользователей и сохраняем их id
	newIDs, err := r.insertUsers(ctx, newUsers)
	if err != nil {
		return 0, err
	}

	// добавляем новые id к существующим
	userIDs = append(userIDs, newIDs...)

	// соотносим всех юзеров с id чата
	err = r.insertUsersInChats(ctx, chatID, userIDs)
	if err != nil {
		return 0, err
	}

	return chatID, nil
}

// DeleteChat удаляет чат по id
func (r *repo) DeleteChat(ctx context.Context, chatID int64) (*emptypb.Empty, error) {
	// проверяем, существует ли чат с указанными id
	exist, err := r.isChatExist(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, fmt.Errorf("chat %d doesn't exist", chatID)
	}

	// удаляем чат (меняем id_deleted на 1)
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

	_, err = r.db.DB().ExecContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// SendMessage отправляет (сохраняет) сообщение пользователя в чат
func (r *repo) SendMessage(ctx context.Context, message *model.Message) (*emptypb.Empty, error) {
	// проверяем, удален ли чат
	exist, err := r.isChatExist(ctx, message.ChatID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("chat %d doesn't exist", message.ChatID)
	}

	// проверяем, существует ли юзер
	exist, err = r.isUserExistByName(ctx, message.From)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("user %s doesn't exist", message.From) // юзер не найден
	}

	userID, err := r.getUserByName(ctx, message.From)
	if err != nil {
		return nil, err
	}

	// проверяем, состоит ли юзер в указанном чате
	inChat, err := r.isUserInChat(ctx, message.ChatID, userID)
	if err != nil {
		return nil, err
	}
	if !inChat {
		return nil, fmt.Errorf("user %v not in chat %d", message.From, message.ChatID)
	}

	// после всех проверок сохраняем сообщение юзера
	query, args, err := sq.Insert(messagesTable).
		PlaceholderFormat(sq.Dollar).
		Columns(chatIDColumn, userIDColumn, messageTextColumn).
		Values(message.ChatID, userID, message.Text).
		ToSql()

	if err != nil {
		return nil, err
	}

	q := db.Query{
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
