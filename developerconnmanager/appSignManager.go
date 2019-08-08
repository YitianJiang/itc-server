package developerconnmanager

import (
	"bytes"
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"code.byted.org/gopkg/context"
)
//tos通用处理逻辑
func uploadProfileToTos(profileContent []byte, tosFilePath string) bool {
	var tosBucket = tos.WithAuth(_const.TOS_BUCKET_NAME_JYT, _const.TOS_BUCKET_TOKEN_JYT)
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tosPutClient, err := tos.NewTos(tosBucket)
	err = tosPutClient.PutObject(context, tosFilePath, int64(len(profileContent)), bytes.NewBuffer(profileContent))
	if err != nil {
		logs.Error("%s", "上传tos失败："+err.Error())
		return false
	}
	return true
}

func deleteTosObj(tosFilePath string) bool {
	var tosBucket = tos.WithAuth(_const.TOS_BUCKET_NAME_JYT, _const.TOS_BUCKET_TOKEN_JYT)
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := tos.NewTos(tosBucket)
	err = client.DelObject(context, tosFilePath)
	if err != nil {
		fmt.Println("Error Delete Tos Object:", err)
		return false
	}
	return true
}
//判断是否请求成功的通用逻辑，根据response status code判断
func AssertResStatusCodeOK(statusCode int) bool{
	if statusCode == http.StatusOK || statusCode == http.StatusCreated || statusCode == http.StatusAccepted || statusCode == http.StatusNonAuthoritativeInfo ||
		statusCode == http.StatusNoContent || statusCode == http.StatusResetContent || statusCode == http.StatusPartialContent ||
		statusCode ==http.StatusMultiStatus || statusCode == http.StatusAlreadyReported || statusCode == http.StatusIMUsed {
		return  true
	}else {
		return false
	}
}

//请求苹果的Delete、Get等接口，不需要拿到苹果返回值
func ReqToAppleNoObjMethod(method,url,tokenString string) bool{
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
	if AssertResStatusCodeOK(response.StatusCode) {
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
//objReq,objRes 请传地址
func ReqToAppleHasObjMethod(method,url,tokenString string,objReq,objRes interface{}) bool{
	var rbodyByte *bytes.Reader
	if objReq != nil {
		bodyByte, _ := json.Marshal(objReq)
		logs.Info(string(bodyByte))
		rbodyByte = bytes.NewReader(bodyByte)
	}else {
		rbodyByte = nil
	}
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
		logs.Info("发送请求失败")
		return false
	}
	defer response.Body.Close()
	if !AssertResStatusCodeOK(response.StatusCode) {
		logs.Info("查看返回状态码")
		logs.Info(string(response.StatusCode))
		responseByte, _ := ioutil.ReadAll(response.Body)
		logs.Info("查看苹果的返回值")
		logs.Info(string(responseByte))
		return false
	} else {
		responseByte, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
			return false
		}
		//logs.Info(string(responseByte))
		json.Unmarshal(responseByte, objRes)
		return true
	}
}

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

func DeleteAppAllInfoFromDB(c *gin.Context){
	logs.Info("在DB中删除该app关联的所有信息")
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
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "删除tt_app_account_cert表数据出错", "failed")
		return
	}
	dbErrorInfo := devconnmanager.DeleteAppBundleProfiles(conditionDB)
	utils.RecordError("删除tt_app_account_cert表数据出错：%v", dbErrorInfo)
	if dbErrorInfo != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "删除tt_app_bundleId_profiles表数据出错", "failed")
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

//接口绑定\换绑签名证书接口
func AppBindCert(c *gin.Context){
	logs.Info("对app进行证书换绑")
	var requestData devconnmanager.AppChangeBindCertRequest
	bindJsonError := c.ShouldBindJSON(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindJsonError)
	if bindJsonError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	conditionDB := map[string]interface{}{"id":requestData.AccountCertId}
	appCertChangeMap := map[string]interface{}{"user_name":requestData.UserName}
	if requestData.CertType == _const.CERT_TYPE_IOS_DEV || requestData.CertType == _const.CERT_TYPE_MAC_DEV {
		appCertChangeMap["dev_cert_id"] = requestData.CertId
	}else if requestData.CertType == _const.CERT_TYPE_IOS_DIST || requestData.CertType == _const.CERT_TYPE_MAC_DIST {
		appCertChangeMap["dist_cert_id"] = requestData.CertId
	}else {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数正证书类型不正确", "failed")
		return
	}
	dbError := devconnmanager.UpdateAppAccountCert(conditionDB,appCertChangeMap)
	utils.RecordError("更新tt_app_account_cert表数据出错：%v", dbError)
	if dbError != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "更新tt_app_account_cert表数据出错", "failed")
		return
	}else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "update success",
			"errorCode": 0,
		})
		return
	}
}

