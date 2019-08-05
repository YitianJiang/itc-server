package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
)

// 机器人邀请人进群
func (abot *BotService) InviteUser2Chat(openChatId string, openIds []string) (map[string]interface{}, error) {
	err := abot.getValidTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("InviteUser2Chat failed because of can not get validTenantAccessToken, error= %v", err)
	}
	var inviteUser2ChatForm form.User2ChatForm
	inviteUser2ChatForm.OpenChatId = &openChatId
	for _, openId := range openIds {
		temp := openId
		inviteUser2ChatForm.OpenIds = append(inviteUser2ChatForm.OpenIds, &temp)
	}
	header := form.GenerateHeader(abot.TenantAccessToken.Token)
	response, err := PostRequest(InviteUser2ChatUrl, header, inviteUser2ChatForm)
	return response, err
}
