package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)

func (abot *BotService) GetUserInfoByOpenId(openId string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("GetUserInfoByOpenId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var getUserInfoForm form.GetUserInfoForm
	getUserInfoForm.OpenId = &openId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(GetUserInfoByOpenIdUrl, header, getUserInfoForm)
	return response, err
}

func (abot *BotService) GetAdminInfoList(openId *string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("GetAdminInfoList failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var getAdminInfoList form.GetAdminInfoListForm
	getAdminInfoList.OpenId = openId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(GetAdminInfoListUrl, header, getAdminInfoList)
	return response, err
}

func (abot *BotService) GetBotInfo() (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("GetBotInfo failed because of can not get validTenantAccessToken, error= %v", err)
	}
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err :=  PostRequest(GetBotInfoUrl, header, nil)
	return response, err
}

func (abot *BotService) AddBotToChat(chatId *string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("AddBotToChat failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var addBotToChat form.BotForm
	addBotToChat.ChatId = chatId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err :=  PostRequest(AddBotToChatUrl, header, addBotToChat)
	return response, err
}

func (abot *BotService) RemoveBotFromChat(chatId *string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("RemoveBotFromChat failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var removeBotFromChat form.BotForm
	removeBotFromChat.ChatId = chatId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err :=  PostRequest(RemoveBotFromChatUrl, header, removeBotFromChat)
	return response, err
}
