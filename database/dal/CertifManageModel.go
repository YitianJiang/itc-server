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

type CertInfo struct {
	gorm.Model
	AccountName         string      `gorm:"column:account_name"            json:"account_name,omitempty"`
	TeamId              string      `gorm:"column:team_id"                 json:"team_id,omitempty"`
	CertName            string      `gorm:"column:cert_name"               json:"cert_name"`
	CertId              string      `gorm:"column:cert_id"                 json:"cert_id"`
	CertExpireDate      string      `gorm:"column:cert_expire_date"        json:"cert_expire_date"`
	CertType            string      `gorm:"column:cert_type"               json:"cert_type"`
	CertDownloadUrl     string      `gorm:"column:cert_download_url"       json:"cert_download_url"`
	PrivKeyUrl          string      `gorm:"column:priv_key_url"            json:"priv_key_url"`
	CsrFileUrl          string      `gorm:"column:csr_file_url"            json:"csr_file_url,omitempty"`
	EffectAppList       []string    `gorm:"-"                              json:"effect_app_list,omitempty"`
}

//创建证书响应
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

//新增证书请求
type InsertCertRequest struct {
	AccountName        string      `json:"account_name"`
	TeamId             string      `json:"team_id"`
	CertName           string      `json:"cert_name"`
	CertType           string      `json:"cert_type"`
}

//创建苹果证书请求
type CreAppleCertReq struct {
	Data Data       `json:"data"`
}

type Data struct {
	Type        string          `json:"type"`
	Attributes  AttributesSend  `json:"attributes"`
}
//todo 这1，2都是和谁学的？！！！！
type AttributesSend struct {
	CsrContent          string      `json:"csrContent"`
	CertificateType     string      `json:"certificateType"`
}
//获取权限请求
type GetPermsResponse struct {
	Data map[string][]string    `json:"data"`
	Errno   int                 `json:"errno"`
	Message string              `json:"message"`
}
//删除证书请求
type DelCertRequest struct {
	CertId   string  `form:"cert_id"`
	TeamId   string  `form:"team_id"`
}
//查询证书请求
type QueryCertRequest struct {
	TeamId         string      `form:"team_id"      json:"team_id"`
	ExpireSoon     string      `form:"expire_soon"  json:"expire_soon"`
	UserName       string      `form:"user_name"    json:"user_name"`
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

type RecAppName struct {
	AppName string
}
//todo 注释里面写清楚表名称！！！
//先根据条件到表tt_apple_conn_account中筛选证书，再根据筛选出来的证书id到表tt_apple_certificate中查询受影响的app
func QueryCertInfo(condition map[string]interface{},expireSoon string) *[]CertInfo {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var certInfos []CertInfo
	db:=conn.LogMode(_const.DB_LOG_MODE).Table(CertInfo{}.TableName()).Where(condition).Find(&certInfos)
	utils.RecordError("Query from DB Failed: ", db.Error)
	//todo certRelatedInfosMap这个玩意到底是啥？在定义个新struct（用CertInfo类型做为新struct其中一列），再新增一列effectAppList不行？最后返回一个[]struct不行？有好好思考？
	//todo 在这玩啥呢，appAccountCerts需要整个塞入取数据嘛，这个不是只取app_name的list作为effectAppList？？不知道你要干啥！！
	var ret []CertInfo
	for i:=0;i<len(certInfos);i++{
		var recAppNames []RecAppName
		db=conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).Where("cert_id=?",certInfos[i].CertId).Select("app_name").Find(&recAppNames)
		utils.RecordError("Update DB Failed: ", db.Error)
		for _,recAppName:=range recAppNames{
			certInfos[i].EffectAppList=append(certInfos[i].EffectAppList, recAppName.AppName)
		}
		if expireSoon=="1"&&isExpired(certInfos[i])==true{
			ret=append(ret, certInfos[i])
		}
		if expireSoon=="0"{
			ret=append(ret, certInfos[i])
		}
	}
	//todo 大的object的传递应该用啥？
	return &ret
}
//todo 注释里面写清楚表名称！！！
//先根据条件到表tt_apple_conn_account中筛选要过期的证书，再根据筛选出来的证书id到表tt_apple_certificate中查询受影响的app
func QueryExpiredCertInfos() *[]CertInfo {
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
	var recAppNames []RecAppName
	for _,certInfo:=range certInfos{
		if isExpired(certInfo)==true{
			db=conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).Select("app_name").Where("cert_id=?",certInfo.CertId).Scan(&recAppNames)
			utils.RecordError("Update DB Failed: ", db.Error)
			if len(recAppNames)==0{
				continue
			}
			for _,recAppName:=range recAppNames{
				certInfo.EffectAppList=append(certInfo.EffectAppList, recAppName.AppName)
			}
			expiredCertInfos=append(expiredCertInfos, certInfo)
		}
	}
	//todo certRelatedInfosMap这个玩意到底是啥？在定义个新struct（用CertInfo类型做为新struct其中一列），再新增一列effectAppList不行？最后返回一个[]struct不行？有好好思考？
	//todo 在这玩啥呢，appAccountCerts需要整个塞入取数据嘛，这个不是只取app_name的list作为effectAppList？？不知道你要干啥！！！
	//todo 大的object的传递应该用啥？
	return &expiredCertInfos
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
