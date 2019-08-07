package developerconnmanager

import (
	"bytes"
	"code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/detect"
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
	"math"
	"mime/multipart"
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

func DeleteCertificate(c *gin.Context) {
	logs.Info("删除证书")
	var delCertRequest devconnmanager.DelCertRequest
	bindQueryError := c.ShouldBindQuery(&delCertRequest)
	utils.RecordError("参数绑定失败", bindQueryError)
	if bindQueryError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}

	condition := make(map[string]interface{})
	condition["cert_id"] = delCertRequest.CertId
	appList := devconnmanager.QueryEffectAppList(delCertRequest.CertId, delCertRequest.CertType)
	if len(appList) == 0 {
		tokenString := GetTokenStringByTeamId(delCertRequest.TeamId)
		delResult := deleteCertInApple(tokenString, delCertRequest.CertId)
		if delResult == -2 {
			utils.AssembleJsonResponse(c, http.StatusBadRequest, "在苹果开发者网站删除对应证书失败,失败原因为不存在该certId对应的证书", "failed")
			return
		}
		if delResult == -1 {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "在苹果开发者网站删除对应证书失败", "failed")
			return
		}
		tosFilePath := "appleConnectFile/" + string(delCertRequest.TeamId) + "/" + delCertRequest.CertType + "/" + delCertRequest.CertId + "/" + dealCertName(delCertRequest.CertName) + ".cer"
		delResultBool := deleteTosCert(tosFilePath)
		if !delResultBool {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "删除tos上的证书失败", "failed")
			return
		}
		delResultBool = devconnmanager.DeleteCertInfo(condition)
		if !delResultBool {
			utils.AssembleJsonResponse(c, http.StatusInternalServerError, "从数据库中删除cert_id对应的证书失败", "failed")
			return
		}
		utils.AssembleJsonResponse(c, _const.SUCCESS, "success", nil)
		return
	} else {
		userNames := devconnmanager.QueryUserNameByAppName(appList)
		var appListStr string
		for _, appName := range appList {
			appListStr += appName
		}
		message := "证书" + delCertRequest.CertId + "将要被删除," + "与该证书关联的app:" + appListStr + " 需要换绑新的证书"
		larkNotifyUsers("证书"+delCertRequest.CertId+"将要被删除", userNames, message)
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "该证书对应的appList不为空,删除失败", "failed")
		return
	}
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
			c.JSON(http.StatusOK, gin.H{
				"data":      certInfo,
				"errorCode": 9,
				"errorInfo": "往数据库中插入证书信息失败",
			})
			return
		}
		//组装lark消息 发送给负责人（用户指定or系统默认）
		botService := service.BotService{}
		botService.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
		if body.CertPrincipal == "" {
			body.CertPrincipal = utils.CreateCertPrincipal
		}
		err := sendNodeAlertToLark(certInfo.AccountName, certInfo.CertType, certInfo.CsrFileUrl, body.CertPrincipal, &botService)
		utils.RecordError("发送新建证书提醒lark失败：", err)

		filterCert(&certInfo)
		c.JSON(http.StatusOK, gin.H{
			"data":      certInfo,
			"errorCode": 0,
			"errorInfo": "",
		})
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
	expireTime := getCertExpireTime(certFileFullName, certFileType, certFileByteInfo, requestData.UserName)
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
	certInfo := devconnmanager.UpdateCertInfoAfterUpload(condition, certInfoInputs)

	utils.AssembleJsonResponse(c, _const.SUCCESS, "success", certInfo)
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
	if response.StatusCode != 201 {
		logs.Info(string(response.StatusCode))
		if response.StatusCode == 409 {
			logs.Info("已经存在类型为IOS_DEVELOPMENT且是通过api创建的证书，创建失败")
		}
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
		}
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
		} else if certName[i] == ' ' || certName[i] == '.' {
			ret += "_"
		} else {
			ret += string(certName[i])
		}
	}
	return ret
}

func sendNodeAlertToLark(accountName string, certType string, csrUrl string, principal string, botService *service.BotService) error {
	cardMessage := generateSendMessageForm(accountName, certType, csrUrl)
	//发送消息
	email := principal
	if !strings.Contains(principal, "@bytedance.com") {
		email += "@bytedance.com"
	}
	cardMessage.Email = &email
	sendMsgResp, err := botService.SendMessage(*cardMessage)
	logs.Info("SendCardMessage response= %v", sendMsgResp)
	return err
}

