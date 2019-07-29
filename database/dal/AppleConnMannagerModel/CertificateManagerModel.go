package devconnmanager

import "code.byted.org/gopkg/gorm"

type CertInfo struct {
	gorm.Model
	AccountName         string      `gorm:"column:account_name"            json:"account_name,omitempty"`
	TeamId              string      `gorm:"column:team_id"                 json:"team_id,omitempty"             form:"team_id"    binding:"required"`
	CertName            string      `gorm:"column:cert_name"               json:"cert_name"`
	CertId              string      `gorm:"column:cert_id"                 json:"cert_id"                       form:"cert_id"    binding:"required"`
	CertExpireDate      string      `gorm:"column:cert_expire_date"        json:"cert_expire_date"`
	CertType            string      `gorm:"column:cert_type"               json:"cert_type"                     form:"cert_type"  binding:"required"`
	CertDownloadUrl     string      `gorm:"column:cert_download_url"       json:"cert_download_url"`
	PrivKeyUrl          string      `gorm:"column:priv_key_url"            json:"priv_key_url"`
	CsrFileUrl          string      `gorm:"column:csr_file_url"            json:"csr_file_url,omitempty"`
	EffectAppList       []string    `gorm:"-"                              json:"effect_app_list,omitempty"`
}

type CreCertResponse struct {
	Data                OutLayer        `json:"data"`
	Links               Links           `json:"links"`
}

type OutLayer struct{
	Type                string              `json:"type"`
	Id                  string              `json:"id"`
	Attributes          Attributes          `json:"attributes"`
	Links               Links               `json:"links"`
}

type Links struct {
	Self string     `json:"self"`
}

type Attributes struct{
	SerialNumber        string      `json:"serialNumber"`
	CertificateContent  string      `json:"certificateContent"`
	DisplayName         string      `json:"displayName"`
	Name                string      `json:"name"`
	CsrContent          string      `json:"csrContent"`
	Platform            string      `json:"platform"`
	ExpirationDate      string      `json:"expirationDate"`
	CertificateType     string      `json:"certificateType"`
}

type InsertCertRequest struct {
	AccountName        string      `json:"account_name" binding:"required"`
	TeamId             string      `json:"team_id"      binding:"required"`
	CertName           string      `json:"cert_name"    binding:"required"`
	CertType           string      `json:"cert_type"    binding:"required"`
}

type CreAppleCertReq struct {
	Data Data       `json:"data"`
}

type Data struct {
	Type        string          `json:"type"`
	Attributes  AttributesSend  `json:"attributes"`
}

type AttributesSend struct {
	CsrContent          string      `json:"csrContent"`
	CertificateType     string      `json:"certificateType"`
}

type GetPermsResponse struct {
	Data map[string][]string    `json:"data"`
	Errno   int                 `json:"errno"`
	Message string              `json:"message"`
}

type DelCertRequest struct {
	CertId   string  `form:"cert_id"        binding:"required"`
	TeamId   string  `form:"team_id"        binding:"required"`
	CertType string  `form:"cert_type"      binding:"required"`
}

type QueryCertRequest struct {
	TeamId         string      `form:"team_id"      json:"team_id"`
	ExpireSoon     string      `form:"expire_soon"  json:"expire_soon"`
	UserName       string      `form:"user_name"    json:"user_name"`
}

type UserName struct {
	UserName string `gorm:"user_name"`
}

type RecAppName struct {
	AppName string  `gorm:"app_name"`
}

type CertName struct {
	CertName string `gorm:"certName"`
}

func (CertInfo) TableName() string{
	return  "tt_apple_certificate"
}