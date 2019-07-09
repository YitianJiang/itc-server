package developerconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"
)

func DeleteByTeamId(c *gin.Context)  {
	logs.Info("根据team_id参数将数据从数据库中删除")
	teamId:=c.Query("team_id")
	dbResult:=dal.DeleteAccountInfo(teamId)
	if !dbResult{
		logs.Error("从数据库中删除数据失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -1,
			"errorInfo" : "从数据库中删除数据失败！",
		})
	}
	c.JSON(http.StatusOK,gin.H{
		"errorCode": "0",
		"message":   "delete success",
	})
}

func GetQueryAccountRet(userName string) dal.AccountQueryRet{
	var resPerms dal.ResourcePermissions
	url:=_const.USER_ALL_RESOURCES_PERMS+"userName="+userName
	QueryRocket(url,&resPerms)
	keys := reflect.ValueOf(resPerms.Data).MapKeys()
	var teamIdsWithPerms []string
	for _,key:=range keys{
		keyStr:=key.String()
		teamIdsWithPerms=append(teamIdsWithPerms,keyStr[:len(keyStr)-14])
	}
	allAccountsInfo:=dal.QueryAccountInfo(nil)
	allAccountInfosMap:=make(map[string]dal.AccountExistRel)
	for _,accountsInfo:=range *allAccountsInfo{
		allAccountInfosMap[accountsInfo.TeamId]=dal.AccountExistRel{false,accountsInfo,nil}
	}
	for _,teamId:=range teamIdsWithPerms{
		allAccountInfosMap[teamId]=dal.AccountExistRel{true,allAccountInfosMap[teamId].AccountInfo,resPerms.Data[teamId+"_space_account"]}
	}
	var retValueWithP8 dal.RetValueWithP8
	var retValueWithoutP8 dal.RetValueWithoutP8
	var accountQueryRet  dal.AccountQueryRet
	for _,accountExistRel:=range allAccountInfosMap{
		if accountExistRel.IsExisted ==true{
			retValueWithP8.TeamId=accountExistRel.AccountInfo.TeamId
			retValueWithP8.AccountP8fileName=accountExistRel.AccountInfo.AccountP8fileName
			retValueWithP8.AccountP8file=accountExistRel.AccountInfo.AccountP8file
			retValueWithP8.UserName=accountExistRel.AccountInfo.UserName
			retValueWithP8.AccountType=accountExistRel.AccountInfo.AccountType
			retValueWithP8.AccountName=accountExistRel.AccountInfo.AccountName
			retValueWithP8.PermissionAction=accountExistRel.Permissions
			accountQueryRet.Data=append(accountQueryRet.Data, retValueWithP8)
		}else{
			retValueWithoutP8.TeamId=accountExistRel.AccountInfo.TeamId
			retValueWithoutP8.UserName=accountExistRel.AccountInfo.UserName
			retValueWithoutP8.AccountType=accountExistRel.AccountInfo.AccountType
			retValueWithoutP8.AccountName=accountExistRel.AccountInfo.AccountName
			retValueWithoutP8.PermissionAction=accountExistRel.Permissions
			accountQueryRet.Data=append(accountQueryRet.Data, retValueWithoutP8)
		}
	}
	return accountQueryRet
}

func QueryAccount(c *gin.Context)  {
	logs.Info("从数据库中查询账户信息")
	userName:=c.DefaultQuery("user_name","")
	if userName==""{
		c.JSON(http.StatusOK,gin.H{
			"errcode" : "-1",
			"errinfo" : "未上传user_name",
		})
		return
	}
	accountQueryRet:=GetQueryAccountRet(userName)
	c.JSON(http.StatusOK,gin.H{
		"data":accountQueryRet.Data,
		"errcode" : "0",
		"errinfo" : "",
	})
}

