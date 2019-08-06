package developerconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
	"net/http"
)

//返回BundleID的能力给前端做展示
type GetCapabilitiesInfoReq struct {
	AppType     string  `form:"app_type" json:"app_type" binding:"required"`
}

type CapabilitiesInfo struct {
	SelectCapabilitiesInfo []string `json:"select_capabilities"`
	SettingsCapabilitiesInfo map[string][]string `json:"settings_capabilities"`
}

func GetBundleIdCapabilitiesInfo(c *gin.Context){
	logs.Info("返回BundleID的能力给前端做展示")
	var requestData GetCapabilitiesInfoReq
	bindQueryError := c.ShouldBindQuery(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	var responseData CapabilitiesInfo
	if bindQueryError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", responseData)
		return
	}
	responseData.SettingsCapabilitiesInfo = make(map[string][]string)
	responseData.SettingsCapabilitiesInfo[_const.ICLOUD] = _const.CloudSettings
	if requestData.AppType == "iOS"{
		responseData.SelectCapabilitiesInfo = _const.IOSSelectCapabilities
		responseData.SettingsCapabilitiesInfo[_const.DATA_PROTECTION] = _const.ProtectionSettings
	}else {
		responseData.SelectCapabilitiesInfo = _const.MacSelectCapabilities
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data": responseData,
	})
}

///**
//API 3-1：根据业务线appid，返回app相关list
//*/
//func GetAppDetailInfo(c *gin.Context) {
//	username, ok := c.GetQuery("user_name")
//	if !ok {
//		utils.AssembleJsonResponse(c, -1, "缺少user_name参数", "")
//		return
//	}
//	app_id, ok := c.GetQuery("app_id")
//	if !ok {
//		utils.AssembleJsonResponse(c, -1, "缺少app_id参数", "")
//		return
//	}
//	app_acc_certs := devconnmanager.QueryAppAccountCert(map[string]interface{}{
//		"app_id": app_id,
//	})
//	if app_acc_certs == nil {
//		utils.AssembleJsonResponse(c, -2, "数据库查询tt_app_account_cert失败！", "")
//		return
//	} else if len(*app_acc_certs) == 0 {
//		utils.AssembleJsonResponse(c, -3, "未查询到该app_id下的账号信息！", "")
//	}
//
//	var fQueryResult []devconnmanager.APPandCert
//	sql := "select aac.app_name,aac.app_type,aac.id as app_acount_id,aac.team_id,aac.account_verify_status,aac.account_verify_user," +
//		"ac.cert_id,ac.cert_type,ac.cert_expire_date,ac.cert_download_url,ac.priv_key_url from tt_app_account_cert acc, tt_apple_certificate ac" +
//		" where acc.app_id = '" + app_id + "' and aac.deleted IS NULL and (aac.dev_cert_id = ac.id or aac.dist_cert_id = ac.id) and ac.deleted_at IS NULL "
//	f_query := devconnmanager.QueryWithSql(sql, &fQueryResult)
//
//	var resourcPerm devconnmanager.GetPermsResponse
//	url := _const.Certain_Resource_All_PERMS_URL + "employeeKey=" + username + "&resourceKeys[]=" + ""
//	result := QueryPerms(url, &resourcPerm)
//
//}

//func GetResourcePermType(c *gin.Context, teamIds []string, username string) (map[string]int, bool) {
//	url := _const.Certain_Resource_All_PERMS_URL + "employeeKey=" + username
//	var resourceMap = make(map[string]string)
//	for _, teamId := range teamIds {
//		lowTeamId := strings.ToLower(teamId)
//		resource := lowTeamId + "_space_account"
//		resourceMap[resource] = teamId
//		url += "&resourceKeys[]=" + resource
//	}
//	var resourcPerm devconnmanager.GetPermsResponse
//	result := QueryPerms(url, &resourcPerm)
//	if !result || resourcPerm.Errno != 0 {
//		utils.AssembleJsonResponse(c, -4, "查询权限失败！", "")
//		return nil, false
//	}
//	for k := range resourcPerm.Data {
//
//	}
//
//}

func DeleteAppAllInfoFromDB (c *gin.Context){
	var requestData devconnmanager.DeleteAppAllInfoRequest
	bindJsonError := c.ShouldBindQuery(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindJsonError)
	if bindJsonError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	conditionDB := map[string]interface{}{"app_id":requestData.AppId,"app_name":requestData.AppName}
	dbError := devconnmanager.DeleteAppAccountCert(conditionDB)
	utils.RecordError("删除tt_app_account_cert表数据出错：%v", dbError)
	if dbError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "删除tt_app_account_cert表数据出错", "failed")
		return
	}
	dbErrorInfo := devconnmanager.DeleteAppBundleProfiles(conditionDB)
	utils.RecordError("删除tt_app_account_cert表数据出错：%v", dbErrorInfo)
	if dbErrorInfo != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "删除tt_app_bundleId_profiles表数据出错", "failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "delete success",
		"errorCode": 0,
	})
	return
}

func CreateAppBindAccount(c *gin.Context) {
	var requestData devconnmanager.CreateAppBindAccountRequest
	//获取请求参数
	bindJsonError := c.ShouldBindJSON(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindJsonError)
	if bindJsonError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	logs.Info("request:%v", requestData)

	/*inputs := map[string]interface{}{
		"app_id": requestData.AppId,
		"app_name": requestData.AppName,
	}*/
	//todo 根据app_id和app_name执行update，如果返回的操作行数为0，则插入数据
	//todo 等待kani提供根据资源和权限获取人员信息的接口，根据该接口获取需要发送审批消息的用户list
	//todo lark消息生成并批量发送
}