//单独Profile创建\更新接口
func UpdateBundleProfilesRelation (bundleId,profileType string,profileId *string) error{
	conditionDb := map[string]interface{}{"bundle_id":bundleId}
	bundleProfilesRelationObj := make(map[string]interface{})
	if profileType == _const.IOS_APP_STORE || profileType ==_const.IOS_APP_INHOUSE || profileType == _const.MAC_APP_STORE{
		bundleProfilesRelationObj["dist_profile_id"] = profileId
	}else if profileType == _const.IOS_APP_DEVELOPMENT || profileType == _const.MAC_APP_DEVELOPMENT {
		bundleProfilesRelationObj["dev_profile_id"] = profileId
	}else {
		bundleProfilesRelationObj["dist_adhoc_profile_id"] = profileId
	}
	dbUpdateErr := devconnmanager.UpdateAppBundleProfiles(conditionDb,bundleProfilesRelationObj)
	return dbUpdateErr
}

func InsertProfileInfoToDB(profileId,profileName,profileType,tosPath string,timeStringDb time.Time) error {
	var profileItem devconnmanager.AppleProfile
	profileItem.ProfileId = profileId
	profileItem.ProfileName = profileName
	profileItem.ProfileExpireDate = timeStringDb
	profileItem.ProfileType = profileType
	profileItem.ProfileDownloadUrl = tosPath
	dbInsertErr := devconnmanager.InsertRecord(&profileItem)
	return dbInsertErr
}

func CreateOrUpdateProfileFromApple(profileName,profileType,bundleidId,certId,token string) *devconnmanager.ProfileDataRes {
	var profileCreateReqObj devconnmanager.ProfileDataReq
	var profileCreateResObj devconnmanager.ProfileDataRes
	profileCreateReqObj.Data.Type = "profiles"
	profileCreateReqObj.Data.Attributes.Name = profileName
	profileCreateReqObj.Data.Attributes.ProfileType = profileType
	profileCreateReqObj.Data.Relationships.BundleId.Data.Type = "bundleIds"
	profileCreateReqObj.Data.Relationships.BundleId.Data.Id = bundleidId
	profileCreateReqObj.Data.Relationships.Certificates.Data = make([]devconnmanager.IdAndTypeItem,1)
	profileCreateReqObj.Data.Relationships.Certificates.Data[0].Type = "certificates"
	profileCreateReqObj.Data.Relationships.Certificates.Data[0].Id = certId
	url := _const.APPLE_PROFILE_MANAGER_URL
	result := ReqToAppleHasObjMethod("POST",url,token,&profileCreateReqObj,&profileCreateResObj)
	if result{
		return &profileCreateResObj
	}else {
		return nil
	}
}

