package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/gopkg/logs"
)

const (
	robotToken = "b-c0baf8bf-b138-4077-8ef9-800476529ca2"
	//user token
	myToken = "u-77a951b4-8a4a-4cb5-8fce-4358c14f5b55"
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
	//userLarkID := GetUserIDinLark(myToken, member)
	//channelID := GetUserChannelID(robotToken, userLarkID)
	//go DoLark(msg, larkAPI, channelID, robotToken)
	m := map[string]interface{}{
		"appID":        "cli_9d8a78c3eff61101",
		"appSecret":    "3kYDkS2M0obuzaEWrArGIc6NOJU6ZVeF",
		"title":        "预审平台消息通知",
		"information":  msg,
		"informMember": member,
		"isAt":         "1",
	}
	if post_body, err := json.Marshal(m); err != nil {
		logs.Error(err.Error())
	} else {
		PostJsonHttp2(post_body)
	}
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
func LarkGroup(msg string, groupId string) {
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
		logs.Error("打开私聊会话失败，如果是0，问题在于Lark平台" + fmt.Sprint(num))
	}
	return retstr
}

//建群发收参数结构体
type CreateGroupRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	OpenIds     []string `json:"open_ids"`
	EmployeeIds []string `json:"employee_ids"`
}

type CreateGroupResponse struct {
	Code               int      `json:"code"`
	Msg                string   `json:"msg"`
	OpenChatId         string   `json:"open_chat_id"`
	InvalidOpenIds     []string `json:"invalid_open_ids"`
	InvalidEmployeeIds []string `json:"invalid_employee_ids"`
}

//获取用户ID发收参数结构体
type GetUserIdsRequest struct {
	Email string `json:"email"`
}

type GetUserIdsResponse struct {
	Code       int    `json:"code"`
	OpenId     string `json:"open_id"`
	EmployeeId string `json:"employee_id"`
}

//发消息结构体
type SendMsgRequest struct {
	OpenChatId string  `json:"open_chat_id"`
	MsgType    string  `json:"msg_type"`
	Content    Content `json:"content"`
}

type Content struct {
	Text string `json:"text"`
}

type SendMsgResponse struct {
	Code          int    `json:"code"`
	Msg           string `json:"msg"`
	OpenMessageId string `json:"open_message_id"`
}

//获取token结构体
type GetTokenRequest struct {
	AppId     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type GetTokenResponse struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}

const (
	CREATE_GROUP_URL            = "https://open.feishu.cn/open-apis/chat/v3/create/"
	GET_USER_IDS_URL            = "https://open.feishu.cn/open-apis/user/v3/email2id"
	SEND_MESSAGE_URL            = "https://open.feishu.cn/open-apis/message/v3/send/"
	GET_Tenant_Access_Token_URL = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/"
	APP_ID                      = "cli_9a2d72678bb8d102"
	APP_SECRET                  = "7aprfnGu8mU3KOTqV4RiSjhIde2gsvAM"
	IOSCertificateBotAppId      = "cli_9dca86fa50ee5101"
	IOSCertificateBotAppSecret  = "XbENqXBQGJeIYaU3oLk3jgdJC5IiuEAW"
	//CreateCertPrincipal         = "zhangmengqi.muki@bytedance.com"
	CreateCertPrincipal       = "gongrui@bytedance.com"
	APPLE_DELETE_CERT_URL     = "https://developer.apple.com/account/resources/certificates/download/"
	DELCERT_FEEDBACK_URL_TEST = "http://10.224.13.149:6789/v1/appleCertManage/asynDeleteFeedback"
	DELCERT_FEEDBACK_URL      = "https://itc.bytedance.net/v1/appleCertManage/asynDeleteFeedback"
)

//新建证书工单卡片基本信息
var CreateCertMessage = "请根据配置信息登录Apple后台手动生成证书并上传至itc证书管理后台"
var CreateCertAccountHeader = "账号名: "
var CreateCertTypeHeader = "证书类型: "
var CsrHeader = "CSR文件: "
var CsrText = "点击链接下载"
var GrayHeaderStyle = "color: gray"

