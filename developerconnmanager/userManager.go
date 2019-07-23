package developerconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"

)

type GetCapabilitiesInfoReq struct {
	AppType     string  `form:"app_type" json:"app_type" binding:"required"`
}

type CapabilitiesInfo struct {
	SelectCapabilitiesInfo []string `json:"select_capabilities"`
	SettingsCapabilitiesInfo map[string][]string `json:"settings_capabilities"`
}

func AssembleJsonResponse(c *gin.Context, errorNo int, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"error_code": errorNo,
		"error_info": message,
		"data":    data,
	})
}

func GetBundleIdCapabilitiesInfo(c *gin.Context){
	logs.Info("返回BundleID的能力给前端做展示")
	var requestData GetCapabilitiesInfoReq
	bindQueryError := c.ShouldBindQuery(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	var responseData CapabilitiesInfo
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", responseData)
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

func ReqToAppleGetInfo(url,tokenString string,obj interface{}) bool{
	//var obj devconnmanager.FromAppleUserInfo
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Info("新建request对象失败")
		return false
	}
	request.Header.Set("Authorization", tokenString)
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送get请求失败")
		return false
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		logs.Info(string(response.StatusCode))
		responseByte, _ := ioutil.ReadAll(response.Body)
		logs.Info(string(responseByte))
		return false
	} else {
		responseByte, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
			return false
		}
		//logs.Info(string(responseByte))
		json.Unmarshal(responseByte, &obj)
		return true
	}
}

//type BundleIDCapabilitiesInfo struct {
//	ICloudInfo string `json:"icloud_action" binding:"required"`
//	PushCapInfo []string `json:"push_list"`
//}
//func AssertBunldIDInfo(c *gin.Context){
//	logs.Info("确认bundle id的能力")
//	var requestData BundleIDCapabilitiesInfo
//	bindQueryError := c.ShouldBindJSON(&requestData)
//	utils.RecordError("请求参数绑定错误: ", bindQueryError)
//	c.JSON(http.StatusOK, gin.H{
//		"message":   "success",
//		"errorCode": 0,
//		"data_push": requestData.PushCapInfo,
//		"data":requestData.ICloudInfo,
//	})
//}

//账号人员管理相关开发
func UserRolesGetInfo(c *gin.Context){
	logs.Info("返回角色列表以供前端筛选")
	var responseData devconnmanager.RolesInfoRes
	responseData.RolesObj = make(map[string]string)
	responseData.RolesObj = _const.RolesInfoMap
	responseData.RolesIndex = _const.RolesIndexList
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"error_code": 0,
		"data": responseData,
	})
}

//账号人员管理相关开发
func UserDetailInfoStructTransfer (itemUserData *devconnmanager.FromAppleUserItemInfo) *devconnmanager.RetUsersDataDetailObj{
	var resItemUserInfo devconnmanager.RetUsersDataDetailObj
	resItemUserInfo.UserIdApple = itemUserData.UserIdApple
	resItemUserInfo.UserName = itemUserData.AttributeUserInfo.UserName
	resItemUserInfo.Email = itemUserData.AttributeUserInfo.Email
	resItemUserInfo.FirstName = itemUserData.AttributeUserInfo.FirstName
	resItemUserInfo.LastName = itemUserData.AttributeUserInfo.LastName
	resItemUserInfo.Roles = itemUserData.AttributeUserInfo.Roles
	resItemUserInfo.AllAppsVisible = itemUserData.AttributeUserInfo.AllAppsVisible
	resItemUserInfo.ProvisioningAllowed = itemUserData.AttributeUserInfo.ProvisioningAllowed
	resItemUserInfo.ExpirationDate = itemUserData.AttributeUserInfo.ExpirationDate
	return &resItemUserInfo
}

