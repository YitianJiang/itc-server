package developerconnmanager

import (
	"bytes"
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/context"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"code.byted.org/yuyilei/bot-api/form"
	"code.byted.org/yuyilei/bot-api/service"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
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
func AssertResStatusCodeOK(statusCode int) bool {
	if statusCode == http.StatusOK || statusCode == http.StatusCreated || statusCode == http.StatusAccepted || statusCode == http.StatusNonAuthoritativeInfo ||
		statusCode == http.StatusNoContent || statusCode == http.StatusResetContent || statusCode == http.StatusPartialContent ||
		statusCode == http.StatusMultiStatus || statusCode == http.StatusAlreadyReported || statusCode == http.StatusIMUsed {
		return true
	} else {
		return false
	}
}

//请求苹果的Delete、Get等接口，不需要拿到苹果返回值
func ReqToAppleNoObjMethod(method, url, tokenString string) bool {
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
		logs.Info("查看返回状态码", response.StatusCode)
		responseByte, _ := ioutil.ReadAll(response.Body)
		logs.Info(string(responseByte))
		return false
	}
}

//objReq,objRes 请传地址
func ReqToAppleHasObjMethod(method, url, tokenString string, objReq, objRes interface{}) bool {
	var rbodyByte *bytes.Reader
	if objReq != nil {
		bodyByte, _ := json.Marshal(objReq)
		logs.Info(string(bodyByte))
		rbodyByte = bytes.NewReader(bodyByte)
	} else {
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
	AppType string `form:"app_type" json:"app_type" binding:"required"`
}

type CapabilitiesInfo struct {
	SelectCapabilitiesInfo   []string            `json:"select_capabilities"`
	SettingsCapabilitiesInfo map[string][]string `json:"settings_capabilities"`
}

//列表联查中间结构体--bundle和appName关联
type BundleResortStruct struct {
	BundleInfo devconnmanager.BundleProfileCert
	AppName    string
}

func GetBundleIdCapabilitiesInfo(c *gin.Context) {
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
	if requestData.AppType == "iOS" {
		responseData.SelectCapabilitiesInfo = _const.IOSSelectCapabilities
		responseData.SettingsCapabilitiesInfo[_const.DATA_PROTECTION] = _const.ProtectionSettings
	} else {
		responseData.SelectCapabilitiesInfo = _const.MacSelectCapabilities
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      responseData,
	})
}

/**
API 3-1：根据业务线appid，返回app相关list
*/
func GetAppSignListDetailInfo(c *gin.Context) {
	logs.Info("获取app签名信息List")
	var requestInfo devconnmanager.AppSignListRequest
	err := c.ShouldBindQuery(&requestInfo)
	if err != nil {
		utils.RecordError("AppSignList请求参数绑定失败,", err)
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "")
		return
	}
	//权限判断，showType为1（超级权限），showType为2(dev权限），showType为0（无权限）
	showType := getPermType(c, requestInfo.Username, requestInfo.TeamId)
	if showType == 0 {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "无权限查看！", "")
		return
	}
	//根据app_id和team_id获取appName基本信息以及证书信息
	var cQueryResult []devconnmanager.APPandCert
	sql := "select aac.app_name,aac.app_type,aac.id as app_acount_id,aac.team_id,aac.account_verify_status,aac.account_verify_user," +
		"ac.cert_id,ac.cert_type,ac.cert_expire_date,ac.cert_download_url,ac.priv_key_url from tt_app_account_cert aac, tt_apple_certificate ac" +
		" where aac.app_id = '" + requestInfo.AppId + "' and aac.team_id = '" + requestInfo.TeamId + "' and aac.deleted_at IS NULL and (aac.dev_cert_id = ac.cert_id or aac.dist_cert_id = ac.cert_id)" +
		" and ac.deleted_at IS NULL "
	query_c := devconnmanager.QueryWithSql(sql, &cQueryResult)
	if query_c != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "查询失败", "")
		return
	} else if len(cQueryResult) == 0 {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "未查询到该app_id在本账号下的信息！", "")
		return
	}
	//以appName为单位重组基本信息，appappNameMap为最终结果的map
	appNameList := "("
	appNameMap := make(map[string]devconnmanager.APPSignManagerInfo)
	for _, fqr := range cQueryResult {
		if v, ok := appNameMap[fqr.AppName]; ok {
			packCertSection(&fqr, showType, &v.CertSection)
			appNameMap[fqr.AppName] = v
		} else {
			var appInfo devconnmanager.APPSignManagerInfo
			packAppNameInfo(&appInfo, &fqr, showType)
			appNameMap[fqr.AppName] = appInfo
			appNameList += "'" + fqr.AppName + "',"
		}
	}
	appNameList = strings.TrimSuffix(appNameList, ",")
	appNameList += ")"
	//根据app_id和app_name获取bundleid信息+profile信息
	var bQueryResult []devconnmanager.APPandBundle
	sql_c := "select abp.app_name,abp.bundle_id as bundle_id_index,abp.bundleid_isdel as bundle_id_is_del,abp.push_cert_id,ap.profile_id,ap.profile_name,ap.profile_expire_date,ap.profile_type,ap.profile_download_url,ab.*" +
		" from tt_app_bundleId_profiles abp, tt_apple_bundleId ab, tt_apple_profile ap " +
		"where abp.app_id = '" + requestInfo.AppId + "' and abp.app_name in " + appNameList + " and abp.bundle_id = ab.bundle_id and (abp.dev_profile_id = ap.profile_id or abp.dist_profile_id = ap.profile_id) " +
		"and abp.deleted_at IS NULL and ab.deleted_at IS NULL and ap.deleted_at IS NULL"
	query_b := devconnmanager.QueryWithSql(sql_c, &bQueryResult)
	if query_b != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "查询失败", "")
		return
	}
	//以bundle为单位重组信息，appName作为附加信息
	bundleMap := make(map[string]BundleResortStruct)
	for _, bqr := range bQueryResult {
		if v, ok := bundleMap[bqr.BundleId]; ok {
			packProfileSection(&bqr, showType, &v.BundleInfo.ProfileCertSection)
			bundleMap[bqr.BundleId] = v
		} else {
			var bundleResort BundleResortStruct
			bundleResort.AppName = bqr.AppName
			bundles := packeBundleProfileCert(c, &bqr, showType)
			//查询push证书失败
			if bundles == nil {
				return
			}
			bundleResort.BundleInfo = (*bundles)
			bundleMap[bqr.BundleId] = bundleResort
		}
	}
	//重组最终结果，appNameMap同bundleMap
	for _, bundleDetail := range bundleMap {
		//1--补全profile中的cert_id
		if bundleDetail.BundleInfo.ProfileCertSection.DistProfile.ProfileId != "" {
			bundleDetail.BundleInfo.ProfileCertSection.DistProfile.UserCertId = appNameMap[bundleDetail.AppName].CertSection.DistCert.CertId
		}
		if bundleDetail.BundleInfo.ProfileCertSection.DevProfile.ProfileId != "" {
			bundleDetail.BundleInfo.ProfileCertSection.DevProfile.UserCertId = appNameMap[bundleDetail.AppName].CertSection.DevCert.CertId
		}
		//2--bundle信息整合到appNameMap中
		appInfo := appNameMap[bundleDetail.AppName]
		appInfo.BundleProfileCertSection = append(appInfo.BundleProfileCertSection, bundleDetail.BundleInfo)
		appNameMap[bundleDetail.AppName] = appInfo
	}
	//结果为appNameMap的value集合
	result := make([]devconnmanager.APPSignManagerInfo, 0)
	for _, info := range appNameMap {
		result = append(result, info)
	}
	utils.AssembleJsonResponse(c, http.StatusOK, "success", result)
}

