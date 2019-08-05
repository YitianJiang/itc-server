package form

import (
	"fmt"
)

func isValidUrgentType(urgentType string) bool {
	switch urgentType {
	case "app",
		"sms",
		"phone": return true
	}
	return false
}

func GenerateUrgentMessageByForm(openMessageId string, urgentType string, openIds []string) (*UrgentMessageForm, error) {
	if !isValidUrgentType(urgentType) {
		return nil, fmt.Errorf("urgentType should in [app, sms, phone]")
	}
	return &UrgentMessageForm{OpenMessageID: openMessageId, UrgentType:urgentType, OpenIDs:openIds}, nil
}
