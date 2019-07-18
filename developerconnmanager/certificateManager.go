package developerconnmanager

import (
	"bytes"
	"code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/context"
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
)

func QueryPerms(url string,resPerms *devconnmanager.GetPermsResponse) bool{
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
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

func QueryResPerms(userName string,resourceKey string) int{
	var resPerms devconnmanager.GetPermsResponse
	url:=_const.Certain_Resource_All_PERMS_URL+"employeeKey="+userName+"&"+"resourceKeys="+resourceKey
	result:=QueryPerms(url,&resPerms)
	println(resourceKey)
	if !result{
		return -1
	}
	hasAdmin:=false
	hasAllCertManager:=false
	hasDevCertManager:=false
	for _,perm :=range resPerms.Data[resourceKey]{
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
	var queryCertRequest devconnmanager.QueryCertRequest
	bindQueryError:=c.ShouldBindQuery(&queryCertRequest)
	if bindQueryError!=nil{
		c.JSON(http.StatusOK, gin.H{
			"message":   "delete fail",
			"errorCode":  1,
			"errorInfo": "请求参数绑定失败",
		})
		return
	}
	if queryCertRequest.TeamId=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 2,
			"errorInfo" : "team_id为空！",
		})
		return
	}
	if queryCertRequest.UserName=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 3,
			"errorInfo" : "user_name为空！",
		})
		return
	}
	condition:=make(map[string]interface{})
	condition["team_id"]=queryCertRequest.TeamId
	teamIdLower:=strings.ToLower(queryCertRequest.TeamId)
	resourceKey:=teamIdLower+"_space_account"
	permsResult:=QueryResPerms(queryCertRequest.UserName,resourceKey)
	if permsResult==-1{
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 4,
			"errorInfo": "查询权限失败",
		})
		return
	}
	if permsResult==3{
		c.JSON(http.StatusOK, gin.H{
			"errorCode": "无权限查看",
			"errorInfo": "无权限查看",
		})
		return
	}
	var certsInfo *[]devconnmanager.CertInfo
	certsInfo=devconnmanager.QueryCertInfo(condition,queryCertRequest.ExpireSoon,permsResult)
	if certsInfo==nil{
		logs.Error("从数据库中查询证书相关信息失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 5,
			"errorInfo" : "从数据库中查询证书相关信息失败！",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      certsInfo,
		"errorCode": 0,
		"errorInfo": "",
	})
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

func CreateCertInApple(tokenString string,certType string,certTypeSufix string) *devconnmanager.CreCertResponse{
	var creAppleCertReq devconnmanager.CreAppleCertReq
	creAppleCertReq.Data.Type=_const.APPLE_RECEIVED_DATA_TYPE
	creAppleCertReq.Data.Attributes.CertificateType= certType
	var csrContent string
	if certTypeSufix=="DEVELOPMENT"{
		csrContent=DownloadTos(_const.TOS_CSR_FILE_FOR_DEV_KEY)
	}
	if certTypeSufix=="DISTRIBUTION"{
		csrContent=DownloadTos(_const.TOS_CSR_FILE_FOR_DIST_KEY)
	}
	creAppleCertReq.Data.Attributes.CsrContent=CutCsrContent(string(csrContent))
	bodyByte, _ := json.Marshal(creAppleCertReq)
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
	var certInfo devconnmanager.CreCertResponse
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
	return  &certInfo
}

