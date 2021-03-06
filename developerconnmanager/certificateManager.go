package developerconnmanager

import (
	"bytes"
	"code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/context"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"code.byted.org/yuyilei/bot-api/form"
	"code.byted.org/yuyilei/bot-api/service"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//todo 重构-数据库操作，错误处理，时间解析
func QueryCertificatesInfo(c *gin.Context) {
	logs.Info("获取证书信息")
	var queryCertRequest devconnmanager.QueryCertRequest

	bindQueryError := c.ShouldBindQuery(&queryCertRequest)
	utils.RecordError("参数绑定失败", bindQueryError)
	if bindQueryError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}

	//查询用户权限
	teamIdLower := strings.ToLower(queryCertRequest.TeamId)
	resourceKey := teamIdLower + "_space_account"
	permsResult := queryResPerms(queryCertRequest.UserName, resourceKey)
	if permsResult == -1 {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "查询权限失败", "failed")
		return
	}
	if permsResult == 3 {
		utils.AssembleJsonResponse(c, http.StatusForbidden, "无权限查看", "failed")
		return
	}

	//获取证书信息
	condition := make(map[string]interface{})
	condition["team_id"] = queryCertRequest.TeamId
	certsInfo := devconnmanager.QueryCertInfo(condition, queryCertRequest.ExpireSoon, permsResult)

	utils.AssembleJsonResponse(c, _const.SUCCESS, "success", certsInfo)
}

func CheckCertExpireDate(c *gin.Context) {
	logs.Info("检查过期证书")
	expiredCertInfos := devconnmanager.QueryExpiredCertInfos()
	if expiredCertInfos == nil {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "查询将要过期的证书信息失败", "failed")
		return
	}
	utils.AssembleJsonResponse(c, _const.SUCCESS, "查询将要过期的证书信息失败", expiredCertInfos)
	for _, expiredCertInfo := range *expiredCertInfos {
		userNames := devconnmanager.QueryUserNameByAppName(expiredCertInfo.EffectAppList)
		larkNotifyUsers("证书将要过期提醒", userNames, "证书"+expiredCertInfo.CertId+"即将过期")
	}
}

func UploadPrivKey(c *gin.Context) {
	p12FileCont, p12filename := receiveP12file(c)
	if len(p12FileCont) == 0 {
		logs.Error("缺少priv_p12_file参数")
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "缺少priv_p12_file参数", "failed")
		return
	}
	var certInfo devconnmanager.CertInfo
	bindError := c.ShouldBind(&certInfo)
	if bindError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	CheckResult := checkUploadRequest(c, &certInfo)
	if !CheckResult {
		return
	}
	chkCertResult := devconnmanager.CheckCertExit(certInfo.TeamId)
	if chkCertResult == -1 {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "数据库查询失败", "failed")
		return
	}
	if chkCertResult == -2 {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "team_id对应的证书记录不存在", "failed")
		return
	}
	tosFilePath := "appleConnectFile/" + string(certInfo.TeamId) + "/" + certInfo.CertType + "/" + certInfo.CertId + "/" + p12filename
	uploadResult := uploadTos(p12FileCont, tosFilePath)
	if !uploadResult {
		logs.Error("上传p12文件到tos失败！")
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "上传p12文件到tos失败", "failed")
		return
	}
	condition := make(map[string]interface{})
	condition["cert_id"] = certInfo.CertId
	privKeyUrl := _const.TOS_BUCKET_URL + tosFilePath
	dbResult := devconnmanager.UpdateCertInfo(condition, privKeyUrl)
	if !dbResult {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "更新数据库中的证书信息失败", "failed")
		logs.Error("更新数据库中的证书信息失败！")
		return
	}
	certInfoNew := devconnmanager.QueryCertInfoByCertId(certInfo.CertId)
	if certInfoNew == nil {
		logs.Error("从数据库中查询证书相关信息失败")
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "从数据库中查询证书相关信息失败", "failed")
		return
	}
	filterCert(certInfoNew)
	utils.AssembleJsonResponse(c, _const.SUCCESS, "success", certInfoNew)
	return
}

