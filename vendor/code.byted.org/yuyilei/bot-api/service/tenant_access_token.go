package service

import (
	"code.byted.org/yuyilei/bot-api/form"
	"fmt"
	"time"
)


type TenantAccessToken struct {
	Token string
	ExpireTime time.Time
}

func GetTenantAccessToken(AppId string, AppSecret string) (*form.AccessTokenResp, error) {
	var requestBody form.AccessTokenForm
	requestBody.AppId = &AppId
	requestBody.AppSecret = &AppSecret
	response, err := PostRequestToken(AccessTokenUrl, nil, requestBody)
	return response, err
}

type BotService struct {
	AppId *string
	AppSecret *string
	TenantAccessToken *TenantAccessToken
}

func (abot *BotService) SetAppIdAndAppSecret(appId, appSecret string) {
	abot.AppId = &appId
	abot.AppSecret = &appSecret
}

func (abot *BotService) getToken() error {
	response, err := GetTenantAccessToken(*(abot.AppId), *(abot.AppSecret))
	if err != nil {
		return fmt.Errorf("invaild appId and appSecret can not get tenantAccessToken, error= %v", err)
	}
	if response.Code != 0 {
		return fmt.Errorf("get tenantAccessToken responseCode is %v, ErrorMsg is %v", response.Code, response.Error)
	}
	expire := time.Duration(response.Expire - 5*60)     //  离过期时间小于 5分钟
	now := time.Now()
	abot.TenantAccessToken = new(TenantAccessToken)
	abot.TenantAccessToken.ExpireTime = now.Add(expire * time.Second)
	abot.TenantAccessToken.Token = response.TenantAccessToken
	return nil
}

func (abot *BotService) getValidTenantAccessToken() error {
	if abot.AppId == nil || abot.AppSecret == nil {
		return fmt.Errorf("empty appId or appSecret, init appId and appSecret first")
	}
	if abot.TenantAccessToken == nil {       // 没有 accessToken
		err := abot.getToken()
		if err != nil {
			return fmt.Errorf("get TenantAccessToken failed, error= %v", err)
		}
		return nil
	}
	if abot.TenantAccessToken.ExpireTime.Before(time.Now()) {    // accessToken 过期
		err := abot.getToken()
		if err != nil {
			return fmt.Errorf("get TenantAccessToken failed, error= %v", err)
		}
		return nil
	}
	return nil
}
