package dal

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type CertificateModel struct {
	gorm.Model
	Appname string `gorm:"column:appname" json:"appname"`
	AppId int `gorm:"column:app_id" json:"app_id"`
	Creator string `gorm:"column:creator" json:"creator"`
	Mails string `gorm:"column:mails" json:"mails"`
	Usage string `gorm:"column:usage" json:"usage"`
	Type string `gorm:"column:type" json:"type"`
	Password string `gorm:"column:password" json:"password"`
	ExpireTime string `gorm:"column:expire_time" json:"expire_time"`
	CertificateFile string `gorm:"column:certificate_file" json:"certificate_file"`
	CertificateFileName string `gorm:"column:certificate_file_name" json:"certificate_file_name"`
	PemFile string `gorm:"column:pem_file" json:"pem_file"`
	PemFileName string `gorm:"column:pem_file_name" json:"pem_file_name"`
}
func (CertificateModel) TableName() string {
	return "tb_certificate"
}

/*
增/查/删 证书
*/
func InsertCertificate(certificate CertificateModel) bool{
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

func QueryCertificate(condition map[string]interface{}) (*[]CertificateModel) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var configs []CertificateModel
	if err := connection.Table(CertificateModel{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Find(&configs).Error; err != nil{
		logs.Error("查询证书列表失败!", err)
	}
	return &configs
}


func DeleteCertificate(condition map[string]interface{}) bool{
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	if err := connection.Table(CertificateModel{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Delete(CertificateModel{}).Error; err != nil{
		logs.Error("删除证书列表失败!", err)
	}
	return true
}
