package devconnmanager

import (
	"code.byted.org/gopkg/gorm"
	"time"
)

//http request model
type UpdateBundleIdIdRequest struct {
	BundleId   string `json:"bundle_id" binding:"required"`
	BundleIdId string `json:"bundleid_id" binding:"required"`
}

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
	AccountName     string `json:"account_name"   binding:"required"`
	AccountType     string `json:"account_type"   binding:"required"`
	TeamId          string `json:"team_id"        binding:"required"`
	BundleId        string `json:"bundle_id"      binding:"required"`
	BundleidId      string `json:"bundleid_id"    binding:"exists"`
	UseCertId       string `json:"use_cert_id"    binding:"required"`
	ProfileId       string `json:"profile_id"     binding:"exists"`
	ProfileName     string `json:"profile_name"   binding:"required"`
	ProfileType     string `json:"profile_type"   binding:"required"`
	UserName        string `json:"user_name"      binding:"required"`
	BundlePrincipal string `json:"principal"`
}

type ProfileUploadRequest struct {
	TeamId      string `form:"team_id"        binding:"required"`
	BundleId    string `form:"bundle_id"      binding:"required"`
	ProfileId   string `form:"profile_id"     binding:"required"`
	ProfileName string `form:"profile_name"   binding:"required"`
	ProfileType string `form:"profile_type"   binding:"required"`
	UserName    string `form:"user_name"      binding:"required"`
}
type ProfileDeleteRequest struct {
	ProfileId   string `form:"profile_id"     binding:"required"`
	UserName    string `form:"user_name"      binding:"required"`
	ProfileType string `form:"profile_type"   binding:"required"`
	TeamId      string `form:"team_id"        binding:"required"`
	AccountName string `form:"account_name"`
	AccountType string `form:"account_type"`
	Operator    string `form:"principal"`
}
type BundleDeleteRequest struct {
	DevProfileId   string `form:"dev_profile_id"`
	DisProfileId   string `form:"dist_profile_id"`
	UserName       string `form:"user_name"      binding:"required"`
	IsDel          string `form:"is_del"         binding:"required"`
	TeamId         string `form:"team_id"        binding:"required"`
	BundleId       string `form:"bundle_id"      binding:"required"`
	BundleidId     string `form:"bundleid_id"`
	AccountName    string `form:"account_name"`
	AccountType    string `form:"account_type"`
	Operator       string `form:"principal"`
	DevProfileName string `form:"dev_profile_name"`
	DisProfileName string `form:"dist_profile_name"`
}

//Profile删除反馈参数struct
type DelProfileFeedback struct {
	CustomerJson DelProfileFeedbackCustomer `json:"customer_parameter"`
}

type DelProfileFeedbackCustomer struct {
	ProfileId string `json:"profile_id"        binding:"required"`
	UserName  string `json:"username"       binding:"required"`
}

//Profile删除反馈参数struct
type DelBundleFeedback struct {
	CustomerJson DelBundleFeedbackCustomer `json:"customer_parameter"`
}

type DelBundleFeedbackCustomer struct {
	BundleIdId    string `json:"bundleid_id"        binding:"required"`
	UserName      string `json:"username"           binding:"required"`
	IsDel         string `json:"is_del"             binding:"required"`
	DistProfileId string `json:"dist_profile_id"`
	DevProfileId  string `json:"dev_profile_id"`
	PushCertId    string `json:"push_cert_id"`
}

type CreateBundleProfileRequest struct {
	AccountType     string `json:"account_type"   binding:"required"`
	TeamId          string `json:"team_id"        binding:"required"`
	AppId           string `json:"app_id"         binding:"required"`
	AppName         string `json:"app_name"       binding:"required"`
	BundlePrincipal string `json:"principal"`
	BundleIdInfo
	UserName                  string            `json:"user_name"      binding:"required"`
	DevProfileInfo            ProfileInfo       `json:"dev_info_data"`
	DistProfileInfo           ProfileInfo       `json:"dist_info_data"`
	EnableCapabilitiesChange  []string          `json:"enable_capabilities_change"`
	DisableCapabilitiesChange []string          `json:"disable_capabilities_change"`
	ConfigCapabilitiesChange  map[string]string `json:"config_capabilities_change"`
}

