package detect

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

/**
json检测信息分析------fj
*/
func JsonInfoAnalysis(info string,mapInfo map[string]int){
	var infoMap = make(map[string]interface{})
	json.Unmarshal([]byte(info),&infoMap)
	appInfos := infoMap["app_info"].(map[string]interface{})
	methodsInfo := infoMap["method_sensitive_infos"].([]interface{})
	strsInfo := infoMap["str_sensitive_infos"].([]interface{})

	var detectInfo dal.DetectInfo
	detectInfo.TaskId = mapInfo["taskId"]
	detectInfo.ToolId = mapInfo["toolId"]
	AppInfoAnalysis(appInfos,&detectInfo)

	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : mapInfo["taskId"],
	})

	data := make(map[string]string)
	data["appId"] = (*detect)[0].AppId
	data["platform"] =  strconv.Itoa((*detect)[0].Platform)

	methodInfo,strInfos,_,err := getIgnoredInfo_2(data)
	if err != nil {
		logs.Error("未查询到该App的增量信息，app信息为：%v",data)
	}

	//敏感method去重
	mRepeat := make(map[string]int)
	newMethods := make([]map[string]interface{},0)
	for _, methodi := range methodsInfo {
		method := methodi.(map[string]interface{})
		var keystr = method["method_name"].(string)+method["method_class_name"].(string)
		if v,ok := mRepeat[keystr]; (!ok||ok&&v==0){
			newMethods = append(newMethods, method)
			mRepeat[keystr]=1
		}
	}
	allMethods := make([]dal.DetectContentDetail,0)
	for _,newMethod := range newMethods {
		var detailContent dal.DetectContentDetail
		var keystr = newMethod["method_class_name"].(string)+"."+newMethod["method_name"].(string)
		if v,ok := methodInfo[keystr]; ok {
			info := v.(map[string]interface{})
			detailContent.Status = info["status"].(int)
		}else{
			detailContent.Status = 0
		}
		detailContent.TaskId = mapInfo["taskId"]
		detailContent.ToolId = mapInfo["toolId"]
		//MethodAnalysis(newMethod,&detailContent)
		allMethods = append(allMethods,*MethodAnalysis_2(newMethod,&detailContent))
	}
	err1 := dal.InsertDetectDetailBatch(&allMethods)
	if err1 != nil {
		//及时报警
		message := "敏感method写入数据库失败，请解决;"+fmt.Sprint(err)
		utils.LarkDingOneInner("fanjuan.xqp", message)
	}

	allStrs := make([]dal.DetectContentDetail,0)
	for _, strInfoi := range strsInfo {
		strInfo := strInfoi.(map[string]interface{})
		var detailContent dal.DetectContentDetail
		detailContent.TaskId = mapInfo["taskId"]
		detailContent.ToolId = mapInfo["toolId"]
		allStrs = append(allStrs,*StrAnalysis_2(strInfo,&detailContent,strInfos))
	}
	err2 := dal.InsertDetectDetailBatch(&allStrs)
	if err2 != nil {
		//及时报警
		message := "敏感str写入数据库失败，请解决;"+fmt.Sprint(err)
		utils.LarkDingOneInner("fanjuan.xqp",message)
	}
	return
}

/**
appInfo解析，并写入数据库-------fj
*/
func AppInfoAnalysis(info map[string]interface{},detectInfo *dal.DetectInfo)  {
	taskId := detectInfo.TaskId
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : taskId,
	})
	if _,ok := info["apk_name"]; ok {
		(*detect)[0].AppName = info["apk_name"].(string)
		detectInfo.ApkName = info["apk_name"].(string)
	}
	if _,ok := info["apk_version_name"]; ok {
		(*detect)[0].AppVersion = info["apk_version_name"].(string)
		detectInfo.Version = info["apk_version_name"].(string)
	}

	if err := dal.UpdateDetectModelNew((*detect)[0]); err != nil {
		logs.Error("任务id:%s信息更新失败，%v",taskId,err)
		return
	}

	_,ok := info["channel"]
	if ok {
		detectInfo.Channel = info["channel"].(string)
	}
	_,ok1 := info["permissions"]

	var permissionArr []interface{}
	if ok1 {
		permissionArr = info["permissions"].([]interface{})
	}

	perStr :=""
	for _,per := range permissionArr {
		perStr += per.(string) +";"
	}
	detectInfo.Permissions = perStr

	err := dal.InsertDetectInfo(*detectInfo)
	if err != nil {
		//及时报警
		message := "appInfo写入数据库失败，请解决;"+fmt.Sprint(err)
		utils.LarkDingOneInner("fanjuan.xqp", message)
	}
	return
}