func CreateOrUpdateProfile(c *gin.Context){
	logs.Info("单独Profile创建&更新接口")
	var requestData devconnmanager.ProfileCreateOrUpdateRequest
	bindJsonError := c.ShouldBindJSON(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindJsonError)
	if bindJsonError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	//todo 企业分发类型账号，通知工单处理人进行处理
	if requestData.AccountType == _const.Enterprise {
		logs.Info("企业分发类型账号，通知工单处理人进行处理")
		logs.Info(requestData.AccountName,requestData.AccountType,requestData.BundleId, requestData.UseCertId,
			requestData.ProfileName,requestData.ProfileType,requestData.UserName)
		//todo 发送Lark消息 @zhangmengqi 如上面logs.Info，Lark消息卡片提供account_name、account_type、bundle_id、use_cert_id、profile_name、profile_type、user_name信息
		c.JSON(http.StatusOK, gin.H{
			"message":   "lark success",
			"errorCode": 0,
		})
		return
	}else {
		logs.Info("普通企业类型账号，苹果api自动处理")
		tokenString := GetTokenStringByTeamId(requestData.TeamId)
		if requestData.ProfileId != "" {
			deleteUrl := _const.APPLE_PROFILE_MANAGER_URL + "/" + requestData.ProfileId
			delRes := ReqToAppleNoObjMethod("DELETE",deleteUrl,tokenString)
			if !delRes{
				logs.Info("delete profile fail from apple server")
			}
			//tos中只删除重名的profile，以免出现覆盖问题
			deleteTosObj("appleConnectFile/" + requestData.TeamId + "/Profile/" + requestData.ProfileType + "/" + requestData.ProfileName + ".mobileprovision")
			dbError := devconnmanager.DeleteAppleProfile(map[string]interface{}{"profile_id":requestData.ProfileId})
			utils.RecordError("删除tt_apple_profile失败：%v", dbError)
			if dbError != nil {
				utils.AssembleJsonResponse(c, http.StatusInternalServerError, "删除tt_apple_profile失败", "failed")
				return
			}
			dbUpdateErr := UpdateBundleProfilesRelation(requestData.BundleId,requestData.ProfileType,nil)
			if dbUpdateErr != nil{
				c.JSON(http.StatusInternalServerError, gin.H{
					"message":   "update apple response to tt_app_bundleId_profiles error",
					"errorCode": 1,
				})
				return
			}
		}
		appleResult := CreateOrUpdateProfileFromApple(requestData.ProfileName,requestData.ProfileType,requestData.BundleidId,requestData.UseCertId,tokenString)
		if appleResult != nil{
			decoded, err := base64.StdEncoding.DecodeString(appleResult.Data.Attributes.ProfileContent)
			if err != nil{
				c.JSON(http.StatusInternalServerError, gin.H{
					"message":   "pp file decoded error",
					"errorCode": 2,
				})
				return
			}else {
				pathTos := "appleConnectFile/" + requestData.TeamId + "/Profile/" + requestData.ProfileType + "/" + requestData.ProfileName + ".mobileprovision"
				uploadResult := uploadProfileToTos(decoded,pathTos)
				if !uploadResult{
					c.JSON(http.StatusInternalServerError, gin.H{
						"message":   "upload profile tos error",
						"errorCode": 3,
					})
					return
				}
				timeString := strings.Split(appleResult.Data.Attributes.ExpirationDate,"+")[0]
				exp, _ := time.Parse("2006-01-02T15:04:05", timeString)
				dbInsertErr := InsertProfileInfoToDB(appleResult.Data.Id,requestData.ProfileName,requestData.ProfileType,_const.TOS_BUCKET_URL + pathTos,exp)
				if dbInsertErr != nil{
					c.JSON(http.StatusInternalServerError, gin.H{
						"message":   "insert tt_apple_profile error",
						"errorCode": 4,
					})
					return
				}
				dbUpdateErr := UpdateBundleProfilesRelation(requestData.BundleId,requestData.ProfileType,&appleResult.Data.Id)
				if dbUpdateErr != nil{
					c.JSON(http.StatusInternalServerError, gin.H{
						"message":   "update apple response to tt_app_bundleId_profiles error",
						"errorCode": 5,
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"message":   "success",
					"errorCode": 0,
					"data": appleResult,
				})
				return
			}
		}else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message":   "apple response error",
				"errorCode": 6,
			})
			return
		}
	}
}

func ProfileUploadFunc(c *gin.Context) {
	logs.Info("单独上传profile描述文件接口")
	var requestData devconnmanager.ProfileUploadRequest
	bindError := c.ShouldBind(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindError)
	if bindError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败，查看是否缺少参数", "failed")
		return
	}
	profileFileByteInfo, profileFileFullName := getFileFromRequest(c, "profile_file")
	pathTos := "appleConnectFile/" + requestData.TeamId + "/Profile/" + requestData.ProfileType + "/" + profileFileFullName
	deleteTosObj(pathTos)
	uploadResult := uploadProfileToTos(profileFileByteInfo,pathTos)
	if !uploadResult{
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "upload profile tos error",
			"errorCode": 3,
		})
		return
	}
	exp := utils.GetFileExpireTime(profileFileFullName,".mobileprovision",profileFileByteInfo,requestData.UserName)
	dbInsertErr := InsertProfileInfoToDB(requestData.ProfileId,requestData.ProfileName,requestData.ProfileType,_const.TOS_BUCKET_URL + pathTos,*exp)
	if dbInsertErr != nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "insert tt_apple_profile error",
			"errorCode": 4,
		})
		return
	}
	dbUpdateErr := UpdateBundleProfilesRelation(requestData.BundleId,requestData.ProfileType,&requestData.ProfileId)
	if dbUpdateErr != nil{
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "update apple response to tt_app_bundleId_profiles error",
			"errorCode": 5,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "upload success",
		"errorCode": 0,
		"profile_download_url": _const.TOS_BUCKET_URL + pathTos,
		"profile_expire_date": exp,
	})
	return
}