type BundleIdInfo struct {
	BundleIdName  string `json:"bundleid_name"      binding:"required"`
	BundleId      string `json:"bundle_id"      binding:"required"`
	BundleType    string `json:"bundleid_type"      binding:"required"`
	BundleIdIsDel string `json:"bundleid_isdel"`
	BundleIdId    string `json:"bundleid_id"`
}

type ProfileInfo struct {
	ProfileId   string `json:"profile_id"`
	ProfileName string `json:"profile_name"`
	ProfileType string `json:"profile_type"`
	CertId      string `json:"cert_id"`
}

//和苹果req res的Model

//Bundle id的苹果req的Model ******Start******
type CreateBundleIdRequest struct {
	Data BundleIdDataObj `json:"data"       binding:"required"`
}

type BundleIdDataObj struct {
	Type       string             `json:"type"       binding:"required"`
	Attributes BundleIdAttributes `json:"attributes" binding:"required"`
}

type BundleIdAttributes struct {
	Identifier string `json:"identifier" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Platform   string `json:"platform" binding:"required"`
}

type CreateBundleIdResponse struct {
	Data BundleData `json:"data"`
}

type BundleData struct {
	IdAndTypeItem
	Relationships BundleIdRelationships `json:"relationships"`
}

type BundleIdRelationships struct {
	BundleIdCapabilities BundleIdCapabilities `json:"bundleIdCapabilities"`
}

type BundleIdCapabilities struct {
	Data []IdAndTypeItem `json:"data"`
}

type OpenBundleIdCapabilityRequest struct {
	Data BundleIdCapabilityReqObj `json:"data"`
}

type BundleIdCapabilityReqObj struct {
	Type          string                       `json:"type"`
	Attributes    BundleCapabilityAttributes   `json:"attributes"`
	Relationships BundleCapabilityRelationship `json:"relationships" binding:"required"`
}

type BundleCapabilityRelationship struct {
	BundleId DataIdAndTypeItemObj `json:"bundleId" binding:"required"`
}

type BundleCapabilityAttributes struct {
	CapabilityType string    `json:"capabilityType" binding:"required"`
	Settings       []Setting `json:"settings"`
}

type Setting struct {
	ConfigKey
	Options []ConfigKey `json:"options"`
}

type ConfigKey struct {
	Key string `json:"key"`
}

type OpenBundleIdCapabilityResponse struct {
	Data IdAndTypeItem `json:"data"`
}

//Bundle id的苹果req的Model ******End******

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
	BundleId     DataIdAndTypeItemObj   `json:"bundleId" binding:"required"`
	Certificates DataIdAndTypeItemList  `json:"certificates" binding:"required"`
	Devices      *DataIdAndTypeItemList `json:"devices,omitempty"`
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
//Devices的苹果res的Model ******Start******
type DevicesAttributesRes struct {
	Name        string `json:"name"           binding:"required"`
	AddedDate   string `json:"addedDate"      binding:"required"`
	DeviceClass string `json:"deviceClass"    binding:"required"`
	Model       string `json:"model"          binding:"required"`
	Udid        string `json:"udid"           binding:"required"`
	Platform    string `json:"platform"       binding:"required"`
	Status      string `json:"status"         binding:"required"`
}

type DevicesDataObjRes struct {
	Id         string               `json:"id"         binding:"required"`
	Type       string               `json:"type"       binding:"required"`
	Attributes DevicesAttributesRes `json:"attributes" binding:"required"`
}

type DevicesDataRes struct {
	Data []DevicesDataObjRes `json:"data"       binding:"required"`
}

//Devices的苹果res的Model ******End******

type ApproveAppBindAccountParamFromLark struct {
	ApproveAppBindAccountCustomerParam `json:"customer_parameter"`
}

type ApproveAppBindAccountCustomerParam struct {
	AppAccountCertId uint   `json:"appAccountCertId" binding:"required"`
	IsApproved       int    `json:"isApproved"       binding:"required"`
	UserName         string `json:"userName"       binding:"required"`
}

type AppSignListRequest struct {
	AppId    string `form:"app_id"   binding:"required"`
	Username string `form:"user_name" binding:"required"`
	TeamId   string `form:"team_id"  binding:"required"`
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
	AppId              string `gorm:"column:app_id"                    json:"app_id"`
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
	ProfileId          string    `gorm:"column:profile_id"              json:"profile_id"`
	ProfileName        string    `gorm:"column:profile_name"            json:"profile_name"`
	ProfileExpireDate  time.Time `gorm:"column:profile_expire_date"     json:"profile_expire_date"`
	ProfileType        string    `gorm:"column:profile_type"            json:"profile_type"`
	ProfileDownloadUrl string    `gorm:"column:profile_download_url"    json:"profile_download_url"`
	OpUser             string    `gorm:"column:op_user"                 json:"op_user"`
}

func (AppAccountCert) TableName() string {
	return "tt_app_account_cert"
}

func (AppBundleProfiles) TableName() string {
	return "tt_app_bundleId_profiles"
}

//todo 线上该dbname为"tt_apple_bundleid"
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
	BundleIdName       string             `json:"bundleid_name"`
	BundleIdIsDel      string             `json:"bundleid_isdel"`
	BundleIdId         string             `json:"bundleid_id"`
	BundleIdType       string             `json:"bundleid_type"`
	EnableCapList      []string           `json:"enable_capabilities_list"`
	ConfigCapObj       map[string]string  `json:"config_capabilities_obj"`
	ProfileCertSection BundleProfileGroup `json:"profile_cert_section"`
	PushCert           AppCertInfo        `json:"push_cert"`
}

type BundleProfileGroup struct {
	DistProfile BundleProfileInfo `json:"dist_profile"`
	DevProfile  BundleProfileInfo `json:"dev_profile"`
}

type BundleProfileInfo struct {
	UserCertId         string     `json:"use_cert_id"`
	ProfileId          string     `json:"profile_id"`
	ProfileName        string     `json:"profile_name"`
	ProfileType        string     `json:"profile_type"`
	ProfileExpireDate  *time.Time `json:"profile_expire_date"`
	ProfileDownloadUrl string     `json:"profile_download_url"`
}

type AppCertGroupInfo struct {
	DistCert AppCertInfo `json:"dist_cert"`
	DevCert  AppCertInfo `json:"dev_cert"`
}

type AppCertInfo struct {
	CertId          string     `json:"cert_id"`
	CertName        string     `json:"cert_name"`
	CertType        string     `json:"cert_type"`
	CertExpireDate  *time.Time `json:"cert_expire_date"`
	CertDownloadUrl string     `json:"cert_download_url"`
	PrivKeyUrl      string     `json:"priv_key_url"`
}

/**
数据库联查struct
*/
//app信息，账号信息和证书信息
type APPandCert struct {
	AppName             string    `json:"app_name"`
	AppType             string    `json:"app_type"`
	AppAcountId         string    `json:"app_account_id"`
	TeamId              string    `json:"team_id"`
	AccountType         string    `json:"account_type"`
	AccountVerifyStatus string    `json:"account_verify_status"`
	AccountVerifyUser   string    `json:"account_verify_user"`
	CertId              string    `json:"cert_id"`
	CertName            string    `json:"cert_name"`
	CertType            string    `json:"cert_type"`
	CertExpireDate      time.Time `json:"cert_expire_date"`
	CertDownloadUrl     string    `json:"cert_download_url"`
	PrivKeyUrl          string    `json:"priv_key_url"`
}

//appname、bundle信息和profile信息
type APPandBundle struct {
	AppName       string `json:"app_name"`
	BundleIdIndex string `json:"bundle_id_index"`
	BundleIdIsDel string `json:"bundleid_isdel"`
	AppleBundleId
	PushCertId         string    `json:"push_cert_id"`
	ProfileId          string    `json:"profile_id"`
	ProfileName        string    `json:"profile_name"`
	ProfileType        string    `json:"profile_type"`
	ProfileExpireDate  time.Time `json:"profile_expire_date"`
	ProfileDownloadUrl string    `json:"profile_download_url"`
}