/**
method解析,并写入数据库-------fj
*/
func MethodAnalysis(method map[string]interface{},detail *dal.DetectContentDetail)  {
	detail.SensiType = 1
	//detail.Status = 0

	if _,ok := method["method_name"]; ok{
		detail.KeyInfo = method["method_name"].(string)
	}
	if _,ok := method["desc"]; ok {
		detail.DescInfo = method["desc"].(string)
	}

	if _,ok := method["method_class_name"]; ok {
		detail.ClassName = method["method_class_name"].(string)
	}

	var call []interface{}
	if _,ok := method["call_location"]; ok {
		call = method["call_location"].([]interface{})
	}

	callLocation := MethodRmRepeat(call)
	detail.CallLoc = callLocation

	err := dal.InsertDetectDetail(*detail)
	if err != nil {
		//及时报警
		message := "敏感method写入数据库失败，请解决;"+fmt.Sprint(err)+"\n敏感方法名："+fmt.Sprint(detail.KeyInfo)
		utils.LarkDingOneInner("fanjuan.xqp", message)
	}
	return
}

func MethodAnalysis_2(method map[string]interface{},detail *dal.DetectContentDetail) *dal.DetectContentDetail {
	detail.SensiType = 1
	//detail.Status = 0

	if _,ok := method["method_name"]; ok{
		detail.KeyInfo = method["method_name"].(string)
	}
	if _,ok := method["desc"]; ok {
		detail.DescInfo = method["desc"].(string)
	}

	if _,ok := method["method_class_name"]; ok {
		detail.ClassName = method["method_class_name"].(string)
	}

	var call []interface{}
	if _,ok := method["call_location"]; ok {
		call = method["call_location"].([]interface{})
	}

	callLocation := MethodRmRepeat(call)
	detail.CallLoc = callLocation

	return detail
}

/**
apk敏感方法内容去重--------fj
*/
func MethodRmRepeat(callInfo []interface{}) string  {
	repeatMap := make(map[string]int)
	result := ""
	for _,info1 := range callInfo {
		info := info1.(map[string]interface{})
		var keystr string
		keystr = info["class_name"].(string)+info["method_name"].(string)+fmt.Sprint(info["line_number"])
		if v,ok := repeatMap[keystr]; (!ok||(ok&&v==0)) {
			repeatMap[keystr] = 1
			mapInfo,_ := json.Marshal(info)
			result += string(mapInfo)+";"
		}
	}
	return result
}

/**
str解析，并写入数据库--------fj
*/
func StrAnalysis(str map[string]interface{},detail *dal.DetectContentDetail,strInfos map[string]interface{})  {
	detail.SensiType = 2
	//detail.Status = 0

	var keys []interface{}
	if _, ok := str["keys"]; ok {
		keys = str["keys"].([]interface{})
	}

	key := ""
	//判断str是否进行状态转变
	key2 := ""
	for _,ks1 := range keys {
		ks := ks1.(string)
		if _,ok := strInfos[ks]; !ok {
			key2 += ks
		}
		key += ks +";"
	}
	detail.KeyInfo = key
	if key2 == ""{
		detail.Status = 1
	}else {
		detail.Status = 0
	}

	if _,ok := str["desc"]; ok {
		detail.DescInfo = str["desc"].(string)
	}

	var callInfo []interface{}
	if _, ok := str["call_location"]; ok {
		callInfo = str["call_location"].([]interface{})
	}

	//敏感字段信息去重
	call_location := StrRmRepeat(callInfo)
	detail.CallLoc = call_location

	//方法和字符串优先级都是0
	//detail.Priority =0
	err := dal.InsertDetectDetail(*detail)
	if err != nil {
		//及时报警
		message := "敏感str写入数据库失败，请解决;"+fmt.Sprint(err)+"\n敏感方法名："+fmt.Sprint(key)
		utils.LarkDingOneInner("fanjuan.xqp",message)
	}
	return

}


