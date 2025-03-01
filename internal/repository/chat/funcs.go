package chat

import (
	"context"
	"errors"

	"github.com/solumD/chat-server/internal/client/db"
	"github.com/solumD/chat-server/internal/model"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
)

// createChat сохраняет чат в БД
func (r *repo) createChat(ctx context.Context, name string) (int64, error) {
	query, args, err := sq.Insert(chatsTable).
		PlaceholderFormat(sq.Dollar).
		Columns(chatNameColumn).
		Values(name).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return 0, err
	}

	q := db.Query{
		Name:     "chat_repository.createChat",
		QueryRaw: query,
	}

	var chatID int64
	err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&chatID)
	if err != nil {
		return 0, err
	}

	return chatID, nil
}

// divideUsers разделяет юзеров на существующих и несуществующих
func (r *repo) divideUsers(ctx context.Context, names []string) ([]int64, []string, error) {
	userIDs := []int64{}   // существующие пользователи
	newUsers := []string{} // новые пользователи (у них пока что нет id)

	for _, user := range names {
		query, args, err := sq.Select(idColumn).
			From(usersTable).
			PlaceholderFormat(sq.Dollar).
			Where(sq.Eq{usernameColumn: user}).ToSql()

		if err != nil {
			return nil, nil, err
		}

		q := db.Query{
			Name:     "chat_repository.divideUsers",
			QueryRaw: query,
		}

		var userID int64
		err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				newUsers = append(newUsers, user) // сохраняем имя нового пользователя
				continue
			}
			return nil, nil, err
		}

		userIDs = append(userIDs, userID) // сохраняем id уже существующего пользователя
	}

	return userIDs, newUsers, nil
}

// insertUsers сохраняет юзеров в БД
func (r *repo) insertUsers(ctx context.Context, names []string) ([]int64, error) {
	userIDs := []int64{}

	for _, name := range names {
		query, args, err := sq.Insert(usersTable).
			PlaceholderFormat(sq.Dollar).
			Columns(usernameColumn).
			Values(name).
			Suffix("RETURNING id").
			ToSql()

		if err != nil {
			return nil, err
		}

		q := db.Query{
			Name:     "chat_repository.insertUsers",
			QueryRaw: query,
		}

		var newUserID int64
		err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&newUserID)
		if err != nil {
			return nil, err
		}

		userIDs = append(userIDs, newUserID) // сохраняем id
	}

	return userIDs, nil
}

// insertUsersInChats сохраняет id чата и его юзеров
func (r *repo) insertUsersInChats(ctx context.Context, chatID int64, userIDs []int64) error {
	builder := sq.Insert(usersInChatsTable).
		PlaceholderFormat(sq.Dollar).
		Columns(chatIDColumn, userIDColumn)

	for _, id := range userIDs {
		builder = builder.Values(chatID, id)
	}

	query, args, err := builder.ToSql()

	if err != nil {
		return err
	}

	q := db.Query{
		Name:     "chat_repository.insertUsersInChats",
		QueryRaw: query,
	}

	_, err = r.db.DB().ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}

// isChatExist проверяет, существует ли в БД чат с указанным id
func (r *repo) isChatExist(ctx context.Context, chatID int64) (bool, error) {
	// выбираем чат с указанным id
	query, args, err := sq.Select(isDeletedColumn).
		From(chatsTable).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{idColumn: chatID}).
		Limit(1).
		ToSql()

	if err != nil {
		return false, err
	}

	q := db.Query{
		Name:     "chat_repository.isChatExist",
		QueryRaw: query,
	}

	var isDeleted int
	err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&isDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	if isDeleted == 1 {
		return false, nil
	}

	return true, nil
}

// isUserExistByName проверяет, существует ли в БД пользователь с указанными именем
func (r *repo) isUserExistByName(ctx context.Context, name string) (bool, error) {
	query, args, err := sq.Select("1").
		From(usersTable).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{usernameColumn: name}).
		ToSql()

	if err != nil {
		return false, err
	}

	q := db.Query{
		Name:     "chat_repository.isUserExistByName",
		QueryRaw: query,
	}

	var one int
	err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// getUserByName получает из БД id юзера с указанными именем
