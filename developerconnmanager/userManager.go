package developerconnmanager

import (
	"bytes"
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"

)

func AssembleJsonResponse(c *gin.Context, errorNo int, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"error_code": errorNo,
		"error_info": message,
		"data":    data,
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
		logs.Info("查看返回状态码",response.StatusCode)
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
		var objRes devconnmanager.FromAppleUserInfo
		resultFromAppleGetHolder := ReqToAppleGetInfo(_const.APPLE_USER_INFO_URL_GET_HOLDER,tokenString,&objRes)
		if !resultFromAppleGetHolder || len(objRes.DataList) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"error_info":   "苹果后台返回账户持有者数据错误",
				"error_code": 1,
				"data": "",
			})
			return
		}
		allVisibleReqUrl := _const.APPLE_USER_INFO_URL_NO_PARAM + "/" +objRes.DataList[0].UserIdApple + "/visibleApps?limit=200&fields[apps]=bundleId,name"
		var obj devconnmanager.RetAllVisibleAppsFromApple
		resultFromApple := ReqToAppleGetInfo(allVisibleReqUrl,tokenString,&obj)
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
		resDataobj,searchRes := devconnmanager.SearchVisibleAppInfos(map[string]interface{}{"team_id": requestData.TeamId,"deleted_at": nil})
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

//编辑指定user的权限
func PostToAppleGetInfo(method,url,tokenString string,obj interface{}) bool{
	bodyByte, _ := json.Marshal(&obj)
	logs.Info(string(bodyByte))
	rbodyByte := bytes.NewReader(bodyByte)
	client := &http.Client{}
	request, err := http.NewRequest(method, url, rbodyByte)
	if err != nil {
		logs.Info("新建request对象失败")
		return false
	}
	request.Header.Set("Authorization", tokenString)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送Post请求失败")
		return false
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated {
		responseByte, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
			return false
		}
		logs.Info(string(responseByte))
		//json.Unmarshal(responseByte, &obj)
		return true
	} else {
		logs.Info("查看返回状态码",response.StatusCode)
		responseByte, _ := ioutil.ReadAll(response.Body)
		logs.Info(string(responseByte))
		return false
	}
}

func EditPermOfUserFunc(c *gin.Context){
	logs.Info("变更指定user的权限")
	var requestData devconnmanager.UserPermEditReq
	bindQueryError := c.ShouldBindJSON(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", map[string]interface{}{})
		return
	}
	logs.Info(requestData.TeamId,requestData.AppleId,requestData.UserId)
	var reqAppleEditUserPerm devconnmanager.UserPermEditReqOfAppleObj
	reqAppleEditUserPerm.DataObj.Type = "users"
	reqAppleEditUserPerm.DataObj.Id = requestData.UserId
	reqAppleEditUserPerm.DataObj.Attributes.Roles = requestData.RolesResult
	reqAppleEditUserPerm.DataObj.Attributes.AllAppsVisible = requestData.AllAppsVisibleResult
	reqAppleEditUserPerm.DataObj.Attributes.ProvisioningAllowed = requestData.ProvisioningAllowedResult
	if requestData.AllAppsVisibleResult == false {
		reqAppleEditUserPerm.DataObj.Relationships = &devconnmanager.VisibleAppObjReqOfApple{}
		reqAppleEditUserPerm.DataObj.Relationships.VisibleApps = &devconnmanager.VisibleAppsReqOfApple{}
		reqAppleEditUserPerm.DataObj.Relationships.VisibleApps.DataList = make([]devconnmanager.VisibleAppItemReqOfApple,0)
		if len(requestData.VisibleAppsResult) > 0{
			for _,itemApp := range requestData.VisibleAppsResult{
				var visibleAppItem devconnmanager.VisibleAppItemReqOfApple
				visibleAppItem.AppType = "apps"
				visibleAppItem.AppAppleId = itemApp
				reqAppleEditUserPerm.DataObj.Relationships.VisibleApps.DataList = append(reqAppleEditUserPerm.DataObj.Relationships.VisibleApps.DataList,visibleAppItem)
			}
		}else {
			c.JSON(http.StatusOK, gin.H{
				"error_info":   "参数传递错误，all_apps_visible_result是false情况下，visible_apps_result不能为空",
				"error_code": 1,
				"data": "",
			})
			return
		}
	}
	tokenString := GetTokenStringByTeamId(requestData.TeamId)
	method := "PATCH"
	userPermEditUrl := _const.APPLE_USER_PERM_EDIT_URL + "/" + requestData.UserId
	resultFromApple := PostToAppleGetInfo(method,userPermEditUrl,tokenString,&reqAppleEditUserPerm)
	if !resultFromApple{
		c.JSON(http.StatusOK, gin.H{
			"error_info":   "苹果后台返回数据错误",
			"error_code": 1,
			"data": "",
		})
		return
	}else {
		dbInsertResult := devconnmanager.InsertUserPermEditHistoryDB(&requestData)
		if dbInsertResult {
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"error_code": 0,
				"data": "变更权限成功",
			})
		}else {
			c.JSON(http.StatusOK, gin.H{
				"error_info":   "数据库插入失败",
				"error_code": 2,
				"data": "db insert error",
			})
		}
		return
	}
}
//用户邀请接口
func UserInvitedFunc(c *gin.Context) {
	logs.Info("邀请用户进入企业账号")
	var requestData devconnmanager.UserInvitedReq
	bindQueryError := c.ShouldBindJSON(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", map[string]interface{}{})
		return
	}
	logs.Info(requestData.TeamId,requestData.AppleId)
	var reqAppleUserInvited devconnmanager.UserInvitedReqOfAppleObj
	reqAppleUserInvited.DataObj.Type = "userInvitations"
	reqAppleUserInvited.DataObj.Attributes.Email = requestData.AppleId
	reqAppleUserInvited.DataObj.Attributes.FirstName = requestData.FirstName
	reqAppleUserInvited.DataObj.Attributes.LastName = requestData.LastName
	reqAppleUserInvited.DataObj.Attributes.ProvisioningAllowed = requestData.ProvisioningAllowedResult
	reqAppleUserInvited.DataObj.Attributes.Roles = requestData.RolesResult
	reqAppleUserInvited.DataObj.Attributes.AllAppsVisible = requestData.AllAppsVisibleResult
	if requestData.AllAppsVisibleResult == false {
		reqAppleUserInvited.DataObj.Relationships = &devconnmanager.VisibleAppObjReqOfApple{}
		reqAppleUserInvited.DataObj.Relationships.VisibleApps = &devconnmanager.VisibleAppsReqOfApple{}
		reqAppleUserInvited.DataObj.Relationships.VisibleApps.DataList = make([]devconnmanager.VisibleAppItemReqOfApple,0)
		if len(requestData.VisibleAppsResult) > 0 {
			for _, itemApp := range requestData.VisibleAppsResult {
				var visibleAppItem devconnmanager.VisibleAppItemReqOfApple
				visibleAppItem.AppType = "apps"
				visibleAppItem.AppAppleId = itemApp
				reqAppleUserInvited.DataObj.Relationships.VisibleApps.DataList = append(reqAppleUserInvited.DataObj.Relationships.VisibleApps.DataList, visibleAppItem)
			}
		}else{
			c.JSON(http.StatusOK, gin.H{
				"error_info":   "参数传递错误，all_apps_visible_result是false情况下，visible_apps_result不能为空",
				"error_code": 1,
				"data": "",
			})
			return
		}
	}
	tokenString := GetTokenStringByTeamId(requestData.TeamId)
	method := "POST"
	resultFromApple := PostToAppleGetInfo(method,_const.APPLE_USER_INVITED_URL,tokenString,&reqAppleUserInvited)
	if !resultFromApple{
		c.JSON(http.StatusOK, gin.H{
			"error_info":   "苹果后台返回数据错误",
			"error_code": 1,
			"data": "",
		})
		return
	}else {
		invitedOrCancel := "1"
		dbInsertResult := devconnmanager.InsertUserInvitedHistoryDB(&requestData,invitedOrCancel)
		if dbInsertResult {
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"error_code": 0,
				"data": "邀请人员成功",
			})
		}else {
			c.JSON(http.StatusOK, gin.H{
				"error_info":   "数据库插入失败",
				"error_code": 2,
				"data": "db insert error",
			})
		}
		return
	}
}
//删除邀请中的成员
func ReqToAppleDeleteUserInvited(method,url,tokenString string) bool{
	client := &http.Client{}
	request, err := http.NewRequest(method, url, nil)
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
	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusNoContent {
		responseByte, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
			return false
		}
		logs.Info(string(responseByte))
		//json.Unmarshal(responseByte, &obj)
		return true
	} else {
		logs.Info("查看返回状态码",response.StatusCode)
		responseByte, _ := ioutil.ReadAll(response.Body)
		logs.Info(string(responseByte))
		return false
	}
}