func StrAnalysis_2(str map[string]interface{},detail *dal.DetectContentDetail,strInfos map[string]interface{}) *dal.DetectContentDetail {
	detail.SensiType = 2
	//detail.Status = 0

	var keys []interface{}
	if _, ok := str["keys"]; ok {
		keys = str["keys"].([]interface{})
	}

	key := ""
	//判断str是否进行状态转变
	key2 := ""
	for _,ks1 := range keys {
		ks := ks1.(string)
		if _,ok := strInfos[ks]; !ok {
			key2 += ks
		}
		key += ks +";"
	}
	detail.KeyInfo = key
	if key2 == ""{
		detail.Status = 1
	}else {
		detail.Status = 0
	}

	if _,ok := str["desc"]; ok {
		detail.DescInfo = str["desc"].(string)
	}

	var callInfo []interface{}
	if _, ok := str["call_location"]; ok {
		callInfo = str["call_location"].([]interface{})
	}

	//敏感字段信息去重
	call_location := StrRmRepeat(callInfo)
	detail.CallLoc = call_location

	//方法和字符串优先级都是0
	//detail.Priority =0
	return detail

}


/**
apk敏感字符串去重--------fj
*/
func StrRmRepeat(callInfo []interface{}) string {
	repeatMap := make(map[string]int)
	result := ""
	for _, info1 := range callInfo {
		info := info1.(map[string]interface{})
		var keystr string
		keystr = info["class_name"].(string) + info["method_name"].(string) + info["key"].(string) + fmt.Sprint(info["line_number"])
		if v, ok := repeatMap[keystr]; (!ok || (ok && v == 0)) {
			repeatMap[keystr] = 1
			mapInfo, _ := json.Marshal(info)
			result += string(mapInfo) + ";"
		}
	}
	return result
}

/**
 *确认安卓二进制包检测结果，更新数据库（包括确认信息入库），并判断是否停止lark消息--------fj
 */
func ConfirmApkBinaryResultv_3(c *gin.Context){
	type confirm struct {
		TaskId  int 	`json:"taskId"`
		Id  	int		`json:"id"`
		Status  int 	`json:"status"`
		Remark  string	`json:"remark"`
		ToolId	int		`json:"toolId"`
	}
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t confirm
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("参数不合法 ，%v",err)
		c.JSON(http.StatusOK, gin.H{
			"message" : "参数不合法！",
			"errorCode" : -1,
			"data" : "参数不合法！",
		})
		return
	}
	//获取确认人信息
	username, _ := c.Get("username")


	//获取详情信息
	condition1 := "id="+strconv.Itoa(t.Id)
	detailInfo, err := dal.QueryDetectContentDetail(condition1)
	if err != nil || detailInfo == nil||len(*detailInfo)==0{
		logs.Error("不存在该检测结果，ID：%d",t.Id)
		c.JSON(http.StatusOK, gin.H{
			"message" : "不存在该检测结果！",
			"errorCode" : -1,
			"data" : "不存在该检测结果！",
		})
		return
	}
	//获取任务信息
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : t.TaskId,
	})
	if (*detect) == nil {
		logs.Error("未查询到该taskid对应的检测任务，%v", t.TaskId)
		c.JSON(http.StatusOK, gin.H{
			"message" : "未查询到该taskid对应的检测任务",
			"errorCode" : -2,
			"data" : "未查询到该taskid对应的检测任务",
		})
		return
	}

	var data map[string]string
	data = make(map[string]string)
	data["id"] = strconv.Itoa(t.Id)
	data["confirmer"] = username.(string)
	data["remark"] = t.Remark
	data["status"] = strconv.Itoa(t.Status)
	flag := dal.ConfirmApkBinaryResultNew(data)
	if !flag {
		logs.Error("二进制检测内容确认失败")
		c.JSON(http.StatusOK, gin.H{
			"message" : "二进制检测内容确认失败！",
			"errorCode" : -1,
			"data" : "二进制检测内容确认失败！",
		})
		return
	}

	//增量忽略结果录入
	if t.Status != 0 {
		senType := (*detailInfo)[0].SensiType
		if senType == 1 {
			var igInfo dal.IgnoreInfoStruct
			igInfo.Platform = (*detect)[0].Platform
			igInfo.AppId,_ = strconv.Atoi((*detect)[0].AppId)
			igInfo.SensiType = 1
			igInfo.KeysInfo = (*detailInfo)[0].ClassName+"."+(*detailInfo)[0].KeyInfo
			igInfo.Confirmer = username.(string)
			igInfo.Remarks = t.Remark
			igInfo.Version = (*detect)[0].AppVersion
			igInfo.Status = t.Status
			err := dal.InsertIgnoredInfo(igInfo)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"message" : "增量信息更新失败！",
					"errorCode" : -1,
					"data" : "增量信息更新失败！",
				})
				return
			}
		}else {
			keys := strings.Split((*detailInfo)[0].KeyInfo,";")
			for _,key := range keys[0:len(keys)-1] {
				var igInfos dal.IgnoreInfoStruct
				igInfos.SensiType = 2
				igInfos.KeysInfo = key
				igInfos.AppId,_ = strconv.Atoi((*detect)[0].AppId)
				igInfos.Platform = (*detect)[0].Platform
				igInfos.Confirmer = username.(string)
				igInfos.Remarks = t.Remark
				igInfos.Status = t.Status
				igInfos.Version = (*detect)[0].AppVersion
				err := dal.InsertIgnoredInfo(igInfos)
				if err != nil {
					c.JSON(http.StatusOK, gin.H{
						"message" : "增量信息更新失败！",
						"errorCode" : -1,
						"data" : "增量信息更新失败！",
					})
					return
				}
			}
		}
	}

	//任务状态更新
	condition := "deleted_at IS NULL and task_id='" + strconv.Itoa(t.TaskId) + "' and tool_id='" + strconv.Itoa(t.ToolId) + "' and status= 0"
	counts := dal.QueryUnConfirmDetectContent(condition)
	if counts == 0 {
		(*detect)[0].Status =1
		err := dal.UpdateDetectModelNew((*detect)[0])
		if err != nil {
			logs.Error("任务确认状态更新失败！%v",err)
			c.JSON(http.StatusOK, gin.H{
				"message" : "任务确认状态更新失败！",
				"errorCode" : -1,
				"data" : "任务确认状态更新失败！",
			})
			return
		}
	}

	logs.Info("confirm success +id :%s",t.Id)
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
	return
}