func InsertCertificate(c *gin.Context) {
	logs.Info("新建证书")
	var body devconnmanager.InsertCertRequest
	err := c.ShouldBindJSON(&body)
	utils.RecordError("参数绑定失败", err)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	strs := strings.Split(body.CertType, "_")
	certTypeSufix := strs[len(strs)-1]

	if certTypeSufix == "PUSH" {
		//组装lark消息 发送给负责人（用户指定or系统默认）
		botService := service.BotService{}
		botService.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
		if body.CertPrincipal == "" {
			body.CertPrincipal = utils.CreateCertPrincipal
		}
		csrFileUrl := _const.TOS_CSR_FILE_URL_PUSH
		var err error
		if body.IsUpdate == "1" {
			if !sendUpdateCertAlertToLark(&body, csrFileUrl, &botService) {
				utils.AssembleJsonResponse(c, http.StatusInternalServerError, "发送新建证书提醒lark失败", nil)
				return
			}
		} else {
			if !sendCreateCertAlertToLark(&body, csrFileUrl, &botService) {
				utils.AssembleJsonResponse(c, http.StatusInternalServerError, "发送更新证书提醒lark失败", nil)
				return
			}
		}
		err = devconnmanager.UpdateAppBundleProfiles(map[string]interface{}{"bundle_id": body.BundleId}, map[string]interface{}{"push_cert_id": _const.NeedUpdate})
		if err != nil {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "更新app_bundle_id_profile表失败", nil)
			return
		}
		utils.AssembleJsonResponse(c, _const.SUCCESS, "success", nil)
		return
	}

	if body.AccountType == "Enterprise" {
		//在数据库插入证书记录
		var certInfo devconnmanager.CertInfo
		certInfo.TeamId = body.TeamId
		certInfo.AccountName = body.AccountName
		certInfo.CertName = body.CertName
		certInfo.CertType = body.CertType
		if certTypeSufix == "DEVELOPMENT" {
			certInfo.PrivKeyUrl = _const.TOS_PRIVATE_KEY_URL_DEV
			certInfo.CsrFileUrl = _const.TOS_CSR_FILE_URL_DEV
		}
		if certTypeSufix == "DISTRIBUTION" {
			certInfo.PrivKeyUrl = _const.TOS_PRIVATE_KEY_URL_DIST
			certInfo.CsrFileUrl = _const.TOS_CSR_FILE_URL_DIST
		}
		dbResult := devconnmanager.InsertCertInfo(&certInfo)
		if !dbResult {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "往数据库中插入证书信息失败", certInfo)
			return
		}
		//组装lark消息 发送给负责人（用户指定or系统默认）
		botService := service.BotService{}
		botService.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
		if body.CertPrincipal == "" {
			body.CertPrincipal = utils.CreateCertPrincipal
		}
		filterCert(&certInfo)
		if !sendCreateCertAlertToLark(&body, certInfo.CsrFileUrl, &botService) {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "发送新建证书lark失败", certInfo)
			return
		}
		utils.AssembleJsonResponse(c, _const.SUCCESS, "success", certInfo)
		return
	}

	tokenString := GetTokenStringByTeamId(body.TeamId)
	creCertResponse := createCertInApple(tokenString, body.CertType, certTypeSufix)
	if creCertResponse == nil || creCertResponse.Data.Attributes.CertificateContent == "" {
		logs.Error("从苹果获取证书失败")
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "从苹果获取证书失败", "failed")
		return
	}
	certContent := creCertResponse.Data.Attributes.CertificateContent
	encryptedCert, err := base64.StdEncoding.DecodeString(certContent)
	if err != nil {
		logs.Error("%s", "base64 decode error"+err.Error())
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "证书格式有误", "failed")
		return
	}
	var certInfo devconnmanager.CertInfo
	certInfo.TeamId = body.TeamId
	certInfo.AccountName = body.AccountName
	certInfo.CertId = creCertResponse.Data.Id
	certInfo.CertType = creCertResponse.Data.Attributes.CertificateType
	certInfo.CertName = creCertResponse.Data.Attributes.Name
	certInfo.CertExpireDate = creCertResponse.Data.Attributes.ExpirationDate
	if certTypeSufix == "DEVELOPMENT" {
		certInfo.PrivKeyUrl = _const.TOS_PRIVATE_KEY_URL_DEV
		certInfo.CsrFileUrl = _const.TOS_CSR_FILE_URL_DEV
	}
	if certTypeSufix == "DISTRIBUTION" {
		certInfo.PrivKeyUrl = _const.TOS_PRIVATE_KEY_URL_DIST
		certInfo.CsrFileUrl = _const.TOS_CSR_FILE_URL_DIST
	}
	tosFilePath := "appleConnectFile/" + string(certInfo.TeamId) + "/" + certInfo.CertType + "/" + certInfo.CertId + "/" + dealCertName(certInfo.CertName) + ".cer"
	uploadResult := uploadTos(encryptedCert, tosFilePath)
	if !uploadResult {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "往tos上传证书信息失败", "failed")
		return
	}
	certInfo.CertDownloadUrl = _const.TOS_BUCKET_URL + tosFilePath
	dbResult := devconnmanager.InsertCertInfo(&certInfo)
	if !dbResult {
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "往数据库中插入证书信息失败", certInfo)
		return
	}
	filterCert(&certInfo)
	utils.AssembleJsonResponse(c, _const.SUCCESS, "请求参数绑定失败", certInfo)
	return
}

