package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)

func (abot *BotService) SendImageMessageByOpenId(image *string, openId *string, rootId *string) (map[string]interface{}, error) {
	if image == nil {
		return nil, fmt.Errorf("SendImageMessageByOpenId failed because of invalid text")
	}
	if openId == nil {
		return nil, fmt.Errorf("SendImageMessageByOpenId failed because of invalid openId")
	}
	response, err := abot.SendImageMessage(image, openId, nil, nil, nil, rootId)
	return response, err
}

func (abot *BotService) SendImageMessageByEmployeeId(image *string, employeeId *string, rootId *string) (map[string]interface{}, error) {
	if image == nil {
		return nil, fmt.Errorf("SendImageMessageByEmployeeId failed because of invalid text")
	}
	if employeeId == nil {
		return nil, fmt.Errorf("SendImageMessageByEmpoyeeId failed because of invalid employeeId")
	}
	response, err := abot.SendImageMessage(image, nil, employeeId, nil, nil, rootId)
	return response, err
}

func (abot *BotService) SendImageMessageByEmail(image *string, email *string, rootId *string) (map[string]interface{}, error) {
	if image == nil {
		return nil, fmt.Errorf("SendImageMessageByEmail failed because of invalid text")
	}
	if email == nil {
		return nil, fmt.Errorf("SendImageMessageByEmail failed because of invalid email")
	}
	response, err := abot.SendImageMessage(image, nil, nil, email, nil, rootId)
	return response, err
}

func (abot *BotService) SendImageMessageByOpenChatId(image *string, openChatId *string, rootId *string) (map[string]interface{}, error) {
	if image == nil {
		return nil, fmt.Errorf("SendImageMessageByOpenChatId failed because of invalid text")
	}
	if openChatId == nil {
		return nil, fmt.Errorf("SendImageMessageByOpenChatId failed because of invalid openChatId")
	}
	response, err := abot.SendImageMessage(image, nil, nil, nil, openChatId, rootId)
	return response, err
}

func (abot *BotService) SendImageMessage(image, openId, employeeId, email, openChatId, rootId *string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("sendImageMessage failed because of can not get validTenantAccessToken, error= %v", err)
	}
	imageMessageContent := form.GenerateImageMessageContent(*image)
	imageMessage, err := form.GenerateMessage("image", imageMessageContent)
	if err != nil {
		return nil, fmt.Errorf("sendImageMessage failed because of GenerateMessage failed, error= %s", err)
	}
	imageMessage.OpenID = openId
	imageMessage.EmployeeID = employeeId
	imageMessage.Email = email
	imageMessage.OpenChatID = openChatId
	imageMessage.RootID = rootId
	response, err := abot.SendMessage(*imageMessage)
	if err != nil {
		return nil, fmt.Errorf("sendImageMessage failed because of sendMessage failed, error= %s", err)
	}
	return response, nil
}
