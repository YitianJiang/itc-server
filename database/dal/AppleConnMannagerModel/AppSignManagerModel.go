package devconnmanager

import (
	"code.byted.org/gopkg/gorm"
	"time"
)

//http request model
type CreateAppBindAccountRequest struct {
	AppId    string `json:"app_id" binding:"required"`
	AppName  string `json:"app_name" binding:"required"`
	AppType  string `json:"app_type" binding:"required"`
	UserName string `json:"user_name" binding:"required"`
	TeamId   string `json:"team_id" binding:"required"`
}

//db model
type AppAccountCert struct {
	gorm.Model
	AppId               string `gorm:"app_id"                           json:"app_id"`
	AppName             string `gorm:"column:app_name"                  json:"app_name"`
	AppType             string `gorm:"column:app_type"                  json:"app_type"`
	UserName            string `gorm:"column:user_name"                 json:"user_name"`
	TeamId              string `gorm:"column:team_id"                   json:"team_id"`
	AccountVerifyStatus string `gorm:"column:account_verify_status"     json:"account_verify_status"`
	AccountVerifyUser   string `gorm:"column:account_verify_user"       json:"account_verify_user"`
	DevCertId           string `gorm:"column:dev_cert_id"               json:"dev_cert_id"`
	DistCertId          string `gorm:"column:dist_cert_id"              json:"dist_cert_id"`
}

type AppBoundleProfiles struct {
	gorm.Model
	AppId              string `json:"app_id"`
	AppName            string `json:"app_name"`
	BundleidId         string `json:"bundleid_id"`
	BundleidIsdel      string `json:"bundleid_isdel"`
	DevProfileId       string `json:"dev_profile_id"`
	UserName           string `json:"user_name"`
	DistAdhocProfileId string `json:"dist_adhoc_profile_id"`
	DistProfileId      string `json:"dist_profile_id"`
}

type AppleBundleId struct {
	gorm.Model
	BundleidId                       string `gorm:"bundleid_id"                         json:"bundleid_id"`
	BundleidName                     string `gorm:"bundleid_name"                       json:"bundleid_name"`
	BundleId                         string `gorm:"bundle_id"                           json:"bundle_id"`
	BundleidType                     string `gorm:"bundleid_type"                       json:"bundleid_type"`
	ICLOUD                           string `gorm:"ICLOUD"                              json:"ICLOUD"`
	IN_APP_PURCHASE                  string `gorm:"IN_APP_PURCHASE"                     json:"IN_APP_PURCHASE"`
	GAME_CENTER                      string `gorm:"GAME_CENTER"                         json:"GAME_CENTER"`
	PUSH_NOTIFICATIONS               string `gorm:"PUSH_NOTIFICATIONS"                  json:"PUSH_NOTIFICATIONS"`
	WALLET                           string `gorm:"WALLET"                              json:"WALLET"`
	INTER_APP_AUDIO                  string `gorm:"INTER_APP_AUDIO"                     json:"INTER_APP_AUDIO"`
	MAPS                             string `gorm:"MAPS"                                json:"MAPS"`
	ASSOCIATED_DOMAINS               string `gorm:"ASSOCIATED_DOMAINS"                  json:"ASSOCIATED_DOMAINS"`
	PERSONAL_VPN                     string `gorm:"PERSONAL_VPN"                        json:"PERSONAL_VPN"`
	APP_GROUPS                       string `gorm:"APP_GROUPS"                          json:"APP_GROUPS"`
	HEALTHKIT                        string `gorm:"HEALTHKIT"                           json:"HEALTHKIT"`
	HOMEKIT                          string `gorm:"HOMEKIT"                             json:"HOMEKIT"`
	WIRELESS_ACCESSORY_CONFIGURATION string `gorm:"WIRELESS_ACCESSORY_CONFIGURATION"    json:"WIRELESS_ACCESSORY_CONFIGURATION"`
	APPLE_PAY                        string `gorm:"APPLE_PAY"                           json:"APPLE_PAY"`
	DATA_PROTECTION                  string `gorm:"DATA_PROTECTION"                     json:"DATA_PROTECTION"`
	SIRIKIT                          string `gorm:"SIRIKIT"                             json:"SIRIKIT"`
	NETWORK_EXTENSIONS               string `gorm:"NETWORK_EXTENSIONS"                  json:"NETWORK_EXTENSIONS"`
	MULTIPATH                        string `gorm:"MULTIPATH"                           json:"MULTIPATH"`
	HOT_SPOT                         string `gorm:"HOT_SPOT"                            json:"HOT_SPOT"`
	NFC_TAG_READING                  string `gorm:"NFC_TAG_READING"                     json:"NFC_TAG_READING"`
	CLASSKIT                         string `gorm:"CLASSKIT"                            json:"CLASSKIT"`
	AUTOFILL_CREDENTIAL_PROVIDER     string `gorm:"AUTOFILL_CREDENTIAL_PROVIDER"        json:"AUTOFILL_CREDENTIAL_PROVIDER"`
	ACCESS_WIFI_INFORMATION          string `gorm:"ACCESS_WIFI_INFORMATION"             json:"ACCESS_WIFI_INFORMATION"`
}

type AppleProfile struct {
	gorm.Model
	ProfileId          string    `json:"profile_id"`
	ProfileName        string    `json:"profile_name"`
	ProfileExpireDate  time.Time `json:"profile_expire_date"`
	ProfileType        string    `json:"profile_type"`
	ProfileDownloadUrl string    `json:"profile_download_url"`
}

func (AppAccountCert) TableName() string {
	return "tt_app_account_cert"
}

func (AppBoundleProfiles) TableName() string {
	return "tt_app_bundleId_profiles"
}

func (AppleBundleId) TableName() string {
	return "tt_apple_bundleId"
}
func (AppleProfile) TableName() string {
	return "tt_apple_profile"
}
