package devconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"strconv"
	"time"
)

func DeleteCertInfo(condition map[string]interface{}) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	if err = connection.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Where(condition).Delete(&CertInfo{}).Error; err != nil {
		logs.Error("Delete Record Failed")
		return false
	}
	return true
}

func InsertCertInfo(CertInfo *CertInfo) bool {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return false
	}
	defer conn.Close()
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo.TableName()).Create(CertInfo).Error; err != nil {
		logs.Error("Insert DB Failed:", err)
		return false
	}
	return true
}

func QueryCertInfo(condition map[string]interface{}, expireSoon string, permsResult int) *[]CertInfo {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var certInfos []CertInfo
	if permsResult == 2 {
		if err = conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).
			Where(condition).Where("cert_type =? or cert_type=?", _const.CERT_TYPE_IOS_DEV, _const.CERT_TYPE_MAC_DEV).Find(&certInfos).
			Error; err != nil {
			logs.Error("Query DB Failed:", err)
			return nil
		}
	}
	if permsResult == 1 {
		if err = conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).
			Where(condition).Find(&certInfos).
			Error; err != nil {
			logs.Error("Query DB Failed:", err)
			return nil
		}
	}
	var ret []CertInfo
	conn.Begin()
	for i := 0; i < len(certInfos); i++ {
		var recAppNames []RecAppName
		recAppNames = GetAppNamesByCertId(conn, certInfos[i].CertType, certInfos[i].CertId)
		for _, recAppName := range recAppNames {
			certInfos[i].EffectAppList = append(certInfos[i].EffectAppList, recAppName.AppName)
		}
		certInfos[i].TeamId = ""
		certInfos[i].AccountName = ""
		certInfos[i].CsrFileUrl = ""
		if expireSoon == "1" && isExpired(certInfos[i]) == true {
			ret = append(ret, certInfos[i])
		}
		if expireSoon == "0" || expireSoon == "" {
			ret = append(ret, certInfos[i])
		}
	}
	conn.Commit()
	return &ret
}

func GetAppNamesByCertId(conn *gorm.DB, certType string, certId string) []RecAppName {
	var recAppNames []RecAppName
	if certType == _const.CERT_TYPE_IOS_DEV || certType == _const.CERT_TYPE_MAC_DEV {
		if err := conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).
			Where("dev_cert_id=?", certId).Find(&recAppNames).
			Error; err != nil {
			logs.Error("Query DB Failed:", err)
			conn.Rollback()
			return nil
		}
	}
	if certType == _const.CERT_TYPE_IOS_DIST || certType == _const.CERT_TYPE_MAC_DIST {
		if err := conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).
			Where("dist_cert_id=?", certId).Find(&recAppNames).
			Error; err != nil {
			logs.Error("Query DB Failed:", err)
			conn.Rollback()
			return nil
		}
	}
	return recAppNames
}

func QueryExpiredCertInfos() *[]CertInfo {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var certInfos []CertInfo
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Find(&certInfos).Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return nil
	}
	var expiredCertInfos []CertInfo
	var recAppNames []RecAppName
	conn.Begin()
	for _, certInfo := range certInfos {
		if isExpired(certInfo) == true {
			recAppNames = GetAppNamesByCertId(conn, certInfo.CertType, certInfo.CertId)
			if len(recAppNames) == 0 {
				continue
			}
			for _, recAppName := range recAppNames {
				certInfo.EffectAppList = append(certInfo.EffectAppList, recAppName.AppName)
			}
			certInfo.TeamId = ""
			certInfo.AccountName = ""
			certInfo.CsrFileUrl = ""
			expiredCertInfos = append(expiredCertInfos, certInfo)
		}
	}
	conn.Commit()
	return &expiredCertInfos
}

func CountDays(y int, m int, d int) int {
	if m < 3 {
		y--
		m += 12
	}
	return 365*y + (y >> 2) - y/100 + y/400 + (153*m-457)/5 + d - 306
}

func GetYMD(date string) []int {
	var first int
	var second int
	var third int
	count := 0
	for i := 0; i < len(date); i++ {
		if date[i] == '-' && count == 0 {
			first = i
			count++
		}
		if date[i] == '-' && count == 1 {
			second = i
		}
		if date[i] == 'T' {
			third = i
			break
		}
	}
	var ret []int
	y, _ := strconv.Atoi(date[0:first])
	m, _ := strconv.Atoi(date[first+1 : second])
	d, _ := strconv.Atoi(date[second+1 : third])
	ret = append(ret, y, m, d)
	return ret
}

func isExpired(certInfo CertInfo) bool {
	timeTemplate := "2006-01-02T15:04:05"
	nowYMD := GetYMD(time.Now().Format(timeTemplate))
	expireYMD := GetYMD(certInfo.CertExpireDate)
	if CountDays(expireYMD[0], expireYMD[1], expireYMD[2])-CountDays(nowYMD[0], nowYMD[1], nowYMD[2]) <= 30 {
		return true
	}
	return false
}

func QueryEffectAppList(certId string, certType string) []string {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var recAppNames []RecAppName
	recAppNames = GetAppNamesByCertId(conn, certType, certId)
	var appList []string
	for _, recAppName := range recAppNames {
		appList = append(appList, recAppName.AppName)
	}
	return appList
}

func QueryUserNameByAppName(appList []string) []string {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var userNames []string
	conn.Begin()
	for _, appName := range appList {
		var userName UserName
		if err = conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).
			Where("app_name=?", appName).Find(&userName).
			Error; err != nil {
			logs.Error("Query DB Failed:", err)
			conn.Rollback()
			return nil
		}
		userNames = append(userNames, userName.UserName)
	}
	conn.Commit()
	return userNames
}

func QueryCertInfoByCertId(certId string) *CertInfo {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var certInfo CertInfo
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Where("cert_id=?", certId).Find(&certInfo).Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return nil
	}
	return &certInfo
}

func UpdateCertInfo(condition map[string]interface{}, priv_key_url string) bool {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return false
	}
	defer conn.Close()
	var certInfo CertInfo
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Where(condition).Find(&certInfo).Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return false
	}
	certInfo.PrivKeyUrl = priv_key_url
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Where(condition).Update(&certInfo).Error; err != nil {
		logs.Error("Update DB Failed:", err)
		return false
	}
	return true
}

func UpdateCertInfoAfterUpload(condition map[string]interface{}, inputs map[string]interface{}) *CertInfo {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()

	var certInfo CertInfo
	err = conn.Debug().
		Table(CertInfo{}.TableName()).
		Where(condition).
		Updates(inputs).
		Find(&certInfo).
		Error
	utils.RecordError("Update DB Failed: ", err)
	return &certInfo
}

func CheckCertExit(teamId string) int {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return -1
	}
	defer conn.Close()
	var teamIds []TeamID
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).
		Where("team_id=?", teamId).Find(&teamIds).
		Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return -1
	}
	if len(teamIds) == 0 {
		return -2
	}
	return 0
}
