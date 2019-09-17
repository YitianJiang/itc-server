package devconnmanager

import (
	"strings"
	"time"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/logs"
)

func QueryExpiredProfileRelatedInfo() (*[]ExpiredProfileCardInput, bool) {
	var expiredProfileInfos []ExpiredProfileInfo
	var expiredProfileCardInputs []ExpiredProfileCardInput
	conn, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return &expiredProfileCardInputs, false
	}
	defer conn.Close()
	nowDate := time.Now().Format("2006-01-02 ")
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(AppleProfile{}.TableName()).
		Where("DATEDIFF(?,profile_expire_date)>=-30", nowDate).
		Find(&expiredProfileInfos).
		Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return &expiredProfileCardInputs, false
	}
	for _, expiredProfileInfo := range expiredProfileInfos {
		condition := map[string]interface{}{}
		if expiredProfileInfo.ProfileType == _const.IOS_APP_STORE || expiredProfileInfo.ProfileType == _const.IOS_APP_INHOUSE || expiredProfileInfo.ProfileType == _const.MAC_APP_STORE {
			condition["dist_profile_id"] = expiredProfileInfo.ProfileId
		} else if expiredProfileInfo.ProfileType == _const.IOS_APP_DEVELOPMENT || expiredProfileInfo.ProfileType == _const.MAC_APP_DEVELOPMENT {
			condition["dev_profile_id"] = expiredProfileInfo.ProfileId
		} else {
			condition["dist_adhoc_profile_id"] = expiredProfileInfo.ProfileId
		}
		var appRelatedInfo AppRelatedInfo
		connOrigin := conn
		if err = conn.LogMode(_const.DB_LOG_MODE).Table(AppBundleProfiles{}.TableName()).
			Where(condition).Find(&appRelatedInfo).
			Error; err != nil {
			logs.Error("Query DB Failed:", err)
			return &expiredProfileCardInputs, false
		}
		conn = connOrigin
		cardExpiredMessageInput := ConbineCardMessafeInputInfo(&expiredProfileInfo, &appRelatedInfo)
		expiredProfileCardInputs = append(expiredProfileCardInputs, *cardExpiredMessageInput)
	}
	return &expiredProfileCardInputs, true
}

func ConbineCardMessafeInputInfo(expiredProfileInfo *ExpiredProfileInfo, appRelatedInfo *AppRelatedInfo) *ExpiredProfileCardInput {
	var expiredProfileCardInput ExpiredProfileCardInput
	expiredProfileCardInput.ProfileId = expiredProfileInfo.ProfileId
	expiredProfileCardInput.ProfileType = expiredProfileInfo.ProfileType
	expiredProfileCardInput.ProfileName = expiredProfileInfo.ProfileName
	expiredProfileCardInput.ProfileExpireDate = expiredProfileInfo.ProfileExpireDate
	expiredProfileCardInput.BundleId = appRelatedInfo.BundleId
	expiredProfileCardInput.AppName = appRelatedInfo.AppName
	expiredProfileCardInput.AppId = appRelatedInfo.AppId
	return &expiredProfileCardInput
}

func QueryExpiredCertRelatedInfo() (*[]ExpiredCertCardInput, bool) {
	var expiredCertCardInputs []ExpiredCertCardInput
	conn, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return &expiredCertCardInputs, false
	}
	defer conn.Close()
	nowDate := time.Now().Format("2006-01-02 ")
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).
		Where("DATEDIFF(?,cert_expire_date)>=-30", nowDate).
		Find(&expiredCertCardInputs).
		Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return &expiredCertCardInputs, false
	}
	var ret []ExpiredCertCardInput
	for _, expiredCertCardInput := range expiredCertCardInputs {
		var affectedApps []AffectedApp
		strs := strings.Split(expiredCertCardInput.CertType, "_")
		connOrigin := conn
		conn = conn.LogMode(_const.DB_LOG_MODE)
		switch strs[len(strs)-1] {
		case _const.PUSH:
			conn = conn.Table(AppBundleProfiles{}.TableName()).Where("push_cert_id=?", expiredCertCardInput.CertId)
		case _const.DEVELOPMENT:
			conn = conn.Table(AppAccountCert{}.TableName()).Where("dev_cert_id=?", expiredCertCardInput.CertId).Where("account_verify_status=1")
		case _const.DISTRIBUTION:
			conn = conn.Table(AppAccountCert{}.TableName()).Where("dist_cert_id=?", expiredCertCardInput.CertId).Where("account_verify_status=1")
		}
		if err := conn.Find(&affectedApps).Error; err != nil {
			logs.Error("Query DB Failed:", err)
			return &ret, false
		}
		conn = connOrigin
		expiredCertCardInput.AffectedApps = affectedApps
		ret = append(ret, expiredCertCardInput)
	}
	return &ret, true
}
