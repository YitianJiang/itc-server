package developerconnmanager

import (
	"bytes"
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/context"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
)


func QueryRocket(url string,resPerms *dal.ResourcePermissions) int{
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Info("新建request对象失败")
		return -1
	}
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送get请求失败")
		return -2
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		logs.Info(string(response.StatusCode))
		return -3
	} else {
		response, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
			return -4
		}
		json.Unmarshal(response, resPerms)
		return 0
	}
}

func QueryResPerms(userName string,resourceKey string) int{
	var resPerms dal.ResourcePermissions
	url:=_const.Certain_Resource_All_PERMS_URL+"employeeKey="+userName+"&"+"resourceKeys="+resourceKey
	println(url)
	QueryRocket(url,&resPerms)
	hasAdmin:=false
	hasAllCertManager:=false
	hasDevCertManager:=false
	println(len(resPerms.Data[resourceKey]))
	for _,perm :=range resPerms.Data[resourceKey]{
		println(perm)
		if perm=="admin"{
			hasAdmin=true
		}
		if perm=="all_cert_manager"{
			hasAllCertManager=true
		}
		if perm=="dev_cert_manager"{
			hasDevCertManager=true
		}
	}
	if hasAdmin||hasAllCertManager{
		return 1
	}
	if !hasAdmin&&!hasAllCertManager&&hasDevCertManager{
		return 2
	}
	return  3
}

func QueryCertificatesInfo(c *gin.Context){
	logs.Info("从数据库中查询证书信息")
	accountName:=c.DefaultQuery("account_name","")
	teamId:=c.DefaultQuery("team_id","")
	expireSoon:=c.DefaultQuery("expire_soon","")
	userName:=c.DefaultQuery("user_name","")
	if teamId=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -1,
			"errorInfo" : "team_id为空！",
		})
		return
	}
	if userName=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -2,
			"errorInfo" : "user_name为空！",
		})
		return
	}
	condition:=make(map[string]interface{})
	condition["team_id"]=teamId
	if accountName != ""{
		condition["account_name"] = accountName
	}
	if expireSoon != ""{
		condition["expire_soon"] = expireSoon
	}
	resourceKey:=teamId+"_space_account"
	permsResult:=QueryResPerms(userName,resourceKey)
	if permsResult==3{
		c.JSON(http.StatusOK, gin.H{
			"data":      "",
			"errorCode": "无权限查看",
			"errorInfo": "无权限查看",
		})
		return
	}
	var certRelatedInfosMap map[dal.CertInfo][]string
	if permsResult==1{
		certRelatedInfosMap=dal.QueryCertInfo(condition)
	}
	if permsResult==2{
		condition["cert_type"] = "IOS_DEVELOPMENT"
		certRelatedInfosMap=dal.QueryCertInfo(condition)
	}
	if certRelatedInfosMap==nil{
		logs.Error("从数据库中查询证书相关信息失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -3,
			"errorInfo" : "从数据库中查询证书相关信息失败！",
		})
		return
	}
	//Todo 表3建完后补上effect_app_list
	certRelatedInfos:=GetCertRelatedInfos(certRelatedInfosMap)
	c.JSON(http.StatusOK, gin.H{
		"data":      certRelatedInfos,
		"errorCode": "0",
		"errorInfo": "",
	})
}

func GetCertRelatedInfos(certRelatedInfoMap map[dal.CertInfo][]string) []dal.CertRelatedInfo{
	var certRelatedInfos []dal.CertRelatedInfo
	var certRelatedInfo dal.CertRelatedInfo
	for certInfo,appNameList:=range certRelatedInfoMap{
		certRelatedInfo.CertId=certInfo.CertId
		certRelatedInfo.CertType=certInfo.CertType
		certRelatedInfo.CertName=certInfo.CertName
		certRelatedInfo.CertExpireDate=certInfo.CertExpireDate
		certRelatedInfo.CertDownloadUrl=certInfo.CertDownloadUrl
		certRelatedInfo.PrivKeyUrl=certInfo.PrivKeyUrl
		certRelatedInfo.EffectAppList=appNameList
		certRelatedInfos=append(certRelatedInfos, certRelatedInfo)
	}
	return  certRelatedInfos
}

func CutCsrContent(csrContent string) string{
	var start,end int
	for i:=0;i<len(csrContent);i++{
		if csrContent[i]=='\n'{
			start=i+1
			break
		}
	}
	count:=0
	for i:=len(csrContent)-1;i>=0;i--{
		if csrContent[i]=='\n'{
			count++
			if count==2{
				end=i-1
				break
			}
		}
	}
	return csrContent[start:end+1]
}

