package developerconnmanager

import (
	"bytes"
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func DeleteByTeamId(c *gin.Context)  {
	logs.Info("根据team_id参数将数据从数据库中删除")
	var delAccRequest devconnmanager.DelAccRequest
	bindQueryError:=c.ShouldBindQuery(&delAccRequest)
	if bindQueryError!=nil{
		c.JSON(http.StatusOK, gin.H{
			"message":   "delete fail",
			"errorCode":  1,
			"errorInfo": "请求参数绑定失败",
		})
		return
	}
	teamId:=delAccRequest.TeamId
	dbResult:=devconnmanager.DeleteAccountInfo(teamId)
	if dbResult==-2{
		logs.Error("从数据库中删除数据失败!,失败原因为数据库中不存在该条数据")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 2,
			"errorInfo" : "从数据库中删除数据失败！,失败原因为数据库中不存在该条数据",
		})
		return
	}
	if dbResult==-1||dbResult==-3{
		logs.Error("从数据库中删除数据失败!")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : 3,
			"errorInfo" : "从数据库中删除数据失败!",
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"errorCode": 0,
		"message":   "delete success",
	})
}

func QueryAccount(c *gin.Context)  {
	logs.Info("从数据库中查询账户信息")
	userName:=c.DefaultQuery("user_name","")
	if userName==""{
		c.JSON(http.StatusOK,gin.H{
			"errcode" : 1,
			"errinfo" : "未上传user_name",
		})
		return
	}
	var resPerms devconnmanager.GetPermsResponse
	url:=_const.USER_ALL_RESOURCES_PERMS_URL+"userName="+userName
	result:=QueryPerms(url,&resPerms)
	if !result{
		c.JSON(http.StatusOK,gin.H{
			"errcode" : 2,
			"errinfo" : "查询权限失败",
		})
		return
	}
	var accountsInfo *[]interface{}
	accountsInfo=devconnmanager.QueryAccInfoWithAuth(&resPerms)
	if accountsInfo==nil{
		c.JSON(http.StatusOK,gin.H{
			"errcode" : 3,
			"errinfo" : "从数据库中查询数据失败",
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"data":accountsInfo,
		"errcode" : 0,
		"errinfo" : "",
	})
}

func ReceiveP8file(c *gin.Context,accountInfo *devconnmanager.AccountInfo) bool{
	file, header, _ :=c.Request.FormFile("account_p8file")
	if header==nil {
		return true
	}
	logs.Info("打印File Name：" + header.Filename)
	p8ByteInfo,err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 2,
			"errorInfo":   "error read p8 file",
		})
		return false
	}
	accountInfo.AccountP8file= string(p8ByteInfo)
	return true
}

func UpdateAccount(c *gin.Context)  {
	logs.Info("更新数据库中的账户信息")
	var accountInfo devconnmanager.AccountInfo
	recResult:=ReceiveP8file(c,&accountInfo)
	if !recResult{
		return
	}
	bindError:=c.ShouldBind(&accountInfo)
	utils.RecordError("请求参数绑定错误: ", bindError)
	if bindError!=nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 1,
			"errorInfo": "请求参数绑定失败",
		})
		return
	}
	dbResult := devconnmanager.UpdateAccountInfo(accountInfo)
	if dbResult==-1 {
		logs.Error("更新数据库中的账户信息失败！")
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 2,
			"errorInfo": "更新数据库中的账户信息失败！",
		})
		return
	}
	if dbResult==-2 {
		logs.Error("更新数据库中的账户信息失败！,数据库中不存在team_id对应的记录")
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 3,
			"errorInfo": "更新数据库中的账户信息失败！，数据库中不存在team_id对应的记录",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errorCode": 0,
		"message":   "update success",
	})
}


func CreateResource(creResRequest devconnmanager.CreateResourceRequest) int{
	bodyByte, _ := json.Marshal(creResRequest)
	rbodyByte := bytes.NewReader(bodyByte)
	client := &http.Client{}
	request, err := http.NewRequest("POST", _const.Create_RESOURCE_URL,rbodyByte)
	if err != nil {
		logs.Info("新建request对象失败")
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送post请求失败")
	}
	defer response.Body.Close()
	var creResResponse devconnmanager.CreResResponse
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Info("读取respose的body内容失败")
	}
	json.Unmarshal(body,&creResResponse)
	if creResResponse.Errno==0{
		return  0
	}
	if creResResponse.Errno==500{
		return -1
	}
	return  -2
}

func InsertAccount(c *gin.Context)  {
	logs.Info("往数据库中添加账户信息")
	var accountInfo =devconnmanager.AccountInfo{}
	recResult:=ReceiveP8file(c,&accountInfo)
	if !recResult{
		return
	}
	bindError:=c.ShouldBind(&accountInfo)
	utils.RecordError("请求参数绑定错误: ", bindError)
	if bindError!=nil{
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 1,
			"errorInfo": "请求参数绑定失败",
		})
		return
	}
	dbResult := devconnmanager.InsertAccountInfo(accountInfo)
	if dbResult==-1 {
		logs.Error("往数据库中插入账号信息失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 2,
			"errorInfo": "往数据库中插入账号信息失败",
		})
		return
	}
	if dbResult==-2 {
		logs.Error("往数据库中插入账号信息失败,数据库中已经存在该记录")
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 3,
			"errorInfo": "往数据库中插入账号信息失败,数据库中已经存在该记录！",
		})
		return
	}
	var creResRequest devconnmanager.CreateResourceRequest
	creResRequest.CreatorKey=accountInfo.UserName
	teamIdLower:=strings.ToLower(accountInfo.TeamId)
	creResRequest.ResourceKey=teamIdLower+"_space_account"
	creResRequest.ResourceName=teamIdLower+"_space_account"
	creResRequest.ResourceType=0
	creResult:=CreateResource(creResRequest)
	if creResult==-1{
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 4,
			"errorInfo": "往数据库中插入账号信息成功，但由于资源已经存在，创建资源失败！",
		})
		return
	}
	if creResult==-2{
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 5,
			"errorInfo": "往数据库中插入账号信息成功，但创建资源失败,错误原因在kani平台！",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"errorCode": 0,
		"message":   "往数据库中插入账号信息成功，同时资源创建成功",
	})
}

func GetTokenStringByAccInfo(accountInfo devconnmanager.AccountInfo) string{
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


func GetTokenStringByTeamId(teamId string) string{
	condition := make(map[string]interface{})
	condition["team_id"] = teamId
	accountsInfo := devconnmanager.QueryAccountInfo(condition)
	if len(*accountsInfo)==0{
		logs.Error("team_id对应的记录不存在,无法签出token")
		return ""
	}else {
		tokenString := GetTokenStringByAccInfo((*accountsInfo)[0])
		return tokenString
	}
}