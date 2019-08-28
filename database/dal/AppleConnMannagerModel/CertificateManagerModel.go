package devconnmanager

import "code.byted.org/gopkg/gorm"

type CertInfo struct {
	gorm.Model
	AccountName     string   `gorm:"column:account_name"            json:"account_name,omitempty"`
	TeamId          string   `gorm:"column:team_id"                 json:"team_id,omitempty"             form:"team_id"    binding:"required"`
	CertName        string   `gorm:"column:cert_name"               json:"cert_name"`
	CertId          string   `gorm:"column:cert_id"                 json:"cert_id"                       form:"cert_id"    binding:"required"`
	CertExpireDate  string   `gorm:"column:cert_expire_date"        json:"cert_expire_date"`
	CertType        string   `gorm:"column:cert_type"               json:"cert_type"                     form:"cert_type"  binding:"required"`
	CertDownloadUrl string   `gorm:"column:cert_download_url"       json:"cert_download_url"`
	PrivKeyUrl      string   `gorm:"column:priv_key_url"            json:"priv_key_url"`
	CsrFileUrl      string   `gorm:"column:csr_file_url"            json:"csr_file_url,omitempty"`
	EffectAppList   []string `gorm:"-"                              json:"effect_app_list,omitempty"`
	OpUser          string   `gorm:"column:op_user"                 json:"op_user"`
}

type UploadCertRequest struct {
	TeamId      string `form:"team_id"       binding:"required"`
	CertName    string `form:"cert_name"`
	CertId      string `form:"cert_id"       binding:"required"`
	CertType    string `form:"cert_type"     binding:"required"`
	Id          string `form:"ID"`
	UserName    string `form:"user_name"     binding:"required"`
	AccountName string `form:"account_name"     binding:"required"`
	BundleId    string `form:"bundle_id"`
}

type CreCertResponse struct {
	Data  OutLayer `json:"data"`
	Links Links    `json:"links"`
}

type OutLayer struct {
	Type       string     `json:"type"`
	Id         string     `json:"id"`
	Attributes Attributes `json:"attributes"`
	Links      Links      `json:"links"`
}

type Links struct {
	Self string `json:"self"`
}

type Attributes struct {
	SerialNumber       string `json:"serialNumber"`
	CertificateContent string `json:"certificateContent"`
	DisplayName        string `json:"displayName"`
	Name               string `json:"name"`
	CsrContent         string `json:"csrContent"`
	Platform           string `json:"platform"`
	ExpirationDate     string `json:"expirationDate"`
	CertificateType    string `json:"certificateType"`
}

type InsertCertRequest struct {
	AccountName   string `json:"account_name" binding:"required"`
	TeamId        string `json:"team_id"      binding:"required"`
	CertName      string `json:"cert_name"`
	CertType      string `json:"cert_type"    binding:"required"`
	CertPrincipal string `json:"cert_principal"`
	AccountType   string `json:"account_type" binding:"required"`
	BundleId      string `json:"bundle_id"`
	UserName      string `json:"user_name"`
	IsUpdate      string `json:"is_update"`
}

type CreAppleCertReq struct {
	Data Data `json:"data"`
}

type Data struct {
	Type       string         `json:"type"`
	Attributes AttributesSend `json:"attributes"`
}

type AttributesSend struct {
	CsrContent      string `json:"csrContent"`
	CertificateType string `json:"certificateType"`
}

type GetPermsResponse struct {
	Data    map[string][]string `json:"data"`
	Errno   int                 `json:"errno"`
	Message string              `json:"message"`
}

type DelCertRequest struct {
	ID           string `form:"ID"`
	CertId       string `form:"cert_id"`
	CertName     string `form:"cert_name"`
	TeamId       string `form:"team_id"        binding:"required"`
	CertType     string `form:"cert_type"      binding:"required"`
	AccountName  string `form:"account_name"   binding:"required"`
	UserName     string `form:"username"       binding:"required"`
	AccType      string `form:"account_type"`
	BundleId     string `form:"bundle_id"`
	CertOperator string `form:"cert_principal"`
}
type DelCertFeedback struct {
	CustomerJson        DelCertFeedbackCustomer `json:"customer_parameter"`
	AdditionalParameter `json:"additional_parameter"`
}

type DelCertFeedbackCustomer struct {
	CertId      string `json:"cert_id"        binding:"required"`
	UserName    string `json:"username"       binding:"required"`
	Bundleid    string `json:"bundle_id"`
	BundleIdId  string `json:"bundleid_id"`
	AccountType string `json:"account_type"`
	TeamId      string `json:"team_id"`
}

type QueryCertRequest struct {
	TeamId     string `form:"team_id"      json:"team_id"`
	ExpireSoon string `form:"expire_soon"  json:"expire_soon"`
	UserName   string `form:"user_name"    json:"user_name"`
}

type UserName struct {
	UserName string `gorm:"user_name"`
}

type RecAppName struct {
	gorm.Model
	AppName string `gorm:"app_name"`
}

func (CertInfo) TableName() string {
	return "tt_apple_certificate"
}

type CreateOrUpdateCertInfoForLark struct {
	//CertName     string
	CertType    string
	TeamId      string
	AccountType string
	CsrUrl      string
	BundleId    string
	UserName    string
}