func GetCsrContent()string {
	client := &http.Client{}
	request, err := http.NewRequest("GET", _const.CSR_FILE_URL, nil)
	if err != nil {
		logs.Info("新建request对象失败")
		return ""
	}
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送get请求失败")
		return ""
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		logs.Info(string(response.StatusCode))
		return ""
	} else {
		response, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
			return ""
		}
		return CutCsrContent(string(response))
	}
}

func CreateCertInApple(tokenString string)dal.RecvCert{
	var certCreate dal.CertCreate
	certCreate.Data.Type="certificates"
	certCreate.Data.Attributes.CertificateType= "IOS_DEVELOPMENT"
	certCreate.Data.Attributes.CsrContent=GetCsrContent()
	bodyByte, _ := json.Marshal(certCreate)
	rbodyByte := bytes.NewReader(bodyByte)
	client := &http.Client{}
	request, err := http.NewRequest("POST", _const.APPLE_CREATE_CERT_URL,rbodyByte)
	if err != nil {
		logs.Info("新建request对象失败")
	}
	request.Header.Set("Authorization", tokenString)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送post请求失败")
	}
	defer response.Body.Close()
	var certInfo dal.RecvCert
	if response.StatusCode != 201 {
		logs.Info(string(response.StatusCode))
		if response.StatusCode==409{
			logs.Info("已经存在类型为IOS_DEVELOPMENT且是通过api创建的证书，创建失败")
		}
	}else{
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.Info("读取respose的body内容失败")
		}
		json.Unmarshal(body, &certInfo)
	}
	return  certInfo
}

func CertInfoGetValues(certInfo *dal.CertInfo,recvCert dal.RecvCert){
	certInfo.CertId=recvCert.Data.Id
	certInfo.CertType=recvCert.Data.Attributes.CertificateType
	certInfo.CertName=recvCert.Data.Attributes.Name
	certInfo.CertExpireDate=recvCert.Data.Attributes.ExpirationDate
	certInfo.PrivKeyUrl=_const.PRIVATE_KEY_URL
	certInfo.CsrFileUrl=_const.CSR_FILE_URL
}

func FormatCertContent(certContent string) []byte{
	start:=0
	end:=64
	var ret string
	for ; ;{
		ret+=certContent[start:end]
		ret+="\n"
		start+=64
		end+=64
		if end>len([]rune(certContent))-1{
			ret+=certContent[start:]
			ret+="\n"
			break
		}
	}
	retByte,err:=base64.StdEncoding.DecodeString(ret)
	if err!=nil{
		logs.Error("%s","base64 decode error"+err.Error())
		return nil
	}
	return  retByte
}

func UploadTos(certContent []byte,tosFilePath string) bool {
	var tosBucket = tos.WithAuth("staticanalysisresult", "C5V4TROQGXMCTPXLIJFT")
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

func CheckParams(c *gin.Context,bodyAddr *dal.RecvParamsInsertCert){
	err:=c.ShouldBindJSON(bodyAddr)
	if err!=nil{
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -1,
			"errorInfo" : "获取前端传过来的参数格式不对！",
		})
		return
	}
	if bodyAddr.CertName=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -2,
			"errorInfo" : "cert_name为空！",
		})
		return
	}
	if bodyAddr.CertType=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -3,
			"errorInfo" : "cert_type为空！",
		})
		return
	}
	if bodyAddr.AccountName=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -4,
			"errorInfo" : "account_name为空！",
		})
		return
	}
	if bodyAddr.TeamId=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -5,
			"errorInfo" : "team_id为空！",
		})
		return
	}
}

func DealCertName(certName string) string{
	var ret string
	for i:=0;i<len(certName);i++{
		if certName[i]==':'{
			continue
		} else if certName[i]==' '||certName[i]=='.'{
			ret+="_"
		}else{
			ret+=string(certName[i])
		}
	}
	return ret
}

func InsertCertificate(c *gin.Context){
	logs.Info("从数据库中查询证书信息")
	var body dal.RecvParamsInsertCert
	CheckParams(c,&body)
	tokenString:=GetTokenStringByTeamId2(body.TeamId)
	recvCert:=CreateCertInApple(tokenString)
	if recvCert.Data.Attributes.CertificateContent==""{
		logs.Error("从苹果获取证书失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -7,
			"errorInfo" : "从苹果获取证书失败！",
		})
		return
	}
	formatedCert:=FormatCertContent(recvCert.Data.Attributes.CertificateContent)
	var certInfo dal.CertInfo
	certInfo.TeamId=body.TeamId
	certInfo.AccountName=body.AccountName
	CertInfoGetValues(&certInfo,recvCert)
	tosFilePath:="appleConnectFile/"+string(certInfo.TeamId)+"/"+certInfo.CertType+"/"+certInfo.CertId+"/"+DealCertName(certInfo.CertName)+".cer"
	UploadTos(formatedCert,tosFilePath)
	certInfo.CertDownloadUrl=_const.TOS_BUCKET_URL+tosFilePath
	dal.InsertCertInfo(certInfo)
	condition:=make(map[string]interface{})
	condition["cert_Id"]=certInfo.CertId
	certRelatedInfosMap:=dal.QueryCertInfo(condition)
	if certRelatedInfosMap==nil{
		logs.Error("从数据库中查询证书相关信息失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -6,
			"errorInfo" : "从数据库中查询证书相关信息失败！",
		})
		return
	}
	//Todo 表3建完后补上effect_app_list
	certRelatedInfos:=GetCertRelatedInfos(certRelatedInfosMap)
	c.JSON(http.StatusOK,gin.H{
		"data":certRelatedInfos,
		"errorCode": "0",
		"errorInfo": "",
	})
}

