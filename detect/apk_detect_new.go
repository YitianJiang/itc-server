package detect

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"fmt"
	"strconv"
)

/**
安卓json检测信息分析----兼容.aab格式检测结果---json到Struct
*/
func ApkJsonAnalysis_2(info string, mapInfo map[string]int) (error, int) {
	logs.Info("新的安卓解析开始～～～～")
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id": mapInfo["taskId"],
	})
	var fisrtResult dal.JSONResultStruct
	err_f := json.Unmarshal([]byte(info), &fisrtResult)
	if err_f != nil {
		logs.Error("taskId:"+fmt.Sprint(mapInfo["taskId"])+",二进制静态包检测返回信息格式错误！,%v", err_f)
		message := "taskId:" + fmt.Sprint(mapInfo["taskId"]) + ",二进制静态包检测返回信息格式错误，请解决;" + fmt.Sprint(err_f)
		DetectTaskErrorHandle((*detect)[0], "1", info)
		utils.LarkDingOneInner("fanjuan.xqp", message)
		return err_f, 0
	}

	//遍历结果数组，并将每组检测结果信息插入数据库
	for index, result := range fisrtResult.Result {
		appInfos := result.AppInfo
		methodsInfo := result.MethodInfos
		strsInfo := result.StrInfos

		var detectInfo dal.DetectInfo
		detectInfo.TaskId = mapInfo["taskId"]
		detectInfo.ToolId = mapInfo["toolId"]
		//检测基础信息解析
		errInfo := AppInfoAnalysis_2(appInfos, &detectInfo, index)
		if errInfo != nil {
			return errInfo, 0
		}

		//获取敏感方法和字符串的确认信息methodInfo,strInfos，为信息初始化做准备
		data := make(map[string]string)
		data["appId"] = (*detect)[0].AppId
		data["platform"] = strconv.Itoa((*detect)[0].Platform)
		methodInfo, strInfos, _, err := getIgnoredInfo_2(data)
		if err != nil {
			logs.Error("未查询到该App的增量信息，app信息为：%v", data)
		}

		//敏感method解析----先外层去重
		mRepeat := make(map[string]int)
		newMethods := make([]dal.MethodInfo, 0) //第一层去重后的敏感方法集
		for _, method := range methodsInfo {
			var keystr = method.MethodName + method.ClassName
			if v, ok := mRepeat[keystr]; !ok || ok && v == 0 {
				newMethods = append(newMethods, method)
				mRepeat[keystr] = 1
			}
		}
		apiMap := GetAllAPIConfigs()
		allMethods := make([]dal.DetectContentDetail, 0) //批量写入数据库的敏感方法struct集合
		for _, newMethod := range newMethods {
			var detailContent dal.DetectContentDetail
			var keystr = newMethod.ClassName + "." + newMethod.MethodName
			if v, ok := methodInfo[keystr]; ok {
				info := v.(map[string]interface{})
				if info["status"].(int) == 1 {
					detailContent.Status = info["status"].(int)
				}
			}
			var extraInfo dal.DetailExtraInfo
			if v, ok := (*apiMap)[keystr]; ok {
				info := v.(map[string]int)
				detailContent.RiskLevel = fmt.Sprint(info["priority"])
				extraInfo.ConfigId = info["id"]
			} else {
				detailContent.RiskLevel = "3"
			}
			detailContent.TaskId = mapInfo["taskId"]
			detailContent.ToolId = mapInfo["toolId"]
			//新增兼容下标
			detailContent.SubIndex = index
			allMethods = append(allMethods, *MethodAnalysis(newMethod, &detailContent, extraInfo)) //内层去重，并放入写库信息数组
		}
		err1 := dal.InsertDetectDetailBatch(&allMethods)
		if err1 != nil {
			//及时报警
			message := "taskId:" + fmt.Sprint(mapInfo["taskId"]) + ",敏感method写入数据库失败，请解决;" + fmt.Sprint(err)
			utils.LarkDingOneInner("fanjuan.xqp", message)
			return err_f, 0
		}

		//敏感方法解析
		allStrs := make([]dal.DetectContentDetail, 0)
		for _, strInfo := range strsInfo {
			var detailContent dal.DetectContentDetail
			detailContent.TaskId = mapInfo["taskId"]
			detailContent.ToolId = mapInfo["toolId"]
			detailContent.SubIndex = index
			allStrs = append(allStrs, *StrAnalysis(strInfo, &detailContent, strInfos))
		}
		err2 := dal.InsertDetectDetailBatch(&allStrs)
		if err2 != nil {
			//及时报警
			message := "taskId:" + fmt.Sprint(mapInfo["taskId"]) + ",敏感str写入数据库失败，请解决;" + fmt.Sprint(err)
			utils.LarkDingOneInner("fanjuan.xqp", message)
			return err_f, 0
		}
	}

	//任务状态更新----该app无需要特别确认的敏感方法、字符串或权限
	errTaskUpdate, unConfirms := taskStatusUpdate(mapInfo["taskId"], mapInfo["toolId"], &(*detect)[0], false, 0)
	if errTaskUpdate != "" {
		return fmt.Errorf(errTaskUpdate), 0
	}
	return nil, unConfirms
}

