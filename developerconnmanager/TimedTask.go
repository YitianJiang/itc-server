package developerconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"code.byted.org/yuyilei/bot-api/form"
	"code.byted.org/yuyilei/bot-api/service"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"strconv"
	"time"
)

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

func NotifyProfileExpired(c *gin.Context) {
	logs.Info("HttpRequest：向app负责人发送消息卡片提醒profile一个月后过期")
	expiredProfileCardInputs,queryResult:=devconnmanager.QueryExpiredProfileRelatedInfo()
	if !queryResult{
		logs.Error("从数据库中查询profile相关信息失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "从数据库中查询profile相关信息失败", "failed")
		return
	}
	abot := service.BotService{}
	abot.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
	for _,expiredProfileCardInput :=range *expiredProfileCardInputs{
		cardElementForms := generateCardOfProfileExpired(&expiredProfileCardInput)
		if err := sendIOSCertLarkMessage(cardElementForms, nil, expiredProfileCardInput.UserName, &abot);err != nil{
			logs.Error("向app: %v负责人%v发送消息卡片提醒一个月后profile过期失败%v",expiredProfileCardInput.AppName,expiredProfileCardInput.UserName, err)
		}
	}
	AssembleJsonResponse(c,  _const.SUCCESS, "向app负责人发送消息卡片提醒一个月后profile过期 成功", nil)
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
	expiredCertInfos,queryResult := devconnmanager.QueryExpiredCertRelatedInfo()
	if !queryResult{
		logs.Error("查询将要过期的证书信息失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "查询将要过期的证书信息失败", "failed")
		return
	}
	abot := service.BotService{}
	abot.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
	for _, expiredCertInfo := range *expiredCertInfos {
		appNamePrincipalMap:=make(map[string]string)
		//把负责人负责的app都放到一张卡片中统一通知
		for _,appAndPrincipal :=range expiredCertInfo.AppAndPrincipals {
			if _, ok := appNamePrincipalMap[appAndPrincipal.UserName]; !ok {
				appNamePrincipalMap[appAndPrincipal.UserName] = appAndPrincipal.AppName
			} else {
				appNamePrincipalMap[appAndPrincipal.UserName] +=  "、"+appAndPrincipal.AppName
			}
		}
		for principal,appNames:=range appNamePrincipalMap{
			cardElementForms := generateCardOfCertExpired(&expiredCertInfo,appNames)
			if err := sendIOSCertLarkMessage(cardElementForms, nil, principal, &abot);err!=nil{
				logs.Error("向app: %v负责人%v发送消息卡片提醒一个月后证书过期失败%v",appNames,principal, err)
			}
		}
	}
	AssembleJsonResponse(c, _const.SUCCESS, "向各app负责人发送消息卡片提醒一个月后证书过期 成功", nil)
	return
}
