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
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
	"strings"
)


func QueryPerms(url string,resPerms *dal.GetPermsResponse) int{
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
//查询个人拥有的某一资源的所有权限
func QueryResPerms(userName string,resourceKey string) int{
	var resPerms dal.GetPermsResponse
	url:=_const.Certain_Resource_All_PERMS_URL+"employeeKey="+userName+"&"+"resourceKeys="+resourceKey
	QueryPerms(url,&resPerms)
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
	//todo accountName不需要传了，也不需要处理了
	//todo 为什么这些数据不用ShouldBindQuery？看看梦琪的ClusterManager代码怎么处理的，且绑定失败是怎么处理的？
	//todo expireSoon传上来非"0"即"1"，1代表你要去数据库过滤近一个月过期的证书，"0"代表不处理
	var queryCertRequest dal.QueryCertRequest
	bindQueryError:=c.ShouldBindQuery(&queryCertRequest)
	if bindQueryError!=nil{
		c.JSON(http.StatusOK, gin.H{
			"message":   "delete fail",
			"errorCode":  -1,
			"errorInfo": "请求参数绑定失败",
		})
		return
	}
	teamId:=queryCertRequest.TeamId
	userName:=queryCertRequest.UserName
	expireSoon:=queryCertRequest.ExpireSoon
	if teamId=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -2,
			"errorInfo" : "team_id为空！",
		})
		return
	}
	if userName=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -3,
			"errorInfo" : "user_name为空！",
		})
		return
	}
	condition:=make(map[string]interface{})
	condition["team_id"]=teamId
	teamId=strings.ToLower(teamId)        //权限平台上的权限名称都是小写
	resourceKey:=teamId+"_space_account"
	permsResult:=QueryResPerms(userName,resourceKey)
	if permsResult==3{
		c.JSON(http.StatusOK, gin.H{
			//todo data怎么又变字符串了呢？空值要不就不返回，要么就返回对应约定的格式！！！
			"errorCode": "无权限查看",
			"errorInfo": "无权限查看",
		})
		return
	}
	//todo certRelatedInfosMap这玩意是啥，把一个对象当key是怎么思考的？
	var certsInfo *[]dal.CertInfo
	if permsResult==1{
		certsInfo=dal.QueryCertInfo(condition,expireSoon)
	}
	if permsResult==2{
		//todo IOS_DEVELOPMENT这种类型不应该定义在const中么？
		condition["cert_type"] = _const.CERT_TYPE_IOS_DEV
		certsInfo=dal.QueryCertInfo(condition,expireSoon)
	}
	if certsInfo==nil{
		logs.Error("从数据库中查询证书相关信息失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -4,
			"errorInfo" : "从数据库中查询证书相关信息失败！",
		})
		return
	}
	//Todo 表3建完后补上effect_app_list
	//todo GetCertRelatedInfos这一步还需要？
	FilterCerts(certsInfo)
	c.JSON(http.StatusOK, gin.H{
		"data":      certsInfo,
		"errorCode": "0",
		"errorInfo": "",
	})
}

func FilterCerts(certsInfo *[]dal.CertInfo){
	for i:=0;i<len(*certsInfo);i++{
		(*certsInfo)[i].TeamId=""
		(*certsInfo)[i].AccountName=""
		(*certsInfo)[i].CsrFileUrl=""
	}
}

