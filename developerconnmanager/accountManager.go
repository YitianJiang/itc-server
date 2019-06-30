package developerconnmanager
//qwe
import (
	"code.byted.org/clientQA/ClusterManager/utils"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
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
		"message":   "delete success",
		"errorCode": "0",
	})
}


func QueryAccount(c *gin.Context)  {
	logs.Info("从数据库中查询账户信息")
	accountName:=c.DefaultQuery("account_name","")
	accountType:=c.DefaultQuery("account_type","")
	condition:=make(map[string]interface{})
	if accountName != ""{
		condition["account_name"] = accountName
	}
	if accountType != ""{
		condition["account_type"] = accountType
	}
	accountsInfo:=dal.QueryAccountInfo(condition)
	if accountsInfo==nil{
		logs.Error("从数据库中查询账户信息失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -1,
			"errorInfo" : "从数据库中查询账户信息失败！",
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"data":accountsInfo,
		"errorCode": "0",
		"message":   "query success",
	})
}

func ReceiveP8file(c *gin.Context,accountInfo *dal.AccountInfo){
	file, header, _ :=c.Request.FormFile("account_p8file")
	logs.Info("打印File Name：" + header.Filename)
	p8ByteInfo,err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   "error read p8 file",
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