func UploadCertificate(c *gin.Context) {
	var requestData devconnmanager.UploadCertRequest
	bindError := c.ShouldBind(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindError)
	if bindError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败，查看是否缺少参数", "failed")
		return
	}

	//获取文件信息
	certFileByteInfo, certFileFullName := getFileFromRequest(c, "cert_file")
	//将文件名中的空格替换为下划线
	certFileFullName = strings.Replace(certFileFullName, " ", "_", -1)
	certFileFullName = strings.Replace(certFileFullName, "(", "", -1)
	certFileFullName = strings.Replace(certFileFullName, ")", "", -1)
	//logs.Info("文件名:%s",certFileFullName)
	if len(certFileByteInfo) == 0 {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "缺少file参数，读取file失败", "failed")
		return
	}

	splits := strings.Split(certFileFullName, ".")
	if len(splits) < 2 {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "文件名解析失败，请提供包含后缀名的证书文件（eg：.cer）", "failed")
		return
	}
	certFileName := splits[0]
	certFileType := "." + splits[len(splits)-1]

	certInfoInputs := make(map[string]interface{})
	if requestData.CertName != "" {
		certInfoInputs["cert_name"] = requestData.CertName
	} else {
		certInfoInputs["cert_name"] = certFileName
	}
	certInfoInputs["cert_id"] = requestData.CertId

	//解析证书获得过期时间
	expireTime := utils.GetFileExpireTime(certFileFullName, certFileType, certFileByteInfo, requestData.UserName)
	certInfoInputs["cert_expire_date"] = expireTime
	logs.Info("exp:", expireTime)

	tosFilePath := "appleConnectFile/" + string(requestData.TeamId) + "/" + requestData.CertType + "/" + requestData.CertId + "/" + certFileFullName
	uploadResult := uploadTos(certFileByteInfo, tosFilePath)
	if !uploadResult {
		utils.RecordError("上传证书文件到tos失败：", nil)
		utils.AssembleJsonResponse(c, http.StatusInternalServerError, "上传证书文件到tos失败", "failed")
		return
	}
	certUrl := _const.TOS_BUCKET_URL + tosFilePath
	certInfoInputs["cert_download_url"] = certUrl

	condition := make(map[string]interface{})
	condition["id"] = requestData.Id

	//logs.Info("%v",condition)
	//logs.Info("%v",certInfoInputs)
	var certInfo *devconnmanager.CertInfo

	if requestData.CertType == _const.IOS_PUSH || requestData.CertType == _const.MAC_PUSH {
		var certInfo devconnmanager.CertInfo
		certInfo.PrivKeyUrl = _const.TOS_PRIVATE_KEY_URL_PUSH
		certInfo.CsrFileUrl = _const.TOS_CSR_FILE_URL_PUSH
		certInfo.CertType = requestData.CertType
		certInfo.TeamId = requestData.TeamId
		certInfo.AccountName = requestData.AccountName
		if requestData.CertName != "" {
			certInfo.CertName = requestData.CertName
		} else {
			certInfo.CertName = certFileName
		}
		certInfo.CertId = requestData.CertId
		certInfo.CertDownloadUrl = certUrl
		certInfo.CertExpireDate = expireTime.Format("2006-01-02 15:04:05")
		err1 := devconnmanager.InsertRecord(&certInfo)
		err2 := devconnmanager.UpdateAppBundleProfiles(map[string]interface{}{"bundle_id": requestData.BundleId}, map[string]interface{}{"push_cert_id": requestData.CertId})
		if err1 != nil {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "证书数据库插入失败", certInfo)
			return
		}
		if err2 != nil {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "bundleIdProfile数据库更新失败", certInfo)
			return
		}
		utils.AssembleJsonResponse(c, _const.SUCCESS, "success", certInfo)
		return
	} else {
		certInfo = devconnmanager.UpdateCertInfoAfterUpload(condition, certInfoInputs)
		utils.AssembleJsonResponse(c, _const.SUCCESS, "success", certInfo)
		return
	}

}

