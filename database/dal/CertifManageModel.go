package dal

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"strconv"
	"time"
)

type RecvCert struct {
	Data                OutLayer        `json:"data"`
	Links               Links           `json:"links"`
}
type CertsInfo struct {
	Data                []OutLayer      `json:"data"`
	Links               Links           `json:"links"`
	Meta                Meta            `json:"meta"`
}
type Meta struct {
	Paging Paging   `json:"paging"`
}
type Paging struct {
	Limit int       `json:"limit"`
	Total int       `json:"total"`
}

type Links struct {
	Self string     `json:"self"`
}
type OutLayer struct{
	Type                string              `json:"type"`
	Id                  string              `json:"id"`
	Attributes          Attributes          `json:"attributes"`
	Links               Links               `json:"links"`
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
type CertInfo struct {
	gorm.Model
	AccountName         string `gorm:"column:account_name"            json:"account_name"`
	TeamId              string `gorm:"column:team_id"                 json:"team_id"`
	CertName            string `gorm:"column:cert_name"               json:"cert_name"`
	CertId              string `gorm:"column:cert_id"                 json:"cert_id"`
	CertExpireDate      string `gorm:"column:cert_expire_date"        json:"cert_expire_date"`
	CertType            string `gorm:"column:cert_type"               json:"cert_type"`
	CertDownloadUrl     string `gorm:"column:cert_download_url"       json:"cert_download_url"`
	PrivKeyUrl          string `gorm:"column:priv_key_url"            json:"priv_key_url"`
	CsrFileUrl          string `gorm:"column:csr_file_url"            json:"csr_file_url"`
}

//用来给前端返回证书相关信息
type CertRelatedInfo struct {
	CertName            string      `json:"cert_name"`
	CertId              string      `json:"cert_id"`
	CertExpireDate      string      `json:"cert_expire_date"`
	CertType            string      `json:"cert_type"`
	CertDownloadUrl     string      `json:"cert_download_url"`
	PrivKeyUrl          string      `json:"priv_key_url"`
	EffectAppList       []string    `json:"effect_app_list"`
}

type RecvParamsInsertCert struct {
	AccountName        string      `json:"account_name"`
	TeamId             string      `json:"team_id"`
	CertName           string      `json:"cert_name"`
	CertType           string      `json:"cert_type"`
}

type CertCreate struct {
	Data Data       `json:"data"`
}

type Data struct {
	Type        string          `json:"type"`
	Attributes  Attributes2     `json:"attributes"`
}

type Attributes2 struct {
	CsrContent          string      `json:"csrContent"`
	CertificateType     string      `json:"certificateType"`
}

type ResourcePermissions struct {
	Data map[string][]string   `json:"data"`
	Errno int       `json:"errno"`
	Message string  `json:"message"`
}

func (CertInfo) TableName() string{
	return  "tt_apple_certificate"
}
func DeleteCertInfo(condition map[string]interface{}) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	if err=connection.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Where(condition).Delete(&CertInfo{}).Error;err!=nil{
		logs.Error("Delete Record Failed")
		return  false
	}
	return true
}

func InsertCertInfo(CertInfo CertInfo) bool {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return false
	}
	defer conn.Close()
	db:= conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo.TableName()).Create(&CertInfo)
	utils.RecordError("Insert into DB Failed: ", db.Error)
	return true
}

//先根据条件到表1中筛选证书，再根据筛选出来的证书id到表2中查询受影响的app
func QueryCertInfo(condition map[string]interface{}) map[CertInfo][]string {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var CertInfos []CertInfo
	db:=conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Where(condition).Find(&CertInfos)
	utils.RecordError("Query from DB Failed: ", db.Error)
	certRelatedInfosMap:=make(map[CertInfo][]string)
	var appAccountCerts []AppAccountCert
	for _,certInfo:=range CertInfos{
		db=conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).Where("cert_id=?",certInfo.CertId).Find(&appAccountCerts)
		utils.RecordError("Update DB Failed: ", db.Error)
		var effectAppList []string
		for _,appAccountCert:=range appAccountCerts{
			effectAppList=append(effectAppList, appAccountCert.AppName)
		}
		certRelatedInfosMap[certInfo]=effectAppList
	}
	return certRelatedInfosMap
}

