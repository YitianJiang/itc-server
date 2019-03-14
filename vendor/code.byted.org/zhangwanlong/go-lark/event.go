package lark

// See https://docs.bytedance.net/doc/IB3XFs06rkTrQyyATX0TKa
const (
	// EventGetMessage ...
	EventGetMessage = 1
)

// EventMessage ...
type EventMessage struct {
	Timestamp string `json:"ts"`
	// Token is shown by Lark to indicate it is not a fake message, check at your own need
	Token     string    `json:"token"`
	EventType string    `json:"type"`
	Event     EventText `json:"event"`
}

// EventText ...
type EventText struct {
	Type          string `json:"type"`
	ChatType      string `json:"chat_type"`
	MsgType       string `json:"msg_type"`
	UserID        string `json:"user"`
	ChatID        string `json:"chat_id"`
	Title         string `json:"title,omitempty"`
	Text          string `json:"text,omitempty"`
	ImageKey      string `json:"image_key,omitempty"`
	ImageURL      string `json:"image_url,omitempty"`
	OpenMessageID string `json:"open_message_id,omitempty"`
}

// EventChallengeReq request of add event hook
type EventChallengeReq struct {
	Token     string `json:"token,omitempty"`
	Challenge string `json:"challenge,omitempty"`
	Type      string `json:"type,omitempty"`
}

// EncryptedReq request of encrypted challagen
type EncryptedReq struct {
	Encrypt string `json:"encrypt,omitempty"`
}

// MsgCallbackFunc for event handling
type MsgCallbackFunc func(EventMessage) error
