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
	"sort"
	"strconv"
	"strings"
)


/**
	安卓json检测信息分析----兼容.aab格式检测结果----json到map
 */

func ApkJsonAnalysis (info string,mapInfo map[string]int){
	var fisrtResult = make(map[string]interface{})
	err_f := json.Unmarshal([]byte(info),&fisrtResult)
	if err_f != nil {
		logs.Error("二进制静态包检测返回信息格式错误！")
		message := "taskId:"+fmt.Sprint(mapInfo["taskId"])+",二进制静态包检测返回信息格式错误，请解决;"+fmt.Sprint(err_f)
		utils.LarkDingOneInner("fanjuan.xqp",message)
		return
	}
	//获取json返回结果数组
	results := fisrtResult["result"].([]interface{})
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : mapInfo["taskId"],
	})
	//遍历结果数组，并将每组检测结果信息插入数据库
	for index,result := range results {
		//结果数组中每个项是map[string]interface{}形式
		infoMap := result.(map[string]interface{})

		appInfos := infoMap["app_info"].(map[string]interface{})
		methodsInfo := infoMap["method_sensitive_infos"].([]interface{})
		strsInfo := infoMap["str_sensitive_infos"].([]interface{})

		var detectInfo dal.DetectInfo
		detectInfo.TaskId = mapInfo["taskId"]
		detectInfo.ToolId = mapInfo["toolId"]
		AppInfoAnalysis(appInfos,&detectInfo,index)

		data := make(map[string]string)
		data["appId"] = (*detect)[0].AppId
		data["platform"] =  strconv.Itoa((*detect)[0].Platform)

		//获取敏感方法和字符串的确认信息，为信息初始化做准备
		methodInfo,strInfos,_,err := getIgnoredInfo_2(data)
		if err != nil {
			logs.Error("未查询到该App的增量信息，app信息为：%v",data)
		}

		//敏感method去重
		mRepeat := make(map[string]int)
		newMethods := make([]map[string]interface{},0)//第一层去重后的敏感方法集
		for _, methodi := range methodsInfo {
			method := methodi.(map[string]interface{})
			var keystr = method["method_name"].(string)+method["method_class_name"].(string)
			if v,ok := mRepeat[keystr]; (!ok||ok&&v==0){
				newMethods = append(newMethods, method)
				mRepeat[keystr]=1
			}
		}
		//批量写入数据库的敏感方法struct集合
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
			//新增兼容下标
			detailContent.SubIndex = index
			//MethodAnalysis(newMethod,&detailContent)
			allMethods = append(allMethods,*MethodAnalysis_2(newMethod,&detailContent))
		}
		err1 := dal.InsertDetectDetailBatch(&allMethods)
		if err1 != nil {
			//及时报警
			message := "taskId:"+fmt.Sprint(mapInfo["taskId"])+",敏感method写入数据库失败，请解决;"+fmt.Sprint(err)
			utils.LarkDingOneInner("fanjuan.xqp", message)
		}

		allStrs := make([]dal.DetectContentDetail,0)
		for _, strInfoi := range strsInfo {
			strInfo := strInfoi.(map[string]interface{})
			var detailContent dal.DetectContentDetail
			detailContent.TaskId = mapInfo["taskId"]
			detailContent.ToolId = mapInfo["toolId"]
			detailContent.SubIndex = index
			allStrs = append(allStrs,*StrAnalysis_2(strInfo,&detailContent,strInfos))
		}
		err2 := dal.InsertDetectDetailBatch(&allStrs)
		if err2 != nil {
			//及时报警
			message := "taskId:"+fmt.Sprint(mapInfo["taskId"])+",敏感str写入数据库失败，请解决;"+fmt.Sprint(err)
			utils.LarkDingOneInner("fanjuan.xqp",message)
		}
	}
	//condition := "deleted_at IS NULL and task_id='" + strconv.Itoa(mapInfo["taskId"]) + "' and tool_id='" + strconv.Itoa(mapInfo["toolId"]) + "' and status= 0"
	//counts := dal.QueryUnConfirmDetectContent(condition)
	//
	//perms,_ := dal.QueryPermAppRelation(map[string]interface{}{
	//	"task_id":mapInfo["taskId"],
	//})
	//
	//var permFlag = true
	//if perms == nil || len(*perms)== 0 {
	//	logs.Error("该任务无权限检测信息！")
	//}else{
	//	P:
	//	for i:=0;i<len(*perms);i++{
	//		permsInfoDB := (*perms)[i].PermInfos
	//		var permList []interface{}
	//		json.Unmarshal([]byte(permsInfoDB),&permList)
	//
	//		for _,m := range permList {
	//			permInfo := m.(map[string]interface{})
	//			if permInfo["status"].(float64) == 0 {
	//				permFlag = false
	//				break P
	//			}
	//		}
	//	}
	//}
	//
	//logs.Notice("当前确认情况，字符串和方法剩余："+fmt.Sprint(counts)+" ,权限是否全部确认："+fmt.Sprint(permFlag))
	//if counts == 0 && permFlag {
	//	(*detect)[0].Status = 1
	//	err := dal.UpdateDetectModelNew((*detect)[0])
	//	if err != nil {
	//		logs.Error("任务确认状态更新失败！%v", err)
	//	}else {
	//		ConfirmCI((*detect)[0].CallBackUrl)
	//	}
	//}

	//任务状态更新----该app无需要特别确认的敏感方法、字符串或权限
	taskStatusUpdate(mapInfo["taskId"],mapInfo["toolId"],&(*detect)[0])
	return
}

