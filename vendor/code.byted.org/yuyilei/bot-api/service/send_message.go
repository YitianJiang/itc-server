package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)

func (abot *BotService) SendMessage(messageForm form.SendMessageForm) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("sendTextMessage failed because of can not get validTenantAccessToken, error= %v", err)
	}
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(SendMessageUrl, header, messageForm)
	return response, err
}