func UploadTos(certContent []byte,tosFilePath string) bool {
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

func DownloadTos(tosFilePath string) string{
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

func DeleteTosCert(tosFilePath string) bool{
	var tosBucket = tos.WithAuth(_const.TOS_BUCKET_NAME_JYT, _const.TOS_BUCKET_TOKEN_JYT)
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	client, err := tos.NewTos(tosBucket)
	err= client.DelObject(context,tosFilePath)
	if err != nil {
		fmt.Println("Error Delete Tos Object:", err)
		return false
	}
	return true
}

func CheckParams(c *gin.Context,bodyAddr *devconnmanager.InsertCertRequest)bool{
	err:=c.ShouldBindJSON(bodyAddr)
	if err!=nil{
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 1,
			"errorInfo" : "请求参数绑定失败！",
		})
		return false
	}
	if bodyAddr.CertName=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 2,
			"errorInfo" : "cert_name为空！",
		})
		return false
	}
	if bodyAddr.CertType=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 3,
			"errorInfo" : "cert_type为空！",
		})
		return false
	}
	if bodyAddr.AccountName=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 4,
			"errorInfo" : "account_name为空！",
		})
		return false
	}
	if bodyAddr.TeamId=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 5,
			"errorInfo" : "team_id为空！",
		})
		return false
	}
	return true
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
	var body devconnmanager.InsertCertRequest
	checkResult:=CheckParams(c,&body)
	if !checkResult{
		return
	}
	tokenString:=GetTokenStringByTeamId(body.TeamId)
	strs:=strings.Split(body.CertType,"_")
	certTypeSufix:=strs[len(strs)-1]
	creCertResponse:=CreateCertInApple(tokenString,body.CertType,certTypeSufix)
	certContent:=creCertResponse.Data.Attributes.CertificateContent
	if certContent==""{
		logs.Error("从苹果获取证书失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 6,
			"errorInfo" : "从苹果获取证书失败！",
		})
		return
	}
	encryptedCert,err:=base64.StdEncoding.DecodeString(certContent)
	if err!=nil {
		logs.Error("%s", "base64 decode error"+err.Error())
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 7,
			"errorInfo" : "证书格式有误！",
		})
		return
	}
	var certInfo devconnmanager.CertInfo
	certInfo.TeamId=body.TeamId
	certInfo.AccountName=body.AccountName
	certInfo.CertId=creCertResponse.Data.Id
	certInfo.CertType=creCertResponse.Data.Attributes.CertificateType
	certInfo.CertName=creCertResponse.Data.Attributes.Name
	certInfo.CertExpireDate=creCertResponse.Data.Attributes.ExpirationDate
	if certTypeSufix=="DEVELOPMENT"{
		certInfo.PrivKeyUrl=_const.TOS_PRIVATE_KEY_URL_DEV
		certInfo.CsrFileUrl=_const.TOS_CSR_FILE_URL_DEV
	}
	if certTypeSufix=="DISTRIBUTION"{
		certInfo.PrivKeyUrl=_const.TOS_PRIVATE_KEY_URL_DIST
		certInfo.CsrFileUrl=_const.TOS_CSR_FILE_URL_DIST
	}
	tosFilePath:="appleConnectFile/"+string(certInfo.TeamId)+"/"+certInfo.CertType+"/"+certInfo.CertId+"/"+DealCertName(certInfo.CertName)+".cer"
	uploadResult:=UploadTos(encryptedCert,tosFilePath)
	if !uploadResult{
		c.JSON(http.StatusOK,gin.H{
			"data":certInfo,
			"errorCode": 8,
			"errorInfo": "往tos上传证书信息失败",
		})
		return
	}
	certInfo.CertDownloadUrl=_const.TOS_BUCKET_URL+tosFilePath
	dbResult:=devconnmanager.InsertCertInfo(certInfo)
	if !dbResult{
		c.JSON(http.StatusOK,gin.H{
			"data":certInfo,
			"errorCode": 9,
			"errorInfo": "往数据库中插入证书信息失败",
		})
		return
	}
	FilterCert(&certInfo)
	c.JSON(http.StatusOK,gin.H{
		"data":certInfo,
		"errorCode": 0,
		"errorInfo": "",
	})
}

func FilterCert(certInfo *devconnmanager.CertInfo){
	certInfo.TeamId=""
	certInfo.AccountName=""
	certInfo.CsrFileUrl=""
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
	responseByte, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Info("读取respose的body内容失败")
	}
	if len(responseByte)==0{
		return true
	}
	return false
}

func CheckDelCertRequest(c *gin.Context,delCertRequest *devconnmanager.DelCertRequest) bool{
	if delCertRequest.TeamId=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 2,
			"errorInfo" : "team_id为空！",
		})
		return false
	}
	if delCertRequest.CertId=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 3,
			"errorInfo": "cert_id为空！",
		})
		return false
	}
	if delCertRequest.CertType=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 4,
			"errorInfo": "cert_type为空！",
		})
		return false
	}
	return true
}