/**
 *安卓增量查询二进制检查结果信息-------fj
 */
func QueryTaskApkBinaryCheckContentWithIgnorance_2(c *gin.Context){
	taskId := c.DefaultQuery("taskId", "")
	if taskId == "" {
		logs.Error("缺少taskId参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少taskId参数",
			"errorCode" : -1,
			"data" : "缺少taskId参数",
		})
		return
	}
	toolId := c.DefaultQuery("toolId", "")
	if toolId == "" {
		logs.Error("缺少toolId参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少toolId参数",
			"errorCode" : -2,
			"data" : "缺少toolId参数",
		})
		return
	}

	//切换到旧版本
	if toolId != "6" {
		QueryTaskBinaryCheckContent(c)
		return
	}

	queryType := c.DefaultQuery("type","")
	var flag = false
	var methodIgs = make(map[string]interface{})
	var strIgs = make(map[string]interface{})
	//var perIgs = make(map[string]interface{})
	var errIg error
	if queryType == ""{
		flag = true
		//获取任务信息
		detect := dal.QueryDetectModelsByMap(map[string]interface{}{
			"id" : taskId,
		})
		if (*detect) == nil || len(*detect)==0{
			logs.Error("未查询到该taskid对应的检测任务，%v", taskId)
			c.JSON(http.StatusOK, gin.H{
				"message" : "未查询到该taskid对应的检测任务",
				"errorCode" : -2,
				"data" : "未查询到该taskid对应的检测任务",
			})
			return
		}
		queryData := make(map[string]string)
		queryData["appId"] = (*detect)[0].AppId
		queryData["platform"] = strconv.Itoa((*detect)[0].Platform)

		//如果可忽略信息没有的话,录入日志但不影响后续操作
		methodIgs,strIgs,_,errIg = getIgnoredInfo_2(queryData)
		if errIg != nil {
			logs.Error("可忽略信息数据库查询失败,%v",errIg)
		}
	}

	condition := "task_id='" + taskId + "' and tool_id='" + toolId + "'"

	content,err := dal.QueryDetectInfo(condition)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message" : "查询检测结果信息数据库操作失败",
			"errorCode" : -1,
			"data" : err,
		})
		return
	}

	details, err := dal.QueryDetectContentDetail(condition)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message" : "查询检测结果详情数据库操作失败",
			"errorCode" : -1,
			"data" : err,
		})
		return
	}

	if (content == nil || details==nil||len(*details)==0){
		logs.Info("未查询到该任务对应的检测内容")
		c.JSON(http.StatusOK, gin.H{
			"message" : "未查询到该任务对应的检测内容",
			"errorCode" : -1,
			"data" : "未查询到该任务对应的检测内容",
		})
		return
	}

	//结果数据重组
	var queryResult dal.DetectQueryStruct
	queryResult.Channel = (*content).Channel
	queryResult.ApkName = (*content).ApkName
	queryResult.Version = (*content).Version

	permission := ""
	perms := strings.Split((*content).Permissions,";")
	for _,perm := range perms[0:(len(perms)-1)]{
		permission += perm +"\n"
	}
	queryResult.Permissions = permission

	//methods := make([]dal.SMethod,0)
	methods_un := make([]dal.SMethod,0)
	methods_con := make([]dal.SMethod,0)
	//strs := make([]dal.SStr,0)
	strs_un := make([]dal.SStr,0)
	strs_con := make([]dal.SStr,0)
	permissions := make([]dal.Permissions,0)
	//perm_un := make([]dal.Permissions,0)
	//perm_con := make([]dal.Permissions,0)

	for _,detail := range (*details) {
		if detail.SensiType == 1 {
			var method dal.SMethod
			method.Status = detail.Status
			method.Confirmer = detail.Confirmer
			method.Remark = detail.Remark
			method.ClassName = detail.ClassName
			method.Desc = detail.DescInfo
			method.Status = detail.Status
			method.Id = detail.ID
			method.MethodName = detail.KeyInfo
			//method.OtherVersion = detail.OtherVersion
			if flag && methodIgs != nil {
				if v,ok := methodIgs[detail.ClassName+"."+detail.KeyInfo]; ok{
					info := v.(map[string]interface{})
					if info["status"] != 0 {
						method.Status = info["status"].(int)
						method.Confirmer = info["confirmer"].(string)
						method.Remark = info["remarks"].(string)
						//detail.UpdatedAt = info["updateTime"].(time.Time)
						method.OtherVersion = info["version"].(string)
					}
				}
			}
			callLocs := strings.Split(detail.CallLoc,";")
			callLoc :=make([]dal.MethodCallJson,0)
			for _,call_loc := range callLocs[0:(len(callLocs)-1)] {
				var call_loc_json dal.MethodCallJson
				err := json.Unmarshal([]byte(call_loc),&call_loc_json)
				if err != nil {
					logs.Error("callLoc数据不符合要求，%v===========%s",err,call_loc)
					c.JSON(http.StatusOK, gin.H{
						"message" : "callLoc数据不符合要求",
						"errorCode" : 0,
						"data" : "callLoc数据不符合要求",
					})
					return
				}
				callLoc = append(callLoc,call_loc_json)
			}
			method.CallLoc = callLoc
			if method.Status ==0 {
				methods_un = append(methods_un,method)
			}else {
				methods_con = append(methods_con,method)
			}
			//methods = append(methods,method)
		}else if detail.SensiType == 2{
			keys2 := make(map[string]int)
			//var keys3 = detail.KeyInfo
			var str dal.SStr
			str.Status = detail.Status
			confirmInfos := make([]dal.ConfirmInfo,0)
			if flag && strIgs != nil{
				keys := strings.Split(detail.KeyInfo,";")
				keys3 := ""
				for _,keyInfo := range keys[0:len(keys)-1] {
					if v,ok := strIgs[keyInfo]; ok{
						info := v.(map[string]interface{})
						if info["status"] != 0 {
							keys2[keyInfo] = 1
							var confirmInfo dal.ConfirmInfo
							confirmInfo.Key = keyInfo
							confirmInfo.Remark = info["remarks"].(string)
							confirmInfo.Confirmer = info["confirmer"].(string)
							confirmInfo.OtherVersion = info["version"].(string)
							confirmInfos = append(confirmInfos,confirmInfo)
						}
					} else{
						keys3 += keyInfo+";"
					}
				}
				if keys3 =="" {
					str.Status = 1
				}
			}
			str.Keys = detail.KeyInfo
			str.Remark = detail.Remark
			str.Confirmer = detail.Confirmer
			str.Desc = detail.DescInfo
			str.Id = detail.ID
			callLocs := strings.Split(detail.CallLoc,";")
			callLoc := make([]dal.StrCallJson,0)
			for _,call_loc := range callLocs[0:(len(callLocs)-1)] {
				var callLoc_json dal.StrCallJson
				err := json.Unmarshal([]byte(call_loc),&callLoc_json)
				if err != nil {
					logs.Error("callLoc数据不符合要求，%v========%s",err,call_loc)
					c.JSON(http.StatusOK, gin.H{
						"message" : "callLoc数据不符合要求",
						"errorCode" : 0,
						"data" : "callLoc数据不符合要求",
					})
					return
				}
				if flag && strIgs != nil {
					if vv,ok := keys2[callLoc_json.Key]; ok&&vv==1{
						continue
					}
				}
				callLoc = append(callLoc,callLoc_json)
			}
			str.CallLoc = callLoc
			str.ConfirmInfos = confirmInfos
			if str.Status == 0 {
				strs_un = append(strs_un,str)
			}else{
				strs_con = append(strs_con,str)
			}
			//strs = append(strs,str)
		}
	}
	for _,m := range methods_con {
		methods_un = append(methods_un,m)
	}
	for _,str := range strs_con {
		strs_un = append(strs_un,str)
	}
	//for _,permInfo := range perm_con {
	//	perm_un = append(perm_un,permInfo)
	//}
	queryResult.SMethods = methods_un
	queryResult.SStrs = strs_un
	queryResult.Permissions_2 = permissions

	logs.Info("query detect result success!")
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : queryResult,
	})
	return
}
/**
 *获取可忽略内容
 */