/**
	appInfo解析，并写入数据库,此处包含权限的处理-------fj
	新增了index下标，兼容.aab结果中新增sub_index，默认为0
*/
func AppInfoAnalysis(info map[string]interface{},detectInfo *dal.DetectInfo,index ...int){
	//数组结果排序标识处理，默认为0
	var realIndex int
	if len(index) == 0 {
		realIndex = 0
	}else {
		realIndex = index[0]
	}

	taskId := detectInfo.TaskId
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : taskId,
	})
	appId,_ := strconv.Atoi((*detect)[0].AppId)
	//判断appInfo信息是否为主要信息，只有主要信息--primary为1才会修改任务的appName和Version
	var taskUpdateFlag = false
	if _,ok := info["primary"];ok {
		if info["primary"].(float64) == 1 {
			taskUpdateFlag = true
		}
	}else {
		//单独测apk文件没有此字段，但是默认为主要的
		taskUpdateFlag = true
	}

	if _,ok := info["apk_name"]; ok {
		//(*detect)[0].AppName = info["apk_name"].(string)
		detectInfo.ApkName = info["apk_name"].(string)
	}
	if _,ok := info["apk_version_name"]; ok {
		//(*detect)[0].AppVersion = info["apk_version_name"].(string)
		detectInfo.Version = info["apk_version_name"].(string)
	}

	if taskUpdateFlag {
		(*detect)[0].AppName = info["apk_name"].(string)
		(*detect)[0].AppVersion = info["apk_version_name"].(string)
		if err := dal.UpdateDetectModelNew((*detect)[0]); err != nil {
			message := "任务ID："+fmt.Sprint(taskId)+"，appName和Version信息更新失败，失败原因："+fmt.Sprint(err)
			logs.Error(message)
			utils.LarkDingOneInner("fanjuan.xqp",message)
			return
		}
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

	//更新任务的权限信息
	perStr :=""//旧版权限信息
	permInfos := make([]map[string]interface{},0)//新版权限信息
	/**
	map[string]interface{}{
		"perm_id":int,
		"key":string,
		"ability":string,
		"priority":int,
		"state":int,//表示是否定义
		"status":int//确认状态
		"first_version"://引入信息
	}
	 */

	larkPerms := ""
	var first_history []dal.PermHistory
	//获取app的权限操作历史map
	impMap := GetImportedPermission(appId)
	//判断是否属于初次引入
	var fhflag bool
	//权限去重map
	permRepeatMap := make(map[string]int)
	for _,per := range permissionArr {
		//增加权限逐条检测后，此处注释掉
		//perStr += per.(string) +";"
		pers := per.(string)

		//权限去重
		if v,okp := permRepeatMap[pers]; okp&&v==1 {
			continue
		}
		permRepeatMap[pers] = 1
		//写app和perm对应关系
		queryResult := dal.QueryDetectConfig(map[string]interface{}{
			"key_info":pers,
			"platform":0,
		})
		fhflag = false
		permInfo := make(map[string]interface{})
		if queryResult == nil || len(*queryResult)==0{
			var conf dal.DetectConfigStruct
			conf.KeyInfo = pers
			//将该权限的优先级定为--3高危
			conf.Priority = 3
			//暂时定为固定人选
			conf.Creator = "kanghuaisong"
			conf.Platform = 0
			perm_id,err := dal.InsertDetectConfig(conf)

			if err != nil {
				logs.Error("update回调时新增权限失败，%v",err)
				//及时报警
				utils.LarkDingOneInner("fanjuan.xqp","taskId:"+fmt.Sprint(taskId)+",update回调新增权限失败")
				return
			}else {
				fhflag = true
				larkPerms += "权限名为："+pers+"\n"
				permInfo["perm_id"] = perm_id
				permInfo["key"] = pers
				permInfo["ability"] = ""
				//优先级默认为3---高危
				permInfo["priority"] = 3
				//此处state表明该权限是自动添加，信息不全，后面query时需要重新读取相关信息
				permInfo["state"] = 0
				permInfo["status"] = 0
				permInfo["first_version"] = detectInfo.Version
			}
		}else{
			permInfo["perm_id"] = (*queryResult)[0].ID
			permInfo["key"] = pers
			permInfo["ability"] = (*queryResult)[0].Ability
			permInfo["priority"] = (*queryResult)[0].Priority
			permInfo["state"] = 1

			//更新确认信息
			if v,ok := impMap[int((*queryResult)[0].ID)]; !ok {
				//logs.Error("未查询到该权限的操作历史")
				permInfo["status"] = 0
				permInfo["first_version"] = detectInfo.Version
				fhflag = true
			}else {
				iMap := v.(map[string]interface{})
				permInfo["status"] = iMap["status"].(int)
				permInfo["first_version"] = iMap["version"].(string)
			}
		}
		//若是初次引入,写入引入信息
		if fhflag {
			var hist dal.PermHistory
			hist.Status = 0
			hist.AppId = appId
			hist.AppVersion = detectInfo.Version
			hist.PermId = int(permInfo["perm_id"].(uint))
			hist.Confirmer = (*detect)[0].Creator
			hist.Remarks = "包检测引入该权限"
			first_history = append(first_history,hist)
		}
		permInfos = append(permInfos,permInfo)
	}
	//若是初次引入,写入引入信息
	if len(first_history)>0 {
		errB := dal.BatchInsertPermHistory(&first_history)
		if errB != nil {
			logs.Error("taskId:"+fmt.Sprint(taskId)+",插入权限第一次引入历史失败")
			//及时报警
			utils.LarkDingOneInner("fanjuan.xqp","taskId:"+fmt.Sprint(taskId)+",插入权限第一次引入历史失败")
		}
	}
	//lark通知创建人完善权限信息-----只发一条消息
	if larkPerms != ""{
		message := "你好，安卓二进制静态包检测出未知权限，请去权限配置页面完善权限信息,需要完善的权限信息有：\n"
		message += larkPerms
		message += "修改链接：http://cloud.bytedance.net/rocket/itc/permission?biz=13"
		utils.LarkDingOneInner("kanghuaisong",message)
		//测试时使用
		//utils.LarkDingOneInner("fanjuan.xqp",message)
		//上线时使用
		//utils.LarkDingOneInner("lirensheng",message)
	}

	//更新权限-app-task关系表
	var relationship dal.PermAppRelation
	relationship.TaskId = taskId
	relationship.AppId = appId
	if taskUpdateFlag {
		relationship.AppVersion = (*detect)[0].AppVersion
	}else {
		relationship.AppVersion = ".aab副包+"+detectInfo.Version
	}
	relationship.AppVersion = detectInfo.Version
	//新增下标兼容.aab结果
	relationship.SubIndex = realIndex
	bytePerms,_ := json.Marshal(permInfos)
	relationship.PermInfos = string(bytePerms)
	//---------------------------失败时处理方式要再仔细看一下
	err1 := dal.InsertPermAppRelation(relationship)
	if err1 != nil {
		utils.LarkDingOneInner("fanjuan.xqp","新增权限App关系失败！taskId:"+fmt.Sprint(taskId)+",appID="+(*detect)[0].AppId)
	}
	detectInfo.Permissions = perStr
	detectInfo.SubIndex = realIndex
	//插入appInfo信息到apk表
	err := dal.InsertDetectInfo(*detectInfo)
	if err != nil {
		//及时报警
		message := "taskId:"+fmt.Sprint(taskId)+",appInfo写入数据库失败，请解决;"+fmt.Sprint(err)
		utils.LarkDingOneInner("fanjuan.xqp", message)
	}
	return
}