func DeleteCertificate(c *gin.Context){
	logs.Info("根据cert_id删除证书")
	var delCertRequest devconnmanager.DelCertRequest
	bindQueryError:=c.ShouldBindQuery(&delCertRequest)
	if bindQueryError!=nil{
		c.JSON(http.StatusOK, gin.H{
			"message":   "delete fail",
			"errorCode": 1,
			"errorInfo": "请求参数绑定失败",
		})
		return
	}
	checkResult:=CheckDelCertRequest(c,&delCertRequest)
	if !checkResult{
		return
	}
	condition:=make(map[string]interface{})
	condition["cert_id"]=delCertRequest.CertId
	appList:=devconnmanager.QueryEffectAppList(delCertRequest.CertId,delCertRequest.CertType)
	if len(appList)==0{
		tokenString:=GetTokenStringByTeamId(delCertRequest.TeamId)
		delResult:=DeleteCertInApple(tokenString,delCertRequest.CertId)
		if !delResult{
			c.JSON(http.StatusOK,gin.H{
				"message": "delete fail",
				"errorCode": 5,
				"errorInfo": "在苹果开发者网站删除对应证书失败",
			})
			return
		}
		certInfo:=devconnmanager.QueryCertInfoByCertId(delCertRequest.CertId)
		tosFilePath:="appleConnectFile/"+string(delCertRequest.TeamId)+"/"+delCertRequest.CertType+"/"+delCertRequest.CertId+"/"+DealCertName(certInfo.CertName)+".cer"
		delResult=DeleteTosCert(tosFilePath)
		if !delResult{
			c.JSON(http.StatusOK,gin.H{
				"message": "delete fail",
				"errorCode": 6,
				"errorInfo": "删除tos上的证书失败",
			})
			return
		}
		delResult=devconnmanager.DeleteCertInfo(condition)
		if !delResult {
			c.JSON(http.StatusOK, gin.H{
				"message":   "delete fail",
				"errorCode": 7,
				"errorInfo": "从数据库中删除cert_id对应的证书失败",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "delete success",
			"errorCode": 0,
			"errorInfo": "",
		})
	}else {
		userNames:=devconnmanager.QueryUserNameByAppName(appList)
		var appListStr string
		for _,appName:=range appList{
			appListStr+=appName
		}
		message:="证书"+delCertRequest.CertId+"将要被删除,"+"与该证书关联的app:"+appListStr+" 需要换绑新的证书"
		LarkNotifyUsers("证书"+delCertRequest.CertId+"将要被删除",userNames,message)
	}
}

func CheckCertExpireDate(c *gin.Context){
	logs.Info("检查过期证书")
	expiredCertInfos:=devconnmanager.QueryExpiredCertInfos()
	c.JSON(http.StatusOK,gin.H{
		"data":expiredCertInfos,
		"errorCode": 0,
		"errorInfo": "",
	})
	for _,expiredCertInfo:=range *expiredCertInfos{
		userNames:=devconnmanager.QueryUserNameByAppName(expiredCertInfo.EffectAppList)
		LarkNotifyUsers("证书将要过期提醒",userNames,"证书"+expiredCertInfo.CertId+"即将过期")
	}

}

func ReceiveP12file(c *gin.Context) ([]byte,string){
	file, header, _ :=c.Request.FormFile("priv_p12_file")
	if header==nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 1,
			"errorInfo":   "没有文件上传",
		})
		return nil,""
	}
	logs.Info("打印File Name：" + header.Filename)
	p12ByteInfo,err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 2,
			"errorInfo":   "error read p12 file",
		})
		return nil,""
	}
	return p12ByteInfo,header.Filename
}

func CheckUploadRequest(c *gin.Context,certInfo *devconnmanager.CertInfo) bool{
	if certInfo.TeamId == "" {
		logs.Error("缺少team_id参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少team_id参数",
			"errorCode" : 3,
			"data" : "缺少team_id参数",
		})
		return false
	}
	if certInfo.CertType == "" {
		logs.Error("缺少cert_type参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少cert_type参数",
			"errorCode" : 4,
			"data" : "缺少cert_type参数",
		})
		return false
	}
	if certInfo.CertId == "" {
		logs.Error("缺少cert_id参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少cert_id参数",
			"errorCode" : 5,
			"data" : "缺少cert_id参数",
		})
		return false
	}
	return true
}

