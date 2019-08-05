package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)

func (abot *BotService) SendShareChatMessageByOpenId(shareOpenChatId *string, openId *string, rootId *string) (map[string]interface{}, error) {
	if shareOpenChatId == nil {
		return nil, fmt.Errorf("SendShareChatMessageByOpenId failed because of invalid shareOpenChatId")
	}
	if openId == nil {
		return nil, fmt.Errorf("SendShareChatMessageByOpenId failed because of invalid openId")
	}
	response, err := abot.SendShareChatMessage(shareOpenChatId, openId, nil, nil, nil, rootId)
	return response, err
}

func (abot *BotService) SendShareChatMessageByEmployeeId(shareOpenChatId *string, employeeId *string, rootId *string) (map[string]interface{}, error) {
	if shareOpenChatId == nil {
		return nil, fmt.Errorf("SendShareChatMessageByEmployeeId failed because of invalid shareOpenChatId")
	}
	if employeeId == nil {
		return nil, fmt.Errorf("SendShareChatMessageByEmployeeId failed because of invalid employeeId")
	}
	response, err := abot.SendShareChatMessage(shareOpenChatId, nil, employeeId, nil, nil, rootId)
	return response, err
}

func (abot *BotService) SendShareChatMessageByEmail(shareOpenChatId *string, email *string, rootId *string) (map[string]interface{}, error) {
	if shareOpenChatId == nil {
		return nil, fmt.Errorf("SendShareChatMessageByEmail failed because of invalid shareOpenChatId")
	}
	if email == nil {
		return nil, fmt.Errorf("SendShareChatMessageByEmail failed because of invalid email")
	}
	response, err := abot.SendShareChatMessage(shareOpenChatId, nil, nil, email, nil, rootId)
	return response, err
}

func (abot *BotService) SendShareChatMessageByOpenChatId(shareOpenChatId *string, openChatId *string, rootId *string) (map[string]interface{}, error) {
	if shareOpenChatId == nil {
		return nil, fmt.Errorf("SendShareChatMessageByOpenChatId failed because of invalid shareOpenChatId")
	}
	if openChatId == nil {
		return nil, fmt.Errorf("SendShareChatMessageByOpenChatId failed because of invalid openChatId")
	}
	response, err := abot.SendShareChatMessage(shareOpenChatId, nil, nil, nil, openChatId, rootId)
	return response, err
}

func (abot *BotService) SendShareChatMessage(shareOpenChatId, openId, employeeId, email, openChatId, rootId *string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("sendShareChatMessage failed because of can not get validTenantAccessToken, error= %v", err)
	}
	shareChatMessageContent := form.GenerateShareChatMessageContent(*shareOpenChatId)
	shareChatMessage, err := form.GenerateMessage("share_chat", shareChatMessageContent)
	if err != nil {
		return nil, fmt.Errorf("sendShareChatMessage failed because of GenerateMessage failed, error= %s", err)
	}
	shareChatMessage.OpenID = openId
	shareChatMessage.EmployeeID = employeeId
	shareChatMessage.Email = email
	shareChatMessage.OpenChatID = openChatId
	shareChatMessage.RootID = rootId
	response, err := abot.SendMessage(*shareChatMessage)
	if err != nil {
		return nil, fmt.Errorf("sendShareChatMessage failed because of sendMessage failed, error= %s", err)
	}
	return response, nil
}