/**
批量method解析-----fj
 */
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
批量str解析---------fj
 */
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
	//通过或不通过的状态表示，true为1，false为2
    var	passFlag = true
	for _,ks1 := range keys {
		ks := ks1.(string)
		if v,ok := strInfos[ks]; !ok {
			key2 += ks
		}else{
			info := v.(map[string]interface{})
			if info["status"].(int) == 2 && passFlag {
				passFlag = false
			}
		}
		key += ks +";"
	}
	detail.KeyInfo = key
	if key2 == ""{
		if passFlag{
			detail.Status = 1
		}else{
			detail.Status = 2
		}
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


type PermSlice [] dal.Permissions

func (a PermSlice) Len() int {         // 重写 Len() 方法
	return len(a)
}
func (a PermSlice) Swap(i, j int){     // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a PermSlice) Less(i, j int) bool {    // 重写 Less() 方法， 从大到小排序
	return a[j].Priority< a[i].Priority
}


/**
	获取权限的确认历史信息------fj
 */

func GetIgnoredPermission(appId int) map[int]interface{}  {
	result := make(map[int]interface{})
	queryResult,err := dal.QueryPermHistory(map[string]interface{}{
		"app_id":appId,
	})
	if err != nil || queryResult == nil || len(*queryResult)== 0 {
		logs.Error("该app暂时没有确认信息")
	}else {
		for _, infoP := range (*queryResult) {
			if _, ok := result[infoP.PermId]; !ok {
				if infoP.Status >0 {//增加引入历史后，将此类信息过滤
					info := make(map[string]interface{})
					info["status"] = infoP.Status
					info["remarks"] = infoP.Remarks
					info["confirmer"] = infoP.Confirmer
					info["version"] = infoP.AppVersion
					result[infoP.PermId] = info
				}
			}
		}
	}
	return result
}

/**
	获取权限表基础信息
 */
func GetPermList() map[int]interface{}{
	result := make(map[int]interface{})
	queryResult := dal.QueryDetectConfig(map[string]interface{}{
		"platform":0,
	})
	if queryResult == nil || len(*queryResult) == 0 {
		logs.Error("权限信息表为空")
	}else {
		for _, infoP := range (*queryResult) {
			if _, ok := result[int(infoP.ID)]; !ok {
					info := make(map[string]interface{})
					info["ability"] = infoP.Ability
					info["priority"] = infoP.Priority
					result[int(infoP.ID)] = info
			}
		}
	}
	return result
}

/**
	获取权限引入历史
 */
func GetImportedPermission(appId int) map[int]interface{}  {
	result := make(map[int]interface{})
	queryResult,err := dal.QueryPermHistory(map[string]interface{}{
		"app_id":appId,
	})
	if err != nil || queryResult == nil || len(*queryResult)== 0 {
		logs.Error("该app暂时没有确认信息")
	}else {
		for _, infoP := range (*queryResult) {
			_, ok := result[infoP.PermId]
			if !ok {
					info := make(map[string]interface{})
					info["version"] = infoP.AppVersion
					info["status"] = infoP.Status
					result[infoP.PermId] = info
			}else if ok && infoP.Status == 0 {
				v := result[infoP.PermId].(map[string]interface{})
				v["version"] = infoP.AppVersion
				result[infoP.PermId] = v
			}
		}
	}
	return result

}

/**
 *获取可忽略内容
 */
func getIgnoredInfo_2(data map[string]string) (map[string]interface{},map[string]interface{},map[string]interface{},error) {
	condition := "app_id ='"+data["appId"]+"' and platform = '"+data["platform"]+"'"
	queryInfo := make(map[string]string)
	queryInfo["condition"] = condition
	result,err := dal.QueryIgnoredInfo(queryInfo)

	//此处如果条件1没有命中，但是23命中了，返回的err其实是nil
	if err != nil || result == nil || len(*result)==0{
		return nil,nil,nil,err
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

		}
	}
	return methodMap,strMap,perMap,nil
}

