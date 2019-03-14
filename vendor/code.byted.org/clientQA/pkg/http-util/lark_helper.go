package http_util

import (
	"code.byted.org/clientQA/pkg/const"
	"code.byted.org/clientQA/pkg/database"
	"code.byted.org/clientQA/pkg/request-processor/request-dal"
	"code.byted.org/gopkg/logs"
	"code.byted.org/zhangwanlong/go-lark"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"math/rand"
	"os"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

/*
发送lark消息
*/
func larkDing2(msg string, api string, chat_id string, msgType string, title string) {
	body := map[string]interface{}{"msg_type": msgType}
	body["token"] = _const.Robot_token
	body["chat_id"] = chat_id
	body["content"] = map[string]string{"text": msg, "title": title}
	body_byte, _ := json.Marshal(body)
	PostJsonHttp(api, body_byte)
}

func LarkDingOneInnerV2(member string, msg string) {
	// 获取user的larkid和会话id
	emailPrefix := getEmailPrefix(member)
	channelID := database.GetLarkChannel(emailPrefix)
	if channelID == "" {
		userLarkID := GetUserIDinLark2(_const.Lark_user_token, emailPrefix)
		channelID = GetUserChannelID(_const.Robot_token, emailPrefix, userLarkID)
	}
	go larkDing2(msg, _const.Larknotice_url, channelID, "text", "")
}

//func LarkDingOneInnerV3(member string, msg string, msgType string) {
//	// 获取user的larkid和会话id
//	userLarkID := GetUserIDinLark2(_const.Lark_user_token, member)
//	channelID := GetUserChannelID(_const.Robot_token, userLarkID)
//	go larkDing2(msg, _const.Larknotice_url, channelID,msgType)
//}

//获取用户的id，不然不知道给谁发消息，如果出了问题，肯定返回的是空串了
func GetUserIDinLark2(token string, email_prefix string) string {
	var retstr string
	//如果缓存里面有，就直接从缓存拿了
	retstr = database.GetLarkID(email_prefix)
	if retstr != "" {
		return retstr
	}

	body := map[string]interface{}{"token": token}
	body["email_prefix"] = email_prefix
	//这里写死了是bytedance吧，反正也用不到别的组织
	body["organization_name"] = "bytednace"
	body_byte, _ := json.Marshal(body)
	err, response := PostJsonHttp(_const.Lark_finduser_url, body_byte)
	var ret_body_json LarkUserStruct
	json.Unmarshal(response, &ret_body_json)
	if err == nil && ret_body_json.Ok == true {
		retstr = fmt.Sprint(ret_body_json.User_id)
	}

	//如果缓存没有用户信息，再set一次
	if _, err := database.SetLarkID(email_prefix, retstr); err != nil {
		logs.Error("%v", err)
	}
	return retstr
}

func getEmailPrefix(email string) string {
	emailPrefix := email
	if len(strings.Split(email, "@")) > 1 {
		emailPrefix = strings.Split(email, "@")[0]
	}
	return emailPrefix
}

//获取和用户会话的channel id
func GetUserChannelID(token string, emailPrefix string, userid string) string {
	var retstr string
	body := map[string]interface{}{"token": token}
	body["user"] = userid
	body_byte, _ := json.Marshal(body)
	err, response := PostJsonHttp(_const.Lark_openim_url, body_byte)
	var ret_body_json IMStruct
	json.Unmarshal(response, &ret_body_json)
	if err == nil && ret_body_json.Ok == true {
		retstr = ret_body_json.Channel.Id
	} else {
		logs.Error("%s", "打开私聊会话失败")
	}
	if _, err := database.SetLarkChannel(emailPrefix, retstr); err != nil {
		logs.Error("%v", err)
	}
	return retstr
}

//lark发消息给群里，内部使用的方法，第二版
//botType：0，通知机器人，1，日历机器人，2,jira机器人
// infotype：0：单人，1：所有人，2：无人,3，部分人
func LarkDingInGroupInner_V2(infoType string, member string, msg string, groupID string) {
	var userFullName string
	var userLarkID string

	if infoType == "0" {
		memberStr := strings.Split(member, ",")
		if len(memberStr) > 0 {
			tempUser := request_dal.GetUserInfo(memberStr[0])
			if &tempUser != nil {
				userFullName = tempUser.Full_name
			} else {
				userFullName = memberStr[0]
			}
			emailPrefix := getEmailPrefix(memberStr[0])
			userLarkID = GetUserIDinLark2(_const.Lark_user_token, emailPrefix)
		}
	}
	switch infoType {
	case "0":
		go larkDing2(msg+"\n"+
			"<at user_id=\""+userLarkID+"\">@"+userFullName+"</at>",
			_const.Larknotice_url, groupID, "text", "")
	case "1":
		go larkDing2(msg+"<at user_id=\"all\">@所有人</at>", _const.Larknotice_url, groupID, "text", "")
	case "2":
		go larkDing2(msg, _const.Larknotice_url, groupID, "text", "")
	case "3":
		informUsers := make(map[string]string)
		for _, memberitem := range strings.Split(member, ",") {
			//这里先从rocket里面获取人名，获取不到再直接赋值传入的内容
			tempUserName := request_dal.GetUserInfo(memberitem)
			emailPrefix := getEmailPrefix(memberitem)
			tempLarkUserID := GetUserIDinLark2(_const.Lark_user_token, emailPrefix)

			if !(tempLarkUserID == "" || tempUserName == nil) {
				//如果没获取到用户以及用户的userid是空，那么就跳过此次
				informUsers[tempLarkUserID] = tempUserName.Full_name
			} else {
				continue
			}
		}
		msg += "\n"
		for key, value := range informUsers {
			msg += "<at user_id=\"" + key + "\">@" + value + "</at>"
		}
		go larkDing2(msg, _const.Larknotice_url, groupID, "text", "")
	}
}

func uploadImg(imgUrl string) (string, error) {
	//脑壳疼，先存一次文件吧
	tempFilename := randStringBytesMaskImprSrc(5) + ".png"
	err := qrcode.WriteFile(imgUrl, qrcode.Medium, 256, tempFilename)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFilename)
	ret := map[string]interface{}{}
	params := map[string]string{"token": _const.Robot_token}
	response, err := PostLocalFileWithParams(params, "image", tempFilename, _const.Lark_upload_img)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal([]byte(response), &ret)
	if err != nil {
		return "", err
	}

	if _, ok := ret["image_key"]; ok {
		return ret["image_key"].(string), nil
	} else {
		return "", errors.New("image_key not found")
	}
}

