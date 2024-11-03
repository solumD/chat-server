package errors

import "fmt"

var (
	ErrDescChatIsNil    = fmt.Errorf("desc chat is nil")    // ErrDescChatIsNil grpc запрос с чатом nil
	ErrDescMessageIsNil = fmt.Errorf("desc message is nil") // ErrDescMessageIsNil grpc запрос с сообщением nil
)