/**
	获取敏感方法和字符串的确认历史
 */

func QueryIgnoredHistory_2(c *gin.Context)  {
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


/*
	获取权限的确认历史，为了和iOS兼容，此处的内容key其实传ID就可以了-----fj
 */
func QueryIgnoredHistory(c *gin.Context)  {
	type queryData struct {
		AppId		int				`json:"appId"`
		Platform 	int				`json:"platform"`
		Key			interface{}		`json:"key"`
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
	if t.Platform == 0 {
		perm_id := int(t.Key.(float64))

		result,err := dal.QueryPermHistory(map[string]interface{}{
			"perm_id":perm_id,
			"app_id":t.AppId,
		})
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
				"key":        res.PermId,
				"updateTime": res.UpdatedAt,
				"remark":     res.Remarks,
				"confirmer":  res.Confirmer,
				"version":    res.AppVersion,
				"status": 	  res.Status,
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
	}else {
		history := dal.QueryPrivacyHistoryModel(map[string]interface{}{
			"app_id":     t.AppId,
			"platform":   t.Platform,
			"permission": t.Key,
		})
		if history == nil || len(*history) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"message":   "该key无确认历史",
				"errorCode": -1,
				"data":      []interface{}{},
			})
			return
		}
		var resList []map[string]interface{}
		for _, hh := range *history {
			temMap := map[string]interface{}{
				"key":        hh.Permission,
				"confirmer":  hh.Confirmer,
				"remark":     hh.ConfirmReason,
				"updateTime": hh.CreatedAt,
				"version":    hh.ConfirmVersion,
			}
			resList = append(resList, temMap)
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data":      resList,
		})
	}
}