func FilterCert(certInfo *dal.CertInfo){
		(*certInfo).TeamId=""
		(*certInfo).AccountName=""
		(*certInfo).CsrFileUrl=""
		(*certInfo).EffectAppList=nil

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

func CreateCertInApple(tokenString string,certType string) *dal.CreCertResponse{
	var creAppleCertReq dal.CreAppleCertReq
	//todo const是用来干啥的？
	creAppleCertReq.Data.Type=_const.APPLE_RECEIVED_DATA_TYPE
	//todo 测试的时候让你写死IOS_DEVELOPMENT，你现在都提交代码了，还在这hard code写死CertificateType？
	creAppleCertReq.Data.Attributes.CertificateType= certType
	//todo tos不能直接读数据？GetObject，需要通过GetCsrContent client.Do下载？
	csrContent:=DownloadTos(_const.TOS_CSR_FILE_KEY)
	creAppleCertReq.Data.Attributes.CsrContent=CutCsrContent(csrContent)
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
	var certInfo dal.CreCertResponse
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
	//todo 用object做返回值？数据要很大怎么办？
	return  &certInfo
}

//todo 看看我的Test64DecodeToString（bundleIdManager.go文件中第12行）方法，证书这个东西也不需要这么多处理啊？

func UploadTos(certContent []byte,tosFilePath string) bool {
	//todo staticanalysisresult C5V4TROQGXMCTPXLIJFT hard code?要const干啥用
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

func CheckParams(c *gin.Context,bodyAddr *dal.InsertCertRequest){
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
	var body dal.InsertCertRequest
	CheckParams(c,&body)
	//todo 解释下GetTokenStringByTeamId2方法为啥要带"2"
	tokenString:=GetTokenStringByTeamId(body.TeamId)
	creCertResponse:=CreateCertInApple(tokenString,body.CertType)
	certContent:=creCertResponse.Data.Attributes.CertificateContent
	if certContent==""{
		logs.Error("从苹果获取证书失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -6,
			"errorInfo" : "从苹果获取证书失败！",
		})
		return
	}
	encryptedCert,err:=base64.StdEncoding.DecodeString(certContent)
	if err!=nil {
		logs.Error("%s", "base64 decode error"+err.Error())
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -7,
			"errorInfo" : "证书格式有误！",
		})
		return
	}
	var certInfo dal.CertInfo
	certInfo.TeamId=body.TeamId
	certInfo.AccountName=body.AccountName
	//todo 怎么一会传地址，一会传对象obj？
	certInfo.CertId=creCertResponse.Data.Id
	certInfo.CertType=creCertResponse.Data.Attributes.CertificateType
	certInfo.CertName=creCertResponse.Data.Attributes.Name
	certInfo.CertExpireDate=creCertResponse.Data.Attributes.ExpirationDate
	certInfo.PrivKeyUrl=_const.TOS_PRIVATE_KEY_URL
	certInfo.CsrFileUrl=_const.TOS_CSR_FILE_URL
	tosFilePath:="appleConnectFile/"+string(certInfo.TeamId)+"/"+certInfo.CertType+"/"+certInfo.CertId+"/"+DealCertName(certInfo.CertName)+".cer"
	uploadResult:=UploadTos(encryptedCert,tosFilePath)
	if !uploadResult{
		c.JSON(http.StatusOK,gin.H{
			"data":certInfo,
			"errorCode": "-8",
			"errorInfo": "往tos上传证书信息失败",
		})
		return
	}
	certInfo.CertDownloadUrl=_const.TOS_BUCKET_URL+tosFilePath
	dbResult:=dal.InsertCertInfo(certInfo)
	if !dbResult{
		c.JSON(http.StatusOK,gin.H{
			"data":certInfo,
			"errorCode": "-9",
			"errorInfo": "往数据库中插入证书信息失败",
		})
		return
	}
	//todo 这块有修改，新增证书接口没有effect_app_list，因为证书新生成的，还没有app用到证书，不需要返回effect_app_list
	FilterCert(&certInfo)
	c.JSON(http.StatusOK,gin.H{
		"data":certInfo,
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


func DeleteCertificate(c *gin.Context){
	logs.Info("根据cert_id删除证书")
	//todo 为什么这些数据不用ShouldBindQuery？看看梦琪的ClusterManager代码怎么处理的，且绑定失败是怎么处理的？
	//todo 这里再让前端传一个team_id，就不用去数据库查找了team_id了
	var delCertRequest dal.DelCertRequest
	bindQueryError:=c.ShouldBindQuery(&delCertRequest)
	if bindQueryError!=nil{
		c.JSON(http.StatusOK, gin.H{
			"message":   "delete fail",
			"errorCode": "-1",
			"errorInfo": "请求参数绑定失败",
		})
		return
	}
	certId:=delCertRequest.CertId
	teamId:=delCertRequest.TeamId
	if teamId=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -2,
			"errorInfo" : "team_id为空！",
		})
		return
	}
	if certId=="" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -3,
			"errorInfo": "cert_id为空！",
		})
		return
	}
	condition:=make(map[string]interface{})
	condition["cert_id"]=delCertRequest.CertId
	appList:=dal.QueryEffectAppList(condition)
	if len(appList)==0{
		tokenString:=GetTokenStringByTeamId(teamId)
		delResult:=DeleteCertInApple(tokenString,certId)
		if !delResult{
			c.JSON(http.StatusOK,gin.H{
				"message": "delete fail",
				"errorCode": "-4",
				"errorInfo": "在苹果开发者网站删除对应证书失败",
			})
			return
		}
		dbRet:=dal.DeleteCertInfo(condition)
		if !dbRet {
			c.JSON(http.StatusOK, gin.H{
				"message":   "delete fail",
				"errorCode": "-5",
				"errorInfo": "从数据库中删除cert_id对应的证书失败",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "delete success",
			"errorCode": "0",
			"errorInfo": "",
		})
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
	expiredCertInfos:=dal.QueryExpiredCertInfos()
	FilterCerts(expiredCertInfos)
	c.JSON(http.StatusOK,gin.H{
		"data":expiredCertInfos,
		"errorCode": "0",
		"errorInfo": "",
	})
	//todo 需要新建lark群（不同证书建立不同群），拉app负责人进群，同步群里，证书即将过期。
	for _,expiredCertInfo:=range *expiredCertInfos{
		userNames:=dal.QueryUserNameAccAppName(expiredCertInfo.EffectAppList)
		LarkNotifyUsers("证书将要过期提醒",userNames,"证书"+expiredCertInfo.CertId+"即将过期")
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


func UploadPrivKey(c *gin.Context){
	p12FileCont,p12filename:=ReceiveP12file(c)
	if len(p12FileCont) ==0 {
		logs.Error("缺少priv_p12_file参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少priv_p12_file参数",
			"errorCode" : -1,
			"data" : "缺少priv_p12_file参数",
		})
		return
	}
	var certInfo dal.CertInfo
	certInfo.TeamId = c.DefaultPostForm("team_id", "")
	if certInfo.TeamId == "" {
		logs.Error("缺少team_id参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少team_id参数",
			"errorCode" : -2,
			"data" : "缺少team_id参数",
		})
		return
	}
	certInfo.CertType = c.DefaultPostForm("cert_type", "")
	if certInfo.CertType == "" {
		logs.Error("缺少cert_type参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少cert_type参数",
			"errorCode" : -3,
			"data" : "缺少cert_type参数",
		})
		return
	}
	certInfo.CertId = c.DefaultPostForm("cert_id", "")
	if certInfo.CertId == "" {
		logs.Error("缺少cert_id参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少cert_id参数",
			"errorCode" : -4,
			"data" : "缺少cert_id参数",
		})
		return
	}
	tosFilePath:="appleConnectFile/"+string(certInfo.TeamId)+"/"+certInfo.CertType+"/"+certInfo.CertId+"/"+p12filename
	//todo ？？？print？
	uploadResult:=UploadTos(p12FileCont,tosFilePath)
	if !uploadResult {
		logs.Error("上传p12文件到tos失败！")
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -5,
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
			"errorCode": -6,
			"errorInfo": "更新数据库中的证书信息失败！",
		})
		return
	}
	//todo 这块有修改，新增证书接口没有effect_app_list，因为证书新生成的，还没有app用到证书，不需要返回effect_app_list
	certInfos:=dal.QueryCertInfo(condition,"0")
	if certInfos==nil{
		logs.Error("从数据库中查询证书相关信息失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -7,
			"errorInfo" : "从数据库中查询证书相关信息失败！",
		})
		return
	}
	FilterCert(&(*certInfos)[0])
	c.JSON(http.StatusOK,gin.H{
		"data":certInfos,
		"errorCode": "0",
		"errorInfo": "",
	})
}

func LarkNotifyUsers(groupName string,userNames []string,message string) bool{
	var getTokenParams utils.GetTokenParams
	getTokenParams.AppId=utils.APP_ID
	getTokenParams.AppSecret=utils.APP_SECRET
	var getTokenRet utils.GetTokenRet
	utils.CallLarkAPI(utils.GET_Tenant_Access_Token_URL,"",getTokenParams,&getTokenRet)
	if getTokenRet.Code!=0{
		logs.Error("获取tenant_access_token失败")
		return false
	}
	token:="Bearer "+getTokenRet.TenantAccessToken

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

	var sendMsgParams utils.SendMsgParams
	var sendMsgRet utils.SendMsgRet
	sendMsgParams.OpenChatId=openChatId
	sendMsgParams.Content.Text=message
	sendMsgParams.MsgType="text"
	utils.CallLarkAPI(utils.SEND_MESSAGE_URL,token,sendMsgParams,&sendMsgRet)
	if sendMsgRet.Code!=0{
		logs.Error("往群里面发送消息失败")
		return false
	}
	return true
}
