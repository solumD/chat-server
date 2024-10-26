package model

// Chat модель чата в репо слое
type Chat struct {
	Name      string
	Usernames []string
}

// Message модель сообщения в репо слое
type Message struct {
	ChatID int64
	From   string
	Text   string
}
