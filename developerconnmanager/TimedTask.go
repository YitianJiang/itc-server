package developerconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"code.byted.org/yuyilei/bot-api/form"
	"code.byted.org/yuyilei/bot-api/service"
	"errors"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func sendExpiredCardMessage(cardInfoFormArray *[][]form.CardElementForm, cardActions *[]form.CardActionForm, groupChatId string, botService *service.BotService, suffixs ...string) error {
	//生成卡片
	cardHeaderTitle := "iOS证书管理通知"
	for _, suffix := range suffixs {
		cardHeaderTitle += suffix
	}
	var cardForm *form.CardForm
	if cardActions != nil {
		cardForm = form.GenerateCardForm(nil, getCardHeader(cardHeaderTitle), *cardInfoFormArray, *cardActions)
	} else {
		cardForm = form.GenerateCardForm(nil, getCardHeader(cardHeaderTitle), *cardInfoFormArray, nil)
	}
	cardMessageContent := form.GenerateCardMessageContent(cardForm)
	cardMessage, err := form.GenerateMessage("interactive", cardMessageContent)
	utils.RecordError("card信息生成出错: ", err)
	if cardMessage == nil {
		return errors.New("card信息生成出错")
	}
	cardMessage.ChatID = &groupChatId
	sendMsgResp, err := botService.SendMessage(*cardMessage)
	logs.Info("发送飞书lark消息响应= %v", sendMsgResp)
	if code, ok := sendMsgResp["code"].(float64); ok && code != 0 {
		if message, ok := sendMsgResp["msg"].(string); ok {
			return errors.New(message)
		}
	}
	return err
}

func generateCardOfProfileExpired(expiredProfileCardInput *devconnmanager.ExpiredProfileCardInput) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm
	//插入提示信息
	messageText := utils.ProfileExpiredCardHeader
	messageForm := form.GenerateTextTag(&messageText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.ProfileNameHeader,expiredProfileCardInput.ProfileName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.ProfileTypeHeader,expiredProfileCardInput.ProfileType))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.ProfileIdHeader,expiredProfileCardInput.ProfileId))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.ProfileExpiredBundleIdHeader,expiredProfileCardInput.BundleId))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.ProfileExpiredAppNameHeader,expiredProfileCardInput.AppName))
	divideText := utils.DivideText
	divideForm := form.GenerateTextTag(&divideText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*divideForm})
	timeToExpired:=strconv.Itoa(int(math.Floor(expiredProfileCardInput.ProfileExpireDate.Sub(time.Now()).Hours()/24)))
	cardFormArray = append(cardFormArray, *generateEmphasisInfoLineOfCard(utils.ProfileExpiredTimeToNow,timeToExpired+_const.DayTimeUnit))
	cardFormArray = append(cardFormArray, *generateEmphasisInfoLineOfCard(utils.ProfileExpiredTipHeader,""))
	cardFormArray = append(cardFormArray, *generateAtLineOfCard(utils.ProfileUpdateHeader,
		_const.ITC_SIGN_SYSTEM_ADDRESS+expiredProfileCardInput.AppId,_const.ITC_SIGN_SYSTEM_ADDRESS+expiredProfileCardInput.AppId))
	return &cardFormArray
}

func generateCardOfQueryDbFail(queryFailTip string) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm
	messageText := queryFailTip
	messageForm := form.GenerateTextTag(&messageText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})
	return &cardFormArray
}