func DeleteAppAllInfoFromDB(c *gin.Context) {
	logs.Info("在DB中删除该app关联的所有信息")
	var requestData devconnmanager.DeleteAppAllInfoRequest
	bindJsonError := c.ShouldBindQuery(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindJsonError)
	if bindJsonError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	conditionDB := map[string]interface{}{"app_id": requestData.AppId, "app_name": requestData.AppName}
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
	logs.Info("创建app并绑定或换绑账号")
	var requestData devconnmanager.CreateAppBindAccountRequest
	//获取请求参数
	bindJsonError := c.ShouldBindJSON(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindJsonError)
	if bindJsonError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	logs.Info("request:%v", requestData)

	// 根据app_id和app_name执行update，如果返回的操作行数为0，则插入数据
	conditions := map[string]interface{}{
		"app_id":   requestData.AppId,
		"app_name": requestData.AppName,
	}
	appAccountCertMap := map[string]interface{}{
		"team_id":               requestData.TeamId,
		"dev_cert_id":           "",
		"dist_cert_id":          "",
		"account_verify_status": "0",
		"app_type":              requestData.AppType,
		"user_name":             requestData.UserName,
	}

	accountCertInfos := devconnmanager.QueryAppAccountCert(conditions)
	var appAccountCert devconnmanager.AppAccountCert

	if len(*accountCertInfos) == 0 {
		//插入数据
		appAccountCert.TeamId = requestData.TeamId
		appAccountCert.DevCertId = ""
		appAccountCert.DistCertId = ""
		appAccountCert.AccountVerifyStatus = "0"
		appAccountCert.AppType = requestData.AppType
		appAccountCert.UserName = requestData.UserName
		appAccountCert.AppId = requestData.AppId
		appAccountCert.AppName = requestData.AppName
		err := devconnmanager.InsertRecord(&appAccountCert)
		if err != nil {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "数据库插入失败", "failed")
			return
		}
	} else if len(*accountCertInfos) == 1 {
		//更新数据
		conditions := map[string]interface{}{"id": (*accountCertInfos)[0].ID}
		err, returnModel := devconnmanager.UpdateAppAccountCertAndGetModelByMap(conditions, appAccountCertMap)
		if err != nil {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "数据库更新失败", "failed")
			return
		}
		appAccountCert = *returnModel
	} else {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "存在多条app_id和app_name都相同的数据，无法更新", "failed")
		return
	}
	//调用根据资源获取admin人员信息的接口，根据该接口获取需要发送审批消息的用户list
	//todo 暂时写死admin list
	//var userList = utils.GetAccountAdminList(requestData.TeamId)
	var userList = &[]string{"zhangmengqi.muki"} //,"fanjuan.xqp"
	//lark消息生成并批量发送 使用go协程
	botService := service.BotService{}
	botService.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
	cardInfos := generateCardOfApproveBindAccount(&appAccountCert)

	for _, adminEmailPrefix := range *userList {
		go alertApproveToUser(adminEmailPrefix, appAccountCert.ID, cardInfos, &botService)
	}

	logs.Info("%v", userList)
	utils.AssembleJsonResponse(c, _const.SUCCESS, "success", appAccountCert)
	return
}

func CreateBundleProfile(c *gin.Context) {
	logs.Info("创建/更新/恢复bundle id")
	var requestData devconnmanager.CreateBundleProfileRequest
	//获取请求参数
	bindJsonError := c.ShouldBindJSON(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindJsonError)
	if bindJsonError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	logs.Info("request:%v", requestData)

	if requestData.AccountType == "Enterprise" {
		//todo 走工单逻辑
	}
	switch requestData.BundleIdIsDel {
	case "0":
		//创建/更新
		if requestData.BundleIdId == "" {
			//新建bundle id
			if requestData.DevProfileInfo == (devconnmanager.ProfileInfo{}) && requestData.DistProfileInfo == (devconnmanager.ProfileInfo{}) {
				utils.AssembleJsonResponse(c, http.StatusBadRequest, "新建bundle id时需要传送至少一个profile info", nil)
				return
			}

			//在苹果后台生成bundle id
			tokenString := GetTokenStringByTeamId(requestData.TeamId)
			var createBundleIdResponseFromApple devconnmanager.CreateBundleIdResponse
			if !createBundleInApple(tokenString, &requestData.BundleIdInfo, &createBundleIdResponseFromApple) {
				utils.AssembleJsonResponse(c, http.StatusInternalServerError, "在苹果后台创建bundle id失败", nil)
				return
			}

			//更新数据库
			resBool, bundleIdRecordId := updateDatabaseAfterCreateBundleId(&requestData, &createBundleIdResponseFromApple, c)
			if !resBool {
				return
			}

			//在苹果后台打开能力
			capabilityNum := len(requestData.EnableCapabilitiesChange)
			var wg sync.WaitGroup
			wg.Add(capabilityNum)
			successChannel := make(chan string, capabilityNum)
			failChannel := make(chan string, capabilityNum)

			for _, capability := range requestData.EnableCapabilitiesChange {
				//go openCapabilityInApple(capability,createBundleIdResponseFromApple.Data.Id,tokenString,failChannel,successChannel)
				go func() {
					var openBundleIdCapabilityRequest devconnmanager.OpenBundleIdCapabilityRequest
					openBundleIdCapabilityRequest.Data.Type = "bundleIdCapabilities"
					openBundleIdCapabilityRequest.Data.Attributes.CapabilityType = capability
					openBundleIdCapabilityRequest.Data.Relationships.BundleId.Data.Type = "bundleIds"
					openBundleIdCapabilityRequest.Data.Relationships.BundleId.Data.Id = createBundleIdResponseFromApple.Data.Id
					if !ReqToAppleHasObjMethod("POST", _const.APPLE_BUNDLE_ID_MANAGER_URL, tokenString, &openBundleIdCapabilityRequest, nil) {
						failChannel <- capability
					} else {
						successChannel <- capability
					}
					wg.Done()
				}()
			}
			wg.Wait()
			//更新能力表
			err, failedList := updateDBAfterOpenCapabilities(failChannel, successChannel, bundleIdRecordId)
			if err != nil {
				utils.AssembleJsonResponse(c, http.StatusInternalServerError, "更新bundle id能力表失败", nil)
				return
			}
			if len(*failedList) > 0 {
				utils.AssembleJsonResponse(c, http.StatusInternalServerError, "苹果后台为bundle id更新能力失败", failedList)
				return
			}
			//todo 在苹果后台新建描述文件
			if requestData.DevProfileInfo != (devconnmanager.ProfileInfo{}) {

			}

			//todo 更新数据库

		} else {
			//todo 更新bundle id
		}
	case "1":
		//todo 恢复bundle id
	}
}

