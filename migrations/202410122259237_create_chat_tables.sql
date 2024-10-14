-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    role INT DEFAULT 0      /* 0 - UNKNOWN, 1 - USER, 2 - ADMIN */
);

CREATE TABLE chats (
    id SERIAL PRIMARY KEY,
    chat_name VARCHAR(255) NOT NULL,
    users VARCHAR[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    chat_id INT REFERENCES chats(id),
    from_id INT REFERENCES users (id),
    message_text TEXT NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE users;
DROP TABLE chats;
DROP TABLE messages;