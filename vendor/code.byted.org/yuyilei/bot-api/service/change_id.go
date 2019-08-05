package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)


func (abot *BotService) OpenId2UserId(OpenId string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("OpenId2UserId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.OpenId2UserIdForm
	requestBody.OpenId = &OpenId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(OpenId2UserIdUrl, header, requestBody)
	return response, err
}

func (abot *BotService) UserId2OpenId(UserId string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("UserId2OpenId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.UserId2OpenIdForm
	requestBody.UserId = &UserId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(UserId2OpenIdUrl, header, requestBody)
	return response, err
}

func (abot *BotService) MessageId2OpenMessageId(MessageId string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("MessageId2OpenMessageId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.MessageId2OpenMessageIdForm
	requestBody.MessageId = &MessageId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(MessageId2OpenMessageIdUrl, header, requestBody)
	return response, err
}

func (abot *BotService) OpenMessageId2MessageId(OpenMessageId string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("OpenMessageId2MessageId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.OpenMessageId2MessageIdForm
	requestBody.OpenMessageId = &OpenMessageId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(OpenMessageId2MessageIdUrl, header, requestBody)
	return response, err
}

func (abot *BotService) ChatId2OpenChatId(ChatId string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("ChatId2OpenChatId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.ChatId2OpenChatIdForm
	requestBody.ChatId = &ChatId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(ChatId2OpenChatIdUrl, header, requestBody)
	return response, err
}

func (abot *BotService) OpenChatId2ChatId(OpenChatId string) (map[string]interface{}, error)  {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("OpenChatId2ChatId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.OpenChatId2ChatIdForm
	requestBody.OpenChatId = &OpenChatId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(OpenChatId2ChatIdUrl, header, requestBody)
	return response, err
}

func (abot *BotService) UserId2EmployeeId(UserId string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("UserId2EmployeeId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.UserId2EmployeeIdForm
	requestBody.UserId = &UserId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(UserId2EmployeeIdUrl, header, requestBody)
	return response, err
}

func (abot *BotService) EmployeeId2UserId(EmployeeId string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("EmployeeId2UserId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.EmployeeId2UserIdForm
	requestBody.EmployeeId = &EmployeeId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(EmployeeId2UserIdUrl, header, requestBody)
	return response, err
}

func (abot *BotService) DepartmentId2OpenDepartmentId(DepartmentId string) (map[string]interface{}, error){
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("DepartmentId2OpenDepartmentId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.DepartmentId2OpenDepartmentIdForm
	requestBody.DepartmentId = &DepartmentId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(DepartmentId2OpenDepartmentIdUrl, header, requestBody)
	return response, err
}

func (abot *BotService) OpenDepartmentId2DepartmentId(OpenDepartmentId string) (map[string]interface{}, error){
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("OpenDepartmentId2DepartmentId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.OpenDepartmentId2DepartmentIdForm
	requestBody.OpenDepartmentId = &OpenDepartmentId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(OpenDepartmentId2DepartmentIdUrl, header, requestBody)
	return response, err
}


func (abot *BotService) Email2OpenIdEmployeeId(email string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("Email2OpenIdEmployeeId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var requestBody form.Email2OpenIdEmployeeIdUrlForm
	requestBody.Email = email
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(Email2OpenIdEmployeeIdUrl, header, requestBody)
	return response, err
}