//先根据条件到表1中筛选要过期的证书，再根据筛选出来的证书id到表2中查询受影响的app
func QueryExpiredCertInfos() map[CertInfo][]string {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var certInfos []CertInfo
	db:=conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Find(&certInfos)
	utils.RecordError("Query from DB Failed: ", db.Error)
	var expiredCertInfos []CertInfo
	for _,certInfo:=range certInfos{
		if isExpired(certInfo)==true{
			expiredCertInfos=append(expiredCertInfos, certInfo)
		}
	}
	certRelatedInfosMap:=make(map[CertInfo][]string)
	var appAccountCerts []AppAccountCert
	for _,expiredCertInfo:=range expiredCertInfos{
		db=conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).Where("cert_id=?",expiredCertInfo.CertId).Find(&appAccountCerts)
		utils.RecordError("Update DB Failed: ", db.Error)
		var effectAppList []string
		for _,appAccountCert:=range appAccountCerts{
			effectAppList=append(effectAppList, appAccountCert.AppName)
		}
		if len(effectAppList)!=0{
			certRelatedInfosMap[expiredCertInfo]=effectAppList
		}
	}
	return certRelatedInfosMap
}

func CountDays(y int, m int, d int) int{
	if m < 3 {
		y--
		m += 12
	}
	return 365 * y + (y >> 2) - y / 100 + y / 400 + (153 * m - 457) / 5 + d - 306
}

func GetYMD(date string)[]int{
	var first int
	var second int
	var third int
	count:=0
	for i:=0;i<len(date);i++{
		if date[i]=='-'&&count==0{
			first=i
			count++
		}
		if date[i]=='-'&&count==1{
			second=i
		}
		if date[i]=='T'{
			third=i
			break
		}
	}
	var ret []int
	y,_:=strconv.Atoi(date[0:first])
	m,_:=strconv.Atoi(date[first+1:second])
	d,_:=strconv.Atoi(date[second+1:third])
	ret=append(ret,y,m,d)
	return ret
}

func isExpired(certInfo CertInfo) bool{
	timeTemplate:= "2006-01-02T15:04:05"
	nowYMD:=GetYMD(time.Now().Format(timeTemplate))
	expireYMD:=GetYMD(certInfo.CertExpireDate)
	if CountDays(expireYMD[0],expireYMD[1],expireYMD[2])-CountDays(nowYMD[0],nowYMD[1],nowYMD[2])<=30{
		return true
	}
	return false
}

func QueryEffectAppList(condition map[string]interface{}) []string{
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var  appAccountCerts []AppAccountCert
	db:=conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).Where(condition).Find(&appAccountCerts)
	utils.RecordError("Query from DB Failed: ", db.Error)
	var appList []string
	for _,appAccountCert:=range appAccountCerts{
		appList=append(appList, appAccountCert.AppName)
	}
	return appList
}

func QueryTeamId(condition map[string]interface{}) string{
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return ""
	}
	defer conn.Close()
	var  certInfo CertInfo
	db:=conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Where(condition).Find(&certInfo)
	utils.RecordError("Query from DB Failed: ", db.Error)
	return certInfo.TeamId
}
//根据app名称来查找用户名
func QueryUserNameAccAppName(appList []string) []string{
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var userNames []string
	for _,appName:=range appList {
		var appAccountCert AppAccountCert
		db := conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).Where("app_name=?",appName).Find(&appAccountCert)
		userNames= append(userNames, appAccountCert.UserName)
		utils.RecordError("Query from DB Failed: ", db.Error)
	}
	return userNames
}

func UpdateCertInfo(condition map[string]interface{},priv_key_url string) bool {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return false
	}
	defer conn.Close()
	var certInfo CertInfo
	conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Where(condition).Find(&certInfo)
	certInfo.PrivKeyUrl=priv_key_url
	db:= conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Where(condition).Update(&certInfo)
	utils.RecordError("Update DB Failed: ", db.Error)
	return true

}