func DeleteCertInApple(tokenString string,certId string)bool{
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", _const.APPLE_CERT_DELETE_ADDR+certId,nil)
	if err != nil {
		logs.Info("新建request对象失败")
	}
	request.Header.Set("Authorization", tokenString)
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送DELETE请求失败")
	}
	defer response.Body.Close()
	r, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Info("读取respose的body内容失败")
	}
	if len(r)==0{
		return true
	}
	return false
}

func GetTokenStringByCertId(CertId string)string{
	condition := make(map[string]interface{})
	condition["cert_id"] = CertId
	teamId := dal.QueryTeamId(condition)
	tokenString:=GetTokenStringByTeamId2(teamId)
	return tokenString
}

func DeleteCertificate(c *gin.Context){
	logs.Info("根据cert_id删除证书")
	certId:=c.Query("cert_id")
	condition:=make(map[string]interface{})
	condition["cert_id"]=certId
	appList:=dal.QueryEffectAppList(condition)
	if len(appList)==0{
		tokenString:=GetTokenStringByCertId(certId)
		delResult:=DeleteCertInApple(tokenString,certId)
		if !delResult{
			c.JSON(http.StatusOK,gin.H{
				"message": "delete fail",
				"errorCode": "-1",
				"errorInfo": "在苹果开发者网站删除对应证书失败",
			})
			return
		}else{
			dbRet:=dal.DeleteCertInfo(condition)
			if dbRet {
				c.JSON(http.StatusOK, gin.H{
					"message":   "delete success",
					"errorCode": "0",
					"errorInfo": "",
				})
			}else{
				c.JSON(http.StatusOK, gin.H{
					"message":   "delete fail",
					"errorCode": "-2",
					"errorInfo": "从数据库中删除cert_id对应的证书失败",
				})
			}
		}
	}else {
		//todo lark拉群拉对应app的负责人，通知换绑证书
		userNames:=dal.QueryUserNameAccAppName(appList)
		var appListStr string
		for _,appName:=range appList{
			appListStr+=appName
		}
		message:="证书"+certId+"将要被删除,"+"与该证书关联的app:"+appListStr+" 需要换绑新的证书"
		LarkNotifyUsers("证书"+certId+"已被删除",userNames,message)
	}
}

func CheckCertExpireDate(c *gin.Context){
	logs.Info("检查过期证书")
	expiredCertInfosMap:=dal.QueryExpiredCertInfos()
	certRelatedInfos:=GetCertRelatedInfos(expiredCertInfosMap)
	c.JSON(http.StatusOK,gin.H{
		"data":certRelatedInfos,
		"errorCode": "0",
		"errorInfo": "",
	})
	//todo 需要新建lark群（不同证书建立不同群），拉app负责人进群，同步群里，证书即将过期。
	for _,certRelatedInfo:=range certRelatedInfos{
		userNames:=dal.QueryUserNameAccAppName(certRelatedInfo.EffectAppList)
		LarkNotifyUsers("证书将要过期提醒",userNames,"证书"+certRelatedInfo.CertId+"即将过期")
	}

}

func ReceiveP12file(c *gin.Context) ([]byte,string){
	file, header, _ :=c.Request.FormFile("priv_p12_file")
	if header==nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"errorInfo":   "没有文件上传",
		})
		return nil,""
	}
	logs.Info("打印File Name：" + header.Filename)
	p12ByteInfo,err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -2,
			"errorInfo":   "error read p12 file",
		})
		return nil,""
	}
	return p12ByteInfo,header.Filename
}

func CheckUpdateParams(c *gin.Context,certInfo *dal.CertInfo) bool{
	certInfo.TeamId = c.DefaultPostForm("team_id", "")
	if certInfo.TeamId == "" {
		logs.Error("缺少team_id参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少team_id参数",
			"errorCode" : -1,
			"data" : "缺少team_id参数",
		})
		return false
	}
	certInfo.CertType = c.DefaultPostForm("cert_type", "")
	if certInfo.CertType == "" {
		logs.Error("缺少cert_type参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少cert_type参数",
			"errorCode" : -2,
			"data" : "缺少cert_type参数",
		})
		return false
	}
	certInfo.CertId = c.DefaultPostForm("cert_id", "")
	if certInfo.CertId == "" {
		logs.Error("缺少cert_id参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少cert_id参数",
			"errorCode" : -3,
			"data" : "缺少cert_id参数",
		})
		return false
	}
	return  true
}

