package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)

func (abot *BotService) SendTextMessageByOpenId(text *string, openId *string, rootId *string) (map[string]interface{}, error) {
	if text == nil {
		return nil, fmt.Errorf("SendTextMessageByOpenId failed because of invalid text")
	}
	if openId == nil {
		return nil, fmt.Errorf("SendTextMessageByOpenId failed because of invalid openId")
	}
	response, err := abot.SendTextMessage(text, openId, nil, nil, nil, rootId)
	return response, err
}

func (abot *BotService) SendTextMessageByEmployeeId(text *string, employeeId *string, rootId *string) (map[string]interface{}, error)  {
	if text == nil {
		return nil, fmt.Errorf("SendTextMessageByOpenId failed because of invalid text")
	}
	if employeeId == nil {
		return nil, fmt.Errorf("SendTextMessageByOpenId failed because of invalid employeeId")
	}
	response, err := abot.SendTextMessage(text, nil, employeeId, nil, nil, rootId)
	return response, err
}

func (abot *BotService) SendTextMessageByEmail(text *string, email *string, rootId *string) (map[string]interface{}, error)  {
	if text == nil {
		return nil, fmt.Errorf("SendTextMessageByOpenId failed because of invalid text")
	}
	if email == nil {
		return nil, fmt.Errorf("SendTextMessageByOpenId failed because of invalid email")
	}
	response, err := abot.SendTextMessage(text, nil, nil, email, nil, rootId)
	return response, err
}

func (abot *BotService) SendTextMessageByOpenChatId(text *string, openChatId *string, rootId *string) (map[string]interface{}, error)  {
	if text == nil {
		return nil, fmt.Errorf("SendTextMessageByOpenId failed because of invalid text")
	}
	if openChatId == nil {
		return nil, fmt.Errorf("SendTextMessageByOpenId failed because of invalid openChatId")
	}
	response, err := abot.SendTextMessage(text, nil, nil, nil, openChatId, rootId)
	return response, err
}


func (abot *BotService) SendTextMessage(text, openId, employeeId, email, openChatId, rootId *string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("sendTextMessage failed because of can not get validTenantAccessToken, error= %v", err)
	}
	textMessageContent := form.GenerateTextMessageContent(*text)
	textMessage, err := form.GenerateMessage("text", textMessageContent)
	if err != nil {
		return nil, fmt.Errorf("sendTextMessage failed because of GenerateMessage failed, error= %s", err)
	}
	textMessage.OpenID = openId
	textMessage.EmployeeID = employeeId
	textMessage.Email = email
	textMessage.OpenChatID = openChatId
	textMessage.RootID = rootId
	response, err := abot.SendMessage(*textMessage)
	if err != nil {
		return nil, fmt.Errorf("sendTextMessage failed because of sendMessage failed, error= %s", err)
	}
	return response, nil
}