/*
生成随机长度字符串
*/
func randStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	var src = rand.NewSource(time.Now().UnixNano())
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}


//func LarkOneWithImg(imgUrl string, member string, msg string, msgType string, title string) error {
//	// 获取user的larkid和会话id
//	imageKey, err := uploadImg(imgUrl)
//	if err != nil {
//		logs.Error("%v", err)
//		return err
//	}
//	msg = "<p><text>" + msg + "</text></p>" +
//		"<p><img key='" + imageKey + "'origin-width='300'origin-height='300'></img></p>"
//	userLarkID := GetUserIDinLark2(_const.Lark_user_token, member)
//	channelID := GetUserChannelID(_const.Robot_token, userLarkID)
//	go larkDing2(msg, _const.Larknotice_url, channelID, msgType, title)
//	return nil
//}

/*
使用https://code.byted.org/zhangwanlong/go-lark中的sdk发送lark消息
imagePath:图片的本地路径
chatID：会话id
msg：需要发送的文本信息
title：富文本标题名字

成功，返回""，nil；失败返回错误原因和err
*/
func LarkOneWithImgGitlab(imagePath string, member string, msg string, title string) (string, error) {
	bot := lark.NewChatBot(_const.Robot_token)
	//上传文件，获取image key
	imageKey, err := uploadImg(imagePath)
	if err != nil {
		logs.Error("Upload Image failed %v\n", err)
		return "", err
	}
	//获取channelID
	emailPrefix := getEmailPrefix(member)
	channelID := database.GetLarkChannel(emailPrefix)
	if channelID == "" {
		userLarkID := GetUserIDinLark2(_const.Lark_user_token, emailPrefix)
		channelID = GetUserChannelID(_const.Robot_token, emailPrefix, userLarkID)
	}
	//富文本消息通过MessageBuilder构建
	mb := lark.NewMsgBuffer(lark.MsgRichText)
	mb.TitleWithImg(title)
	mb.Text("<p><text>" + msg + "</text></p>" +
		"<p><img key= '" + imageKey + "'origin-width='300'origin-height='300'></img></p>")
	sendMessage := mb.BindChannel(channelID).Build()
	responseData, err := bot.PostMessage(sendMessage)
	if err == nil && responseData.Ok == true {
		return "", nil
	} else {
		return responseData.Error, err
	}

}
