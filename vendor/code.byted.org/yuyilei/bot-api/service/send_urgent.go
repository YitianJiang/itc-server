package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)

func (abot *BotService) UrgentMessage(openMessageId string, urgentType string, openIds []string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("UrgentMessage failed because of can not get validTenantAccessToken, error= %v", err)
	}
	urgentMessage, err := form.GenerateUrgentMessageByForm(openMessageId, urgentType, openIds)
	if err != nil {
		return nil,  fmt.Errorf("UrgentMessage failed because of GenerateUrgentMessage failed, error= %v", err)
	}
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(UrgentMessageUrl, header, urgentMessage)
	return response, err
}

func (abot *BotService) UrgentMessageByForm(urgentMessageForm *form.UrgentMessageForm) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("UrgentMessageByForm failed because of can not get validTenantAccessToken, error= %v", err)
	}
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(UrgentMessageUrl, header, urgentMessageForm)
	return response, err
}