func createBundleInApple(tokenString string, bundleIdInfo *devconnmanager.BundleIdInfo, res *devconnmanager.CreateBundleIdResponse) bool {
	var createBundleIdRequest devconnmanager.CreateBundleIdRequest
	//var createBundleIdResponseFromApple devconnmanager.CreateBundleIdResponse
	createBundleIdRequest.Data.Type = "bundleIds"
	createBundleIdRequest.Data.Attributes.Name = bundleIdInfo.BundleIdName
	createBundleIdRequest.Data.Attributes.Identifier = bundleIdInfo.BundleId
	createBundleIdRequest.Data.Attributes.Platform = bundleIdInfo.BundleType
	if !ReqToAppleHasObjMethod("POST", _const.APPLE_BUNDLE_ID_MANAGER_URL, tokenString, &createBundleIdRequest, res) {
		return false
	}
	return true
}

func updateDatabaseAfterCreateBundleId(requestData *devconnmanager.CreateBundleProfileRequest, res *devconnmanager.CreateBundleIdResponse, c *gin.Context) (bool, uint) {
	var appBundleProfiles devconnmanager.AppBundleProfiles
	appBundleProfiles.BundleId = requestData.BundleId
	appBundleProfiles.BundleidId = res.Data.Id
	appBundleProfiles.BundleidIsdel = "0"
	appBundleProfiles.UserName = requestData.UserName
	err := devconnmanager.InsertRecord(&appBundleProfiles)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "bundle id数据插入数据库失败", nil)
		return false, 0
	}
	appleBundleId := devconnmanager.AppleBundleId{}
	appleBundleIdElem := reflect.ValueOf(&appleBundleId).Elem()

	for _, capability := range res.Data.Relationships.BundleIdCapabilities.Data {
		begin := strings.Index(capability.Id, "_")
		end := strings.LastIndex(capability.Id, "_")
		if begin == -1 || end == -1 {
			logs.Error("")
			continue
		}
		capabilityName := capability.Id[begin:end]
		//写入数据库
		field := appleBundleIdElem.FieldByName(capabilityName)
		field.SetString("1")
	}

	logs.Info("bundle id能力对象：%v", appleBundleId)
	err = devconnmanager.InsertRecord(&appleBundleId)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "bundle id能力数据插入数据库失败", nil)
		return false, 0
	}
	return true, appleBundleId.ID
}

func updateDBAfterOpenCapabilities(failChannel chan string, successChannel chan string, bundleIdRecordId uint) (error, *[]string) {
	close(failChannel)
	close(successChannel)

	var failedList []string

	changedCapabilities := make(map[string]interface{})
	for capability := range successChannel {
		changedCapabilities[capability] = "1"
	}
	err := devconnmanager.UpdateAppleBundleId(map[string]interface{}{"id": bundleIdRecordId}, changedCapabilities)
	if err != nil {
		return err, &failedList
	}

	//存在更新失败的能力
	if len(failChannel) != 0 {
		for capability := range failChannel {
			failedList = append(failedList, capability)
		}
	}
	return nil, &failedList
}

/*func openCapabilityInApple(capability string,bundleIdId string,tokenString string,failChannel chan string,successChannel chan string,) {
	var openBundleIdCapabilityRequest devconnmanager.OpenBundleIdCapabilityRequest
	openBundleIdCapabilityRequest.Data.Type="bundleIdCapabilities"
	openBundleIdCapabilityRequest.Data.Attributes.CapabilityType=capability
	openBundleIdCapabilityRequest.Data.Relationships.BundleId.Data.Type="bundleIds"
	openBundleIdCapabilityRequest.Data.Relationships.BundleId.Data.Id=bundleIdId
	if !ReqToAppleHasObjMethod("POST", _const.APPLE_BUNDLE_ID_MANAGER_URL, tokenString, &openBundleIdCapabilityRequest, nil){
		failChannel<-capability
		return
	}
	successChannel<-capability
}*/

func alertApproveToUser(adminEmailPrefix string, id uint, cardInfos *[][]form.CardElementForm, botService *service.BotService) {
	cardActions := generateActionsOfApproveBindAccount(id, adminEmailPrefix)
	err := sendIOSCertLarkMessage(cardInfos, cardActions, adminEmailPrefix, botService)
	utils.RecordError("发送lark消息错误", err)
}

func ApproveAppBindAccountFeedback(c *gin.Context) {
	var requestData devconnmanager.ApproveAppBindAccountParamFromLark
	err := c.ShouldBindJSON(&requestData)
	utils.RecordError("绑定post请求body出错：%v", err)
	if err != nil {
		utils.AssembleJsonResponseWithStatusCode(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	logs.Info("请求参数：%v", requestData)
	accountCertInfos := devconnmanager.QueryAppAccountCert(map[string]interface{}{"id": requestData.AppAccountCertId})
	if len(*accountCertInfos) < 1 {
		logs.Error("appAccountCertId出错，找不到对应记录")
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "appAccountCertId出错，找不到对应记录", "failed")
		return
	}

	//logs.Info("accountCertInfos:%v", *accountCertInfos)
	switch (*accountCertInfos)[0].AccountVerifyStatus {
	case "0":
		//还未进行审核
		switch requestData.IsApproved {
		case -1:
			if !devconnmanager.UpdateAppAccountCertByMap(map[string]interface{}{"id": requestData.AppAccountCertId}, map[string]interface{}{"account_verify_status": -1, "account_verify_user": requestData.UserName}) {
				utils.AssembleJsonResponseWithStatusCode(c, http.StatusInternalServerError, "审核失败:内部服务器错误", nil)
				return
			}
			utils.AssembleJsonResponse(c, _const.SUCCESS, "success", "审核成功：绑定请求已被拒绝")
			return
		case 1:
			err, appAccountCert := devconnmanager.UpdateAppAccountCertAndGetModelByMap(map[string]interface{}{"id": requestData.AppAccountCertId}, map[string]interface{}{"account_verify_status": 1, "account_verify_user": requestData.UserName})
			if err != nil {
				utils.AssembleJsonResponseWithStatusCode(c, http.StatusInternalServerError, "审核失败:内部服务器错误", nil)
				return
			}
			//给user赋予权限
			teamId := strings.ToLower(appAccountCert.TeamId) + "_space_account"
			if !utils.GiveUsersPermission(&[]string{requestData.UserName}, teamId, &[]string{"all_cert_manager"}) {
				utils.AssembleJsonResponseWithStatusCode(c, http.StatusInternalServerError, "审核失败:权限赋予失败", nil)
			}
			utils.AssembleJsonResponse(c, _const.SUCCESS, "success", "审核成功：绑定请求已被通过")
			return
		}
	case "1":
		//已经通过审核
		switch requestData.IsApproved {
		case -1:
			utils.AssembleJsonResponseWithStatusCode(c, http.StatusBadRequest, "审核失败：绑定请求已被审核过[通过]", "failed")
			return
		case 1:
			utils.AssembleJsonResponseWithStatusCode(c, http.StatusBadRequest, "审核失败：绑定请求已被审核过[通过]", "failed")
			return
		}
	case "-1":
		//审核已经被拒绝
		switch requestData.IsApproved {
		case -1:
			utils.AssembleJsonResponseWithStatusCode(c, http.StatusBadRequest, "审核失败：绑定请求已被审核过[拒绝]", "failed")
			return
		case 1:
			utils.AssembleJsonResponseWithStatusCode(c, http.StatusBadRequest, "审核失败，绑定请求已被审核过[拒绝]", "failed")
			return
		}
	}
}

//profile删除接口
func DeleteProfile(c *gin.Context) {
	logs.Info("删除单个Profile")
	var deleteRequest devconnmanager.ProfileDeleteRequest
	err := c.ShouldBindQuery(&deleteRequest)
	if err != nil {
		utils.RecordError("delete profile error:", err)
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "")
		return
	}
	//企业分发账号删除，提工单处理---此if内容，待apple openAPI ready，可删除
	if deleteRequest.AccountType == _const.Enterprise {
		//向负责人发送lark消息
		abot := service.BotService{}
		abot.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
		appleUrl := utils.APPLE_DELETE_PROFILE_URL + deleteRequest.ProfileId
		if deleteRequest.Operator == "" {
			deleteRequest.Operator = utils.CreateCertPrincipal
		}
		param := map[string]interface{}{
			"profile_id": deleteRequest.ProfileId,
			"username":   deleteRequest.Operator,
		}
		cardElementForms := generateCardOfProfileDelete(&deleteRequest, appleUrl)
		cardActions := generateActionOfProfileDelete(&param)
		err := sendIOSCertLarkMessage(cardElementForms, cardActions, deleteRequest.Operator, &abot)
		if err != nil {
			utils.RecordError("发送lark消息通知负责人删除证书失败，", err)
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "发送lark消息通知负责人删除证书失败", "")
			return
		}
		//profile tos+db删除
		updateData := map[string]interface{}{
			"deleted_at":time.Now(),
		}
		if isDel := deleteProfileDBandTos(c, deleteRequest.ProfileId, deleteRequest.ProfileName, deleteRequest.ProfileType, deleteRequest.TeamId, deleteRequest.BundleId,updateData,deleteRequest.UserName); !isDel {
			return
		}
		utils.AssembleJsonResponse(c, 0, "success", "")
		return
	}
	//普通账号处理
	deleteUrl := _const.APPLE_PROFILE_MANAGER_URL + "/" + deleteRequest.ProfileId
	deleteInfoFromAppleApi(deleteRequest.TeamId,deleteUrl)
	updateData := map[string]interface{}{
		"deleted_at":time.Now(),
		"op_user":deleteRequest.UserName,
	}
	if isDel := deleteProfileDBandTos(c, deleteRequest.ProfileId, deleteRequest.ProfileName, deleteRequest.ProfileType, deleteRequest.TeamId, deleteRequest.BundleId,updateData,deleteRequest.UserName); !isDel {
		return
	}
	utils.AssembleJsonResponse(c, 0, "success", "")
}

