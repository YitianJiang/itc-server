package devconnmanager

import "code.byted.org/gopkg/gorm"

type ExpiredProfileInfo struct {
	gorm.Model
	ProfileId          string `gorm:"column:profile_id"`
	ProfileName        string `gorm:"column:profile_name"`
	ProfileType        string `gorm:"column:profile_type"`
}

type AppRelatedInfo struct {
	gorm.Model
	AppName          string     `gorm:"column:app_name"`
	BundleId         string     `gorm:"column:bundle_id"`
	UserName         string     `gorm:"column:user_name"`
}

type ExpiredProfileCardInput struct {
	AppName          string
	BundleId         string
	ProfileId        string
	ProfileName      string
	ProfileType      string
	UserName         string
}

type ExpiredCertCardInput struct {
	gorm.Model
	TeamId              string `gorm:"column:team_id"`
	AccountName         string `gorm:"column:account_name"`
	CertType            string `gorm:"column:cert_type"`
	CertName            string `gorm:"column:cert_name"`
	CertId              string `gorm:"column:cert_id"`
	AppAndPrincipals    []AppAndPrincipal
}

type AppAndPrincipal struct {
	gorm.Model
	AppName          string `gorm:"column:app_name"`
	UserName         string `gorm:"column:user_name"`
}