//删除证书工单卡片基本信息
var DeleteCertMessage = "请根据账号信息登陆Apple后台，删除指定证书，点击删除链接可以直接跳转；删除完成后，请点击卡片\"已删除\"按钮。"
var DeleteCertIdHeader = "证书ID："
var DeleteCertNameHeader = "证书名称："
var AppleUrlHeader = "删除链接："
var AppleUrlText = "点击跳转"
var DeleteButtonText = "已删除"

func CallLarkAPI(url string, token string, paramsIn interface{}, paramsOut interface{}) {
	bodyByte, _ := json.Marshal(paramsIn)
	rbodyByte := bytes.NewReader(bodyByte)
	client := &http.Client{}
	request, err := http.NewRequest("POST", url, rbodyByte)
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

/*
 *lark机器人发消息给个人，内部调用
 */
func LarkDingOneInnerWithUrl(member string, msg string, urlTitle string, larkUrl string) {
	titleArr := strings.Split(urlTitle, ",")
	urlArr := strings.Split(larkUrl, ",")
	m := map[string]interface{}{
		"appID":        "cli_9d8a78c3eff61101",
		"appSecret":    "3kYDkS2M0obuzaEWrArGIc6NOJU6ZVeF",
		"title":        "预审平台消息通知",
		"information":  msg,
		"informMember": member,
		"hyper":        titleArr,
		"hyperLink":    urlArr,
		"isAt":         "1",
	}
	if post_body, err := json.Marshal(m); err != nil {
		logs.Error(err.Error())
	} else {
		PostJsonHttp2(post_body)
	}
}

//yy add
type LarkCardMessage struct {
	ChatID  string      `json:"open_chat_id"`
	Email   string      `json:"email"`
	MsgType string      `json:"msg_type"`
	Content LarkContent `json:"content"`
}
type LarkContent struct {
	Card LarkCard `json:"card"`
}
type LarkCard struct {
	Link    CardLink      `json:"card_link"`
	Header  CardHeader    `json:"header"`
	Content []interface{} `json:"content"`
}
type CardLink struct {
	Herf string `json:"href"`
}
type CardHeader struct {
	Title string `json:"title"`
	Color string `json:"image_color"`
	Lines int    `json:"lines"`
}
type LarkText struct {
	Tag   string `json:"tag"`
	Text  string `json:"text"`
	Style string `json:"style"`
}
type LarkField struct {
	Tag    string  `json:"tag"`
	Fields []Field `json:"fields"`
}
type Field struct {
	Title FieldInner `json:"title"`
	Value FieldInner `json:"value"`
	Short bool       `json:"short"`
}
type FieldInner struct {
	Tag   string `json:"tag"`
	Text  string `json:"text"`
	Lines int    `json:"lines"`
}

//初始化消息中的不变量
func initLarkStruct(lark_people, rd_bm, qa_bm, lark_message, detect_num, self_item_num, url string, groupFlag bool) LarkCardMessage {
	//分割线
	var divid LarkText
	divid.Tag = "text"
	divid.Text = "------------------------------------------\n"
	divid.Style = "width: 100%; fontSize: 13;color: #C2C2C2"
	//内容
	//content第一部分
	var arr1 []interface{}
	var text1 LarkText
	text1.Tag = "text"
	text1.Text = lark_message + "\n"
	arr1 = append(arr1, text1)
	arr1 = append(arr1, divid)
	var fields1 Field
	fields1.Title.Tag = "text"
	fields1.Title.Text = "项目"
	fields1.Title.Lines = 1
	fields1.Value.Tag = "text"
	fields1.Value.Text = "静态检查项"
	fields1.Value.Lines = 1
	fields1.Short = true
	var fields2 Field
	fields2.Title.Tag = "text"
	fields2.Title.Text = "待确认数量"
	fields2.Title.Lines = 1
	fields2.Value.Tag = "text"
	fields2.Value.Text = detect_num
	fields2.Value.Lines = 1
	fields2.Short = true
	var fields3 Field
	fields3.Title.Tag = "text"
	fields3.Title.Text = "项目"
	fields3.Title.Lines = 1
	fields3.Value.Tag = "text"
	fields3.Value.Text = "自查项"
	fields3.Value.Lines = 1
	fields3.Short = true
	var fields4 Field
	fields4.Title.Tag = "text"
	fields4.Title.Text = "待确认数量"
	fields4.Title.Lines = 1
	fields4.Value.Tag = "text"
	fields4.Value.Text = self_item_num
	fields4.Value.Lines = 1
	fields4.Short = true
	fieldArr := []Field{fields1, fields2, fields3, fields4}
	var lark_field LarkField
	lark_field.Tag = "field"
	lark_field.Fields = fieldArr
	arr1 = append(arr1, lark_field)
	//content第二部分
	var arr2 []interface{}
	if detect_num == "0" && self_item_num == "0" {
		var noConfirm LarkText
		noConfirm.Tag = "text"
		noConfirm.Text = "本次检测未发现新增危险项！\n业务方无需点击确认~\n"
		noConfirm.Style = "width: 100%; fontSize: 15; color:#6495ED"
		arr2 = append(arr2, divid)
		arr2 = append(arr2, noConfirm)
	} else {
		var text2 LarkText
		text2.Tag = "text"
		text2.Text = "预审平台结果确认建议\n"
		text2.Style = "width: 100%; fontSize: 15"
		arr2 = append(arr2, divid)
		arr2 = append(arr2, text2)
		var fields5 Field
		fields5.Title.Tag = "text"
		fields5.Title.Text = "检测结果"
		fields5.Title.Lines = 1
		fields5.Value.Tag = "text"
		fields5.Value.Text = "静态检查项"
		fields5.Value.Lines = 1
		fields5.Short = true
		var fields6 Field
		fields6.Title.Tag = "text"
		fields6.Title.Text = "确认人"
		fields6.Title.Lines = 1
		fields6.Value.Tag = "text"
		if rd_bm != "" {
			fields6.Value.Text = "RD BM" + "(" + rd_bm + ")"
		} else {
			fields6.Value.Text = "RD BM"
		}
		fields6.Value.Lines = 1
		fields6.Short = true
		var fields7 Field
		fields7.Title.Tag = "text"
		fields7.Title.Text = "检测结果"
		fields7.Title.Lines = 1
		fields7.Value.Tag = "text"
		fields7.Value.Text = "自查项 Binary"
		fields7.Value.Lines = 1
		fields7.Short = true
		var fields8 Field
		fields8.Title.Tag = "text"
		fields8.Title.Text = "确认人"
		fields8.Title.Lines = 1
		fields8.Value.Tag = "text"
		if qa_bm != "" {
			fields8.Value.Text = "QA BM" + "(" + qa_bm + ")"
		} else {
			fields8.Value.Text = "QA BM"
		}
		fields8.Value.Lines = 1
		fields8.Short = true
		var fields9 Field
		fields9.Title.Tag = "text"
		fields9.Title.Text = "检测结果"
		fields9.Title.Lines = 1
		fields9.Value.Tag = "text"
		fields9.Value.Text = "自查项 Metedate"
		fields9.Value.Lines = 1
		fields9.Short = true
		var fields10 Field
		fields10.Title.Tag = "text"
		fields10.Title.Text = "确认人"
		fields10.Title.Lines = 1
		fields10.Value.Tag = "text"
		fields10.Value.Text = "产品线提审负责人"
		fields10.Value.Lines = 1
		fields10.Short = true
		var notice Field
		notice.Title.Tag = "text"
		notice.Title.Text = "---------------------------------------------\n"
		notice.Title.Lines = 1
		notice.Value.Tag = "text"
		notice.Value.Text = "！！！点击卡片跳转详情页！！！"
		notice.Value.Lines = 1
		notice.Short = false
		fieldArr2 := []Field{fields5, fields6, fields7, fields8, fields9, fields10, notice}
		var lark_field2 LarkField
		lark_field2.Tag = "field"
		lark_field2.Fields = fieldArr2
		arr2 = append(arr2, lark_field2)
	}
	//卡片
	var card LarkCard
	card.Link.Herf = url
	card.Header.Title = "预审平台消息通知"
	card.Header.Color = "blue"
	card.Header.Lines = 1
	var temp_arr []interface{}
	temp_arr = append(temp_arr, arr1)
	temp_arr = append(temp_arr, arr2)
	card.Content = temp_arr
	//内容
	var larkCC LarkContent
	larkCC.Card = card
	//消息
	var message LarkCardMessage
	message.MsgType = "interactive"
	if groupFlag {
		message.ChatID = lark_people
	} else {
		message.Email = lark_people + "@bytedance.com"
	}
	message.Content = larkCC

	return message
}

func LarkDetectResult(lark_people, rd_bm, qa_bm, lark_message, url string, detect_num, self_item_num int, groupFlag bool) bool {
	detect := strconv.Itoa(detect_num)
	self := strconv.Itoa(self_item_num)
	larkStruct := initLarkStruct(lark_people, rd_bm, qa_bm, lark_message, detect, self, url, groupFlag)
	larkBody, err := json.Marshal(larkStruct)
	if err != nil {
		logs.Error(err.Error())
		return false
	}
	if groupFlag {
		m := make(map[string]interface{})
		json.Unmarshal(larkBody, &m)
		delete(m, "email")
		larkBody, err = json.Marshal(m)
		if err != nil {
			logs.Error(err.Error())
		}
	}
	token := GetLarkToken()
	res, _ := PostJsonHttp3(larkBody, token, _const.OFFICE_LARK_URL)
	return res
}

func GetUserOpenId(email string) string {
	requestMap := map[string]interface{}{
		"email": email,
	}
	requestBody, _ := json.Marshal(requestMap)
	isPost, response := PostJsonHttp3(requestBody, GetLarkToken(), _const.LARK_Email2Id_URL)
	if isPost {
		m := make(map[string]interface{})
		json.Unmarshal([]byte(response), &m)
		if m != nil {
			return m["open_id"].(string)
		}
	}
	return ""
}

//获取用户的全部信息，包括中文名
func GetUserAllInfo(open_id string) string {
	requestUrl := "https://open.feishu.cn/open-apis/user/v3/info"
	params := map[string]string{
		"open_id": open_id,
	}
	return GetLarkInfo(requestUrl, params)
}

//强制拉用户到预审平台的用户群
func UserInGroup(username string) {
	inGroupUrl := "https://open.feishu.cn/open-apis/chat/v3/chatter/add/"
	user_id := GetUserOpenId(username + "@bytedance.com")
	allInfo := GetUserAllInfo(user_id)
	m := make(map[string]interface{})
	json.Unmarshal([]byte(allInfo), &m)
	user_ids := []string{user_id}
	if user_id != "" {
		m := map[string]interface{}{
			"open_chat_id": "oc_5226ab6b46ad51fc1a8926d15003b490",
			"open_ids":     user_ids,
		}
		request_body, _ := json.Marshal(m)
		PostJsonHttp3(request_body, GetLarkToken(), inGroupUrl)
	}
}

//群里拉预审机器人入群，必须robot的创建者在群内，暂时不用
func Bot2Group(groupId string) {
	bot2GroupUrl := "https://oapi.zjurl.cn/open-apis/api/v2/bot/chat/join"
	m := map[string]interface{}{
		"token":   myToken,
		"bot":     robotToken,
		"chat_id": groupId,
	}
	request_body, _ := json.Marshal(m)
	fmt.Println(request_body)
	PostJsonHttp3(request_body, "", bot2GroupUrl)
}

//初始化消息中的不变量
func resultLarkStruct(lark_people, lark_message, detect_no_pass, self_no_pass, url string, groupFlag bool) LarkCardMessage {
	//分割线
	var divid LarkText
	divid.Tag = "text"
	divid.Text = "------------------------------------------\n"
	divid.Style = "width: 100%; fontSize: 13;color: #C2C2C2"
	//内容
	//content第一部分
	var arr1 []interface{}
	var text1 LarkText
	text1.Tag = "text"
	text1.Text = lark_message + "\n"
	arr1 = append(arr1, text1)
	arr1 = append(arr1, divid)
	var fields1 Field
	fields1.Title.Tag = "text"
	fields1.Title.Text = "项目"
	fields1.Title.Lines = 1
	fields1.Value.Tag = "text"
	fields1.Value.Text = "静态检查项"
	fields1.Value.Lines = 1
	fields1.Short = true
	var fields2 Field
	fields2.Title.Tag = "text"
	fields2.Title.Text = "未通过数量"
	fields2.Title.Lines = 1
	fields2.Value.Tag = "text"
	fields2.Value.Text = detect_no_pass
	fields2.Value.Lines = 1
	fields2.Short = true
	var fields3 Field
	fields3.Title.Tag = "text"
	fields3.Title.Text = "项目"
	fields3.Title.Lines = 1
	fields3.Value.Tag = "text"
	fields3.Value.Text = "自查项"
	fields3.Value.Lines = 1
	fields3.Short = true
	var fields4 Field
	fields4.Title.Tag = "text"
	fields4.Title.Text = "未通过数量"
	fields4.Title.Lines = 1
	fields4.Value.Tag = "text"
	fields4.Value.Text = self_no_pass
	fields4.Value.Lines = 1
	fields4.Short = true
	fieldArr := []Field{fields1, fields2, fields3, fields4}
	var lark_field LarkField
	lark_field.Tag = "field"
	lark_field.Fields = fieldArr
	arr1 = append(arr1, lark_field)
	//content第二部分
	var arr2 []interface{}
	var noConfirm LarkText
	noConfirm.Tag = "text"
	if detect_no_pass == "0" && self_no_pass == "0" {
		noConfirm.Text = "Notice:检测结果已全部确认通过!\n业务方可正常执行下一步操作！"
		noConfirm.Style = "width: 100%; fontSize: 15; color:#DAA520"
	} else {
		noConfirm.Text = "Notice:存在不通过项，无法进行下一步\n点击卡片查看详情，更改后重新上传检测"
		noConfirm.Style = "width: 100%; fontSize: 15; color:#DC143C"
	}
	arr2 = append(arr2, divid)
	arr2 = append(arr2, noConfirm)
	//卡片
	var card LarkCard
	card.Link.Herf = url
	card.Header.Title = "预审确认结果通知"
	card.Header.Color = "yellow"
	card.Header.Lines = 1
	var temp_arr []interface{}
	temp_arr = append(temp_arr, arr1)
	temp_arr = append(temp_arr, arr2)
	card.Content = temp_arr
	//内容
	var larkCC LarkContent
	larkCC.Card = card
	//消息
	var message LarkCardMessage
	message.MsgType = "interactive"
	if groupFlag {
		message.ChatID = lark_people
	} else {
		message.Email = lark_people + "@bytedance.com"
	}
	message.Content = larkCC

	return message
}
func LarkConfirmResult(lark_people, lark_message, url string, detect_no_pass, self_no_pass int, groupFlag bool) bool {
	larkStruct := resultLarkStruct(lark_people, lark_message, strconv.Itoa(detect_no_pass), strconv.Itoa(self_no_pass), url, groupFlag)
	larkBody, err := json.Marshal(larkStruct)
	if err != nil {
		logs.Error(err.Error())
		return false
	}
	if groupFlag {
		m := make(map[string]interface{})
		json.Unmarshal(larkBody, &m)
		delete(m, "email")
		larkBody, err = json.Marshal(m)
		if err != nil {
			logs.Error(err.Error())
		}
	}
	token := GetLarkToken()
	res, _ := PostJsonHttp3(larkBody, token, _const.OFFICE_LARK_URL)
	return res
}