//证书删除异步确认
func AsynProfileDeleteFeedback(c *gin.Context) {
	var feedbackInfo devconnmanager.DelProfileFeedback
	err := c.ShouldBindJSON(&feedbackInfo)
	if err != nil {
		utils.RecordError("请求参数绑定失败！", err)
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "")
		return
	}
	//更新对应cert_id的op_user信息
	condition := map[string]interface{}{
		"profile_id": feedbackInfo.CustomerJson.ProfileId,
	}
	updateInfo := map[string]interface{}{
		"op_user":    feedbackInfo.CustomerJson.UserName,
	}
	errInfo := devconnmanager.UpdateAppleProfile(condition, updateInfo)
	if errInfo != nil {
		utils.RecordError("异步更新删除信息操作人失败，证书ID："+feedbackInfo.CustomerJson.ProfileId, errInfo)
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "数据库异步更新删除信息操作人失败", "")
		return
	}
	utils.AssembleJsonResponse(c, 0, "success", "")
	return
}

//接口绑定\换绑签名证书接口
func AppBindCert(c *gin.Context) {
	logs.Info("对app进行证书换绑")
	var requestData devconnmanager.AppChangeBindCertRequest
	bindJsonError := c.ShouldBindJSON(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindJsonError)
	if bindJsonError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	conditionDB := map[string]interface{}{"id": requestData.AccountCertId}
	appCertChangeMap := map[string]interface{}{"user_name": requestData.UserName}
	if requestData.CertType == _const.CERT_TYPE_IOS_DEV || requestData.CertType == _const.CERT_TYPE_MAC_DEV {
		appCertChangeMap["dev_cert_id"] = requestData.CertId
	} else if requestData.CertType == _const.CERT_TYPE_IOS_DIST || requestData.CertType == _const.CERT_TYPE_MAC_DIST {
		appCertChangeMap["dist_cert_id"] = requestData.CertId
	} else {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数正证书类型不正确", "failed")
		return
	}
	dbError := devconnmanager.UpdateAppAccountCert(conditionDB, appCertChangeMap)
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

//单独Profile创建\更新接口
func UpdateBundleProfilesRelation(bundleId, profileType string, profileId *string,username string) error {
	conditionDb := map[string]interface{}{"bundle_id": bundleId}
	bundleProfilesRelationObj := make(map[string]interface{})
	if profileType == _const.IOS_APP_STORE || profileType == _const.IOS_APP_INHOUSE || profileType == _const.MAC_APP_STORE {
		bundleProfilesRelationObj["dist_profile_id"] = profileId
	} else if profileType == _const.IOS_APP_DEVELOPMENT || profileType == _const.MAC_APP_DEVELOPMENT {
		bundleProfilesRelationObj["dev_profile_id"] = profileId
	} else {
		bundleProfilesRelationObj["dist_adhoc_profile_id"] = profileId
	}
	bundleProfilesRelationObj["user_name"] = username
	dbUpdateErr := devconnmanager.UpdateAppBundleProfiles(conditionDb, bundleProfilesRelationObj)
	return dbUpdateErr
}
//bundleid_id删除接口，is_del=1保配删，is_del=2完全删除
func DeleteBundleid(c *gin.Context)  {
	logs.Info("bundleId删除")
	var delRequest devconnmanager.BundleDeleteRequest
	err := c.ShouldBindQuery(&delRequest)
	if err!= nil {
		utils.RecordError("请求参数绑定失败，err:",err)
		utils.AssembleJsonResponse(c,http.StatusBadRequest,"请求参数绑定失败","")
		return
	}
	if isDelOk := checkIsDel(c,delRequest.IsDel); !isDelOk{
		return
	}

	bundleid := devconnmanager.QueryAppleBundleId(map[string]interface{}{
		"bundleid_id":delRequest.BundleidId,
	})
	if bundleid == nil || len(*bundleid)==0{
		utils.AssembleJsonResponse(c,http.StatusInternalServerError,"查询bundleid_id失败","")
		return
	}
	//获取bundleID下profile信息
	profileIdList := make([]interface{},0)
	var profiles *[]devconnmanager.AppleProfile
	if delRequest.DevProfileId != "" || delRequest.DisProfileId != "" {
		if delRequest.DevProfileId != "" {
			profileIdList = append(profileIdList,delRequest.DevProfileId)
		}
		if delRequest.DisProfileId != "" {
			profileIdList = append(profileIdList,delRequest.DisProfileId)
		}
	}
	if len(profileIdList) >0 {
		profiles = devconnmanager.QueryAppleProfileWithList("profile_id",&profileIdList)
	}

	//发起工单删，待apple openAPI ready，此if下可删除
	if delRequest.AccountType == _const.Enterprise {
		abot := service.BotService{}
		abot.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
		url := utils.APPLE_DELETE_BUNDLE_URL +delRequest.BundleidId
		param := map[string]interface{}{
			"bundleid_id":delRequest.BundleidId,
			"username":delRequest.UserName,
			"dist_profile_id":delRequest.DisProfileId,
			"dev_profile_id":delRequest.DevProfileId,
			"is_del":delRequest.IsDel,
		}
		cardContent := generateCardOfBundleDelete(&delRequest,url,(*bundleid)[0].BundleId)
		cardAction := generateActionOfBundleDelete(&param)
		if delRequest.Operator == "" {
			delRequest.Operator = utils.CreateCertPrincipal
		}
		sendErr :=sendIOSCertLarkMessage(cardContent,cardAction,delRequest.Operator,&abot)
		if sendErr != nil {
			utils.AssembleJsonResponse(c,http.StatusInternalServerError,"发送工单lark消息失败","")
			return
		}
		//删除profile信息
		if profiles!= nil && len(*profiles)>0 {
			updateData := map[string]interface{}{
				"deleted_at":time.Now(),
			}
			for _,profile :=range (*profiles){
				deleteProfileDBandTos(c,profile.ProfileId,profile.ProfileName,profile.ProfileType,delRequest.TeamId,(*bundleid)[0].BundleId,updateData,delRequest.UserName)
			}
		}
		//bundleid删除
		queryData := map[string]interface{}{
			"bundleid_id":delRequest.BundleidId,
		}
		if delRequest.IsDel == "1" {
			updateData := map[string]interface{}{
				"bundleid_isdel":"1",
			}
			updateErr := devconnmanager.UpdateAppBundleProfiles(queryData,updateData)
			if updateErr != nil {
				utils.AssembleJsonResponse(c,http.StatusInternalServerError,"bundleId软删除失败","")
				return
			}
			utils.AssembleJsonResponse(c,http.StatusOK,"bundleId软删除成功","")
		}else {
			delErr1 := devconnmanager.DeleteAppBundleProfiles(queryData)
			delErr2 := devconnmanager.DeleteAppleBundleId(queryData)
			if delErr1 != nil || delErr2 != nil {
				utils.AssembleJsonResponse(c, http.StatusInternalServerError, "bundleId完全删除失败", "")
				return
			}
			utils.AssembleJsonResponse(c, http.StatusOK, "bundleId完全删除成功", "")
		}
		return
	}

	//apple openAPI删除
	//profile删除
	if profiles != nil && len(*profiles)>0 {
		for _,profile := range (*profiles) {
			updateData := map[string]interface{}{
				"deleted_at":time.Now(),
				"op_user":delRequest.UserName,
			}
			deleteUrl := _const.APPLE_PROFILE_MANAGER_URL + "/" + profile.ProfileId
			go deleteInfoFromAppleApi(delRequest.TeamId,deleteUrl)
			if ok := deleteProfileDBandTos(c,profile.ProfileId,profile.ProfileName,profile.ProfileType,delRequest.TeamId,(*bundleid)[0].BundleId,updateData,delRequest.UserName);!ok {
				return
			}
		}
	}
	//bundleid删除
	bundleAppleUrl := _const.APPLE_BUNDLE_MANAGER_URL +delRequest.BundleidId
	if ok := deleteInfoFromAppleApi(delRequest.TeamId,bundleAppleUrl); !ok {
		utils.RecordError("delete bundleid failed from apple open api",nil)
		utils.AssembleJsonResponse(c,http.StatusInternalServerError,"bundleid_id苹果后台删除失败","")
		return
	}

	queryData := map[string]interface{}{
		"bundleid_id":delRequest.BundleidId,
	}
	if delRequest.IsDel == "1" {
		updateData := map[string]interface{}{
			"bundleid_isdel":"1",
			"user_name":delRequest.UserName,
		}
		updateErr := devconnmanager.UpdateAppBundleProfiles(queryData,updateData)
		if updateErr != nil {
			utils.AssembleJsonResponse(c,http.StatusInternalServerError,"bundleId软删除失败","")
			return
		}
		utils.AssembleJsonResponse(c,http.StatusOK,"bundleId软删除成功","")
	}else {
		//删除并更新操作人
		updateData := map[string]interface{}{
			"deleted_at":time.Now(),
			"user_name":delRequest.UserName,
		}
		delErr1 := devconnmanager.UpdateAppBundleProfiles(queryData,updateData)
		delErr2 := devconnmanager.UpdateAppleBundleId(queryData,updateData)
		if delErr1 != nil || delErr2 != nil {
			utils.AssembleJsonResponse(c,http.StatusInternalServerError,"bundleId完全删除失败","")
			return
		}
		utils.AssembleJsonResponse(c,http.StatusOK,"bundleId完全删除成功","")
	}
}

//bundle删除异步确认
func AsynBundleDeleteFeedback(c *gin.Context)  {
	var feedbackInfo devconnmanager.DelBundleFeedback
	err := c.ShouldBindJSON(&feedbackInfo)
	if err != nil {
		utils.RecordError("请求参数绑定失败！", err)
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "")
		return
	}
	//更新对应profile_id的op_user信息
	profileIdList := make([]interface{},0)
	if feedbackInfo.CustomerJson.DevProfileId != "" {
		profileIdList = append(profileIdList,feedbackInfo.CustomerJson.DevProfileId)
	}
	if feedbackInfo.CustomerJson.DistProfileId != "" {
		profileIdList = append(profileIdList,feedbackInfo.CustomerJson.DistProfileId)
	}
	var errInfo,errInfo3 error
	if len(profileIdList)>0 {
		updateInfo := map[string]interface{}{
			"op_user":    feedbackInfo.CustomerJson.UserName,
		}
		errInfo = devconnmanager.UpdateAppleProfileBatch("profile_id",&profileIdList, updateInfo)
	}
	condition := map[string]interface{}{
		"bundleid_id":feedbackInfo.CustomerJson.BundleIdId,
	}
	updateInfo := map[string]interface{}{
		"user_name": feedbackInfo.CustomerJson.UserName,
	}
	errInfo2 := devconnmanager.UpdateAppBundleProfiles(condition,updateInfo)
	if feedbackInfo.CustomerJson.IsDel == "2" {
		errInfo3 = devconnmanager.UpdateAppleBundleId(condition,updateInfo)
	}
	if errInfo != nil || errInfo2 != nil || errInfo3 != nil {
		utils.RecordError("异步更新bundleid删除信息操作人失败,bundleID："+feedbackInfo.CustomerJson.BundleIdId, errInfo)
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "数据库异步更新删除信息操作人失败", "")
		return
	}
	utils.AssembleJsonResponse(c, 0, "success", "")
	return
}

