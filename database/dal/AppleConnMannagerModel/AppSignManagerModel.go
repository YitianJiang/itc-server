package devconnmanager

import "code.byted.org/gopkg/gorm"

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
	BundleidId   string `json:"bundleid_id"`
	BundleidName string `json:"bundleid_name"`
	BundleId     string `json:"bundle_id"`
}

type AppleProfile struct {
	gorm.Model
	ProfileId string `json:"profile_id"`
}

func (AppAccountCert) TableName() string {
	return "tt_app_account_cert"
}

func (AppBoundleProfiles) TableName() string {
	return "tt_app_bundleId_profiles"
}
