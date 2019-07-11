package utils

import (
	"bytes"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

//建群发收参数结构体
type CreateGroupRequest struct{
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	OpenIds         []string           `json:"open_ids"`
	EmployeeIds     []string           `json:"employee_ids"`
}

type CreateGroupResponse struct {
	Code                int         `json:"code"`
	Msg                 string      `json:"msg"`
	OpenChatId          string      `json:"open_chat_id"`
	InvalidOpenIds      []string    `json:"invalid_open_ids"`
	InvalidEmployeeIds  []string    `json:"invalid_employee_ids"`
}

//获取用户ID发收参数结构体
type GetUserIdsRequest struct {
	Email       string      `json:"email"`
}

type GetUserIdsResponse struct {
	Code            int     `json:"code"`
	OpenId          string  `json:"open_id"`
	EmployeeId      string  `json:"employee_id"`
}

//拉用户进群发收参数结构体
type PullUserParams struct {
	OpenChatId      string      `json:"open_chat_id"`
	OpenIds         []string    `json:"open_ids"`
	EmployeeIds     []string    `json:"employee_ids"`
}

type PullUserRet struct {
	Code                int         `json:"code"`
	Msg                 string      `json:"msg"`
	InvalidOpenIds      string      `json:"invalid_open_ids"`
	InvalidEmployeeIds  string      `json:"invalid_employee_ids"`
}

//发消息结构体
type SendMsgParams struct {
	OpenChatId      string      `json:"open_chat_id"`
	MsgType         string      `json:"msg_type"`
	Content         Content     `json:"content"`
}

type Content struct {
	Text    string      `json:"text"`
}

type SendMsgRet struct {
	Code                int         `json:"code"`
	Msg                 string      `json:"msg"`
	OpenMessageId       string      `json:"open_message_id"`
}

type GetTokenParams struct {
	AppId                string      `json:"app_id"`
	AppSecret            string      `json:"app_secret"`
}

type GetTokenRet struct {
	Code                int         `json:"code"`
	Msg                 string      `json:"msg"`
	TenantAccessToken   string      `json:"tenant_access_token"`
	Expire              int         `json:"expire"`
}

const (
	CREATE_GROUP_URL="https://open.feishu.cn/open-apis/chat/v3/create/"
	GET_USER_IDS_URL="https://open.feishu.cn/open-apis/user/v3/email2id"
	PULL_USER_TO_GROUP_URL="https://open.feishu.cn/open-apis/chat/v3/chatter/add/"
	SEND_MESSAGE_URL="https://open.feishu.cn/open-apis/message/v3/send/"
	GET_Tenant_Access_Token_URL="https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/"
	APP_ID="cli_9a2d72678bb8d102"
	APP_SECRET="7aprfnGu8mU3KOTqV4RiSjhIde2gsvAM"
)

func CallLarkAPI(url string,token string,paramsIn interface{} ,paramsOut interface{} ) {
	bodyByte, _ := json.Marshal(paramsIn)
	rbodyByte := bytes.NewReader(bodyByte)
	client := &http.Client{}
	request, err := http.NewRequest("POST", url,rbodyByte)
	if err != nil {
		logs.Info("新建request对象失败")
	}
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送post请求失败")
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Info("读取respose的body内容失败")
	}
	json.Unmarshal(body, paramsOut)
}