func ReceiveP8file(c *gin.Context,accountInfo *dal.AccountInfo){
	file, header, _ :=c.Request.FormFile("account_p8file")
	if header==nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"errorInfo":   "没有文件上传",
		})
		return
	}
	logs.Info("打印File Name：" + header.Filename)
	p8ByteInfo,err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -2,
			"errorInfo":   "error read p8 file",
		})
		return
	}
	accountInfo.AccountP8file= string(p8ByteInfo)
}

func UpdateAccount(c *gin.Context)  {
	logs.Info("更新数据库中的账户信息")
	var accountInfo dal.AccountInfo
	ReceiveP8file(c,&accountInfo)
	bindError:=c.ShouldBind(&accountInfo)
	utils.RecordError("请求参数绑定错误: ", bindError)
	if bindError==nil {
		dbResult := dal.UpdateAccountInfo(accountInfo)
		if !dbResult {
			logs.Error("更新数据库中的账户信息失败！")
			c.JSON(http.StatusOK, gin.H{
				"errorCode": -1,
				"errorInfo": "更新数据库中的账户信息失败！",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"errorCode": "0",
			"message":   "update success",
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -2,
			"errorInfo": "请求参数绑定失败",
		})
	}
}


func InsertAccount(c *gin.Context)  {
	logs.Info("往数据库中添加账户信息")
	var accountInfo =dal.AccountInfo{}
	ReceiveP8file(c,&accountInfo)
	bindError:=c.ShouldBind(&accountInfo)
	utils.RecordError("请求参数绑定错误: ", bindError)
	if bindError==nil {
		dbResult := dal.InsertAccountInfo(accountInfo)
		if !dbResult {
			logs.Error("往数据库中插入数据失败")
			c.JSON(http.StatusOK, gin.H{
				"errorCode": -1,
				"errorInfo": "往数据库中插入数据失败！",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"errorCode": "0",
			"message":   "insert success",
		})
	}else{
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -2,
			"errorInfo": "请求参数绑定失败",
		})
	}
}

func GetTokenStringByAccInfo(accountInfo dal.AccountInfo) string{
	authKey, error := AuthKeyFromBytes([]byte(accountInfo.AccountP8file))
	if error != nil{
		logs.Info("读取authKey失败")
		return ""
	}
	token := jwt.New(jwt.SigningMethodES256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(15 * time.Minute).Unix()
	claims["iss"] = accountInfo.IssueId
	claims["aud"] = "appstoreconnect-v1"
	token.Claims = claims
	token.Header["kid"] = accountInfo.KeyId
	tokenString, err := token.SignedString(authKey)
	if err != nil{
		logs.Info("签token失败")
	}
	return tokenString
}

func GetTokenStringByTeamId(c *gin.Context){
	logs.Info("获取TokenString")
	teamId:=c.DefaultQuery("team_id","")
	if teamId==""{
		logs.Error("获取team_id参数失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -1,
			"errorInfo" : "获取team_id参数失败！",
		})
		return
	}else {
		condition := make(map[string]interface{})
		condition["team_id"] = teamId
		accountsInfo := dal.QueryAccountInfo(condition)
		if len(*accountsInfo)==0{
			logs.Error("team_id对应的记录不存在")
			c.JSON(http.StatusOK, gin.H{
				"errorCode" : -2,
				"errorInfo" : "team_id对应的记录不存在",
			})
			return
		}else {
			TokenString := GetTokenStringByAccInfo((*accountsInfo)[0])
			c.JSON(http.StatusOK, gin.H{
				"errorCode":   "0",
				"TokenString": TokenString,
			})
		}
	}
}
func GetTokenStringByTeamId2(teamId string) string{
	condition := make(map[string]interface{})
	condition["team_id"] = teamId
	accountsInfo := dal.QueryAccountInfo(condition)
	if len(*accountsInfo)==0{
		logs.Error("team_id对应的记录不存在")
		return ""
	}else {
		tokenString := GetTokenStringByAccInfo((*accountsInfo)[0])
		return tokenString
	}
}