func InsertProfileInfoToDB(profileId, profileName, profileType, tosPath , opUser string, timeStringDb time.Time) error {
	var profileItem devconnmanager.AppleProfile
	profileItem.ProfileId = profileId
	profileItem.ProfileName = profileName
	profileItem.ProfileExpireDate = timeStringDb
	profileItem.ProfileType = profileType
	profileItem.ProfileDownloadUrl = tosPath
	profileItem.OpUser = opUser
	dbInsertErr := devconnmanager.InsertRecord(&profileItem)
	return dbInsertErr
}

func CreateOrUpdateProfileFromApple(profileName, profileType, bundleidId, certId, token string) *devconnmanager.ProfileDataRes {
	var profileCreateReqObj devconnmanager.ProfileDataReq
	var profileCreateResObj devconnmanager.ProfileDataRes
	profileCreateReqObj.Data.Type = "profiles"
	profileCreateReqObj.Data.Attributes.Name = profileName
	profileCreateReqObj.Data.Attributes.ProfileType = profileType
	profileCreateReqObj.Data.Relationships.BundleId.Data.Type = "bundleIds"
	profileCreateReqObj.Data.Relationships.BundleId.Data.Id = bundleidId
	profileCreateReqObj.Data.Relationships.Certificates.Data = make([]devconnmanager.IdAndTypeItem, 1)
	profileCreateReqObj.Data.Relationships.Certificates.Data[0].Type = "certificates"
	profileCreateReqObj.Data.Relationships.Certificates.Data[0].Id = certId
	if profileType == _const.IOS_APP_DEVELOPMENT || profileType == _const.MAC_APP_DEVELOPMENT {
		var devicesResObj devconnmanager.DevicesDataRes
		platformType := "IOS"
		if profileType == _const.MAC_APP_DEVELOPMENT{
			platformType = "MAC_OS"
		}
		deviceResult := GetAllEnableDevicesObj(platformType,"ENABLED",token,&devicesResObj)
		if deviceResult{
			if len(devicesResObj.Data) != 0{
				profileCreateReqObj.Data.Relationships.Devices = &devconnmanager.DataIdAndTypeItemList{}
				profileCreateReqObj.Data.Relationships.Devices.Data = make([]devconnmanager.IdAndTypeItem,0)
				for _,deviceItem :=range devicesResObj.Data {
					var itemId devconnmanager.IdAndTypeItem
					itemId.Id = deviceItem.Id
					itemId.Type = deviceItem.Type
					profileCreateReqObj.Data.Relationships.Devices.Data = append(profileCreateReqObj.Data.Relationships.Devices.Data,itemId)
				}
			}
		}else {
			logs.Info("获取Devices信息失败")
			return nil
		}
	}
	url := _const.APPLE_PROFILE_MANAGER_URL
	result := ReqToAppleHasObjMethod("POST", url, token, &profileCreateReqObj, &profileCreateResObj)
	if result {
		return &profileCreateResObj
	} else {
		return nil
	}
}