/**
	安卓获取检测任务详情---数组形式
 */
func QueryTaskApkBinaryCheckContentWithIgnorance_3(c *gin.Context){
	taskId := c.DefaultQuery("taskId", "")
	if taskId == "" {
		logs.Error("缺少taskId参数")
		errorReturn(c,"缺少taskId参数")
		return
	}
	toolId := c.DefaultQuery("toolId", "")
	if toolId == "" {
		logs.Error("缺少toolId参数")
		errorReturn(c,"缺少toolId参数")
		return
	}

	//切换到旧版本
	if toolId != "6" {
		QueryTaskBinaryCheckContent(c)
		return
	}
	//一次查所有
	var methodIgs = make(map[string]interface{})
	var strIgs = make(map[string]interface{})
	var perIgs = make(map[int]interface{})
	allPermList := GetPermList()
	var appId int
	var errIg error
	//获取任务信息
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : taskId,
	})
	if detect == nil || len(*detect)==0{
		logs.Error("未查询到该taskid对应的检测任务，%v", taskId)
		errorReturn(c,"未查询到该taskid对应的检测任务")
		return
	}
	//查询增量信息
	queryData := make(map[string]string)
	queryData["appId"] = (*detect)[0].AppId
	appId,_ = strconv.Atoi((*detect)[0].AppId)
	queryData["platform"] = strconv.Itoa((*detect)[0].Platform)
	methodIgs,strIgs,_,errIg = getIgnoredInfo_2(queryData)
	if errIg != nil {
		logs.Error("可忽略信息数据库查询失败,%v",errIg)//如果可忽略信息没有的话,录入日志但不影响后续操作
	}
	condition := "task_id='" + taskId + "' and tool_id='" + toolId + "'"
	perIgs = GetIgnoredPermission(appId)

	//查询基础信息和敏感信息
	contents,err := dal.QueryDetectInfo_2(condition)
	if err != nil {
		errorReturn(c,"查询检测结果信息数据库操作失败")
		return
	}
	details,err2 := dal.QueryDetectContentDetail(condition)
	if err2 != nil {
		errorReturn(c,"查询检测结果信息数据库操作失败")
		return
	}
	if contents == nil || len(*contents) == 0 || details == nil || len(*details)==0{
		logs.Info("未查询到该任务对应的检测内容,taskId:"+taskId)
		errorReturn(c,"未查询到该任务对应的检测内容")
		return
	}
	//查询权限信息，此处为空--代表旧版无权限确认部分
	info, err3 := dal.QueryPermAppRelation(map[string]interface{}{
		"task_id": taskId,
	})
	var hasPermListFlag = true//标识是否进行权限分条确认的结果标识
	if err3 != nil || info == nil || len(*info)==0 {
		logs.Error("未查询到该任务的权限确认信息")
		hasPermListFlag = false
	}

	detailMap := make(map[int][]dal.DetectContentDetail)
	permsMap := make(map[int]dal.PermAppRelation)
	var midResult []dal.DetectQueryStruct
	var firstResult dal.DetectQueryStruct//主要包检测结果
	for num,content := range (*contents) {
		var queryResult dal.DetectQueryStruct
		queryResult.Channel = content.Channel
		queryResult.ApkName = content.ApkName
		queryResult.Version = content.Version

		permission := ""
		//增量的时候，此处一般为""
		perms := strings.Split(content.Permissions,";")
		for _,perm := range perms[0:(len(perms)-1)]{
			permission += perm +"\n"
		}
		queryResult.Permissions = permission
		queryResult.Index = num+1
		detailListIndex := make([]dal.DetectContentDetail,0)
		for i := 0;i<len(*details);i++ {
			if (*details)[i].SubIndex == num {
				detailListIndex = append(detailListIndex,(*details)[i])
			}
		}
		detailMap[num+1] = detailListIndex

		if hasPermListFlag {
			var permInfo dal.PermAppRelation
			for i:= 0;i<len(*info);i++{
				if (*info)[i].SubIndex == num {
					permInfo = (*info)[i]
					break
				}
			}
			permsMap[num+1] = permInfo
		}
		if queryResult.ApkName == (*detect)[0].AppName {
			firstResult = queryResult
		}else {
			midResult =append(midResult,queryResult)
		}
	}
	finalResult := make([]dal.DetectQueryStruct,0)
	finalResult = append(finalResult,firstResult)
	finalResult = append(finalResult,midResult...)

	//任务检测结果组输出重组
	for i:= 0;i<len(finalResult);i++ {
		details := detailMap[finalResult[i].Index]
		permissions := make([]dal.Permissions,0)

		//获取敏感信息输出结果
		methods_un,strs_un := GetDetectDetailOutInfo(details,c,methodIgs,strIgs)
		if methods_un == nil && strs_un == nil {
			return
		}
		finalResult[i].SMethods = methods_un
		finalResult[i].SStrs = make([]dal.SStr,0)
		finalResult[i].SStrs_new = strs_un

		//权限结果重组
		if hasPermListFlag{
			thePerm := permsMap[finalResult[i].Index]
			permissionsP,errP := GetTaskPermissions_2(thePerm,perIgs,allPermList)
			if errP != nil || permissionsP == nil || len(*permissionsP) == 0 {
				finalResult[i].Permissions_2 = permissions
			}else {
				finalResult[i].Permissions_2 = (*permissionsP)
			}
		}else {
			finalResult[i].Permissions_2 = permissions
		}

	}

	logs.Info("query detect result success!")
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : finalResult,
	})
	return
}

