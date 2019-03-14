package lark

// OutcomingMessage struct of an outcoming message
type OutcomingMessage struct {
	MsgType       MessageType
	ChatID        string
	RootID        string
	MetionUserIDs []string
	Content       MessageContent
	Email         string
}

// MessageContent struct of message content
type MessageContent struct {
	Title    string
	Text     string
	ImageKey string
	ChatID   string
}

// MessageType message type
type MessageType string

const (
	// MsgText simple text message
	MsgText MessageType = "text"
	// MsgPost rich text message
	MsgPost MessageType = "post"
	// MsgImage simple image message
	MsgImage MessageType = "image"
	// MsgShareCard share chat group card
	MsgShareCard MessageType = "share_chat"
	// MsgInteractive interactive widget
	MsgInteractive MessageType = "interactive"
	//MsgRichText rich text with image message
	MsgRichText MessageType = "rich_text"
)