func CreateOrUpdateProfile(c *gin.Context) {
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
		logs.Info(requestData.AccountName, requestData.AccountType, requestData.BundleId, requestData.UseCertId,
			requestData.ProfileName, requestData.ProfileType, requestData.UserName)
		//todo 发送Lark消息 @zhangmengqi 如上面logs.Info，Lark消息卡片提供account_name、account_type、bundle_id、use_cert_id、profile_name、profile_type、user_name信息
		c.JSON(http.StatusOK, gin.H{
			"message":   "lark success",
			"errorCode": 0,
		})
		return
	} else {
		logs.Info("普通企业类型账号，苹果api自动处理")
		tokenString := GetTokenStringByTeamId(requestData.TeamId)
		if requestData.ProfileId != "" {
			deleteUrl := _const.APPLE_PROFILE_MANAGER_URL + "/" + requestData.ProfileId
			delRes := ReqToAppleNoObjMethod("DELETE", deleteUrl, tokenString)
			if !delRes {
				logs.Info("delete profile fail from apple server")
			}
			updateData := map[string]interface{}{
				"deleted_at":time.Now(),
				"op_user":requestData.UserName,
			}
			if isDel := deleteProfileDBandTos(c, requestData.ProfileId, requestData.ProfileName, requestData.ProfileType, requestData.TeamId, requestData.BundleId,updateData,requestData.UserName); !isDel {
				return
			}
		}
		appleResult := CreateOrUpdateProfileFromApple(requestData.ProfileName, requestData.ProfileType, requestData.BundleidId, requestData.UseCertId, tokenString)
		if appleResult != nil {
			decoded, err := base64.StdEncoding.DecodeString(appleResult.Data.Attributes.ProfileContent)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message":   "pp file decoded error",
					"errorCode": 2,
				})
				return
			} else {
				pathTos := "appleConnectFile/" + requestData.TeamId + "/Profile/" + requestData.ProfileType + "/" + requestData.ProfileName + ".mobileprovision"
				uploadResult := uploadProfileToTos(decoded, pathTos)
				if !uploadResult {
					c.JSON(http.StatusInternalServerError, gin.H{
						"message":   "upload profile tos error",
						"errorCode": 3,
					})
					return
				}
				timeString := strings.Split(appleResult.Data.Attributes.ExpirationDate, "+")[0]
				exp, _ := time.Parse("2006-01-02T15:04:05", timeString)
				dbInsertErr := InsertProfileInfoToDB(appleResult.Data.Id, requestData.ProfileName, requestData.ProfileType, _const.TOS_BUCKET_URL+pathTos, requestData.UserName, exp)
				if dbInsertErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"message":   "insert tt_apple_profile error",
						"errorCode": 4,
					})
					return
				}
				dbUpdateErr := UpdateBundleProfilesRelation(requestData.BundleId, requestData.ProfileType, &appleResult.Data.Id,requestData.UserName)
				if dbUpdateErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"message":   "update apple response to tt_app_bundleId_profiles error",
						"errorCode": 5,
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"message":   "success",
					"errorCode": 0,
					"data":      appleResult,
				})
				return
			}
		} else {
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
	uploadResult := uploadProfileToTos(profileFileByteInfo, pathTos)
	if !uploadResult {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "upload profile tos error",
			"errorCode": 3,
		})
		return
	}
	exp := utils.GetFileExpireTime(profileFileFullName, ".mobileprovision", profileFileByteInfo, requestData.UserName)
	dbInsertErr := InsertProfileInfoToDB(requestData.ProfileId, requestData.ProfileName, requestData.ProfileType, _const.TOS_BUCKET_URL+pathTos, requestData.UserName, *exp)
	if dbInsertErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "insert tt_apple_profile error",
			"errorCode": 4,
		})
		return
	}
	dbUpdateErr := UpdateBundleProfilesRelation(requestData.BundleId, requestData.ProfileType, &requestData.ProfileId,requestData.UserName)
	if dbUpdateErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "update apple response to tt_app_bundleId_profiles error",
			"errorCode": 5,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":              "upload success",
		"errorCode":            0,
		"profile_download_url": _const.TOS_BUCKET_URL + pathTos,
		"profile_expire_date":  exp,
	})
	return
}

func generateInfoLineOfCard(header string, content string) *[]form.CardElementForm {
	var infoLineFormList []form.CardElementForm

	headerForm := form.GenerateTextTag(&header, false, nil)
	headerForm.Style = &utils.GrayHeaderStyle
	infoLineFormList = append(infoLineFormList, *headerForm)

	appIdForm := form.GenerateTextTag(&content, false, nil)
	infoLineFormList = append(infoLineFormList, *appIdForm)

	return &infoLineFormList
}
func generateAtLineOfCard(header,atTest,url string) *[]form.CardElementForm {
	var csrInfoFormList []form.CardElementForm
	csrHeader := header
	csrHeaderForm := form.GenerateTextTag(&csrHeader, false, nil)
	csrHeaderForm.Style = &utils.GrayHeaderStyle
	csrInfoFormList = append(csrInfoFormList, *csrHeaderForm)
	csrText := atTest
	csrUrlForm := form.GenerateATag(&csrText, false, url)
	csrInfoFormList = append(csrInfoFormList, *csrUrlForm)
	return &csrInfoFormList
}