func CreateUsersInfoResData(obj *devconnmanager.FromAppleUserInfo) *devconnmanager.RetUsersDataObj{
	var resDataobj devconnmanager.RetUsersDataObj
	resDataobj.EmployData = make(map[string][]devconnmanager.RetUsersDataDetailObj)
	for _,itemUserData := range obj.DataList {
		for _,roleItem := range itemUserData.AttributeUserInfo.Roles {
			if _, ok := resDataobj.EmployData[roleItem]; ok {
				resItemUserInfo := UserDetailInfoStructTransfer(&itemUserData)
				resDataobj.EmployData[roleItem] = append(resDataobj.EmployData[roleItem],*resItemUserInfo)
			}else {
				roleUserinfoList := make([]devconnmanager.RetUsersDataDetailObj,0)
				resItemUserInfo := UserDetailInfoStructTransfer(&itemUserData)
				roleUserinfoList = append(roleUserinfoList,*resItemUserInfo)
				resDataobj.EmployData[roleItem] = roleUserinfoList
			}
		}
	}
	return &resDataobj
}

func UserDetailInfoGet(c *gin.Context){
	logs.Info("返回全部人员信息")
	var requestData devconnmanager.UserDetailInfoReq
	bindQueryError := c.ShouldBindQuery(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", map[string]interface{}{})
		return
	}
	logs.Info(requestData.TeamId)
	tokenString := GetTokenStringByTeamId(requestData.TeamId)
	var obj devconnmanager.FromAppleUserInfo
	resultFromApple := ReqToAppleGetInfo(_const.APPLE_USER_INFO_URL,tokenString,&obj)
	if resultFromApple {
		resDataobj := CreateUsersInfoResData(&obj)
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"error_code": 0,
			"data": resDataobj,
		})
	}else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "苹果后台返回数据错误",
			"error_code": 1,
			"data": "",
		})
	}
	return
}

func CreateUsersInvitedInfoResData(obj *devconnmanager.FromAppleUserInfo) *devconnmanager.RetUsersInvitedDataObj{
	var resDataobj devconnmanager.RetUsersInvitedDataObj
	resDataobj.InvitedData = make(map[string][]devconnmanager.RetUsersDataDetailObj)
	for _,itemUserData := range obj.DataList {
		for _,roleItem := range itemUserData.AttributeUserInfo.Roles {
			if _, ok := resDataobj.InvitedData[roleItem]; ok {
				resItemUserInfo := UserDetailInfoStructTransfer(&itemUserData)
				resDataobj.InvitedData[roleItem] = append(resDataobj.InvitedData[roleItem],*resItemUserInfo)
			}else {
				roleUserinfoList := make([]devconnmanager.RetUsersDataDetailObj,0)
				resItemUserInfo := UserDetailInfoStructTransfer(&itemUserData)
				roleUserinfoList = append(roleUserinfoList,*resItemUserInfo)
				resDataobj.InvitedData[roleItem] = roleUserinfoList
			}
		}
	}
	return &resDataobj
}

func UserInvitedDetailInfoGet(c *gin.Context){
	logs.Info("返回全部被邀请的人员信息")
	var requestData devconnmanager.UserDetailInfoReq
	bindQueryError := c.ShouldBindQuery(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", map[string]interface{}{})
		return
	}
	logs.Info(requestData.TeamId)
	tokenString := GetTokenStringByTeamId(requestData.TeamId)
	var obj devconnmanager.FromAppleUserInfo
	resultFromApple := ReqToAppleGetInfo(_const.APPLE_USER_INVITED_INFO_URL,tokenString,&obj)
	if resultFromApple  {
		resDataobj := CreateUsersInvitedInfoResData(&obj)
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"error_code": 0,
			"data": resDataobj,
		})
	}else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "苹果后台返回数据错误",
			"error_code": 1,
			"data": "",
		})
	}
	return
}
//返回账号下全部app信息
func CreateVisibleAppsInfoResData(obj *devconnmanager.RetAllVisibleAppsFromApple,teamId string) *[]devconnmanager.RetAllVisibleAppItem{
	resDataListobj := make([]devconnmanager.RetAllVisibleAppItem,0)
	for _,itemUserData := range obj.DataList {
		var resDataobj devconnmanager.RetAllVisibleAppItem
		resDataobj.AppAppleId = itemUserData.AppAppleId
		resDataobj.AppName = itemUserData.AppsAttribute.AppName
		resDataobj.BundleID = itemUserData.AppsAttribute.BundleID
		resDataListobj = append(resDataListobj,resDataobj)
		devconnmanager.InsertVisibleAppInfo(resDataobj,teamId)
	}
	return &resDataListobj
}