func UserDeleteFunc(c *gin.Context)  {
	logs.Info("删除企业账号中已有的人员")
	var requestData devconnmanager.UserDeleteReq
	bindQueryError := c.ShouldBindJSON(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", map[string]interface{}{})
		return
	}
	logs.Info(requestData.TeamId,requestData.AppleId)
	tokenString := GetTokenStringByTeamId(requestData.TeamId)
	deletedUrl := _const.APPLE_USER_PERM_EDIT_URL + "/" + requestData.UserId
	method := "DELETE"
	resultFromApple := ReqToAppleDeleteUserInvited(method,deletedUrl,tokenString)
	if !resultFromApple{
		c.JSON(http.StatusOK, gin.H{
			"error_info":   "苹果后台返回数据错误",
			"error_code": 1,
			"data": "",
		})
		return
	}else {
		invitedOrCancel := "2"
		dbInsertResult := devconnmanager.DeleteUserHistoryDB(&requestData,invitedOrCancel)
		if dbInsertResult {
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"error_code": 0,
				"data": "删除已有人员成功",
			})
		}else {
			c.JSON(http.StatusOK, gin.H{
				"error_info":   "数据库插入失败",
				"error_code": 2,
				"data": "db insert error",
			})
		}
		return
	}
}

func UserInvitedDeleteFunc(c *gin.Context)  {
	logs.Info("删除邀请用户")
	var requestData devconnmanager.UserDeleteReq
	bindQueryError := c.ShouldBindJSON(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", map[string]interface{}{})
		return
	}
	logs.Info(requestData.TeamId,requestData.AppleId)
	tokenString := GetTokenStringByTeamId(requestData.TeamId)
	deleteInvitedUrl := _const.APPLE_USER_INVITED_URL + "/" + requestData.UserId
	method := "DELETE"
	resultFromApple := ReqToAppleDeleteUserInvited(method,deleteInvitedUrl,tokenString)
	if !resultFromApple{
		c.JSON(http.StatusOK, gin.H{
			"error_info":   "苹果后台返回数据错误",
			"error_code": 1,
			"data": "",
		})
		return
	}else {
		invitedOrCancel := "0"
		dbInsertResult := devconnmanager.DeleteUserHistoryDB(&requestData,invitedOrCancel)
		if dbInsertResult {
			c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"error_code": 0,
				"data": "删除邀请人员成功",
			})
		}else {
			c.JSON(http.StatusOK, gin.H{
				"error_info":   "数据库插入失败",
				"error_code": 2,
				"data": "db insert error",
			})
		}
		return
	}
}