func UploadPrivKey(c *gin.Context){
	p12FileCont,p12filename:=ReceiveP12file(c)
	if len(p12FileCont) ==0 {
		logs.Error("缺少priv_p12_file参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少priv_p12_file参数",
			"errorCode" : 1,
			"data" : "缺少priv_p12_file参数",
		})
		return
	}
	var certInfo devconnmanager.CertInfo
	bindError:=c.ShouldBind(&certInfo)
	if bindError!=nil{
		c.JSON(http.StatusOK, gin.H{
			"message":   "请求参数绑定失败",
			"errorCode": 2,
			"errorInfo": "请求参数绑定失败",
		})
		return
	}
	CheckResult:=CheckUploadRequest(c,&certInfo)
	if !CheckResult{
		return
	}
	chkCertResult:=devconnmanager.CheckCertExit(certInfo.TeamId)
	if chkCertResult==-1{
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 6,
			"errorInfo": "数据库连接失败",
		})
		return
	}
	if chkCertResult==-2{
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 7,
			"errorInfo": "team_id对应的证书记录不存在",
		})
		return
	}
	tosFilePath:="appleConnectFile/"+string(certInfo.TeamId)+"/"+certInfo.CertType+"/"+certInfo.CertId+"/"+p12filename
	uploadResult:=UploadTos(p12FileCont,tosFilePath)
	if !uploadResult {
		logs.Error("上传p12文件到tos失败！")
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 8,
			"errorInfo": "上传p12文件到tos失败！",
		})
		return
	}
	condition := make(map[string]interface{})
	condition["cert_id"] = certInfo.CertId
	privKeyUrl:=_const.TOS_BUCKET_URL+tosFilePath
	dbResult := devconnmanager.UpdateCertInfo(condition,privKeyUrl)
	if !dbResult {
		logs.Error("更新数据库中的证书信息失败！")
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 9,
			"errorInfo": "更新数据库中的证书信息失败！",
		})
		return
	}
	certInfoNew:=devconnmanager.QueryCertInfoByCertId(certInfo.CertId)
	if certInfoNew==nil{
		logs.Error("从数据库中查询证书相关信息失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 10,
			"errorInfo" : "从数据库中查询证书相关信息失败！",
		})
		return
	}
	FilterCert(certInfoNew)
	c.JSON(http.StatusOK,gin.H{
		"data":*certInfoNew,
		"errorCode": 0,
		"errorInfo": "",
	})
}

func LarkNotifyUsers(groupName string,userNames []string,message string) bool{
	var getTokenRequest utils.GetTokenRequest
	getTokenRequest.AppId=utils.APP_ID
	getTokenRequest.AppSecret=utils.APP_SECRET
	var getTokenResponse utils.GetTokenResponse
	utils.CallLarkAPI(utils.GET_Tenant_Access_Token_URL,"",getTokenRequest,&getTokenResponse)
	if getTokenResponse.Code!=0{
		logs.Error("获取tenant_access_token失败")
		return false
	}
	token:="Bearer "+getTokenResponse.TenantAccessToken

	var getUserIdsRequest utils.GetUserIdsRequest
	var getUserIdsResponse utils.GetUserIdsResponse
	var openIds []string
	var employeeIds []string
	for _,userName:=range userNames{
		getUserIdsRequest.Email=userName
		utils.CallLarkAPI(utils.GET_USER_IDS_URL,token,getUserIdsRequest,&getUserIdsResponse)
		openIds=append(openIds, getUserIdsResponse.OpenId)
		employeeIds= append(employeeIds, getUserIdsResponse.EmployeeId)
	}
	if getUserIdsResponse.Code!=0{
		logs.Error("获取用open_id和employee_id失败")
		return false
	}

	var createGroupRequest utils.CreateGroupRequest
	createGroupRequest.Name=groupName
	createGroupRequest.Description= groupName
	createGroupRequest.EmployeeIds=employeeIds
	createGroupRequest.OpenIds=openIds
	var createGroupResponse utils.CreateGroupResponse
	utils.CallLarkAPI(utils.CREATE_GROUP_URL,token,createGroupRequest,&createGroupResponse)
	openChatId:=createGroupResponse.OpenChatId
	if createGroupResponse.Code!=0{
		logs.Error("机器人建群失败")
		return false
	}

	var sendMsgRequest utils.SendMsgRequest
	var sendMsgResponse utils.SendMsgResponse
	sendMsgRequest.OpenChatId=openChatId
	sendMsgRequest.Content.Text=message
	sendMsgRequest.MsgType="text"
	utils.CallLarkAPI(utils.SEND_MESSAGE_URL,token,sendMsgRequest,&sendMsgResponse)
	if sendMsgResponse.Code!=0{
		logs.Error("往群里面发送消息失败")
		return false
	}
	return true
}