//生成绑定账号审核消息卡片内容
func generateCardOfApproveBindAccount(appAccountCert *devconnmanager.AppAccountCert) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm

	//插入提示信息
	messageText := utils.ApproveBindAccountMessage
	messageForm := form.GenerateTextTag(&messageText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})

	//插入userName, appId, appName, appType, teamId, accountName
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.UserNameHeader, appAccountCert.UserName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.AppIdHeader, appAccountCert.AppId))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.AppNameHeader, appAccountCert.AppName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.AppTypeHeader, appAccountCert.AppType))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.TeamIdHeader, appAccountCert.TeamId))

	accountInfos := devconnmanager.QueryAccountInfo(map[string]interface{}{"team_id": appAccountCert.TeamId})
	if len(*accountInfos) != 1 {
		logs.Error("获取teamId对应的account失败：%s 错误原因：teamId对应的account记录数不等于1", appAccountCert.TeamId)
	} else {
		cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.AccountNameHeader, (*accountInfos)[0].AccountName))
	}
	return &cardFormArray
}

//生成绑定账号审核消息卡片action
func generateActionsOfApproveBindAccount(appAccountCertId uint, userName string) *[]form.CardActionForm {
	var cardActions []form.CardActionForm
	var cardAction form.CardActionForm
	var buttons []form.CardButtonForm
	var approveButtonText = utils.ApproveButtonText
	var rejectButtonText = utils.RejectButtonText
	var hideOther = false
	var url = utils.ApproveAppBindAccountUrl

	approveButtonParams := map[string]interface{}{"isApproved": 1, "appAccountCertId": appAccountCertId, "userName": userName}
	rejectButtonParams := map[string]interface{}{"isApproved": -1, "appAccountCertId": appAccountCertId, "userName": userName}

	approveButton, err := form.GenerateButtonForm(&approveButtonText, nil, nil, nil, "post", url, false, false, &approveButtonParams, nil, &hideOther)
	if err != nil {
		utils.RecordError("生成审核卡片同意button失败，", err)
	}
	rejectButton, err := form.GenerateButtonForm(&rejectButtonText, nil, nil, nil, "post", url, false, false, &rejectButtonParams, nil, &hideOther)
	if err != nil {
		utils.RecordError("生成审核卡片拒绝button失败，", err)
	}
	buttons = append(buttons, *approveButton)
	buttons = append(buttons, *rejectButton)
	cardAction.Buttons = buttons
	cardActions = append(cardActions, cardAction)
	return &cardActions
}

//根据team_id获取权限类型，返回值：0--无任何权限，1--admin权限，2--dev权限
func getPermType(c *gin.Context, username string, team_id string) int {
	var resourcPerm devconnmanager.GetPermsResponse
	resourceKey := strings.ToLower(team_id) + "_space_account"
	url := _const.Certain_Resource_All_PERMS_URL + "employeeKey=" + username + "&resourceKeys=" + resourceKey
	result := queryPerms(url, &resourcPerm)
	if !result {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "权限获取失败！", "")
		return 0
	}
	var showType = 0
	for _, permInfo := range resourcPerm.Data[resourceKey] {
		if permInfo == "admin" || permInfo == "all_cert_manager" {
			showType = 1
			break
		} else if permInfo == "dev_cert_manager" {
			showType = 2
			break
		}
	}
	return showType
}

//API3-1，重组app和账号信息
func packAppNameInfo(appInfo *devconnmanager.APPSignManagerInfo, fqr *devconnmanager.APPandCert, showType int) {
	appInfo.AppName = fqr.AppName
	appInfo.TeamId = fqr.TeamId
	appInfo.AccountType = fqr.AccountType
	appInfo.AccountVerifyStatus = fqr.AccountVerifyStatus
	appInfo.AccountVerifyUser = fqr.AccountVerifyUser
	appInfo.AppAcountId = fqr.AppAcountId
	appInfo.AppType = fqr.AppType
	appInfo.BundleProfileCertSection = make([]devconnmanager.BundleProfileCert, 0)
	packCertSection(fqr, showType, &appInfo.CertSection)
}

//API3-1，重组bundle信息
func packeBundleProfileCert(c *gin.Context, bqr *devconnmanager.APPandBundle, showType int) *devconnmanager.BundleProfileCert {
	var bundleInfo devconnmanager.BundleProfileCert
	bundleInfo.BoundleId = bqr.BundleIdIndex
	bundleInfo.BundleIdIsDel = bqr.BundleIdIsDel
	bundleInfo.BundleIdId = bqr.BundleidId
	bundleInfo.BundleIdName = bqr.BundleidName
	bundleInfo.BundleIdType = bqr.BundleidType
	packProfileSection(bqr, showType, &bundleInfo.ProfileCertSection)
	//push_cert信息整合--
	if bqr.PushCertId != "" {
		pushCert := devconnmanager.QueryCertInfoByCertId(bqr.PushCertId)
		if pushCert == nil {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "数据库查询push证书信息失败", "")
			return nil
		}
		bundleInfo.PushCert.CertId = (*pushCert).CertId
		bundleInfo.PushCert.CertType = (*pushCert).CertType
		bundleInfo.PushCert.CertExpireDate, _ = time.Parse((*pushCert).CertExpireDate, "2006-01-02 15：04：05")
		bundleInfo.PushCert.CertDownloadUrl = (*pushCert).CertDownloadUrl
		bundleInfo.PushCert.PrivKeyUrl = (*pushCert).PrivKeyUrl
	}
	//enablelist重组+capacity_obj重组
	bundleCapacityRepack(bqr, &bundleInfo)
	return &bundleInfo
}

//API3-1，重组bundle能力
func bundleCapacityRepack(bundleStruct *devconnmanager.APPandBundle, bundleInfo *devconnmanager.BundleProfileCert) {
	//config_capacibilitie_obj
	bundleInfo.ConfigCapObj["ICLOUD"] = bundleStruct.ICLOUD
	bundleInfo.ConfigCapObj["DATA_PROTECTION"] = bundleStruct.DATA_PROTECTION

	//enableList
	param, _ := json.Marshal(bundleStruct)
	bundleMap := make(map[string]interface{})
	json.Unmarshal(param, &bundleMap)
	for k, v := range bundleMap {
		if _, ok := _const.IOSSelectCapabilitiesMap[k]; ok && v == "1" {
			bundleInfo.EnableCapList = append(bundleInfo.EnableCapList, k)
		}
	}
}

//API3-1，重组profile信息
func packProfileSection(bqr *devconnmanager.APPandBundle, showType int, profile *devconnmanager.BundleProfileGroup) {
	if strings.Contains(bqr.ProfileType, "APP_DEVELOPMENT") {
		profile.DevProfile.ProfileType = bqr.ProfileType
		profile.DevProfile.ProfileId = bqr.ProfileId
		profile.DevProfile.ProfileName = bqr.ProfileName
		profile.DevProfile.ProfileDownloadUrl = bqr.ProfileDownloadUrl
		profile.DevProfile.ProfileExpireDate = bqr.ProfileExpireDate
	} else if showType == 1 {
		profile.DistProfile.ProfileType = bqr.ProfileType
		profile.DistProfile.ProfileName = bqr.ProfileName
		profile.DistProfile.ProfileId = bqr.ProfileId
		profile.DistProfile.ProfileDownloadUrl = bqr.ProfileDownloadUrl
		profile.DistProfile.ProfileExpireDate = bqr.ProfileExpireDate
	}
}

