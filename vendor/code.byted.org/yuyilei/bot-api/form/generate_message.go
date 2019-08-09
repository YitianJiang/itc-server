package form

import (
	"fmt"
)


func GenerateHeader(AccessToken string) map[string]string {
	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", AccessToken)
	header["Content-Type"] = "application/json"
	return header
}

func GenerateTextMessageContent(text string) *MessageContent {
	return &MessageContent{Text: &text}
}

func GenerateImageMessageContent(imageKey string) *MessageContent {
	return &MessageContent{ImageKey: &imageKey}
}

func GenerateShareChatMessageContent(shareChatId string) *MessageContent {
	return &MessageContent{ShareOpenChatID: &shareChatId}
}
func isValidMsgType(msgType string) bool {
	switch msgType {
	case  "text",
		"image",
		"post",
		"share_chat",
		"interactive": return true
	}
	return false
}

func GenerateMessage(msgType string, content *MessageContent)  (*SendMessageForm, error){
	if !isValidMsgType(msgType) {
		return nil, fmt.Errorf("invalid msgType")
	}
	var textMessageForm SendMessageForm
	textMessageForm.MsgType = &msgType
	textMessageForm.Content = *content
	return &textMessageForm, nil
}