func UploadPrivKey(c *gin.Context){
	p12FileCont,p12filename:=ReceiveP12file(c)
	if len(p12FileCont) ==0 {
		logs.Error("缺少priv_p12_file参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少priv_p12_file参数",
			"errorCode" : -5,
			"data" : "缺少priv_p12_file参数",
		})
		return
	}
	var certInfo dal.CertInfo
	checkResult:=CheckUpdateParams(c,&certInfo)
	if  checkResult {
		tosFilePath:="appleConnectFile/"+string(certInfo.TeamId)+"/"+certInfo.CertType+"/"+certInfo.CertId+"/"+p12filename
		println(tosFilePath)
		uploadResult:=UploadTos(p12FileCont,tosFilePath)
		if !uploadResult {
			logs.Error("上传p12文件到tos失败！")
			c.JSON(http.StatusOK, gin.H{
				"errorCode": -1,
				"errorInfo": "上传p12文件到tos失败！",
			})
			return
		}
		condition := make(map[string]interface{})
		condition["cert_id"] = certInfo.CertId
		privKeyUrl:=_const.TOS_BUCKET_URL+tosFilePath
		dbResult := dal.UpdateCertInfo(condition,privKeyUrl)
		if !dbResult {
			logs.Error("更新数据库中的证书信息失败！")
			c.JSON(http.StatusOK, gin.H{
				"errorCode": -2,
				"errorInfo": "更新数据库中的证书信息失败！",
			})
			return
		}
		certRelatedInfosMap:=dal.QueryCertInfo(condition)
		if certRelatedInfosMap==nil{
			logs.Error("从数据库中查询证书相关信息失败")
			c.JSON(http.StatusOK, gin.H{
				"errorCode" : -3,
				"errorInfo" : "从数据库中查询证书相关信息失败！",
			})
			return
		}
		certRelatedInfos:=GetCertRelatedInfos(certRelatedInfosMap)
		c.JSON(http.StatusOK,gin.H{
			"data":certRelatedInfos,
			"errorCode": "0",
			"errorInfo": "",
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -4,
			"errorInfo": "请求参数缺失",
		})
	}
}

func LarkNotifyUsers(groupName string,userNames []string,message string) bool{
	var getTokenParams utils.GetTokenParams
	getTokenParams.AppId=utils.APP_ID
	getTokenParams.AppSecret=utils.APP_SECRET
	var getTokenRet utils.GetTokenRet
	utils.PushGetLarkParams(utils.GET_Tenant_Access_Token_URL,"",getTokenParams,&getTokenRet)
	if getTokenRet.Code!=0{
		logs.Error("获取tenant_access_token失败")
		return false
	}
	token:="Bearer "+getTokenRet.TenantAccessToken

	var  getUserIdsParams utils.GetUserIdsParams
	var getUserIdsRet utils.GetUserIdsRet
	var openIds []string
	var employeeIds []string
	for _,userName:=range userNames{
		getUserIdsParams.Email=userName
		utils.PushGetLarkParams(utils.GET_USER_IDS_URL,token,getUserIdsParams,&getUserIdsRet)
		openIds=append(openIds, getUserIdsRet.OpenId)
		employeeIds= append(employeeIds, getUserIdsRet.EmployeeId)
	}
	if getUserIdsRet.Code!=0{
		logs.Error("获取用open_id和employee_id失败")
		return false
	}

	var createGroupParams utils.CreateGroupParams
	createGroupParams.Name=groupName
	createGroupParams.Description= groupName
	createGroupParams.EmployeeIds=employeeIds
	createGroupParams.OpenIds=openIds
	var createGroupRet utils.CreateGroupRet
	utils.PushGetLarkParams(utils.CREATE_GROUP_URL,token,createGroupParams,&createGroupRet)
	openChatId:=createGroupRet.OpenChatId
	if createGroupRet.Code!=0{
		logs.Error("机器人建群失败")
		return false
	}

	var sendMsgParams utils.SendMsgParams
	var sendMsgRet utils.SendMsgRet
	sendMsgParams.OpenChatId=openChatId
	sendMsgParams.Content.Text=message
	sendMsgParams.MsgType="text"
	utils.PushGetLarkParams(utils.SEND_MESSAGE_URL,token,sendMsgParams,&sendMsgRet)
	if sendMsgRet.Code!=0{
		logs.Error("往群里面发送消息失败")
		return false
	}
	return true
}