/**
appInfo解析，并写入数据库,此处包含权限的处理-------fj
新增了index下标，兼容.aab结果中新增sub_index，默认为0
*/
func AppInfoAnalysis_2(info dal.AppInfoStruct, detectInfo *dal.DetectInfo, index ...int) error {
	//数组结果排序标识处理，默认为0
	var realIndex int
	if len(index) == 0 {
		realIndex = 0
	} else {
		realIndex = index[0]
	}

	taskId := detectInfo.TaskId
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id": taskId,
	})
	appId, _ := strconv.Atoi((*detect)[0].AppId)

	//判断appInfo信息是否为主要信息，只有主要信息--primary为1才会修改任务的appName和Version,或者primary为nil---只有一个信息
	var taskUpdateFlag = false
	if info.Primary == nil || info.Primary.(float64) == 1 {
		taskUpdateFlag = true
	}
	detectInfo.ApkName = info.ApkName
	detectInfo.Version = info.ApkVersionName
	detectInfo.Channel = info.Channel

	if taskUpdateFlag {
		(*detect)[0].AppName = info.ApkName
		(*detect)[0].AppVersion = info.ApkVersionName
		if err := dal.UpdateDetectModelNew((*detect)[0]); err != nil {
			message := "任务ID：" + fmt.Sprint(taskId) + "，appName和Version信息更新失败，失败原因：" + fmt.Sprint(err)
			logs.Error(message)
			utils.LarkDingOneInner("fanjuan.xqp", message)
			return err
		}
	}

	//更新任务的权限信息
	var permissionArr = info.PermsInAppInfo
	permAppInfos, errP := permUpdate(&permissionArr, detectInfo, detect)
	if errP != nil {
		return errP
	}

	//更新权限-app-task关系表
	var relationship dal.PermAppRelation
	relationship.TaskId = taskId
	relationship.AppId = appId
	if taskUpdateFlag {
		relationship.AppVersion = (*detect)[0].AppVersion
	} else {
		relationship.AppVersion = ".aab副包+" + detectInfo.Version
	}
	relationship.AppVersion = detectInfo.Version
	relationship.SubIndex = realIndex //新增下标兼容.aab结果
	relationship.PermInfos = permAppInfos
	err1 := dal.InsertPermAppRelation(relationship)
	if err1 != nil {
		utils.LarkDingOneInner("fanjuan.xqp", "新增权限App关系失败！taskId:"+fmt.Sprint(taskId)+",appID="+(*detect)[0].AppId)
		return err1
	}

	//插入appInfo信息到apk表
	perStr := "" //旧版权限信息
	detectInfo.Permissions = perStr
	detectInfo.SubIndex = realIndex
	err := dal.InsertDetectInfo(*detectInfo)
	if err != nil {
		//及时报警
		message := "taskId:" + fmt.Sprint(taskId) + ",appInfo写入数据库失败，请解决;" + fmt.Sprint(err)
		utils.LarkDingOneInner("fanjuan.xqp", message)
		return err
	}
	return nil
}