func (r *repo) getUserIDByName(ctx context.Context, name string) (int64, error) {
	query, args, err := sq.Select(idColumn).
		From(usersTable).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{usernameColumn: name}).
		ToSql()

	if err != nil {
		return 0, err
	}

	q := db.Query{
		Name:     "chat_repository.getUserIDByName",
		QueryRaw: query,
	}

	var userID int64
	err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// getUserNameByID получает из БД имя юзера по указанному id
func (r *repo) getUserNameByID(ctx context.Context, userID int64) (string, error) {
	query, args, err := sq.Select(usernameColumn).
		From(usersTable).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{idColumn: userID}).
		ToSql()

	if err != nil {
		return "", err
	}

	q := db.Query{
		Name:     "chat_repository.getUserNameByID",
		QueryRaw: query,
	}

	var username string
	err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

// isUserInChat проверяет, находится ли юзер в указанном чате
func (r *repo) isUserInChat(ctx context.Context, chatID int64, userID int64) (bool, error) {
	query, args, err := sq.Select("1").
		From(usersInChatsTable).
		PlaceholderFormat(sq.Dollar).
		Where(sq.And{sq.Eq{chatIDColumn: chatID}, sq.Eq{userIDColumn: userID}}).
		ToSql()

	if err != nil {
		return false, err
	}

	q := db.Query{
		Name:     "chat_repository.isUserInChat",
		QueryRaw: query,
	}

	var one int
	err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// getUserChatsIDs выбирает id тех чатов, в которых состоит юзер
func (r *repo) getUserChatsIDs(ctx context.Context, userID int64) ([]int64, error) {
	query, args, err := sq.Select(chatIDColumn).
		From(usersInChatsTable).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{userIDColumn: userID}).
		ToSql()

	if err != nil {
		return nil, err
	}

	q := db.Query{
		Name:     "chat_repository.getUserChatsIDs",
		QueryRaw: query,
	}

	rows, err := r.db.DB().QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chatIDs := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		chatIDs = append(chatIDs, id)
	}

	return chatIDs, nil
}

func (r *repo) getChatUsersIDs(ctx context.Context, chatID int64) ([]int64, error) {
	query, args, err := sq.Select(userIDColumn).
		From(usersInChatsTable).
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{chatIDColumn: chatID}).
		ToSql()

	if err != nil {
		return nil, err
	}

	q := db.Query{
		Name:     "chat_repository.getChatUsersIDs",
		QueryRaw: query,
	}

	rows, err := r.db.DB().QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}

	return userIDs, nil
}

func (r *repo) getChatsInfo(ctx context.Context, chatIDs []int64) ([]*model.Chat, error) {
	chatsInfo := []*model.Chat{}

	for _, id := range chatIDs {
		exist, err := r.isChatExist(ctx, id)
		if err != nil {
			return nil, err
		}

		if !exist {
			continue
		}

		query, args, err := sq.Select(idColumn, chatNameColumn).
			From(chatsTable).
			PlaceholderFormat(sq.Dollar).
			Where(sq.Eq{idColumn: id}).ToSql()

		if err != nil {
			return nil, err
		}

		q := db.Query{
			Name:     "chat_repository.getChatsInfo",
			QueryRaw: query,
		}

		chatInfo := &model.Chat{}
		err = r.db.DB().QueryRowContext(ctx, q, args...).Scan(&chatInfo.ID, &chatInfo.Name)
		if err != nil {
			return nil, err
		}

		userIDs, err := r.getChatUsersIDs(ctx, id)
		if err != nil {
			return nil, err
		}

		usernames := []string{}
		for _, id := range userIDs {
			username, err := r.getUserNameByID(ctx, id)
			if err != nil {
				return nil, err
			}

			usernames = append(usernames, username)
		}

		chatInfo.Usernames = usernames

		chatsInfo = append(chatsInfo, chatInfo)
	}

	return chatsInfo, nil
}
