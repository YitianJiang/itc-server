package utils

import (
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	robotToken = "b-e57916f0-f644-472e-a9d4-087b39ce70fc"
	//user token
	myToken = "u-725a4ebd-483d-4175-9ca1-16920f94dc17"
	larkAPI = "https://oapi.zjurl.cn/open-apis/api/v2/message"
)
type LarkUser struct {
	Ok     bool   `json:"ok"`
	UserID string `json:"user_id"`
}
type IM struct {
	Channel IMChannel `json:"channel"`
	Ok      bool      `json:"ok"`
}
type IMChannel struct {
	Id string `json:"id"`
}
/*
 *获取用户的id，不然不知道给谁发消息，如果出了问题，肯定返回的是空串了
 */
func GetUserIDinLark(token string, emailPrefix string) string {
	//其实这里需要判断下有没有加上后缀。。。
	if len(strings.Split(emailPrefix, "@")) > 1 {
		emailPrefix = strings.Split(emailPrefix, "@")[0]
	}
	var retstr string
	body := map[string]interface{}{"token": token}
	body["email_prefix"] = emailPrefix
	//这里写死了是bytedance吧，反正也用不到别的组织
	body["organization_name"] = "bytednace"
	bodyByte, _ := json.Marshal(body)
	num, response := PostJsonHttp("https://oapi.zjurl.cn/open-apis/api/v1/user.user_id", bodyByte)
	var retBodyJson LarkUser
	json.Unmarshal(response, &retBodyJson)
	if num == 0 && retBodyJson.Ok == true {
		retstr = fmt.Sprint(retBodyJson.UserID)
	} else {
		logs.Error("获取UserID失败，如果是0，问题在于Lark平台" + fmt.Sprint(num))
	}
	return retstr
}
/*
 *lark机器人发消息给个人，内部调用
 */
func LarkDingOneInner(member string, msg string) {
	// 获取user的larkID和会话ID
	userLarkID := GetUserIDinLark(myToken, member)
	channelID := GetUserChannelID(robotToken, userLarkID)
	go DoLark(msg, larkAPI, channelID, robotToken)
}
/*
 *发送lark消息，chatId为个人id或者群组id
 */
func DoLark(msg string, api string, chatId string, token string) {

	body := map[string]interface{}{"msg_type": "text"}
	body["token"] = token
	body["chat_id"] = chatId
	body["content"] = map[string]string{"text": msg}
	bodyByte, _ := json.Marshal(body)
	PostJsonHttp(api, bodyByte)
}
/**
 * 发送lark群消息
 */
func LarkGroup(msg string, groupId string){
	go DoLark(msg, larkAPI, groupId, robotToken)
}
/**
 *发送lark富文本消息
 */
func DoRichLark(chatId string, token string, msg string, title string) {

	body := map[string]interface{}{"msg_type": "rich_text"}
	body["token"] = token
	body["chat_id"] = chatId
	body["content"] = map[string]string{"text": msg, "title": title}
	bodyByte, _ := json.Marshal(body)
	PostJsonHttp(larkAPI, bodyByte)
}
/*
 *获取和用户会话的channel id
 */
func GetUserChannelID(token string, userid string) string {
	var retstr string
	body := map[string]interface{}{"token": token}
	body["user"] = userid
	bodyByte, _ := json.Marshal(body)
	num, response := PostJsonHttp("https://oapi.zjurl.cn/open-apis/api/v1/im.open", bodyByte)
	var retBodyJson IM
	json.Unmarshal(response, &retBodyJson)
	if num == 0 && retBodyJson.Ok == true {
		retstr = retBodyJson.Channel.Id
	} else {
		fmt.Println("打开私聊会话失败，如果是0，问题在于Lark平台" + fmt.Sprint(num))
	}
	return retstr
}