func generateSendMessageForm(accountName string, certType string, csrUrl string) *form.SendMessageForm {
	cardInfoFormArray := generateCardInfoOfCreateCert(accountName, certType, csrUrl)
	cardHeaderTitle := "iOS证书管理通知"
	cardForm := form.GenerateCardForm(nil, getCardHeader(cardHeaderTitle), *cardInfoFormArray, nil)
	cardMessageContent := form.GenerateCardMessageContent(cardForm)
	cardMessage, err := form.GenerateMessage("interactive", cardMessageContent)
	utils.RecordError("card信息生成出错: ", err)
	return cardMessage
}

func getCardHeader(headerTitle string) *form.CardElementForm {
	// 生成cardHeader
	imageColor := "orange"
	cardHeader, err := form.GenerateCardHeader(&headerTitle, nil, &imageColor, nil)
	utils.RecordError("生成cardHeader错误: ", err)

	return cardHeader
}

func generateCardInfoOfCreateCert(accountName string, certType string, csrUrl string) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm

	//插入提示信息
	messageText := utils.CreateCertMessage
	messageForm := form.GenerateTextTag(&messageText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})

	//插入账号信息
	var accountFormList []form.CardElementForm

	accountHeader := utils.CreateCertAccountHeader
	accountHeaderForm := form.GenerateTextTag(&accountHeader, false, nil)
	accountHeaderForm.Style = &utils.GrayHeaderStyle
	accountFormList = append(accountFormList, *accountHeaderForm)

	accountNameForm := form.GenerateTextTag(&accountName, false, nil)
	accountFormList = append(accountFormList, *accountNameForm)

	cardFormArray = append(cardFormArray, accountFormList)

	//插入证书类型信息
	var certTypeFormList []form.CardElementForm

	certTypeHeader := utils.CreateCertTypeHeader
	certTypeHeaderForm := form.GenerateTextTag(&certTypeHeader, false, nil)
	certTypeHeaderForm.Style = &utils.GrayHeaderStyle
	certTypeFormList = append(certTypeFormList, *certTypeHeaderForm)

	certTypeTextForm := form.GenerateTextTag(&certType, false, nil)
	certTypeFormList = append(certTypeFormList, *certTypeTextForm)

	cardFormArray = append(cardFormArray, certTypeFormList)

	//插入csr文件url信息
	var csrInfoFormList []form.CardElementForm

	csrHeader := utils.CsrHeader
	csrHeaderForm := form.GenerateTextTag(&csrHeader, false, nil)
	csrHeaderForm.Style = &utils.GrayHeaderStyle
	csrInfoFormList = append(csrInfoFormList, *csrHeaderForm)

	csrText := utils.CsrText
	csrUrlForm := form.GenerateATag(&csrText, false, csrUrl)
	csrInfoFormList = append(csrInfoFormList, *csrUrlForm)

	cardFormArray = append(cardFormArray, csrInfoFormList)

	return &cardFormArray
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

func getCertExpireTime(certFileName string, certFileType string, certFileBytes []byte, userName string) *time.Time {
	getCertExpUrl := "http://" + detect.DETECT_URL_PRO + "/query_certificate_expire_date" //过期日期访问地址
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("certificate", certFileName)
	utils.RecordError("访问过期日期POST请求create form file错误！", err)

	_, err = fileWriter.Write(certFileBytes)
	utils.RecordError("访问过期日期POST请求复制文件错误！", err)

	_ = writer.WriteField("username", userName)
	_ = writer.WriteField("type", certFileType)
	contentType := writer.FormDataContentType()
	err = writer.Close()
	utils.RecordError("关闭writer出错！！", err)

	response, err := http.Post(getCertExpUrl, contentType, body)
	utils.RecordError("获取证书过期信息失败！", err)

	responseByte, err := ioutil.ReadAll(response.Body)

	responseMap := make(map[string]interface{})
	err = json.Unmarshal(responseByte, &responseMap)
	utils.RecordError("证书过期信息结果解析失败！", err)

	if _, ok := responseMap["expire_time"]; !ok {
		return nil
	}

	expTimeStamp := int64(math.Floor(responseMap["expire_time"].(float64)))
	exp := time.Unix(expTimeStamp, 0)

	return &exp
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
