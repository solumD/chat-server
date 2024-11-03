package errors

import "fmt"

var (
	ErrDescChatIsNil    = fmt.Errorf("desc chat is nil")
	ErrDescMessageIsNil = fmt.Errorf("desc message is nil")
)
