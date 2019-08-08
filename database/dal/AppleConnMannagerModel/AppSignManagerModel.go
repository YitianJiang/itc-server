package devconnmanager

import (
	"code.byted.org/gopkg/gorm"
	"time"
)

//http request model
type CreateAppBindAccountRequest struct {
	AppId    string `json:"app_id"    binding:"required"`
	AppName  string `json:"app_name"  binding:"required"`
	AppType  string `json:"app_type"  binding:"required"`
	UserName string `json:"user_name" binding:"required"`
	TeamId   string `json:"team_id"   binding:"required"`
}

type DeleteAppAllInfoRequest struct {
	AppId   string `form:"app_id"   binding:"required"`
	AppName string `form:"app_name" binding:"required"`
}

type AppChangeBindCertRequest struct {
	AccountCertId int    `json:"account_cert_id" binding:"required"`
	CertId        string `json:"cert_id"         binding:"required"`
	CertType      string `json:"cert_type"       binding:"required"`
	UserName      string `json:"user_name"       binding:"required"`
}

type ProfileCreateOrUpdateRequest struct {
	AccountName string `json:"account_name"   binding:"required"`
	AccountType string `json:"account_type"   binding:"required"`
	TeamId      string `json:"team_id"        binding:"required"`
	BundleId    string `json:"bundle_id"      binding:"required"`
	BundleidId  string `json:"bundleid_id"    binding:"exists"`
	UseCertId   string `json:"use_cert_id"    binding:"required"`
	ProfileId   string `json:"profile_id"     binding:"exists"`
	ProfileName string `json:"profile_name"   binding:"required"`
	ProfileType string `json:"profile_type"   binding:"required"`
	UserName    string `json:"user_name"      binding:"required"`
}

type ProfileUploadRequest struct {
	TeamId      string    `form:"team_id"        binding:"required"`
	BundleId    string    `form:"bundle_id"      binding:"required"`
	ProfileId   string    `form:"profile_id"     binding:"required"`
	ProfileName string    `form:"profile_name"   binding:"required"`
	ProfileType string    `form:"profile_type"   binding:"required"`
	UserName    string    `form:"user_name"      binding:"required"`
}

//和苹果req res的Model

//Profile的苹果req的Model ******Start******
type IdAndTypeItem struct {
	Type string `json:"type"   binding:"required"`
	Id   string `json:"id"     binding:"required"`
}

type DataIdAndTypeItemList struct {
	Data []IdAndTypeItem `json:"data" binding:"required"`
}

type DataIdAndTypeItemObj struct {
	Data IdAndTypeItem `json:"data" binding:"required"`
}

type ProfileRelationShipSec struct {
	BundleId     DataIdAndTypeItemObj  `json:"bundleId" binding:"required"`
	Certificates DataIdAndTypeItemList `json:"certificates" binding:"required"`
}

type ProfileAttributes struct {
	Name        string `json:"name" binding:"required"`
	ProfileType string `json:"profileType" binding:"required"`
}

type ProfileDataObj struct {
	Type          string                 `json:"type"       binding:"required"`
	Attributes    ProfileAttributes      `json:"attributes" binding:"required"`
	Relationships ProfileRelationShipSec `json:"relationships" binding:"required"`
}

type ProfileDataReq struct {
	Data ProfileDataObj `json:"data"       binding:"required"`
}

//Profile的苹果req的Model ******End******
//Profile的苹果res的Model ******Start******
type ProfileAttributesRes struct {
	Name           string `json:"name"           binding:"required"`
	ProfileContent string `json:"profileContent" binding:"required"`
	ExpirationDate string `json:"expirationDate" binding:"required"`
}

type ProfileDataObjRes struct {
	Id         string               `json:"id"         binding:"required"`
	Attributes ProfileAttributesRes `json:"attributes" binding:"required"`
}

type ProfileDataRes struct {
	Data ProfileDataObjRes `json:"data"       binding:"required"`
}

//Profile的苹果res的Model ******End******

type ApproveAppBindAccountParamFromLark struct {
	ApproveAppBindAccountCustomerParam `json:"customer_parameter"`
}

