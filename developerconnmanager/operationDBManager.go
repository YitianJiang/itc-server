package developerconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"fmt"
	"code.byted.org/clientQA/itc-server/utils"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"github.com/gin-gonic/gin"
	"net/http"
	"code.byted.org/gopkg/logs"
)

type DeleteBundleIdReqModel struct {
	BundleId           string		`json:"bundle_id"      					binding:"required"`
	CapName            string 	    `json:"cap_name"       					binding:"required"`
}

type DBItemFromBundleId struct {
	gorm.Model
	BundleId           string 	    `gorm:"column:bundle_id"                json:"bundle_id"`
	InAppPurchase      string       `gorm:"column:IN_APP_PURCHASE"          json:"IN_APP_PURCHASE"`
}

type DeleteCertPrivReqModel struct {
	CertId    		   string	    `json:"cert_id"        					binding:"required"`
	ColumnName		   string		`json:"colum_name"     					binding:"required"`
}

type InsertCertInfoReqModel struct {
	AccountName        string		`json:"account_name"        			binding:"required"`
	TeamId			   string		`json:"team_id"        					binding:"required"`
	CertName		   string		`json:"cert_name"        				binding:"required"`
	PrivKeyUrl		   string       `json:"priv_key_url"        			binding:"required"`
	CsrFileUrl		   string       `json:"csr_file_url"        			binding:"required"`
	CertId    		   string	    `json:"cert_id"        					binding:"required"`
	CertExpireDate	   string		`json:"cert_expire_date"        		binding:"required"`
	CertType		   string       `json:"cert_type"        		        binding:"required"`
	CertDownloadUrl	   string		`json:"cert_download_url"        		binding:"required"`
}


func queryAppAccountCert(tableName string,queryData map[string]interface{}) {
	conn, err := database.GetDBConnection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
	}
	defer conn.Close()
	db := conn.Table(tableName).LogMode(_const.DB_LOG_MODE)
	var result = make([]DBItemFromBundleId, 0)
	if err := db.Where(queryData).Find(&result).Error; err != nil {
		utils.RecordError("?????? tt_app_account_cert????????????????????????"+fmt.Sprint(queryData)+",errInfo???", err)
	}
	logs.Info("?????????",result)
}

func deleteTableColumnDb(tableName string,queryData, item map[string]interface{}) error {
	conn, err := database.GetDBConnection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(tableName).LogMode(_const.DB_LOG_MODE)
	if err1 := db.Where(queryData).Update(item).Error; err1 != nil {
		utils.RecordError("?????? tt_app_account_cert????????????????????????"+fmt.Sprint(queryData)+",errInfo???", err1)
		return err1
	}
	return nil
}

func DeleteBundleIdCap(c *gin.Context){
	logs.Info("???????????????bundleid?????????")
	var body DeleteBundleIdReqModel
	err := c.ShouldBindJSON(&body)
	utils.RecordError("??????????????????", err)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "????????????????????????", "failed")
		return
	}
	conditionData := map[string]interface{}{"bundle_id":body.BundleId}
	//queryAppAccountCert(conditionData)
	capDelete := map[string]interface{}{body.CapName:""}
	dbError := deleteTableColumnDb(devconnmanager.AppleBundleId{}.TableName(),conditionData,capDelete)
	utils.RecordError("??????tt_app_account_cert??????????????????%v", dbError)
	if dbError != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "??????tt_app_account_cert???????????????", "failed")
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "update success",
			"errorCode": 0,
		})
		return
	}
}

func DeleteCertPrivKey(c *gin.Context){
	logs.Info("???????????????Cert????????????")
	var body DeleteCertPrivReqModel
	err := c.ShouldBindJSON(&body)
	utils.RecordError("??????????????????", err)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "????????????????????????", "failed")
		return
	}
	conditionData := map[string]interface{}{"cert_id":body.CertId}
	capDelete := map[string]interface{}{body.ColumnName:""}
	dbError := deleteTableColumnDb("tt_apple_certificate",conditionData,capDelete)
	utils.RecordError("??????tt_app_account_cert??????????????????%v", dbError)
	if dbError != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "??????tt_apple_certificate???????????????", "failed")
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "update success",
			"errorCode": 0,
		})
		return
	}
}

func InsertCertInfoTest(c *gin.Context){
	logs.Info("???????????????Cert????????????")
	var body InsertCertInfoReqModel
	err := c.ShouldBindJSON(&body)
	utils.RecordError("??????????????????", err)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "????????????????????????", "failed")
		return
	}
	var insertObj devconnmanager.CertInfo
	insertObj.AccountName = body.AccountName
	insertObj.TeamId = body.TeamId
	insertObj.CertName = body.CertName
	insertObj.PrivKeyUrl = body.PrivKeyUrl
	insertObj.CsrFileUrl = body.CsrFileUrl
	insertObj.CertId = body.CertId
	insertObj.CertExpireDate = body.CertExpireDate
	insertObj.CertType = body.CertType
	insertObj.CertDownloadUrl = body.CertDownloadUrl
	dbResult := devconnmanager.InsertCertInfo(&insertObj)
	if !dbResult {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "???????????????????????????????????????", body)
		return
	}
	utils.AssembleJsonResponse(c, _const.SUCCESS, "????????????", body)
	return
}