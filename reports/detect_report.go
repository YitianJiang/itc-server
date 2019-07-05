package reports

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func QueryDetectReports(c *gin.Context){
	param,_ := ioutil.ReadAll(c.Request.Body)
	var t dal.QueryReportInfo
	if err := json.Unmarshal(param,&t);err != nil {
		logs.Error("报告查询参数格式不合法，errInfo:%v",err)
		errorReturn(c,"报告查询参数格式不合法")
		return
	}
	//时间数据判断
	if t.StartTime.After(t.EndTime) {
		message := "时间选择错误（开始时间在结束时间之后）"
		logs.Error(message)
		errorReturn(c,message)
		return
	}

	condition := "platform = '"+fmt.Sprint(t.Platform)+"'"
	if t.AppId != 0 {
		condition += " app_id = '"+fmt.Sprint(t.AppId)+"'"
	}

	infos,errs := dal.QueryReportsInfo(condition)
	if errs != nil {
		logs.Error("获取信息失败")
		errorReturn(c,"获取信息失败")
		return
	}
	//此处返回是否为error还是空数组
	if infos == nil || len(*infos)==0 {
		message := "所选条件内无符合条件检测任务，条件为" +fmt.Sprint(t)
		logs.Error(message)
		errorReturn(c,message)
		return
	}










}

///**
//	组织检测数据报表返回数据---检测数量+确认
// */
//func getTaskChartsData(infos *[]dal.DetectStruct, t dal.QueryReportInfo) []map[string]interface{}{
//
//
//}
//
///**
//	组织权限数据报表返回数据
// */
//func getPermChartsData(infos *[]dal.DetectConfigStruct,t dal.QueryReportInfo) []map[string]interface{}{
//
//}

/**
	请求错误信息返回统一格式
*/
func errorReturn(c *gin.Context,message string,errs ... error){
	c.JSON(http.StatusOK, gin.H{
		"message" : message,
		"errorCode" : -1,
		"data" : message,
	})
	return
}
