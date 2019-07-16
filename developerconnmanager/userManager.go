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

type BundleIDCapabilitiesInfo struct {
	ICloudInfo string `json:"icloud_action" binding:"required"`
	PushCapInfo []string `json:"push_list"`
}
func AssertBunldIDInfo(c *gin.Context){
	logs.Info("确认bundle id的能力")
	var requestData BundleIDCapabilitiesInfo
	bindQueryError := c.ShouldBindJSON(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data_push": requestData.PushCapInfo,
		"data":requestData.ICloudInfo,
	})
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
func ReqFromAppleGetUserInfo(url,tokenString string) (bool,*devconnmanager.FromAppleUserInfo){
	var obj devconnmanager.FromAppleUserInfo
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Info("新建request对象失败")
		return false,&obj
	}
	request.Header.Set("Authorization", tokenString)
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送get请求失败")
		return false,&obj
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		logs.Info(string(response.StatusCode))
		return false,&obj
	} else {
		responseByte, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
			return false,&obj
		}
		//logs.Info(string(responseByte))
		json.Unmarshal(responseByte, &obj)
		return true,&obj
	}
}
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
	logs.Info("返回角色列表以供前端筛选")
	var requestData devconnmanager.UserDetailInfoReq
	bindQueryError := c.ShouldBindQuery(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", map[string]interface{}{})
		return
	}
	logs.Info(requestData.TeamId)
	tokenString := GetTokenStringByTeamId(requestData.TeamId)
	resultFromApple,userObjInfo := ReqFromAppleGetUserInfo(_const.APPLE_USER_INFO_URL,tokenString)
	if resultFromApple {
		resDataobj := CreateUsersInfoResData(userObjInfo)
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