func DeleteCertificate(c *gin.Context) {
	logs.Info("根据cert_id删除证书")
	var delCertRequest devconnmanager.DelCertRequest
	bindQueryError := c.ShouldBindQuery(&delCertRequest)
	if bindQueryError != nil {
		logs.Error(bindQueryError.Error())
		c.JSON(http.StatusOK, gin.H{
			"message":   "delete fail",
			"errorCode": 1,
			"errorInfo": "请求参数绑定失败",
		})
		return
	}
	//push_cert未在苹果后台生成删除操作
	if delCertRequest.ID == "" {
		if delCertRequest.CertType == _const.IOS_PUSH || delCertRequest.CertType == _const.MAC_PUSH {
			if delCertRequest.BundleId == "" || delCertRequest.BundleId == _const.UNDEFINED {
				utils.AssembleJsonResponse(c, 1, "bundle_id为空", "bundle_id不可为空")
				return
			}
			if err := devconnmanager.UpdateAppBundleProfiles(map[string]interface{}{"bundle_id": delCertRequest.BundleId},
				map[string]interface{}{"push_cert_id": nil}); err != nil {
				utils.AssembleJsonResponse(c, 1, "更新app_bundle_profile信息失败", "更新app_bundle_profile信息失败")
				return
			}
			utils.AssembleJsonResponse(c, 0, "success", "删除成功")
			return
		} else {
			utils.AssembleJsonResponse(c, 1, "ID为空", "ID不可为空")
			return
		}
	}

	//删除未在苹果后台生成的证书---此if下操作待apple open API ready后可删除或不执行
	if delCertRequest.CertId == "" {
		condition := map[string]interface{}{
			"id": delCertRequest.ID,
		}
		updateInfo := map[string]interface{}{
			"op_user":    delCertRequest.UserName,
			"deleted_at": time.Now(),
		}
		certDBDelete(c, &condition, &updateInfo)
		return
	}
	condition := make(map[string]interface{})
	condition["cert_id"] = delCertRequest.CertId
	appList := devconnmanager.QueryEffectAppList(delCertRequest.CertId, delCertRequest.CertType)
	if len(appList) == 0 {

		//企业分发账号和push证书工单处理逻辑---此if下操作待apple open API ready后可删除或不执行
		if delCertRequest.AccType == _const.Enterprise || delCertRequest.CertType == _const.IOS_PUSH || delCertRequest.CertType == _const.MAC_PUSH {
			var bundleid = ""   //判断是否为push证书
			var bundleIdId = "" //用于点击按钮后去苹果后台查询能力
			if delCertRequest.CertType == _const.IOS_PUSH || delCertRequest.CertType == _const.MAC_PUSH {
				condition := map[string]interface{}{
					"push_cert_id": delCertRequest.CertId,
				}
				abpInfo := devconnmanager.QueryAppBundleProfiles(condition)
				if abpInfo == nil || len(*abpInfo) == 0 {
					utils.AssembleJsonResponse(c, http.StatusInternalServerError, "查询Push证书对应的bundleID信息失败", "")
					return
				}
				bundleid = (*abpInfo)[0].BundleId
				bundleIdId = (*abpInfo)[0].BundleidId
			}
			//向负责人发送lark消息
			abot := service.BotService{}
			abot.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
			appleUrl := utils.APPLE_DELETE_CERT_URL + delCertRequest.CertId
			cardElementForms := generateCardOfCertDelete(delCertRequest.AccountName, delCertRequest.CertId, delCertRequest.CertName, appleUrl, delCertRequest.UserName)
			if delCertRequest.CertOperator == "" || delCertRequest.CertOperator == _const.UNDEFINED {
				delCertRequest.CertOperator = utils.CreateCertPrincipal
			}
			//卡片参数增加bundleIdId、teamId、accountType，用于在点击已删除时在苹果后台查询能力
			param := map[string]interface{}{
				"cert_id":      delCertRequest.CertId,
				"username":     delCertRequest.CertOperator,
				"bundle_id":    bundleid,
				"bundleid_id":  bundleIdId,
				"team_id":      delCertRequest.TeamId,
				"account_type": delCertRequest.AccType,
			}
			cardActions := generateActionsOfCertDelete(&param)
			err := sendIOSCertLarkMessage(cardElementForms, cardActions, delCertRequest.CertOperator, &abot, "--删除证书")
			if err != nil {
				utils.RecordError("发送lark消息通知负责人删除证书失败，", err)
				c.JSON(http.StatusOK, gin.H{
					"message":   "delete fail",
					"errorCode": 11,
					"errorInfo": "发送lark消息通知负责人删除证书失败",
				})
				return
			}
			//删除tos文件
			tosFilePath := "appleConnectFile/" + string(delCertRequest.TeamId) + "/" + delCertRequest.CertType + "/" + delCertRequest.CertId + "/" + dealCertName(delCertRequest.CertName) + ".cer"
			delResultBool := deleteTosCert(tosFilePath)
			if !delResultBool {
				//此处不阻塞，只打log
				logs.Error("删除tos文件失败，路径：" + tosFilePath)
			}
			//db删除，只更新deleted_at
			updateInfo := map[string]interface{}{
				"deleted_at": time.Now(),
			}
			//push证书删除时，更新和bundle之间关系
			if delCertRequest.CertType == _const.IOS_PUSH || delCertRequest.CertType == _const.MAC_PUSH {
				condition := map[string]interface{}{
					"push_cert_id": delCertRequest.CertId,
				}
				var updateInfo map[string]interface{}
				if delCertRequest.AccType != _const.Enterprise { //organization账号下处理逻辑新增deleting状态
					updateInfo = map[string]interface{}{
						"push_cert_id": _const.Deleting,
						"user_name":    delCertRequest.UserName,
					}
				} else {
					updateInfo = map[string]interface{}{
						"push_cert_id": nil,
						"user_name":    delCertRequest.UserName,
					}
				}
				if updateErr := devconnmanager.UpdateAppBundleProfiles(condition, updateInfo); updateErr != nil {
					utils.AssembleJsonResponse(c, http.StatusInternalServerError, "删除Push证书时更新bundle-profile信息失败", "")
					return
				}
			}
			certDBDelete(c, &condition, &updateInfo)
			return
		}
		//原有逻辑---Apple open api
		tokenString := GetTokenStringByTeamId(delCertRequest.TeamId)
		delResult := deleteCertInApple(tokenString, delCertRequest.CertId)
		if delResult == -2 {
			c.JSON(http.StatusOK, gin.H{
				"message":   "delete fail",
				"errorCode": 6,
				"errorInfo": "在苹果开发者网站删除对应证书失败,失败原因为不存在该certId对应的证书",
			})
			return
		}
		if delResult == -1 {
			c.JSON(http.StatusOK, gin.H{
				"message":   "delete fail",
				"errorCode": 7,
				"errorInfo": "在苹果开发者网站删除对应证书失败",
			})
			return
		}
		tosFilePath := "appleConnectFile/" + string(delCertRequest.TeamId) + "/" + delCertRequest.CertType + "/" + delCertRequest.CertId + "/" + dealCertName(delCertRequest.CertName) + ".cer"
		delResultBool := deleteTosCert(tosFilePath)
		if !delResultBool {
			c.JSON(http.StatusOK, gin.H{
				"message":   "delete fail",
				"errorCode": 8,
				"errorInfo": "删除tos上的证书失败",
			})
			return
		}
		updateInfo := map[string]interface{}{
			"op_user":    delCertRequest.UserName,
			"deleted_at": time.Now(),
		}
		certDBDelete(c, &condition, &updateInfo)
	} else {
		userNames := devconnmanager.QueryUserNameByAppName(appList)
		var appListStr string
		for _, appName := range appList {
			appListStr += appName
		}
		message := "证书" + delCertRequest.CertId + "将要被删除," + "与该证书关联的app:" + appListStr + " 需要换绑新的证书"
		larkNotifyUsers("证书"+delCertRequest.CertId+"将要被删除", userNames, message)
		c.JSON(http.StatusOK, gin.H{
			"message":   "delete fail",
			"errorCode": 10,
			"errorInfo": "该证书对应的appList不为空,删除失败",
		})
	}
}

