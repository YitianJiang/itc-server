package developerconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

//返回BundleID的能力给前端做展示
type GetCapabilitiesInfoReq struct {
	AppType     string  `form:"app_type" json:"app_type" binding:"required"`
}

type CapabilitiesInfo struct {
	SelectCapabilitiesInfo []string `json:"select_capabilities"`
	SettingsCapabilitiesInfo map[string][]string `json:"settings_capabilities"`
}
//列表联查中间结构体--bundle和appName关联
type BundleResortStruct struct {
	BundleInfo 			devconnmanager.BundleProfileCert
	AppName				string
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

/**
API 3-1：根据业务线appid，返回app相关list
*/
func GetAppSignListDetailInfo(c *gin.Context) {
	logs.Info("获取app签名信息List")
	var requestInfo devconnmanager.AppSignListRequest
	err := c.ShouldBindQuery(&requestInfo)
	if err != nil {
		utils.RecordError("AppSignList请求参数绑定失败,",err)
		utils.AssembleJsonResponse(c,http.StatusBadRequest,"请求参数绑定失败","")
		return
	}
	//权限判断，showType为1（超级权限），showType为2(dev权限），showType为0（无权限）
	showType := getPermType(c,requestInfo.Username,requestInfo.TeamId)
	if showType == 0 {
		AssembleJsonResponse(c,http.StatusBadRequest,"无权限查看！","")
		return
	}
	//根据app_id和team_id获取appName基本信息以及证书信息
	var cQueryResult []devconnmanager.APPandCert
	sql := "select aac.app_name,aac.app_type,aac.id as app_acount_id,aac.team_id,aac.account_verify_status,aac.account_verify_user," +
		"ac.cert_id,ac.cert_type,ac.cert_expire_date,ac.cert_download_url,ac.priv_key_url from tt_app_account_cert aac, tt_apple_certificate ac" +
		" where aac.app_id = '" + requestInfo.AppId + "' and aac.team_id = '"+requestInfo.TeamId+"' and aac.deleted_at IS NULL and (aac.dev_cert_id = ac.cert_id or aac.dist_cert_id = ac.cert_id)" +
		" and ac.deleted_at IS NULL "
	query_c := devconnmanager.QueryWithSql(sql, &cQueryResult)
	if query_c != nil {
		AssembleJsonResponse(c,http.StatusInternalServerError,"查询失败","")
		return
	}else if len(cQueryResult)== 0{
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "未查询到该app_id在本账号下的信息！", "")
		return
	}
	//以appName为单位重组基本信息，appappNameMap为最终结果的map
	appNameList := "("
	appNameMap := make(map[string]devconnmanager.APPSignManagerInfo)
	for _,fqr := range cQueryResult {
		if v,ok := appNameMap[fqr.AppName]; ok {
			packCertSection(&fqr,showType,&v.CertSection)
			appNameMap[fqr.AppName] = v
		}else{
			var appInfo devconnmanager.APPSignManagerInfo
			packAppNameInfo(&appInfo,&fqr,showType)
			appNameMap[fqr.AppName]= appInfo
			appNameList += "'"+fqr.AppName +"',"
		}
	}
	appNameList = strings.TrimSuffix(appNameList,",")
	appNameList += ")"
	//根据app_id和app_name获取bundleid信息+profile信息
	var bQueryResult []devconnmanager.APPandBundle
	sql_c := "select abp.app_name,abp.bundle_id as bundle_id_index,abp.push_cert_id,ap.profile_id,ap.profile_name,ap.profile_expire_date,ap.profile_type,ap.profile_download_url,ab.*" +
		" from tt_app_bundleId_profiles abp, tt_apple_bundleId ab, tt_apple_profile ap " +
		"where abp.app_id = '"+requestInfo.AppId+"' and abp.app_name in "+appNameList+" and abp.bundle_id = ab.bundle_id and (abp.dev_profile_id = ap.profile_id or abp.dist_profile_id = ap.profile_id) " +
		"and abp.deleted_at IS NULL and ab.deleted_at IS NULL and ap.deleted_at IS NULL"
	query_b := devconnmanager.QueryWithSql(sql_c,&bQueryResult)
	if query_b != nil {
		AssembleJsonResponse(c,http.StatusInternalServerError,"查询失败","")
		return
	}
	//以bundle为单位重组信息，appName作为附加信息
	bundleMap := make(map[string]BundleResortStruct)
	for _,bqr := range bQueryResult{
		if v,ok := bundleMap[bqr.BundleId];ok{
			packProfileSection(&bqr,showType,&v.BundleInfo.ProfileCertSection)
			bundleMap[bqr.BundleId] = v
		}else {
			var bundleResort BundleResortStruct
			bundleResort.AppName = bqr.AppName
			bundles :=packeBundleProfileCert(c,&bqr,showType)
			//查询push证书失败
			if bundles == nil {
				return
			}
			bundleResort.BundleInfo =(*bundles)
			bundleMap[bqr.BundleId]=bundleResort
		}
	}
	//重组最终结果，appNameMap同bundleMap
	for _,bundleDetail := range bundleMap {
		//1--补全profile中的cert_id
		if bundleDetail.BundleInfo.ProfileCertSection.DistProfile.ProfileId != "" {
			bundleDetail.BundleInfo.ProfileCertSection.DistProfile.UserCertId = appNameMap[bundleDetail.AppName].CertSection.DistCert.CertId
		}
		if bundleDetail.BundleInfo.ProfileCertSection.DevProfile.ProfileId != "" {
			bundleDetail.BundleInfo.ProfileCertSection.DevProfile.UserCertId = appNameMap[bundleDetail.AppName].CertSection.DevCert.CertId
		}
		//2--bundle信息整合到appNameMap中
		appInfo := appNameMap[bundleDetail.AppName]
		appInfo.BundleProfileCertSection = append(appInfo.BundleProfileCertSection,bundleDetail.BundleInfo)
		appNameMap[bundleDetail.AppName]= appInfo
	}
	//结果为appNameMap的value集合
	result := make([]devconnmanager.APPSignManagerInfo,0)
	for _,info := range appNameMap {
		result = append(result,info)
	}
	AssembleJsonResponse(c,http.StatusOK,"success",result)
}

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
	var appCertChange devconnmanager.AppAccountCert
	appCertChange.UserName = requestData.UserName
	if requestData.CertType == _const.CERT_TYPE_IOS_DEV || requestData.CertType == _const.CERT_TYPE_MAC_DEV {
		appCertChange.DevCertId = requestData.CertId
	}else if requestData.CertType == _const.CERT_TYPE_IOS_DIST || requestData.CertType == _const.CERT_TYPE_MAC_DIST {
		appCertChange.DistCertId = requestData.CertId
	}else {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数正证书类型不正确", "failed")
		return
	}
	dbError := devconnmanager.UpdateAppAccountCert(conditionDB,appCertChange)
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