func NotifyProfileExpired(c *gin.Context) {
	logs.Info("HttpRequest：向app负责人发送消息卡片提醒profile一个月后过期")
	abot := service.BotService{}
	abot.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
	expiredProfileCardInputs,queryResult:=devconnmanager.QueryExpiredProfileRelatedInfo()
	if !queryResult{
		logs.Error("从数据库中查询profile相关信息失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "从数据库中查询profile相关信息失败", "failed")
		cardElementForms :=generateCardOfQueryDbFail(utils.QueryExpiredProfileFailTip)
		if err := sendExpiredCardMessage(cardElementForms, nil, _const.IOS_CERT_MANAGE_GROUP_CHAT_ID , &abot);err != nil{
			logs.Error("向iOS证书管理群发送消息卡片提醒从数据库中查询将要过期的描述文件失败%v", err)
		}
		return
	}
	for _,expiredProfileCardInput :=range *expiredProfileCardInputs{
		cardElementForms := generateCardOfProfileExpired(&expiredProfileCardInput)
		if err := sendExpiredCardMessage(cardElementForms, nil, _const.IOS_CERT_MANAGE_GROUP_CHAT_ID , &abot);err != nil{
			logs.Error("向iOS证书管理群发送消息卡片提醒一个月后描述文件过期失败%v", err)
		}
	}
	AssembleJsonResponse(c,  _const.SUCCESS, "向iOS证书管理群发送消息卡片提醒一个月后描述文件过期 成功", nil)
	return
}

func generateEmphasisInfoLineOfCard(header string, content string) *[]form.CardElementForm {
	var infoLineFormList []form.CardElementForm

	headerForm := form.GenerateTextTag(&header, false, nil)
	headerForm.Style = &utils.RedHeaderStyle
	infoLineFormList = append(infoLineFormList, *headerForm)

	appIdForm := form.GenerateTextTag(&content, false, nil)
	appIdForm.Style = &utils.RedHeaderStyle
	infoLineFormList = append(infoLineFormList, *appIdForm)

	return &infoLineFormList
}

func generateCardOfCertExpired(expiredCertCardInput *devconnmanager.ExpiredCertCardInput,appNames string) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm
	//插入提示信息
	messageTextFront := utils.CertExpiredCardHeader
	messageFormFront := form.GenerateTextTag(&messageTextFront, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageFormFront})
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.DeleteCertNameHeader,expiredCertCardInput.CertName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.CreateCertTypeHeader,expiredCertCardInput.CertType))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.DeleteCertIdHeader,expiredCertCardInput.CertId))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.CertExpiredAccountNameHeader,expiredCertCardInput.AccountName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.CertExpiredTeamIdHeader,expiredCertCardInput.TeamId))
	divideText := utils.DivideText
	divideForm := form.GenerateTextTag(&divideText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*divideForm})
	timeToExpired:=strconv.Itoa(int(math.Floor(expiredCertCardInput.CertExpireDate.Sub(time.Now()).Hours()/24)))
	cardFormArray = append(cardFormArray, *generateEmphasisInfoLineOfCard(utils.CertExpiredTimeToNow,timeToExpired+_const.DayTimeUnit))
	cardFormArray = append(cardFormArray, *generateEmphasisInfoLineOfCard(utils.CertExpiredAppHeader,appNames))
	cardFormArray = append(cardFormArray, *generateEmphasisInfoLineOfCard(utils.CertExpiredTipHeader,""))
	cardFormArray = append(cardFormArray, *generateAtLineOfCard(utils.CertBindChangeHeader,_const.ITC_CERT_SYSTEM_ADDRESS,_const.ITC_CERT_SYSTEM_ADDRESS))
	return &cardFormArray
}

func NotifyCertExpired(c *gin.Context) {
	logs.Info("检查一个月内将要过期的证书并发送卡片通知")
	abot := service.BotService{}
	abot.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
	expiredCertInfos,queryResult := devconnmanager.QueryExpiredCertRelatedInfo()
	if !queryResult{
		logs.Error("查询将要过期的证书信息失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "查询将要过期的证书信息失败", "failed")
		cardElementForms :=generateCardOfQueryDbFail(utils.QueryExpiredCertFailTip)
		if err := sendExpiredCardMessage(cardElementForms, nil, _const.IOS_CERT_MANAGE_GROUP_CHAT_ID , &abot);err != nil{
			logs.Error("向iOS证书管理群发送消息卡片提醒从数据库中查询将要过期的证书失败%v", err)
		}
		return
	}
	for _, expiredCertInfo := range *expiredCertInfos {
		    var affectedAppNamesTogether string
		    if len(expiredCertInfo.AffectedApps)==0{
		    	continue
		    }
		    for _,affectedApp:=range expiredCertInfo.AffectedApps {
			    affectedAppNamesTogether += affectedApp.AppName + "、"
		    }
		    affectedAppNamesTogether=strings.TrimRight(affectedAppNamesTogether,"、")
			cardElementForms := generateCardOfCertExpired(&expiredCertInfo,affectedAppNamesTogether)
			if err := sendExpiredCardMessage(cardElementForms, nil, _const.IOS_CERT_MANAGE_GROUP_CHAT_ID, &abot);err!=nil{
				logs.Error("向iOS证书管理群发送消息卡片提醒一个月后证书过期失败%v", err)
			}
	}
	AssembleJsonResponse(c, _const.SUCCESS, "向iOS证书管理群发送消息卡片提醒一个月后证书过期 成功", nil)
	return
}