type ApproveAppBindAccountCustomerParam struct {
	AppAccountCertId uint `json:"appAccountCertId" binding:"required"`
	IsApproved       int  `json:"isApproved"       binding:"required"`
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

type AppBundleProfiles struct {
	gorm.Model
	AppId              string `gorm:"app_id"                           json:"app_id"`
	AppName            string `gorm:"column:app_name"                  json:"app_name"`
	BundleId           string `gorm:"column:bundle_id"                 json:"bundle_id"`
	BundleidId         string `gorm:"column:bundleid_id"               json:"bundleid_id"`
	BundleidIsdel      string `gorm:"column:bundleid_isdel"            json:"bundleid_isdel"`
	PushCertId         string `gorm:"column:push_cert_id"              json:"push_cert_id"`
	DevProfileId       string `gorm:"column:dev_profile_id"            json:"dev_profile_id"`
	UserName           string `gorm:"column:user_name"                 json:"user_name"`
	DistAdhocProfileId string `gorm:"column:dist_adhoc_profile_id"     json:"dist_adhoc_profile_id"`
	DistProfileId      string `gorm:"column:dist_profile_id"           json:"dist_profile_id"`
}

type AppleBundleId struct {
	gorm.Model
	BundleidId                       string `gorm:"column:bundleid_id"                         json:"bundleid_id"`
	BundleidName                     string `gorm:"column:bundleid_name"                       json:"bundleid_name"`
	BundleId                         string `gorm:"column:bundle_id"                           json:"bundle_id"`
	BundleidType                     string `gorm:"column:bundleid_type"                       json:"bundleid_type"`
	ICLOUD                           string `gorm:"column:ICLOUD"                              json:"ICLOUD"`
	IN_APP_PURCHASE                  string `gorm:"column:IN_APP_PURCHASE"                     json:"IN_APP_PURCHASE"`
	GAME_CENTER                      string `gorm:"column:GAME_CENTER"                         json:"GAME_CENTER"`
	PUSH_NOTIFICATIONS               string `gorm:"column:PUSH_NOTIFICATIONS"                  json:"PUSH_NOTIFICATIONS"`
	WALLET                           string `gorm:"column:WALLET"                              json:"WALLET"`
	INTER_APP_AUDIO                  string `gorm:"column:INTER_APP_AUDIO"                     json:"INTER_APP_AUDIO"`
	MAPS                             string `gorm:"column:MAPS"                                json:"MAPS"`
	ASSOCIATED_DOMAINS               string `gorm:"column:ASSOCIATED_DOMAINS"                  json:"ASSOCIATED_DOMAINS"`
	PERSONAL_VPN                     string `gorm:"column:PERSONAL_VPN"                        json:"PERSONAL_VPN"`
	APP_GROUPS                       string `gorm:"column:APP_GROUPS"                          json:"APP_GROUPS"`
	HEALTHKIT                        string `gorm:"column:HEALTHKIT"                           json:"HEALTHKIT"`
	HOMEKIT                          string `gorm:"column:HOMEKIT"                             json:"HOMEKIT"`
	WIRELESS_ACCESSORY_CONFIGURATION string `gorm:"column:WIRELESS_ACCESSORY_CONFIGURATION"    json:"WIRELESS_ACCESSORY_CONFIGURATION"`
	APPLE_PAY                        string `gorm:"column:APPLE_PAY"                           json:"APPLE_PAY"`
	DATA_PROTECTION                  string `gorm:"column:DATA_PROTECTION"                     json:"DATA_PROTECTION"`
	SIRIKIT                          string `gorm:"column:SIRIKIT"                             json:"SIRIKIT"`
	NETWORK_EXTENSIONS               string `gorm:"column:NETWORK_EXTENSIONS"                  json:"NETWORK_EXTENSIONS"`
	MULTIPATH                        string `gorm:"column:MULTIPATH"                           json:"MULTIPATH"`
	HOT_SPOT                         string `gorm:"column:HOT_SPOT"                            json:"HOT_SPOT"`
	NFC_TAG_READING                  string `gorm:"column:NFC_TAG_READING"                     json:"NFC_TAG_READING"`
	CLASSKIT                         string `gorm:"column:CLASSKIT"                            json:"CLASSKIT"`
	AUTOFILL_CREDENTIAL_PROVIDER     string `gorm:"column:AUTOFILL_CREDENTIAL_PROVIDER"        json:"AUTOFILL_CREDENTIAL_PROVIDER"`
	ACCESS_WIFI_INFORMATION          string `gorm:"column:ACCESS_WIFI_INFORMATION"             json:"ACCESS_WIFI_INFORMATION"`
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

func (AppBundleProfiles) TableName() string {
	return "tt_app_bundleId_profiles"
}

func (AppleBundleId) TableName() string {
	return "tt_apple_bundleId"
}
func (AppleProfile) TableName() string {
	return "tt_apple_profile"
}

/**
单个app_name的App签名管理信息
*/
type APPSignManagerInfo struct {
	AppName                  string              `json:"app_name"`
	AppType                  string              `json:"app_type"`
	AppAcountId              string              `json:"app_account_id"`
	TeamId                   string              `json:"team_id"`
	AccountType              string              `json:"account_type"`
	AccountVerifyStatus      string              `json:"account_verify_status"`
	AccountVerifyUser        string              `json:"account_verify_user"`
	BundleProfileCertSection []BundleProfileCert `json:"bundle_profile_cert_section"`
	CertSection              AppCertGroupInfo    `json:"cert_section"`
}
type BundleProfileCert struct {
	BoundleId          string             `json:"bundle_id"`
	BundleIdIsDel      string             `json:"bundleid_isdel"`
	BundleIdId         string             `json:"bundleid_id"`
	BundleIdType       string             `json:"bundleid_type"`
	EnableCapList      []string           `json:"enable_capabilities_list"`
	ConfigCapObj       BundleConfigCap    `json:"config_capabilities_obj"`
	ProfileCertSection BundleProfileGroup `json:"profile_cert_section"`
}
type BundleConfigCap struct {
	ICLOUD         BundleConfigCapInfo `json:"ICLOUD"`
	DataProtection BundleConfigCapInfo `json:"DATA_PROTECTION"`
}
type BundleConfigCapInfo struct {
	KeyInfo string         `json:"key"`
	Options []OptionStruct `json:"options"`
}

type OptionStruct struct {
	Key string `json:"key"`
}

type BundleProfileGroup struct {
	DistProfile BundleProfileInfo `json:"dist_profile"`
	DevProfile  BundleProfileInfo `json:"dev_profile"`
}

type BundleProfileInfo struct {
	UserCertId         string    `json:"user_cert_id"`
	ProfileId          string    `json:"profile_id"`
	ProfileName        string    `json:"profile_name"`
	ProfileType        string    `json:"profile_type"`
	ProfileExpireDate  time.Time `json:"profile_expire_date"`
	ProfileDownloadUrl string    `json:"profile_download_url"`
}

type AppCertGroupInfo struct {
	DistCert AppCertInfo `json:"dist_cert"`
	DevCert  AppCertInfo `json:"dev_cert"`
	PushCert AppCertInfo `json:"push_cert"`
}

type AppCertInfo struct {
	CertId          string    `json:"cert_id"`
	CertType        string    `json:"cert_type"`
	CertExpireDate  time.Time `json:"cert_expire_date"`
	CertDownloadUrl string    `json:"cert_download_url"`
	PrivKeyUrl      string    `json:"priv_key_url"`
}

/**
数据库联查struct
*/

type APPandCert struct {
	AppName             string    `json:"app_name"`
	AppType             string    `json:"app_type"`
	AppAcountId         string    `json:"app_account_id"`
	TeamId              string    `json:"team_id"`
	AccountType         string    `json:"account_type"`
	AccountVerifyStatus string    `json:"account_verify_status"`
	AccountVerifyUser   string    `json:"account_verify_user"`
	CertId              string    `json:"cert_id"`
	CertType            string    `json:"cert_type"`
	CertExpireDate      time.Time `json:"cert_expire_date"`
	CertDownloadUrl     string    `json:"cert_download_url"`
	PrivKeyUrl          string    `json:"priv_key_url"`
}