//根据team_id获取权限类型，返回值：0--无任何权限，1--admin权限，2--dev权限
func getPermType(c *gin.Context,username string,team_id string) int {
	var resourcPerm devconnmanager.GetPermsResponse
	resourceKey := strings.ToLower(team_id)+"_space_account"
	url := _const.Certain_Resource_All_PERMS_URL + "employeeKey=" + username + "&resourceKeys=" + resourceKey
	result := queryPerms(url, &resourcPerm)
	if !result {
		AssembleJsonResponse(c,http.StatusInternalServerError,"权限获取失败！","")
		return 0
	}
	var showType = 0
	for _,permInfo := range resourcPerm.Data[resourceKey] {
		if permInfo == "admin" || permInfo == "all_cert_manager" {
			showType = 1
			break
		}else if permInfo == "dev_cert_manager" {
			showType = 2
			break
		}
	}
	return showType
}
//API3-1，重组app和账号信息
func packAppNameInfo(appInfo *devconnmanager.APPSignManagerInfo,fqr *devconnmanager.APPandCert,showType int) {
	appInfo.AppName = fqr.AppName
	appInfo.TeamId = fqr.TeamId
	appInfo.AccountType = fqr.AccountType
	appInfo.AccountVerifyStatus = fqr.AccountVerifyStatus
	appInfo.AccountVerifyUser = fqr.AccountVerifyUser
	appInfo.AppAcountId = fqr.AppAcountId
	appInfo.AppType = fqr.AppType
	appInfo.BundleProfileCertSection = make([]devconnmanager.BundleProfileCert,0)
	packCertSection(fqr,showType,&appInfo.CertSection)
}
//API3-1，重组bundle信息
func packeBundleProfileCert(c *gin.Context,bqr *devconnmanager.APPandBundle,showType int) *devconnmanager.BundleProfileCert{
	var bundleInfo devconnmanager.BundleProfileCert
	bundleInfo.BoundleId = bqr.BundleIdIndex
	bundleInfo.BundleIdIsDel = bqr.BundleIdIsDel
	bundleInfo.BundleIdId = bqr.BundleidId
	bundleInfo.BundleIdName = bqr.BundleidName
	bundleInfo.BundleIdType = bqr.BundleidType
	packProfileSection(bqr,showType,&bundleInfo.ProfileCertSection)
	//push_cert信息整合--
	if bqr.PushCertId != ""{
		pushCert := devconnmanager.QueryCertInfoByCertId(bqr.PushCertId)
		if pushCert == nil {
			AssembleJsonResponse(c,http.StatusInternalServerError,"数据库查询push证书信息失败","")
			return nil
		}
		bundleInfo.PushCert.CertId = (*pushCert).CertId
		bundleInfo.PushCert.CertType = (*pushCert).CertType
		bundleInfo.PushCert.CertExpireDate,_ = time.Parse((*pushCert).CertExpireDate,"2006-01-02 15：04：05")
		bundleInfo.PushCert.CertDownloadUrl = (*pushCert).CertDownloadUrl
		bundleInfo.PushCert.PrivKeyUrl = (*pushCert).PrivKeyUrl
	}
	//enablelist重组+capacity_obj重组
	bundleCapacityRepack(bqr,&bundleInfo)
	return &bundleInfo
}
//API3-1，重组bundle能力
func bundleCapacityRepack(bundleStruct *devconnmanager.APPandBundle,bundleInfo *devconnmanager.BundleProfileCert)  {
	//config_capacibilitie_obj
	var icloud devconnmanager.BundleConfigCapInfo
	icloud.KeyInfo = "ICLOUD_VERSION"
	icloud.Options = bundleStruct.ICLOUD
	var dataProtect devconnmanager.BundleConfigCapInfo
	dataProtect.KeyInfo = "DATA_PROTECTION_PERMISSION_LEVEL"
	dataProtect.Options = bundleStruct.DATA_PROTECTION
	bundleInfo.ConfigCapObj = make(map[string]devconnmanager.BundleConfigCapInfo)
	bundleInfo.ConfigCapObj["ICLOUD"] = icloud
	bundleInfo.ConfigCapObj["DATA_PROTECTION"] = dataProtect

	//enableList
	param,_:=json.Marshal(bundleStruct)
	bundleMap := make(map[string]interface{})
	json.Unmarshal(param,&bundleMap)
	for k,v := range bundleMap{
		if _,ok := _const.IOSSelectCapabilitiesMap[k];ok && v=="1" {
			bundleInfo.EnableCapList = append(bundleInfo.EnableCapList,k)
		}
	}
}
//API3-1，重组profile信息
func packProfileSection(bqr *devconnmanager.APPandBundle,showType int,profile *devconnmanager.BundleProfileGroup)  {
	if strings.Contains(bqr.ProfileType,"APP_DEVELOPMENT"){
		profile.DevProfile.ProfileType = bqr.ProfileType
		profile.DevProfile.ProfileId = bqr.ProfileId
		profile.DevProfile.ProfileName = bqr.ProfileName
		profile.DevProfile.ProfileDownloadUrl = bqr.ProfileDownloadUrl
		profile.DevProfile.ProfileExpireDate = bqr.ProfileExpireDate
	}else if showType == 1{
		profile.DistProfile.ProfileType = bqr.ProfileType
		profile.DistProfile.ProfileName = bqr.ProfileName
		profile.DistProfile.ProfileId = bqr.ProfileId
		profile.DistProfile.ProfileDownloadUrl = bqr.ProfileDownloadUrl
		profile.DistProfile.ProfileExpireDate = bqr.ProfileExpireDate
	}
}
//API3-1，重组证书信息
func packCertSection(fqr *devconnmanager.APPandCert,showType int,certSection *devconnmanager.AppCertGroupInfo)  {
	if strings.Contains(fqr.CertType,"DISTRIBUTION") && showType == 1{
		certSection.DistCert.CertType = fqr.CertType
		certSection.DistCert.CertId = fqr.CertId
		certSection.DistCert.CertDownloadUrl = fqr.CertDownloadUrl
		certSection.DistCert.CertExpireDate = fqr.CertExpireDate
	}else if strings.Contains(fqr.CertType,"DEVELOPMENT"){
		certSection.DevCert.CertType = fqr.CertType
		certSection.DevCert.CertId = fqr.CertId
		certSection.DevCert.CertDownloadUrl = fqr.CertDownloadUrl
		certSection.DevCert.CertExpireDate = fqr.CertExpireDate
	}
	//else {
	//	certSection.PushCert.CertType = fqr.CertType
	//	certSection.PushCert.CertId = fqr.CertId
	//	certSection.PushCert.CertDownloadUrl = fqr.CertDownloadUrl
	//	certSection.PushCert.CertExpireDate = fqr.CertExpireDate
	//}
}