func VisibleAppsInfoGet(c *gin.Context) {
	logs.Info("返回账号下全部app信息")
	var requestData devconnmanager.UserDetailInfoReq
	bindQueryError := c.ShouldBindQuery(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", map[string]interface{}{})
		return
	}
	logs.Info(requestData.TeamId)
	tokenString := GetTokenStringByTeamId(requestData.TeamId)
	if requestData.UpdateDBControl == "1"{
		delStatus := devconnmanager.DeleteVisibleAppInfo(map[string]interface{}{"team_id": requestData.TeamId})
		if !delStatus{
			c.JSON(http.StatusOK, gin.H{
				"error_info":   "数据库删除数据失败",
				"error_code": 2,
				"data": "",
			})
			return
		}
		var obj devconnmanager.RetAllVisibleAppsFromApple
		resultFromApple := ReqToAppleGetInfo(_const.APPLE_USER_VISIBLE_APPS_URL,tokenString,&obj)
		if !resultFromApple{
			c.JSON(http.StatusOK, gin.H{
				"error_info":   "苹果后台返回数据错误",
				"error_code": 1,
				"data": "",
			})
			return
		}
		resDataobj := CreateVisibleAppsInfoResData(&obj,requestData.TeamId)
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"error_code": 0,
			"data": resDataobj,
		})
	}else {
		resDataobj,searchRes := devconnmanager.SearchVisibleAppInfos(map[string]interface{}{"team_id": requestData.TeamId})
		if !searchRes{
			c.JSON(http.StatusOK, gin.H{
				"error_info":   "数据库查询数据失败",
				"error_code": 3,
				"data": "",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"error_code": 0,
			"data": resDataobj,
		})
	}
	return
}

//返回个人可见的全部app信息
func CreateVisibleAppsInfoResDataNoDB(obj *devconnmanager.RetAllVisibleAppsFromApple) *[]devconnmanager.RetAllVisibleAppItem{
	resDataListobj := make([]devconnmanager.RetAllVisibleAppItem,0)
	for _,itemUserData := range obj.DataList {
		var resDataobj devconnmanager.RetAllVisibleAppItem
		resDataobj.AppAppleId = itemUserData.AppAppleId
		resDataobj.AppName = itemUserData.AppsAttribute.AppName
		resDataobj.BundleID = itemUserData.AppsAttribute.BundleID
		resDataListobj = append(resDataListobj,resDataobj)
	}
	return &resDataListobj
}

func VisibleAppsOfUserGet(c *gin.Context)  {
	logs.Info("返回个人可见的全部app信息")
	var requestData devconnmanager.UserVisibleAppsReq
	bindQueryError := c.ShouldBindQuery(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", map[string]interface{}{})
		return
	}
	logs.Info(requestData.TeamId)
	tokenString := GetTokenStringByTeamId(requestData.TeamId)
	var obj devconnmanager.RetAllVisibleAppsFromApple
	var visibleUrlChoice string
	if requestData.OrInvited == "1"{
		visibleUrlChoice = _const.APPLE_USER_INVITED_INFO_URL_NO_PARAM
	}else {
		visibleUrlChoice = _const.APPLE_USER_INFO_URL_NO_PARAM
	}
	userIdVisibleAppUrl := visibleUrlChoice + "/" + requestData.UserId + "/visibleApps?limit=200&fields[apps]=bundleId,name"
	resultFromApple := ReqToAppleGetInfo(userIdVisibleAppUrl,tokenString,&obj)
	if !resultFromApple{
		c.JSON(http.StatusOK, gin.H{
			"error_info":   "苹果后台返回数据错误",
			"error_code": 1,
			"data": "",
		})
		return
	}
	resDataobj := CreateVisibleAppsInfoResDataNoDB(&obj)
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"error_code": 0,
		"data": resDataobj,
	})
}