//lark卡片点击已删除后异步更新db操作人信息，查询苹果后台并更新bundleId能力，给申请者回执
func AsynDeleteCertFeedback(c *gin.Context) {
	logs.Notice("点击已删除")
	var feedbackInfo devconnmanager.DelCertFeedback
	err := c.ShouldBindJSON(&feedbackInfo)
	if err != nil {
		utils.RecordError("请求参数绑定失败！", err)
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 1,
			"errorInfo": "请求参数绑定失败！",
		})
		return
	}
	certInfo := devconnmanager.QueryDeletedCertInfoByCertId(feedbackInfo.CustomerJson.CertId)
	if certInfo == nil {
		utils.AssembleJsonResponseWithStatusCode(c, http.StatusInternalServerError, "无匹配的证书信息", nil)
		return
	}
	go sendDeleteCertResultToApplicant(feedbackInfo.OpenId, feedbackInfo.CustomerJson.UserName, certInfo.CertName, certInfo.CertType, certInfo.AccountName)
	//更新对应cert_id的op_user信息
	condition := map[string]interface{}{
		"cert_id": feedbackInfo.CustomerJson.CertId,
	}
	updateInfo := map[string]interface{}{
		"deleted_at": time.Now(),
		"op_user":    feedbackInfo.CustomerJson.UserName,
	}
	okU := devconnmanager.UpdateCertInfoByMap(condition, updateInfo)
	var okU2 error
	if feedbackInfo.CustomerJson.Bundleid != "" {
		queryData := map[string]interface{}{
			"bundle_id": feedbackInfo.CustomerJson.Bundleid,
		}
		updateData := map[string]interface{}{
			"user_name":    feedbackInfo.CustomerJson.UserName,
			"push_cert_id": nil,
		}
		okU2 = devconnmanager.UpdateAppBundleProfiles(queryData, updateData)
	}
	if !okU || okU2 != nil {
		utils.RecordError("异步更新删除证书信息操作人失败，证书ID："+feedbackInfo.CustomerJson.CertId, nil)
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 4,
			"errorInfo": "数据库异步更新删除信息操作人失败",
		})
		return
	}
	if feedbackInfo.CustomerJson.AccountType == _const.Organization {
		logs.Notice("查询bundleId能力")
		//去苹果后台查询能力并更新数据库
		go updateBundleIdCapabilities(feedbackInfo.CustomerJson.BundleIdId, feedbackInfo.CustomerJson.TeamId)
	}
	utils.AssembleJsonResponse(c, _const.SUCCESS, "", nil)
	return
}

func updateBundleIdCapabilities(bundleIdId string, teamId string) {
	capabilitiesMap := *queryBundleIdCapabilities(bundleIdId, teamId)
	_ = devconnmanager.UpdateAppleBundleId(map[string]interface{}{"bundleid_id": bundleIdId}, capabilitiesMap)
}

func queryBundleIdCapabilities(bundleIdId string, teamId string) *map[string]interface{} {
	queryUrl := fmt.Sprintf(_const.APPLE_BUNDLE_ID_CAPABILITIES_QUERY_URL, bundleIdId)
	tokenString := GetTokenStringByTeamId(teamId)
	capabilitiesMap := make(map[string]interface{})
	var responseBody devconnmanager.QueryBundleIdCapabilityResponse
	if !ReqToAppleHasObjMethod("GET", queryUrl, tokenString, nil, &responseBody) {
		logs.Error("查询bundleId能力失败。bundleIdId: %s, teamId: %s", bundleIdId, teamId)
		return &capabilitiesMap
	}
	logs.Notice("%v", responseBody)
	for capabilityName := range _const.IOSSelectCapabilitiesMap {
		capabilitiesMap[capabilityName] = ""
	}
	for _, capability := range responseBody.Data {
		//非配置能力
		if len(capability.Attributes.Settings) == 0 || len(capability.Attributes.Settings[0].Options) == 0 {
			capabilitiesMap[capability.Attributes.CapabilityType] = capability.Id
		} else {
			//配置能力
			if capability.Attributes.CapabilityType != "APPLE_ID_AUTH"{
				capabilitiesMap[capability.Attributes.CapabilityType] = capability.Attributes.Settings[0].Options[0].Key
			}
		}
	}
	return &capabilitiesMap
}

func queryPerms(url string, resPerms *devconnmanager.GetPermsResponse) bool {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", "Basic "+_const.KANI_APP_ID_AND_SECRET_BASE64)
	if err != nil {
		logs.Info("新建request对象失败")
		return false
	}
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送get请求失败")
		return false
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		logs.Info(string(response.StatusCode))
		return false
	} else {
		responseByte, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
			return false
		}
		json.Unmarshal(responseByte, resPerms)
		return true
	}
}

func queryResPerms(userName string, resourceKey string) int {
	var resPerms devconnmanager.GetPermsResponse
	url := _const.Certain_Resource_All_PERMS_URL + "employeeKey=" + userName + "&" + "resourceKeys=" + resourceKey
	result := queryPerms(url, &resPerms)
	if !result {
		return -1
	}
	hasAdmin := false
	hasAllCertManager := false
	hasDevCertManager := false
	for _, perm := range resPerms.Data[resourceKey] {
		if perm == "admin" {
			hasAdmin = true
		}
		if perm == "all_cert_manager" {
			hasAllCertManager = true
		}
		if perm == "dev_cert_manager" {
			hasDevCertManager = true
		}
	}
	if hasAdmin || hasAllCertManager {
		return 1
	}
	if !hasAdmin && !hasAllCertManager && hasDevCertManager {
		return 2
	}
	return 3
}