//API3-1，重组证书信息
func packCertSection(fqr *devconnmanager.APPandCert, showType int, certSection *devconnmanager.AppCertGroupInfo) {
	if strings.Contains(fqr.CertType, "DISTRIBUTION") && showType == 1 {
		certSection.DistCert.CertType = fqr.CertType
		certSection.DistCert.CertId = fqr.CertId
		certSection.DistCert.CertDownloadUrl = fqr.CertDownloadUrl
		certSection.DistCert.CertExpireDate = fqr.CertExpireDate
	} else if strings.Contains(fqr.CertType, "DEVELOPMENT") {
		certSection.DevCert.CertType = fqr.CertType
		certSection.DevCert.CertId = fqr.CertId
		certSection.DevCert.CertDownloadUrl = fqr.CertDownloadUrl
		certSection.DevCert.CertExpireDate = fqr.CertExpireDate
	}
}

//删除profile--tos删除+db删除
func deleteProfileDBandTos(c *gin.Context, profileId string, profileName string, profileType string, teamId string, bundleId string,updateData map[string]interface{},username string) bool {
	//tos中只删除重名的profile，以免出现覆盖问题
	deleteTosObj("appleConnectFile/" + teamId + "/Profile/" + profileType + "/" + profileName + ".mobileprovision")
	dbError := devconnmanager.UpdateAppleProfile(map[string]interface{}{"profile_id": profileId},updateData)
	utils.RecordError("删除tt_apple_profile失败：%v", dbError)
	if dbError != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "删除tt_apple_profile失败", "failed")
		return false
	}
	dbUpdateErr := UpdateBundleProfilesRelation(bundleId, profileType, nil,username)
	if dbUpdateErr != nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "update apple response to tt_app_bundleId_profiles error", "")
		return false
	}
	return true
}

//profile删除工单lark消息---消息卡片文字内容
func generateCardOfProfileDelete(deleteInfo *devconnmanager.ProfileDeleteRequest, appleUrl string) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm
	//插入提示信息
	messageText := utils.DeleteProfileMessage
	messageForm := form.GenerateTextTag(&messageText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})
	//插入提示：账号，profileID，profileName，申请人
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.CreateCertAccountHeader, deleteInfo.AccountName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.DeleteProfileIdHeader, deleteInfo.ProfileId))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.DeleteProfileNameHeader, deleteInfo.ProfileName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.UserNameHeader, deleteInfo.UserName))
	//插入删除跳转链接
	cardFormArray = append(cardFormArray,*generateAtLineOfCard(utils.AppleUrlHeader,utils.AppleUrlText,appleUrl))
	return &cardFormArray
}

//profile删除工单lark消息--action
func generateActionOfProfileDelete(param *map[string]interface{}) *[]form.CardActionForm {
	var cardActions []form.CardActionForm
	var cardAction form.CardActionForm
	var buttons []form.CardButtonForm
	var text = utils.DeleteButtonText
	var hideOther = false
	//online
	var url = utils.DELPROFILE_FEEDBACK_URL
	//test
	//var url = utils.DELPROFILE_FEEDBACK_URL_TEST
	button, err := form.GenerateButtonForm(&text, nil, nil, nil, "post", url, false, false, param, nil, &hideOther)
	if err != nil {
		utils.RecordError("生成卡片button失败，", err)
	}
	buttons = append(buttons, *button)
	cardAction.Buttons = buttons
	cardActions = append(cardActions, cardAction)
	return &cardActions
}
func GetAllEnableDevicesObj(devicePlatform,enableStatus,tokenString string,devicesResObj *devconnmanager.DevicesDataRes) bool{
	urlGet := _const.APPLE_DEVICES_MANAGER_URL + "?limit=200"
	if devicePlatform != "ALL"{
		urlGet = urlGet + "&filter[platform]=" + devicePlatform
	}
	if enableStatus != "ALL" {
		urlGet = urlGet + "&filter[status]=" + enableStatus
	}
	//var devicesResObj devconnmanager.DevicesDataRes
	reqResult := ReqToAppleHasObjMethod("GET",urlGet,tokenString,nil,devicesResObj)
	return reqResult
}

func checkIsDel(c *gin.Context,isDel string) bool  {
	if isDel == "1" || isDel == "2" {
		return true
	}else {
		utils.AssembleJsonResponse(c,http.StatusBadRequest,"is_del参数不正确！","")
		return false
	}
}

func deleteInfoFromAppleApi (teamId string,url string) bool{
	tokenString := GetTokenStringByTeamId(teamId)
	delRes := ReqToAppleNoObjMethod("DELETE", url, tokenString)
	if !delRes {
		utils.RecordError("delete fail from apple server",nil)
		return false
	}
	return true
}

func generateCardOfBundleDelete(deleteInfo *devconnmanager.BundleDeleteRequest, appleUrl,bundleId string) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm
	//插入提示信息
	messageText := utils.DeleteBundleMessage
	messageForm := form.GenerateTextTag(&messageText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})
	//插入提示：账号，bundleid_id，bundleid，申请人
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.CreateCertAccountHeader, deleteInfo.AccountName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.BundleAppleId, deleteInfo.BundleidId))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.BundleId, bundleId))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.UserNameHeader, deleteInfo.UserName))
	//插入删除跳转链接
	cardFormArray = append(cardFormArray,*generateAtLineOfCard(utils.AppleUrlHeader,utils.AppleUrlText,appleUrl))
	//插入profile删除提示
	if deleteInfo.DevProfileId != "" || deleteInfo.DisProfileId != ""{
		divideText := "--------------------------------------------------\n"
		divideForm := form.GenerateTextTag(&divideText, false, nil)
		cardFormArray = append(cardFormArray, []form.CardElementForm{*divideForm})
		messageText := utils.ProfileDeleteWithBundelMessage
		messageForm := form.GenerateTextTag(&messageText, false, nil)
		cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})
		if deleteInfo.DisProfileId != "" {
			url := utils.APPLE_DELETE_PROFILE_URL+deleteInfo.DisProfileId
			cardFormArray = append(cardFormArray,*generateAtLineOfCard(utils.DistProfileTitle,deleteInfo.DisProfileId,url))
		}
		if deleteInfo.DevProfileId != ""{
			url := utils.APPLE_DELETE_PROFILE_URL+deleteInfo.DevProfileId
			cardFormArray = append(cardFormArray,*generateAtLineOfCard(utils.DevProfileTitle,deleteInfo.DevProfileId,url))
		}
	}
	return &cardFormArray
}
func generateActionOfBundleDelete(param *map[string]interface{}) *[]form.CardActionForm {
	var cardActions []form.CardActionForm
	var cardAction form.CardActionForm
	var buttons []form.CardButtonForm
	var text = utils.DeleteButtonText
	var hideOther = false
	//online
	//var url = utils.DELBUNDLE_FEEDBACK_URL
	//test
	var url = utils.DELBUNDLE_FEEDBACK_URL_TEST
	button, err := form.GenerateButtonForm(&text, nil, nil, nil, "post", url, false, false, param, nil, &hideOther)
	if err != nil {
		utils.RecordError("生成卡片button失败，", err)
	}
	buttons = append(buttons, *button)
	cardAction.Buttons = buttons
	cardActions = append(cardActions, cardAction)
	return &cardActions
}