/**
	敏感方法和字符串的结果输出解析---新版
 */
func GetDetectDetailOutInfo(details []dal.DetectContentDetail, c *gin.Context,methodIgs map[string]interface{},strIgs map[string]interface{})([]dal.SMethod,[]dal.SStr){
	methods_un := make([]dal.SMethod,0)
	methods_con := make([]dal.SMethod,0)
	strs_un := make([]dal.SStr,0)
	strs_con := make([]dal.SStr,0)

	//敏感方法和字符串增量形式检测结果重组
	for _,detail := range details {
		if detail.SensiType == 1 {//敏感方法
			var method dal.SMethod
			method.Status = detail.Status
			method.Confirmer = detail.Confirmer
			method.Remark = detail.Remark
			method.ClassName = detail.ClassName
			method.Desc = detail.DescInfo
			method.Status = detail.Status
			method.Id = detail.ID
			method.MethodName = detail.KeyInfo
			if detail.ExtraInfo != "" {
				var t dal.DetailExtraInfo
				json.Unmarshal([]byte(detail.ExtraInfo),&t)
				method.GPFlag = int(t.GPFlag)
			}
			if methodIgs != nil {
				if v,ok := methodIgs[detail.ClassName+"."+detail.KeyInfo]; ok{
					info := v.(map[string]interface{})
					if info["status"] != 0 && method.Status != 0{
						method.Status = info["status"].(int)
						method.Confirmer = info["confirmer"].(string)
						method.Remark = info["remarks"].(string)
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
					errorReturn(c,"callLoc数据不符合要求")
					return nil,nil
				}
				callLoc = append(callLoc,call_loc_json)
			}
			method.CallLoc = callLoc
			if method.Status ==0 {
				methods_un = append(methods_un,method)
			}else {
				methods_con = append(methods_con,method)
			}
		}else if detail.SensiType == 2{//敏感字符串
			keys2 := make(map[string]int)
			//var keys3 = detail.KeyInfo
			var str dal.SStr
			str.Status = detail.Status
			confirmInfos := make([]dal.ConfirmInfo,0)
				keys := strings.Split(detail.KeyInfo,";")
				keys3 := ""
				for _,keyInfo := range keys[0:len(keys)-1] {
					var confirmInfo dal.ConfirmInfo
					if v,ok := strIgs[keyInfo]; ok{
						info := v.(map[string]interface{})
						if info["status"] != 0 {
							keys2[keyInfo] = 1
							confirmInfo.Key = keyInfo
							confirmInfo.Status = info["status"].(int)
							confirmInfo.Remark = info["remarks"].(string)
							confirmInfo.Status = info["status"].(int)
							confirmInfo.Confirmer = info["confirmer"].(string)
							confirmInfo.OtherVersion = info["version"].(string)
							confirmInfos = append(confirmInfos,confirmInfo)
						}
					} else{
						keys3 += keyInfo+";"
						confirmInfo.Key = keyInfo
						confirmInfos = append(confirmInfos,confirmInfo)
					}
				}
				if keys3 =="" {
					str.Status = 1
				}
			str.Keys = detail.KeyInfo
			str.Remark = detail.Remark
			str.Confirmer = detail.Confirmer
			str.Desc = detail.DescInfo
			str.Id = detail.ID
			if detail.ExtraInfo != "" {
				var t dal.DetailExtraInfo
				json.Unmarshal([]byte(detail.ExtraInfo),&t)
				str.GPFlag = int(t.GPFlag)
			}
			callLocs := strings.Split(detail.CallLoc,";")
			callLoc := make([]dal.StrCallJson,0)
			for _,call_loc := range callLocs[0:(len(callLocs)-1)] {
				var callLoc_json dal.StrCallJson
				err := json.Unmarshal([]byte(call_loc),&callLoc_json)
				if err != nil {
					logs.Error("callLoc数据不符合要求，%v========%s",err,call_loc)
					errorReturn(c,"callLoc数据不符合要求")
					return nil,nil
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
		}
	}
	//保证结果未确认结果在前
	for _,m := range methods_con {
		methods_un = append(methods_un,m)
	}
	for _,str := range strs_con {
		strs_un = append(strs_un,str)
	}
	return methods_un,strs_un
}


/**
	权限结果输出解析
 */
func GetTaskPermissions_2(info dal.PermAppRelation,perIgs map[int]interface{},allPermList map[int]interface{}) (*PermSlice,error) {
	bytePerms := []byte(info.PermInfos)

	var infos []interface{}
	if err := json.Unmarshal(bytePerms,&infos); err != nil {
		logs.Error("该任务的权限信息存储格式出错")
		return nil,err
	}

	var result PermSlice
	var reulst_con PermSlice
	for v,permInfo := range infos {
		var permOut dal.Permissions
		permMap := permInfo.(map[string]interface{})
		//更新权限信息
		permMap["priority"] = int(permMap["priority"].(float64))
		if v,ok := allPermList[int(permMap["perm_id"].(float64))];ok {
			info := v.(map[string]interface{})
			permMap["priority"] = info["priority"].(int)
			permMap["ability"] = info["ability"].(string)
		}
		permOut.Id = uint(v)+1
		permOut.Priority = permMap["priority"].(int)
		permOut.Status = int(permMap["status"].(float64))
		permOut.Key = permMap["key"].(string)
		permOut.PermId = int(permMap["perm_id"].(float64))
		permOut.Desc = permMap["ability"].(string)
		permOut.OtherVersion = permMap["first_version"].(string)

		if v,ok := perIgs[int(permMap["perm_id"].(float64))]; ok {
			perm := v.(map[string]interface{})
			permOut.Status = perm["status"].(int)
			permOut.Remark = perm["remarks"].(string)
			permOut.Confirmer = perm["confirmer"].(string)
		}

		if permOut.Status == 0 {
			result = append(result,permOut)
		}else {
			reulst_con = append(reulst_con,permOut)
		}
	}
	sort.Sort(PermSlice(result))
	sort.Sort(PermSlice(reulst_con))
	for _,outInfo := range reulst_con {
		result = append(result,outInfo)
	}
	return &result,nil
}


/**
	安卓确认检测结果----兼容.aab结果
 */
func ConfirmApkBinaryResultv_5(c *gin.Context){
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.PostConfirm
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("参数不合法 ，%v",err)
		errorReturn(c,"参数不合法")
		return
	}
	//获取确认人信息
	username, _ := c.Get("username")
	usernameStr := username.(string)

	//获取任务信息
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : t.TaskId,
	})

	//获取该任务的权限信息
	perms,errPerm := dal.QueryPermAppRelation(map[string]interface{}{
		"task_id":t.TaskId,
	})

	if t.Type == 0 {//敏感方法和字符串确认
		condition1 := "id="+strconv.Itoa(t.Id)
		detailInfo, err := dal.QueryDetectContentDetail(condition1)
		if err != nil || detailInfo == nil||len(*detailInfo)==0{
			logs.Error("不存在该检测结果，ID：%d",t.Id)
			errorReturn(c,"不存在该检测结果")
			return
		}
		if detect == nil || len(*detect)== 0{
			logs.Error("未查询到该taskid对应的检测任务，%v", t.TaskId)
			errorReturn(c,"未查询到该taskid对应的检测任务")
			return
		}

		var data map[string]string
		data = make(map[string]string)
		data["id"] = strconv.Itoa(t.Id)
		data["confirmer"] = usernameStr
		data["remark"] = t.Remark
		data["status"] = strconv.Itoa(t.Status)
		flag := dal.ConfirmApkBinaryResultNew(data)
		if !flag {
			logs.Error("二进制检测内容确认失败")
			errorReturn(c,"二进制检测内容确认失败")
			return
		}

		//增量忽略结果录入
		if t.Status != 0 {
			senType := (*detailInfo)[0].SensiType
			if senType == 1 {
				keyInfo := (*detailInfo)[0].ClassName+"."+(*detailInfo)[0].KeyInfo
				//结果写入
				if err := createIgnoreInfo(c,&t,&(*detect)[0],usernameStr,keyInfo,1); err != nil {
					return
				}
			}else {
				keys := strings.Split((*detailInfo)[0].KeyInfo,";")
				for _,key := range keys[0:len(keys)-1] {
					//结果写入
					if err := createIgnoreInfo(c,&t,&(*detect)[0],usernameStr,key,2);err != nil {
						return
					}
				}
			}
		}
	}else {
		if errPerm != nil || perms == nil || len(*perms) == 0 {
			logs.Error("未查询到该任务的检测信息")
			errorReturn(c,"未查询到该任务的检测信息")
			return
		}
		if t.Index >len(*perms){
			logs.Error("权限结果数组下标越界")
			errorReturn(c,"权限结果数组下标越界")
			return
		}
		permsInfoDB := (*perms)[t.Index-1].PermInfos
		var permList []interface{}
		if err := json.Unmarshal([]byte(permsInfoDB),&permList); err != nil {
			logs.Error("该任务的权限存储信息格式出错")
			errorReturn(c,"该任务的权限存储信息格式出错")
			return
		}
		//增加数组越界维护
		if t.Id >len(permList){
			logs.Error("权限ID越界")
			errorReturn(c,"权限ID越界")
			return
		}
		permMap := permList[t.Id-1].(map[string]interface{})
		permMap["status"] = t.Status

		permId := int(permMap["perm_id"].(float64))
		newPerms,_ := json.Marshal(permList)
		(*perms)[t.Index-1].PermInfos = string(newPerms)
		if err := dal.UpdataPermAppRelation(&(*perms)[t.Index-1]); err != nil {
			logs.Error("更新任务权限确认情况失败")
			errorReturn(c,"更新任务权限确认情况失败")
			return
		}
		//写入操作历史
		var history dal.PermHistory
		history.Status = t.Status
		history.AppVersion = (*perms)[t.Index-1].AppVersion
		history.AppId = (*perms)[t.Index-1].AppId
		history.PermId = permId
		history.Remarks = t.Remark
		history.Confirmer = usernameStr
		history.TaskId = t.TaskId
		if err := dal.InsertPermOperationHistory(history); err != nil {
			logs.Error("权限操作历史写入失败！")
			errorReturn(c,"权限操作历史写入失败！")
			return
		}
	}

	//任务状态更新
	updateInfo := taskStatusUpdate(t.TaskId,t.ToolId,&(*detect)[0])
	if updateInfo != "" {
		errorReturn(c,updateInfo)
	}
	logs.Info("confirm success +id :%d",t.Id)
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
	return
}