func getIgnoredInfo_2(data map[string]string) (map[string]interface{},map[string]interface{},map[string]interface{},error) {
	condition := "app_id ='"+data["appId"]+"' and platform = '"+data["platform"]+"'"
	queryInfo := make(map[string]string)
	queryInfo["condition"] = condition
	result,err := dal.QueryIgnoredInfo(queryInfo)
	if err != nil {
		return nil,nil,nil,err
	}
	if result == nil || len(*result)==0{
		return nil,nil,nil,nil
	}
	methodMap := make(map[string]interface{})
	strMap := make(map[string]interface{})
	perMap := make(map[string]interface{})
	for i:=0;i<len(*result);i++{
		if (*result)[i].SensiType == 1 {
			if _,ok := methodMap[(*result)[i].KeysInfo];!ok {
				info := make(map[string]interface{})
				info["status"] = (*result)[i].Status
				info["remarks"] = (*result)[i].Remarks
				info["confirmer"] = (*result)[i].Confirmer
				info["version"] = (*result)[i].Version
				methodMap[(*result)[i].KeysInfo]=info
			}
		}else if (*result)[i].SensiType == 2{
			if _,ok := strMap[(*result)[i].KeysInfo];!ok {
				info := make(map[string]interface{})
				info["status"] = (*result)[i].Status
				info["remarks"] = (*result)[i].Remarks
				info["confirmer"] = (*result)[i].Confirmer
				info["version"] = (*result)[i].Version
				//info["updateTime"] = (*result)[i].UpdatedAt
				strMap[(*result)[i].KeysInfo]=info
			}
		}else {
			if _,ok := perMap[(*result)[i].KeysInfo];!ok {
				info := make(map[string]interface{})
				info["status"] = (*result)[i].Status
				info["remarks"] = (*result)[i].Remarks
				info["confirmer"] = (*result)[i].Confirmer
				info["version"] = (*result)[i].Version
				//info["updateTime"] = (*result)[i].UpdatedAt
				perMap[(*result)[i].KeysInfo]=info
			}
		}
	}
	return methodMap,strMap,perMap,nil
}

