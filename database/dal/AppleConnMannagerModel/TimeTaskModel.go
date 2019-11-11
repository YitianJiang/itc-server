package devconnmanager

import (
	"code.byted.org/gopkg/gorm"
	"time"
)

type ExpiredProfileInfo struct {
	gorm.Model
	ProfileId          string    `gorm:"column:profile_id"`
	ProfileName        string    `gorm:"column:profile_name"`
	ProfileType        string    `gorm:"column:profile_type"`
	ProfileExpireDate  time.Time `gorm:"column:profile_expire_date"`
}

type AppRelatedInfo struct {
	gorm.Model
	AppName          string     `gorm:"column:app_name"`
	BundleId         string     `gorm:"column:bundle_id"`
	AppId            string     `gorm:"column:app_id"`
}

type ExpiredProfileCardInput struct {
	AppName             string
	BundleId            string
	AppId               string
	ProfileId           string
	ProfileName         string
	ProfileType         string
	ProfileExpireDate   time.Time
}

type ExpiredCertCardInput struct {
	gorm.Model
	TeamId              string              `gorm:"column:team_id"`
	AccountName         string              `gorm:"column:account_name"`
	CertType            string              `gorm:"column:cert_type"`
	CertName            string              `gorm:"column:cert_name"`
	CertId              string              `gorm:"column:cert_id"`
	AffectedApps        []AffectedApp       `gorm:"-"`
	CertExpireDate      time.Time           `gorm:"column:cert_expire_date"`
}

type AffectedApp struct {
	gorm.Model
	AppName          string `gorm:"column:app_name"`
}

type AppIdType struct {
	AppId              string              `gorm:"column:app_id"`
}