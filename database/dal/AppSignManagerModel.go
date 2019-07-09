package dal

import "code.byted.org/gopkg/gorm"

type AppAccountCert struct {
	gorm.Model
	AppId                   string `gorm:"app_id"                           json:"app_id"`
	AppName                 string `gorm:"column:app_name"                  json:"app_name"`
	AppType                 string `gorm:"column:app_type"                  json:"app_type"`
	UserName                string `gorm:"column:user_name"                 json:"user_name"`
	TeamId                  string `gorm:"column:team_id"                   json:"team_id"`
	AccountVerifyStatus     string `gorm:"column:account_verify_status"     json:"account_verify_status"`
	AccountVerifyUser       string `gorm:"column:account_verify_user"       json:"account_verify_user"`
	CertId                  string `gorm:"column:cert_id"                   json:"cert_id"`
}

func (AppAccountCert) TableName()string{
	return "app_account_cert"
}