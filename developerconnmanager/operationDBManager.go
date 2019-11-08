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
	BundleId           string		`json:"bundle_id"      binding:"required"`
	CapName            string 	    `json:"cap_name"       binding:"required"`
}

type DBItemFromBundleId struct {
	gorm.Model
	BundleId           string 	    `gorm:"column:bundle_id"                 json:"bundle_id"	   `
	InAppPurchase      string       `gorm:"column:IN_APP_PURCHASE"           json:"IN_APP_PURCHASE"`
}

type DeleteCertPrivReqModel struct {
	CertId    		   string	    `json:"cert_id"        binding:"required"`
	ColumnName		   string		`json:"colum_name"     binding:"required"`
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
		utils.RecordError("查询 tt_app_account_cert失败，查询条件："+fmt.Sprint(queryData)+",errInfo：", err)
	}
	logs.Info("看这里",result)
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
		utils.RecordError("更新 tt_app_account_cert失败，更新条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}

func DeleteBundleIdCap(c *gin.Context){
	logs.Info("删除指定的bundleid的能力")
	var body DeleteBundleIdReqModel
	err := c.ShouldBindJSON(&body)
	utils.RecordError("参数绑定失败", err)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	conditionData := map[string]interface{}{"bundle_id":body.BundleId}
	//queryAppAccountCert(conditionData)
	capDelete := map[string]interface{}{body.CapName:""}
	dbError := deleteTableColumnDb(devconnmanager.AppleBundleId{}.TableName(),conditionData,capDelete)
	utils.RecordError("更新tt_app_account_cert表数据出错：%v", dbError)
	if dbError != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "更新tt_app_account_cert表数据出错", "failed")
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
	logs.Info("删除指定的Cert表中的列")
	var body DeleteCertPrivReqModel
	err := c.ShouldBindJSON(&body)
	utils.RecordError("参数绑定失败", err)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	conditionData := map[string]interface{}{"cert_id":body.CertId}
	capDelete := map[string]interface{}{body.ColumnName:""}
	dbError := deleteTableColumnDb("tt_apple_certificate",conditionData,capDelete)
	utils.RecordError("更新tt_app_account_cert表数据出错：%v", dbError)
	if dbError != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "更新tt_apple_certificate表数据出错", "failed")
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "update success",
			"errorCode": 0,
		})
		return
	}
}