func QueryIgnoredHistory(c *gin.Context)  {
	type queryData struct {
		AppId		int			`json:"appId"`
		Platform 	int			`json:"platform"`
		Key			string		`json:"key"`
	}
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t queryData
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("参数不合法 ，%v",err)
		c.JSON(http.StatusOK, gin.H{
			"message" : "参数不合法！",
			"errorCode" : -1,
			"data" : "参数不合法！",
		})
		return
	}
	logs.Info(t.Key)
	queryDatas := make(map[string]string)
	queryDatas["condition"] = "app_id='"+strconv.Itoa(t.AppId)+"' and platform='"+strconv.Itoa(t.Platform)+"' and keys_info='"+t.Key+"'"
	result,err := dal.QueryIgnoredInfo(queryDatas)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message" : "查询确认历史失败",
			"errorCode" : -1,
			"data" : "查询确认历史失败",
		})
		return
	}

	data := make([]map[string]interface{},0)

	for _,res := range (*result) {
		dd := map[string]interface{}{
			"key":        res.KeysInfo,
			"updateTime": res.UpdatedAt,
			"remark":     res.Remarks,
			"confirmer":  res.Confirmer,
			"version":    res.Version,
		}
		data = append(data,dd)
	}
	logs.Info("查询确认历史成功")
	c.JSON(http.StatusOK,gin.H{
		"message":"success",
		"errorCode":0,
		"data":data,
	})
	return

}


