package dal

import (
	"strconv"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type CertificateModel struct {
	gorm.Model
	Appname             string `gorm:"column:appname"                    json:"appName"`
	AppId               int    `gorm:"column:app_id"                     json:"appId"`
	Creator             string `gorm:"column:creator"                    json:"creator"`
	Mails               string `gorm:"column:mails"                      json:"mails"`
	Usage               string `gorm:"column:certificate_usage"          json:"usage"`
	Type                string `gorm:"column:certificate_style"          json:"type"`
	Password            string `gorm:"column:certificate_password"       json:"password"`
	ExpireTime          string `gorm:"column:expire_time"                json:"expireTime"`
	CertificateFile     string `gorm:"column:certificate_file"           json:"certificateFile"`
	CertificateFileName string `gorm:"column:certificate_file_name"      json:"certificateName"`
	PemFile             string `gorm:"column:pem_file"                   json:"pemFile"`
	PemFileName         string `gorm:"column:pem_file_name"              json:"pemFileName"`
}

//表命
func (c CertificateModel) TableName() string {
	return "tb_certificate"
}

/*
增/查/删 证书
*/
func InsertCertificate(certificate CertificateModel) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	err = connection.Table(CertificateModel{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&certificate).Error
	if err != nil {
		logs.Error("添加证书失败!", err)
		return false
	}
	return true
}

func QueryCertificate(condition map[string]interface{}) *[]CertificateModel {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var configs []CertificateModel
	if err := connection.Table(CertificateModel{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Find(&configs).Error; err != nil {
		logs.Error("查询证书列表失败!", err)
	}
	return &configs
}

func QueryLikeCertificate(condition map[string]interface{}) *[]CertificateModel {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var configs []CertificateModel
	db := connection.Table(CertificateModel{}.TableName()).LogMode(_const.DB_LOG_MODE)
	//appname 模糊查询，大概率被废弃了
	if _, ok := condition["appname"]; ok {
		appname := "%"
		appname += string(condition["appname"].(string)[:])
		appname += "%"
		db = db.Where("appname LIKE ?", appname)
	}
	if _, ok := condition["certificate_style"]; ok == true {
		db = db.Where("certificate_style = ?", condition["certificate_style"].(string))
	}
	if _, ok := condition["creator"]; ok {
		db = db.Where("creator = ?", condition["creator"].(string))
	}
	if _, ok := condition["appId"]; ok {
		db = db.Where("app_id = ?", condition["appId"].(string))
	}
	db = db.Order("id DESC", true)  //按照id递减排序
	_, ok1 := condition["pageSize"]
	_, ok2 := condition["page"]
	if ok1 && ok2 {
		pageSize, _ := strconv.Atoi(condition["pageSize"].(string))
		page, _ := strconv.Atoi(condition["page"].(string))
		if pageSize > 0 && page > 0 {
			db = db.Limit(pageSize).Offset((page - 1) * pageSize)
		}
	}
	if err := db.Find(&configs).Error; err != nil {
		logs.Error("查询证书列表失败!", err)
	}
	return &configs
}

func DeleteCertificate(condition map[string]interface{}) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	if err := connection.Table(CertificateModel{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Delete(CertificateModel{}).Error; err != nil {
		logs.Error("删除证书列表失败!", err)
	}
	return true
}