func cutCsrContent(csrContent string) string {
	var start, end int
	for i := 0; i < len(csrContent); i++ {
		if csrContent[i] == '\n' {
			start = i + 1
			break
		}
	}
	count := 0
	for i := len(csrContent) - 1; i >= 0; i-- {
		if csrContent[i] == '\n' {
			count++
			if count == 2 {
				end = i - 1
				break
			}
		}
	}
	return csrContent[start : end+1]
}

func createCertInApple(tokenString string, certType string, certTypeSufix string) *devconnmanager.CreCertResponse {
	var creAppleCertReq devconnmanager.CreAppleCertReq
	creAppleCertReq.Data.Type = _const.APPLE_RECEIVED_DATA_TYPE
	creAppleCertReq.Data.Attributes.CertificateType = certType
	var csrContent string
	if certTypeSufix == "DEVELOPMENT" {
		csrContent = downloadTos(_const.TOS_CSR_FILE_FOR_DEV_KEY)
	}
	if certTypeSufix == "DISTRIBUTION" {
		csrContent = downloadTos(_const.TOS_CSR_FILE_FOR_DIST_KEY)
	}
	creAppleCertReq.Data.Attributes.CsrContent = cutCsrContent(string(csrContent))
	bodyByte, _ := json.Marshal(creAppleCertReq)
	rbodyByte := bytes.NewReader(bodyByte)
	client := &http.Client{}
	request, err := http.NewRequest("POST", _const.APPLE_CREATE_CERT_URL, rbodyByte)
	if err != nil {
		logs.Info("新建request对象失败")
		return nil
	}
	request.Header.Set("Authorization", tokenString)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送post请求失败")
		return nil
	}
	defer response.Body.Close()
	var certInfo devconnmanager.CreCertResponse
	logs.Info("苹果返回状态码：%d", response.StatusCode)
	if response.StatusCode != 201 {
		body, _ := ioutil.ReadAll(response.Body)
		logs.Info("苹果返回response\n：%s", string(body))
		if response.StatusCode == 409 {
			logs.Info("已经存在类型为IOS_DEVELOPMENT且是通过api创建的证书，创建失败")
		}
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
		}
		logs.Info("苹果返回response\n：%s", string(body))
		json.Unmarshal(body, &certInfo)
	}
	return &certInfo
}

func uploadTos(certContent []byte, tosFilePath string) bool {
	var tosBucket = tos.WithAuth(_const.TOS_BUCKET_NAME_JYT, _const.TOS_BUCKET_TOKEN_JYT)
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tosPutClient, err := tos.NewTos(tosBucket)
	err = tosPutClient.PutObject(context, tosFilePath, int64(len(certContent)), bytes.NewBuffer(certContent))
	if err != nil {
		logs.Error("%s", "上传tos失败："+err.Error())
		return false
	}
	return true
}

func downloadTos(tosFilePath string) string {
	var tosBucket = tos.WithAuth(_const.TOS_BUCKET_NAME_JYT, _const.TOS_BUCKET_TOKEN_JYT)
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := tos.NewTos(tosBucket)
	obj, err := client.GetObject(context, tosFilePath)
	if err != nil {
		fmt.Println("Error:", err)
	}
	content, _ := ioutil.ReadAll(obj.R)
	defer obj.R.Close()
	return string(content)
}

func deleteTosCert(tosFilePath string) bool {
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

func dealCertName(certName string) string {
	var ret string
	for i := 0; i < len(certName); i++ {
		if certName[i] == ':' {
			continue
		} else if certName[i] == ' ' || certName[i] == '.' || certName[i] == '(' || certName[i] == ')' || certName[i] == ',' {
			ret += "_"
		} else {
			ret += string(certName[i])
		}
	}
	return ret
}

func generateCardOfCreateOrUpdateCert(message string, certInfo *devconnmanager.CreateOrUpdateCertInfoForLark) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm

	//插入提示信息
	messageForm := form.GenerateTextTag(&message, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})

	//插入基本信息：用户名，账户名，[bundleId]
	cardFormArray = append(cardFormArray, generateCenterText("基本信息部分"))
	accountInfos := devconnmanager.QueryAccountInfo(map[string]interface{}{"team_id": certInfo.TeamId})
	if len(*accountInfos) != 1 {
		cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.TeamIdHeader, certInfo.TeamId))
		logs.Error("获取teamId对应的account失败：%s 错误原因：teamId对应的account记录数不等于1", certInfo.TeamId)
	} else {
		cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.AccountHeader, (*accountInfos)[0].AccountName))
	}

	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.UserNameHeader, certInfo.UserName))
	if certInfo.BundleId != "" {
		cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.BundleIdHeader, certInfo.BundleId))
	}
	cardFormArray = append(cardFormArray, getDividerOfCard())

	//插入证书相关：csr，类型
	cardFormArray = append(cardFormArray, generateCenterText("证书信息部分"))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.CreateCertTypeHeader, certInfo.CertType))
	cardFormArray = append(cardFormArray, *generateAtLineOfCard(utils.CsrHeader, utils.CsrText, certInfo.CsrUrl))
	return &cardFormArray
}

