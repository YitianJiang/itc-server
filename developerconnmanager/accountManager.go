package developerconnmanager

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
	"net/http"
)

func DeleteByTeamId(c *gin.Context)  {
	logs.Info("根据team_id参数将数据从数据库中删除")
	teamId:=c.Query("team_id")
	dbResult:=dal.DeleteAccount(teamId)
	if !dbResult{
		logs.Error("从数据库中删除数据失败")
		c.JSON(http.StatusOK, gin.H{
			"errorCode" : -1,
			"errorInfo" : "从数据库中删除数据失败！",
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"message":   "delete success",
		"errorCode": "0",
	})
}


//func UpdataAccountInfo(c *gin.Context){
//	logs.Info("更新数据")
//	var account dal.Account
//	c.Bind(&account)
//
//
//}