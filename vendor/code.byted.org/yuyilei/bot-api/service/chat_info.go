package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)

// 获得机器人所在的群信息
func (abot *BotService) GetChatInfo(openChatId string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("GetChatInfo failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var getChatInfoFom form.GetChatInfoForm
	getChatInfoFom.OpenChatId = &openChatId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(GetChatInfoUrl, header, getChatInfoFom)
	return response, err
}



// 获得机器人所在的群列表
// page必填，pageSize 不必填 默认为 30，范围为 [1,100]
func (abot *BotService) GetChatList(page int16, pageSize *int8) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("GetChatList failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var getChatListForm form.GetChatListForm
	getChatListForm.Page = &page
	if pageSize == nil {
		var temp int8 = 30
		pageSize = &temp
	}
	getChatListForm.PageSize = pageSize
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(GetChatListUrl, header, getChatListForm)
	return response, err
}


// 机器人创建群
// name 不必填， description 不必填 openIds 必填
func (abot *BotService) CreateChat(name *string, description *string, openIds []string) (map[string]interface{}, error){
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("CreateChat failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var createChatForm form.CreateChatForm
	createChatForm.Name = name
	createChatForm.Description = description
	for _, openId := range openIds {
		temp := openId
		createChatForm.OpenIds = append(createChatForm.OpenIds, &temp)
	}
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(CreateChatUrl, header, createChatForm)
	return response, err
}

// 更新群名称、转让群主
func (abot *BotService) UpdateChatInfo(openChatId string, ownerId *string, name *string, i18nNames map[string]string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("UpdateChatInfo failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var updateChatInfoForm form.UpdateChatInfoForm
	updateChatInfoForm.OpenChatId = &openChatId
	updateChatInfoForm.OwnerId = ownerId
	updateChatInfoForm.Name = name
	updateChatInfoForm.I18nNames = i18nNames
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(UpdateChatInfoUrl, header, updateChatInfoForm)
	return response, err
}


// 获取用户和机器人的ChatID/OpenChatID、获取用户和用户之间的ChatID/OpenChatID 
func (abot *BotService) GetP2pChatId(userId *string, openId *string, chatter *string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("GetP2pChatId failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var getP2pChatIdForm form.GetP2pChatIdForm
	getP2pChatIdForm.UserId = userId
	getP2pChatIdForm.OpenId = openId
	getP2pChatIdForm.Chatter = chatter
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(GetP2pChatIdUrl, header, getP2pChatIdForm)
	return response, err
}

// 机器人解散群(机器人为群主的时候)
func (abot *BotService) DisbandChat(chatId string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("DisbandChat failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var disbandChatForm form.DisbandChatForm
	disbandChatForm.ChatId = &chatId
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(DisbandChatUrl, header, disbandChatForm)
	return response, err
}