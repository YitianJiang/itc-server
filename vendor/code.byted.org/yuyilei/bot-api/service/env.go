package service

import (
	"os"
)

var UrlPrefix string
var OpenId2UserIdUrl string
var UserId2OpenIdUrl string
var MessageId2OpenMessageIdUrl string
var OpenMessageId2MessageIdUrl string
var ChatId2OpenChatIdUrl string
var OpenChatId2ChatIdUrl string
var UserId2EmployeeIdUrl string
var EmployeeId2UserIdUrl string
var DepartmentId2OpenDepartmentIdUrl string
var OpenDepartmentId2DepartmentIdUrl string
var Email2OpenIdEmployeeIdUrl string
var UrgentMessageUrl string
var SendMessageUrl string
var SendBatchMessageUrl string
var UploadImageUrl string
var AccessTokenUrl string
var GetChatInfoUrl string
var UpdateChatInfoUrl string
var GetChatListUrl string
var CreateChatUrl string
var GetP2pChatIdUrl string
var InviteUser2ChatUrl string
var DeleteUserForChatUrl string
var GetUserInfoByOpenIdUrl string
var GetAdminInfoListUrl string
var GetBotInfoUrl string

var DisbandChatUrl string
var AddBotToChatUrl string
var RemoveBotFromChatUrl string

func IsOnline() bool {
	env := os.Getenv("ENV")
	return env == "online" || env == "pre_release"
}

func IsStaging() bool {
	env := os.Getenv("ENV")
	return env == "staging"
}

// 海外
func IsVaAws() bool {
	idc := os.Getenv("IDC")
	return idc == "va_aws"
}


func init()  {
	// 不设置环境变量的话，默认为online
	if IsVaAws() {
		if IsStaging() {
			initVaAwsStagingUrlPrefix()
		} else {
			initVaAwsOnlineUrlPrefix()
		}
	} else {
		if IsStaging() {
			initStagingUrlPrefix()
		} else {
			initOnlineUrlPrefix()
		}
	}
	initUrl()
}


func initOnlineUrlPrefix()  {
	UrlPrefix = "https://open.feishu.cn"
}

func initStagingUrlPrefix()  {
	UrlPrefix = "https://open.feishu-staging.cn"
}

func initVaAwsOnlineUrlPrefix()  {
	UrlPrefix = "https://open.larksuite.com"
}

func initVaAwsStagingUrlPrefix() {
	UrlPrefix = "https://open.larksuite-staging.com"
}

func initUrl()  {
	OpenId2UserIdUrl = UrlPrefix + "/open-apis/exchange/v3/openid2uid/"
	UserId2OpenIdUrl = UrlPrefix + "/open-apis/exchange/v3/uid2openid/"
	MessageId2OpenMessageIdUrl = UrlPrefix + "/open-apis/exchange/v3/mid2omid/"
	OpenMessageId2MessageIdUrl = UrlPrefix + "/open-apis/exchange/v3/omid2mid/"
	ChatId2OpenChatIdUrl = UrlPrefix + "/open-apis/exchange/v3/cid2ocid/"
	OpenChatId2ChatIdUrl = UrlPrefix + "/open-apis/exchange/v3/ocid2cid/"
	UserId2EmployeeIdUrl = UrlPrefix + "/open-apis/exchange/v3/uid2eid/"
	EmployeeId2UserIdUrl = UrlPrefix + "/open-apis/exchange/v3/eid2uid/"
	DepartmentId2OpenDepartmentIdUrl = UrlPrefix + "/open-apis/exchange/v3/did2odid/"
	OpenDepartmentId2DepartmentIdUrl = UrlPrefix + "/open-apis/exchange/v3/odid2did/"
	Email2OpenIdEmployeeIdUrl = UrlPrefix + "/open-apis/user/v3/email2id"
	UrgentMessageUrl = UrlPrefix + "/open-apis/message/v3/urgent/"
	SendMessageUrl = UrlPrefix + "/open-apis/message/v3/send/"
	SendBatchMessageUrl = UrlPrefix + "/open-apis/message/v3/batch_send/"
	UploadImageUrl = UrlPrefix + "/open-apis/image/v4/upload/"//v4接口
	AccessTokenUrl = UrlPrefix + "/open-apis/auth/v3/tenant_access_token/internal/"
	GetChatInfoUrl = UrlPrefix + "/open-apis/chat/v3/info/"
	UpdateChatInfoUrl = UrlPrefix + "/open-apis/chat/v3/update/"
	GetChatListUrl = UrlPrefix + "/open-apis/chat/v3/list/"
	CreateChatUrl = UrlPrefix + "/open-apis/chat/v3/create/"
	GetP2pChatIdUrl = UrlPrefix + "/open-apis/chat/v3/p2p/id"
	InviteUser2ChatUrl = UrlPrefix + "/open-apis/chat/v3/chatter/add/"
	DeleteUserForChatUrl = UrlPrefix + "/open-apis/chat/v3/chatter/delete/"
	GetUserInfoByOpenIdUrl = UrlPrefix + "/open-apis/user/v3/info"
	GetAdminInfoListUrl = UrlPrefix + "/open-apis/user/v3/app_admin/list/"
	GetBotInfoUrl = UrlPrefix + "/open-apis/bot/v3/info/"

	DisbandChatUrl = UrlPrefix + "/open-apis/chat/v4/disband"
	AddBotToChatUrl = UrlPrefix + "/open-apis/bot/v4/add"
	RemoveBotFromChatUrl = UrlPrefix + "/open-apis/bot/v4/remove"
}
