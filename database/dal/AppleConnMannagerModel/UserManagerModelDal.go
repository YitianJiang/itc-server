package devconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"github.com/astaxie/beego/logs"
)

func DeleteVisibleAppInfo(condition map[string]interface{}) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	if err := connection.Table(AllVisibleAppDB{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Delete(AllVisibleAppDB{}).Error; err != nil {
		logs.Error("删除可见app信息失败!", err)
	}
	return true
}

func InsertVisibleAppInfo (appInfo RetAllVisibleAppItem,teamId string) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	var appInfoTran AllVisibleAppDB
	appInfoTran.AppAppleId = appInfo.AppAppleId
	appInfoTran.BundleID = appInfo.BundleID
	appInfoTran.AppName = appInfo.AppName
	appInfoTran.TeamId = teamId
	err = connection.Table(AllVisibleAppDB{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&appInfoTran).Error
	if err != nil {
		logs.Error("添加appInfo失败!", err)
		return false
	}
	return true
}

func SearchVisibleAppInfos (condition map[string]interface{}) (*[]RetAllVisibleAppItem,bool) {
	connection, err := database.GetConneection()
	var AllVisibleAppObjResponse []RetAllVisibleAppItem
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return &AllVisibleAppObjResponse,false
	}
	defer connection.Close()
	if err := connection.Table(AllVisibleAppDB{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Find(&AllVisibleAppObjResponse).Error; err != nil {
		logs.Error("查询appInfo失败!", err)
		return &AllVisibleAppObjResponse,false
	}
	return &AllVisibleAppObjResponse,true
}