/**
处理权限信息，包括（初次引入写入配置表，历史表，lark通知）
*/
func permUpdate(permissionArr *[]string, detectInfo *dal.DetectInfo, detect *[]dal.DetectStruct) (string, error) {
	appId, _ := strconv.Atoi((*detect)[0].AppId)
	taskId := detectInfo.TaskId
	//更新任务的权限信息
	permInfos := make([]map[string]interface{}, 0) //新版权限信息，结构如下
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

	larkPerms := "" //lark消息通知的权限内容
	var first_history []dal.PermHistory
	//获取app的权限操作历史map
	impMap := GetImportedPermission(appId)
	//判断是否属于初次引入
	var fhflag bool
	//权限去重map
	permRepeatMap := make(map[string]int)
	for _, pers := range *permissionArr {
		//权限去重
		if v, okp := permRepeatMap[pers]; okp && v == 1 {
			continue
		}
		permRepeatMap[pers] = 1
		//写app和perm对应关系
		queryResult := dal.QueryDetectConfig(map[string]interface{}{
			"key_info":   pers,
			"platform":   0,
			"check_type": 0,
		})
		fhflag = false
		permInfo := make(map[string]interface{})
		if queryResult == nil || len(*queryResult) == 0 {
			var conf dal.DetectConfigStruct
			conf.KeyInfo = pers
			//将该权限的优先级定为--3高危
			conf.Priority = 3
			//暂时定为固定---标识itc检测新增
			conf.Creator = "itc"
			conf.Platform = 0
			perm_id, err := dal.InsertDetectConfig(conf)

			if err != nil {
				logs.Error("taskId:"+fmt.Sprint(taskId)+",update回调时新增权限失败，%v", err)
				//及时报警
				utils.LarkDingOneInner("fanjuan.xqp", "taskId:"+fmt.Sprint(taskId)+",update回调新增权限失败")
				return "", err
			} else {
				fhflag = true
				larkPerms += "权限名为：" + pers + "\n"
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
		} else {
			permInfo["perm_id"] = (*queryResult)[0].ID
			permInfo["key"] = pers
			permInfo["ability"] = (*queryResult)[0].Ability
			permInfo["priority"] = (*queryResult)[0].Priority
			permInfo["state"] = 1

			//更新确认信息
			if v, ok := impMap[int((*queryResult)[0].ID)]; !ok {
				//logs.Error("未查询到该权限的操作历史")
				permInfo["status"] = 0
				permInfo["first_version"] = detectInfo.Version
				fhflag = true
			} else {
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
			hist.TaskId = taskId
			first_history = append(first_history, hist)
		}
		permInfos = append(permInfos, permInfo)
	}
	bytePerms, _ := json.Marshal(permInfos)
	//若存在初次引入权限，批量写入引入信息
	if len(first_history) > 0 {
		errB := dal.BatchInsertPermHistory(&first_history)
		if errB != nil {
			logs.Error("taskId:" + fmt.Sprint(taskId) + ",插入权限第一次引入历史失败")
			//及时报警
			utils.LarkDingOneInner("fanjuan.xqp", "taskId:"+fmt.Sprint(taskId)+",插入权限第一次引入历史失败")
			return "", errB
		}
	}
	//lark通知创建人完善权限信息-----只发一条消息
	if larkPerms != "" {
		message := "你好，安卓二进制静态包检测出未知权限，请去权限配置页面完善权限信息,需要完善的权限信息有：\n"
		message += larkPerms
		message += "修改链接：http://cloud.bytedance.net/rocket/itc/permission?biz=13"

		for _,people := range _const.PermLarkPeople {
			utils.LarkDingOneInner(people,message)
		}
	}

	return string(bytePerms), nil
}

/**
批量method解析-----fj
*/
func MethodAnalysis(method dal.MethodInfo, detail *dal.DetectContentDetail, extraInfo dal.DetailExtraInfo) *dal.DetectContentDetail {
	detail.SensiType = 1
	//detail.Status = 0

	detail.KeyInfo = method.MethodName
	detail.DescInfo = method.Desc
	detail.ClassName = method.ClassName
	//增加flag标识
	if method.Flag != 0 {
		extraInfo.GPFlag = method.Flag
	}
	byteExtra, _ := json.Marshal(extraInfo)
	detail.ExtraInfo = string(byteExtra)
	var call = method.CallLocation

	callLocation := MethodRmRepeat_2(call)
	detail.CallLoc = callLocation

	return detail
}

/**
apk敏感方法内容去重--------fj
*/
func MethodRmRepeat_2(callInfo []dal.CallLocInfo) string {
	repeatMap := make(map[string]int)
	result := ""
	for _, info := range callInfo {
		var keystr string
		keystr = info.ClassName + info.CallMethodName + fmt.Sprint(info.LineNum)
		if v, ok := repeatMap[keystr]; !ok || (ok && v == 0) {
			repeatMap[keystr] = 1
			mapInfo, _ := json.Marshal(info)
			result += string(mapInfo) + ";"
		}
	}
	return result
}

/**
批量str解析---------fj
*/
func StrAnalysis(str dal.StrInfo, detail *dal.DetectContentDetail, strInfos map[string]interface{}) *dal.DetectContentDetail {
	detail.SensiType = 2

	var keys = str.Keys
	key := ""
	//判断str是否进行状态转变
	key2 := ""
	//通过或不通过的状态表示，true为1，false为2
	var passFlag = true
	for _, ks := range keys {
		if v, ok := strInfos[ks]; !ok {
			key2 += ks
		} else {
			info := v.(map[string]interface{})
			if info["status"].(int) == 2 && passFlag {
				passFlag = false
			}
		}
		key += ks + ";"
	}
	detail.KeyInfo = key
	if key2 == "" {
		if passFlag {
			detail.Status = 1
		}
	} else {
		detail.Status = 0
	}
	detail.DescInfo = str.Desc
	//增加敏感字符串的gp标识
	if str.Flag != 0 {
		var extraInfo dal.DetailExtraInfo
		extraInfo.GPFlag = str.Flag
		byteExtra, _ := json.Marshal(extraInfo)
		detail.ExtraInfo = string(byteExtra)
	}
	var callInfo = str.CallLocation
	//敏感字段信息去重
	call_location := StrRmRepeat_2(callInfo)
	detail.CallLoc = call_location
	return detail

}

/**
apk敏感字符串去重--------fj
*/
func StrRmRepeat_2(callInfo []dal.CallLocInfo) string {
	repeatMap := make(map[string]int)
	result := ""
	for _, info := range callInfo {
		var keystr string
		keystr = info.ClassName + info.CallMethodName + info.Key + fmt.Sprint(info.LineNum)
		if v, ok := repeatMap[keystr]; !ok || (ok && v == 0) {
			repeatMap[keystr] = 1
			mapInfo, _ := json.Marshal(info)
			result += string(mapInfo) + ";"
		}
	}
	return result
}

func GetAllAPIConfigs() *map[string]interface{} {
	queryData := map[string]interface{}{
		"check_type": 1,
	}
	apiMap := make(map[string]interface{})
	result := dal.QueryDetectConfig(queryData)
	if result == nil {
		return &apiMap
	}
	for _, api := range *result {
		if _, ok := apiMap[api.KeyInfo]; !ok {
			info := map[string]int{
				"priority": api.Priority,
				"id":       int(api.ID),
			}
			apiMap[api.KeyInfo] = info
		}
	}
	return &apiMap

}

/**
检测任务发生问题逻辑处理
*/
func DetectTaskErrorHandle(detect dal.DetectStruct, errCode string, errInfo string) error {
	var errStruct dal.ErrorStruct
	errStruct.ErrCode = errCode
	errStruct.ErrInfo = errInfo
	errBytes, _ := json.Marshal(errStruct)
	var errString = string(errBytes)
	detect.ErrInfo = &errString
	err := dal.UpdateDetectModelNew(detect)
	if err != nil {
		return err
	} else {
		return nil
	}
}
