package model

// Chat модель чата в сервисном слое
type Chat struct {
	Name      string
	Usernames []string
}

// Message модель сообщения в сервисном слое
type Message struct {
	ChatID int64
	From   string
	Text   string
}
