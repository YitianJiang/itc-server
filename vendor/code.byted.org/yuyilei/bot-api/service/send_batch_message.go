package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)

func (abot *BotService) SendBatchMessage(text *string, departmentIds []string, openIds []string) (map[string]interface{}, error){
	if text == nil {
		return nil, fmt.Errorf("SendBatchMessage failed because of invalid text")
	}
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("sendBatchMessage failed because of can not get validTenantAccessToken, error= %v", err)
	}
	batchMessageContent := form.GenerateTextMessageContent(*text)
	batchMessage, err := form.GenerateMessage("text", batchMessageContent)
	if err != nil {
		return nil, fmt.Errorf("sendBatchMessage failed because of GenerateMessage failed, error= %s", err)
	}
	batchMessage.DepartmentIds = departmentIds
	batchMessage.OpenIds = openIds
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(SendBatchMessageUrl, header, batchMessage)
	return response, err
}