func sendCreateCertAlertToLark(requestData *devconnmanager.InsertCertRequest, csrUrl string, botService *service.BotService) bool {
	certInfo := devconnmanager.CreateOrUpdateCertInfoForLark{
		UserName:    requestData.UserName,
		BundleId:    requestData.BundleId,
		AccountType: requestData.AccountType,
		TeamId:      requestData.TeamId,
		CertType:    requestData.CertType,
		CsrUrl:      csrUrl,
	}
	cardInfos := generateCardOfCreateOrUpdateCert(utils.CreateCertMessage, &certInfo)
	//logs.Info("%v",*cardInfos)
	err := sendIOSCertLarkMessage(cardInfos, nil, requestData.CertPrincipal, botService, "--创建证书")
	utils.RecordError("发送创建证书工单失败：", err)
	if err != nil {
		return false
	}
	return true
}

func sendUpdateCertAlertToLark(requestData *devconnmanager.InsertCertRequest, csrUrl string, botService *service.BotService) bool {
	certInfo := devconnmanager.CreateOrUpdateCertInfoForLark{
		UserName:    requestData.UserName,
		BundleId:    requestData.BundleId,
		AccountType: requestData.AccountType,
		TeamId:      requestData.TeamId,
		CertType:    requestData.CertType,
		CsrUrl:      csrUrl,
	}
	cardInfos := generateCardOfCreateOrUpdateCert(utils.UpdateCertMessage, &certInfo)
	//logs.Info("%v",*cardInfos)
	err := sendIOSCertLarkMessage(cardInfos, nil, requestData.CertPrincipal, botService, "--更新证书")
	utils.RecordError("发送更新证书工单失败：", err)
	if err != nil {
		return false
	}
	return true
}

func getCardHeader(headerTitle string) *form.CardElementForm {
	// 生成cardHeader
	imageColor := "orange"
	cardHeader, err := form.GenerateCardHeader(&headerTitle, nil, &imageColor, nil)
	utils.RecordError("生成cardHeader错误: ", err)

	return cardHeader
}

func filterCert(certInfo *devconnmanager.CertInfo) {
	certInfo.TeamId = ""
	certInfo.AccountName = ""
}

func deleteCertInApple(tokenString string, certId string) int {
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", _const.APPLE_CERT_DELETE_ADDR+certId, nil)
	if err != nil {
		logs.Info("新建request对象失败")
		return -1
	}
	request.Header.Set("Authorization", tokenString)
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送DELETE请求失败")
		return -1
	}
	defer response.Body.Close()
	if response.StatusCode == 409 {
		logs.Info("苹果不存在该certId对应的证书")
		return -2
	}
	responseByte, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Info("读取respose的body内容失败")
		return -1
	}
	if len(responseByte) == 0 {
		return 0
	}
	return -1
}

//证书数据库删除操作--前端交互版
func certDBDelete(c *gin.Context, condition *map[string]interface{}, updateInfo *map[string]interface{}) {
	delResultBool := devconnmanager.UpdateCertInfoByMap(*condition, *updateInfo)
	if !delResultBool {
		c.JSON(http.StatusOK, gin.H{
			"message":   "delete fail",
			"errorCode": 9,
			"errorInfo": "从数据库中删除cert_id对应的证书失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "delete success",
		"errorCode": 0,
		"errorInfo": "",
	})
}

//发送IOS证书管理卡片消息
func sendIOSCertLarkMessage(cardInfoFormArray *[][]form.CardElementForm, cardActions *[]form.CardActionForm, certOperator string, botService *service.BotService, suffixs ...string) error {
	//生成卡片
	cardHeaderTitle := "iOS证书管理通知"
	for _, suffix := range suffixs {
		cardHeaderTitle += suffix
	}
	var cardForm *form.CardForm
	if cardActions != nil {
		cardForm = form.GenerateCardForm(nil, getCardHeader(cardHeaderTitle), *cardInfoFormArray, *cardActions)
	} else {
		cardForm = form.GenerateCardForm(nil, getCardHeader(cardHeaderTitle), *cardInfoFormArray, nil)
	}
	cardMessageContent := form.GenerateCardMessageContent(cardForm)
	cardMessage, err := form.GenerateMessage("interactive", cardMessageContent)
	utils.RecordError("card信息生成出错: ", err)
	if cardMessage == nil {
		return errors.New("card信息生成出错")
	}
	//发送消息
	email := certOperator
	if !strings.Contains(certOperator, "@bytedance.com") {
		email += "@bytedance.com"
	}
	cardMessage.Email = &email
	sendMsgResp, err := botService.SendMessage(*cardMessage)
	logs.Info("发送飞书lark消息响应= %v", sendMsgResp)
	if code, ok := sendMsgResp["code"].(float64); ok && code != 0 {
		if message, ok := sendMsgResp["msg"].(string); ok {
			return errors.New(message)
		}
	}
	return err
}

//生成删除工单通知卡片---文字信息
func generateCardOfCertDelete(accountName string, certId string, certName string, appleUrl string, username string) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm

	//插入提示信息
	messageText := utils.DeleteCertMessage
	messageForm := form.GenerateTextTag(&messageText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})

	//插入账号信息,证书ID，证书name,申请人
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.CreateCertAccountHeader, accountName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.DeleteCertIdHeader, certId))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.DeleteCertNameHeader, certName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.UserNameHeader, username))

	//插入apple后台url信息
	var csrInfoFormList []form.CardElementForm

	csrHeader := utils.AppleUrlHeader
	csrHeaderForm := form.GenerateTextTag(&csrHeader, false, nil)
	csrHeaderForm.Style = &utils.GrayHeaderStyle
	csrInfoFormList = append(csrInfoFormList, *csrHeaderForm)

	csrText := utils.AppleUrlText
	csrUrlForm := form.GenerateATag(&csrText, false, appleUrl)
	csrInfoFormList = append(csrInfoFormList, *csrUrlForm)

	cardFormArray = append(cardFormArray, csrInfoFormList)

	return &cardFormArray

}