/**
	任务确认状态更新
 */
func taskStatusUpdate(taskId int, toolId int, detect *dal.DetectStruct) string {
	condition := "deleted_at IS NULL and task_id='" + strconv.Itoa(taskId) + "' and tool_id='" + strconv.Itoa(toolId) + "' and status= 0"
	counts := dal.QueryUnConfirmDetectContent(condition)

	perms,_ := dal.QueryPermAppRelation(map[string]interface{}{
		"task_id":taskId,
	})
	var permFlag = true
	if perms == nil || len(*perms)== 0 {
		logs.Error("该任务无权限检测信息！")
	}else{
	P:
		for i:=0;i<len(*perms);i++{
			permsInfoDB := (*perms)[i].PermInfos
			var permList []interface{}
			if err := json.Unmarshal([]byte(permsInfoDB),&permList);err!=nil {
				return "该任务的权限存储信息格式出错"
			}
			for _,m := range permList {
				permInfo := m.(map[string]interface{})
				if permInfo["status"].(float64) == 0 {
					permFlag = false
					break P
				}
			}
		}
	}

	logs.Notice("当前确认情况，字符串和方法剩余："+fmt.Sprint(counts)+" ,权限是否全部确认："+fmt.Sprint(permFlag))
	if counts == 0 && permFlag {
		detect.Status = 1
		err := dal.UpdateDetectModelNew(*detect)
		if err != nil {
			logs.Error("任务确认状态更新失败！%v", err)
			utils.LarkDingOneInner("fanjuan.xqp","任务ID："+fmt.Sprint(taskId)+"确认状态更新失败，失败原因："+err.Error())
			return "任务确认状态更新失败！"
		}
		CICallBack(&(*detect)[0])
	}
	return ""

}

/**
	新增敏感字符or敏感方法确认历史
 */
func createIgnoreInfo(c *gin.Context,t *dal.PostConfirm,detect *dal.DetectStruct,usernameStr string,key string,senType int) error  {
	var igInfo dal.IgnoreInfoStruct
	igInfo.Platform = detect.Platform
	igInfo.AppId,_ = strconv.Atoi(detect.AppId)
	igInfo.SensiType = senType
	igInfo.KeysInfo = key
	igInfo.Confirmer = usernameStr
	igInfo.Remarks = t.Remark
	igInfo.Version = detect.AppVersion
	igInfo.Status = t.Status
	igInfo.TaskId = t.TaskId
	err := dal.InsertIgnoredInfo(igInfo)
	if err != nil {
		errorReturn(c,"增量信息更新失败！")
		return err
	}
	return nil
}

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