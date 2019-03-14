package lark

import (
	"fmt"
	"log"
	"strings"
)

// MsgBuffer stores all the messages attached
// You can call every function, but some of which is only available for specific condition
type MsgBuffer struct {
	message OutcomingMessage
	buf     []msgElem
	msgType MessageType
}

// MsgType of a msg buf
type MsgType int

type msgElem struct {
	elemType MsgType
	content  string
}

const (
	// MsgText text only message
	msgText MsgType = iota
	// MsgAt @somebody
	msgAt
	// MsgAtAll @all
	msgAtAll
	// msgSpace space
	msgSpace
)

// NewMsgBuffer create a message buffer
func NewMsgBuffer(newMsgType MessageType) *MsgBuffer {
	return &MsgBuffer{
		message: OutcomingMessage{
			MsgType: newMsgType,
		},
		buf:     make([]msgElem, 0),
		msgType: newMsgType,
	}
}

// Title set buffer's title
// for MsgPost only
func (m *MsgBuffer) Title(title string) *MsgBuffer {
	if m.msgType != MsgPost {
		log.Println("`Title` is only available to MsgPost")
	}
	m.message.Content.Title = title
	return m
}

// Title set buffer's title
// for MsgRichText only
func (m *MsgBuffer) TitleWithImg(title string) *MsgBuffer {
	if m.msgType != MsgRichText {
		log.Println("`Title` is only available to MsgPost")
	}
	m.message.Content.Title = title
	return m
}

// Image attaches image key
// for MsgImage only
func (m *MsgBuffer) Image(imageKey string) *MsgBuffer {
	if m.msgType != MsgImage {
		log.Println("`Image` is only available to MsgImage")
	}
	m.message.Content.ImageKey = imageKey
	return m
}

// BindChannel binds channel
func (m *MsgBuffer) BindChannel(Channel string) *MsgBuffer {
	m.message.ChatID = Channel
	return m
}

// ShareChat attaches chat id
// for MsgShareChat only
func (m *MsgBuffer) ShareChat(ChatID string) *MsgBuffer {
	if m.msgType != MsgShareCard {
		log.Println("`ShareChat` is only available to MsgShareChat")
	}
	m.message.Content.ChatID = ChatID
	return m
}

// BindMention binds metion user id
// This is for official style @at
// `Mention` is based on text message, which is hackable in earlier versions
func (m *MsgBuffer) BindMention(metionUserIDs []string) *MsgBuffer {
	m.message.MetionUserIDs = metionUserIDs
	return m
}

// BindReply binds root id for reply
// rootID is OpenMessageID of the message you reply
func (m *MsgBuffer) BindReply(rootID string) *MsgBuffer {
	m.message.RootID = rootID
	return m
}

// BindEmail binds email
// Used for private chat message
func (m *MsgBuffer) BindEmail(email string) *MsgBuffer {
	m.message.Email = email
	return m
}

// Text add simple texts
func (m *MsgBuffer) Text(text string) *MsgBuffer {
	msg := msgElem{
		elemType: msgText,
		content:  text,
	}
	m.buf = append(m.buf, msg)
	return m
}

// Textln add simple texts with a newline
func (m *MsgBuffer) Textln(text string) *MsgBuffer {
	msg := msgElem{
		elemType: msgText,
		content:  text + "\n",
	}
	m.buf = append(m.buf, msg)
	return m
}

// Space attaches spaces
func (m *MsgBuffer) Space(n int) *MsgBuffer {
	msg := msgElem{
		elemType: msgSpace,
		content:  strings.Repeat(" ", n),
	}
	m.buf = append(m.buf, msg)
	return m
}

// Newline attaches \n
func (m *MsgBuffer) Newline() *MsgBuffer {
	msg := msgElem{
		elemType: msgSpace,
		content:  "\n",
	}
	m.buf = append(m.buf, msg)
	return m
}

// Mention @somebody
func (m *MsgBuffer) Mention(userID, userName string) *MsgBuffer {
	msg := msgElem{
		elemType: msgAt,
		content:  fmt.Sprintf("<at user_id=\"%s\">@%s</at>", userID, userName),
	}
	m.buf = append(m.buf, msg)
	return m
}

// MentionAll @all
func (m *MsgBuffer) MentionAll(allName string) *MsgBuffer {
	msg := msgElem{
		elemType: msgAtAll,
		content:  fmt.Sprintf("<at user_id=\"all\">@%s</at>", allName),
	}
	m.buf = append(m.buf, msg)
	return m
}

// Clear all message
func (m *MsgBuffer) Clear() *MsgBuffer {
	m.message = OutcomingMessage{
		MsgType: m.msgType,
	}
	m.buf = make([]msgElem, 0)
	return m
}

// Build final message
func (m *MsgBuffer) Build() OutcomingMessage {
	for _, msg := range m.buf {
		m.message.Content.Text += msg.content
	}
	return m.message
}

// Len returns buf len
func (m *MsgBuffer) Len() int {
	return len(m.buf)
}