//生成删除工单通知卡片---action信息
func generateActionsOfCertDelete(param *map[string]interface{}) *[]form.CardActionForm {
	var cardActions []form.CardActionForm
	var cardAction form.CardActionForm
	var buttons []form.CardButtonForm
	var text = utils.DeleteButtonText
	var hideOther = false
	var url = utils.DELCERT_FEEDBACK_URL
	button, err := form.GenerateButtonForm(&text, nil, nil, nil, "post", url, true, false, param, nil, &hideOther)
	if err != nil {
		utils.RecordError("生成卡片button失败，", err)
	}
	buttons = append(buttons, *button)
	cardAction.Buttons = buttons
	cardActions = append(cardActions, cardAction)
	return &cardActions
}

func receiveP12file(c *gin.Context) ([]byte, string) {
	file, header, _ := c.Request.FormFile("priv_p12_file")
	if header == nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 1,
			"errorInfo": "没有文件上传",
		})
		return nil, ""
	}
	logs.Info("打印File Name：" + header.Filename)
	p12ByteInfo, err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 2,
			"errorInfo": "error read p12 file",
		})
		return nil, ""
	}
	return p12ByteInfo, header.Filename
}

func checkUploadRequest(c *gin.Context, certInfo *devconnmanager.CertInfo) bool {
	if certInfo.TeamId == "" {
		logs.Error("缺少team_id参数")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少team_id参数",
			"errorCode": 3,
			"data":      "缺少team_id参数",
		})
		return false
	}
	if certInfo.CertType == "" {
		logs.Error("缺少cert_type参数")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少cert_type参数",
			"errorCode": 4,
			"data":      "缺少cert_type参数",
		})
		return false
	}
	if certInfo.CertId == "" {
		logs.Error("缺少cert_id参数")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少cert_id参数",
			"errorCode": 5,
			"data":      "缺少cert_id参数",
		})
		return false
	}
	return true
}

func getFileFromRequest(c *gin.Context, paramName string) ([]byte, string) {
	file, header, _ := c.Request.FormFile(paramName)

	if header == nil {
		utils.RecordError("没有文件上传", nil)
		return nil, ""
	}
	logs.Info("上传File Name：" + header.Filename)
	fileByteInfo, err := ioutil.ReadAll(file)
	if err != nil {
		utils.RecordError("读取上传文件失败", nil)
		return nil, ""
	}
	return fileByteInfo, header.Filename
}

func larkNotifyUsers(groupName string, userNames []string, message string) bool {
	var getTokenRequest utils.GetTokenRequest
	getTokenRequest.AppId = utils.APP_ID
	getTokenRequest.AppSecret = utils.APP_SECRET
	var getTokenResponse utils.GetTokenResponse
	utils.CallLarkAPI(utils.GET_Tenant_Access_Token_URL, "", getTokenRequest, &getTokenResponse)
	if getTokenResponse.Code != 0 {
		logs.Error("获取tenant_access_token失败")
		return false
	}
	token := "Bearer " + getTokenResponse.TenantAccessToken

	var getUserIdsRequest utils.GetUserIdsRequest
	var getUserIdsResponse utils.GetUserIdsResponse
	var openIds []string
	var employeeIds []string
	for _, userName := range userNames {
		getUserIdsRequest.Email = userName
		utils.CallLarkAPI(utils.GET_USER_IDS_URL, token, getUserIdsRequest, &getUserIdsResponse)
		openIds = append(openIds, getUserIdsResponse.OpenId)
		employeeIds = append(employeeIds, getUserIdsResponse.EmployeeId)
	}
	if getUserIdsResponse.Code != 0 {
		logs.Error("获取用open_id和employee_id失败")
		return false
	}

	var createGroupRequest utils.CreateGroupRequest
	createGroupRequest.Name = groupName
	createGroupRequest.Description = groupName
	createGroupRequest.EmployeeIds = employeeIds
	createGroupRequest.OpenIds = openIds
	var createGroupResponse utils.CreateGroupResponse
	utils.CallLarkAPI(utils.CREATE_GROUP_URL, token, createGroupRequest, &createGroupResponse)
	openChatId := createGroupResponse.OpenChatId
	if createGroupResponse.Code != 0 {
		logs.Error("机器人建群失败")
		return false
	}

	var sendMsgRequest utils.SendMsgRequest
	var sendMsgResponse utils.SendMsgResponse
	sendMsgRequest.OpenChatId = openChatId
	sendMsgRequest.Content.Text = message
	sendMsgRequest.MsgType = "text"
	utils.CallLarkAPI(utils.SEND_MESSAGE_URL, token, sendMsgRequest, &sendMsgResponse)
	if sendMsgResponse.Code != 0 {
		logs.Error("往群里面发送消息失败")
		return false
	}
	return true
}

func sendDeleteCertResultToApplicant(openId, userName, certName, certType, accountName string) {
	botService := service.BotService{}
	botService.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
	name := getUserNameByOpenId(openId, &botService)
	title := "证书删除结果通知"
	message := fmt.Sprintf("你提交的证书删除申请\n[账号：%s，证书名：%s，类型：%s]\n已于%s 被%s 完成", accountName, certName, certType, time.Now().Format("2006-01-02 15:04:05"), name)
	alertTextWithTitleToPeople(title, message, userName+"@bytedance.com", &botService)
}
