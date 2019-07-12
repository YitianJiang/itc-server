package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"code.byted.org/gopkg/logs"
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
		fmt.Println("打开私聊会话失败，如果是0，问题在于Lark平台" + fmt.Sprint(num))
	}
	return retstr
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
func initLarkStruct(lark_people, lark_message, detect_num, self_item_num, url string) LarkCardMessage {
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
	fields6.Value.Text = "RD BM"
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
	fields8.Value.Text = "QA BM"
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
	message.Email = lark_people + "@bytedance.com"
	message.Content = larkCC

	return message
}

func LarkDetectResult(lark_people, lark_message, url string, detect_num, self_item_num int) bool {
	detect := strconv.Itoa(detect_num)
	self := strconv.Itoa(self_item_num)
	larkStruct := initLarkStruct(lark_people, lark_message, detect, self, url)
	larkBody, err := json.Marshal(larkStruct)
	if err != nil {
		fmt.Println("error", err)
		return false
	}
	token := GetLarkToken()
	res := PostJsonHttp3(larkBody, token)
	return res
}
