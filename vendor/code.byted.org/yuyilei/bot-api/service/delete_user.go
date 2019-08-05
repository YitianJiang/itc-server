package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)

// 机器人踢人出群
// 机器人必须是群主
func (abot *BotService) DeleteUserFromChat(openChatId string, openIds []string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("DeleteUserFormChat failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var deleteUserFromChatForm form.User2ChatForm
	deleteUserFromChatForm.OpenChatId = &openChatId
	for _, openId := range openIds {
		temp := openId
		deleteUserFromChatForm.OpenIds = append(deleteUserFromChatForm.OpenIds, &temp)
	}
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(DeleteUserForChatUrl, header, deleteUserFromChatForm)
	return response, err
}
