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
//记录user的权限变更历史

func addStringFromSlice(x []string) string{
	var sliceString string
	sliceString = x[0]
	x = x[1:]
	for _,item := range x {
		sliceString = sliceString + "," + item
	}
	stringIntLen := len(sliceString)
	sliceString = sliceString[:stringIntLen]
	return sliceString
}

func InsertUserPermEditHistoryDB(userPermInfoReq *UserPermEditReq) bool{
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	var userPermInfoDB InsertUserPermEditHistoryDBModel
	userPermInfoDB.OperateUserName = userPermInfoReq.OperateUserName
	userPermInfoDB.TeamId = userPermInfoReq.TeamId
	userPermInfoDB.UserId = userPermInfoReq.UserId
	userPermInfoDB.AppleId = userPermInfoReq.AppleId
	if userPermInfoReq.ProvisioningChangeSign == "1"{
		if userPermInfoReq.ProvisioningAllowedResult {
			userPermInfoDB.ProvisioningChange = "open"
		}else {
			userPermInfoDB.ProvisioningChange = "close"
		}
	}
	if userPermInfoReq.AllappsVisibleChangeSign == "1"{
		if userPermInfoReq.AllAppsVisibleResult {
			userPermInfoDB.AllappsVisibleChange = "open"
		}else {
			userPermInfoDB.AllappsVisibleChange = "close"
		}
	}
	if len(userPermInfoReq.RolesAdd) > 0 {
		userPermInfoDB.RolesAdd = addStringFromSlice(userPermInfoReq.RolesAdd)
	}
	if len(userPermInfoReq.RolesMin) > 0 {
		userPermInfoDB.RolesMin = addStringFromSlice(userPermInfoReq.RolesMin)
	}
	if len(userPermInfoReq.VisibleAppsAdd) > 0 {
		userPermInfoDB.VisibleAppsAdd = addStringFromSlice(userPermInfoReq.VisibleAppsAdd)
	}
	if len(userPermInfoReq.VisibleAppsMin) > 0 {
		userPermInfoDB.VisibleAppsMin = addStringFromSlice(userPermInfoReq.VisibleAppsMin)
	}
	err = connection.Table(InsertUserPermEditHistoryDBModel{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&userPermInfoDB).Error
	if err != nil {
		logs.Error("记录user的权限变更历史失败", err)
		return false
	}
	return true
}

func InsertUserInvitedHistoryDB(userInvitedInfoReq *UserInvitedReq,invitedOrCancel string) bool{
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	var userInvitedDB InsertUserInvitedHistoryDBModel
	userInvitedDB.OperateUserName = userInvitedInfoReq.OperateUserName
	userInvitedDB.TeamId = userInvitedInfoReq.TeamId
	userInvitedDB.AppleId = userInvitedInfoReq.AppleId
	userInvitedDB.InvitedOrCancel = invitedOrCancel
	err = connection.Table(InsertUserInvitedHistoryDBModel{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&userInvitedDB).Error
	if err != nil {
		logs.Error("记录user的邀请历史失败", err)
		return false
	}
	return true
}