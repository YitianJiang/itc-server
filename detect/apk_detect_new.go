package detect

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"code.byted.org/clientQA/itc-server/database"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
)

// The error type of detect service.
const (
	DetectServiceScriptError         = 1
	DetectServiceInfrastructureError = 2

	informer = "hejiahui.2019"
)

// ParseResultAndroid supports apk and aab format.
func ParseResultAndroid(task *dal.DetectStruct, resultJson *string, toolID int) (error, int) {

	msgHeader := fmt.Sprintf("task id: %v", task.ID)

	var result dal.JSONResultStruct
	if err := json.Unmarshal([]byte(*resultJson), &result); err != nil {
		logs.Error("%s unmarshal error: %v", msgHeader, err)
		handleDetectTaskError(task, DetectServiceScriptError, *resultJson)
		return err, 0
	}

	// Handle the basic information and permission of apk and aab package.
	for i := range result.Result {
		if err := AppInfoAnalysis_2(task, result.Result[i].AppInfo, toolID, i); err != nil {
			logs.Error("%s analysis app information failed: %v", msgHeader, err)
			return err, 0
		}
	}

	//遍历结果数组，并将每组检测结果信息插入数据库
	for index, result := range result.Result {

		//获取敏感方法和字符串的确认信息methodInfo,strInfos，为信息初始化做准备
		methodInfo, strInfos, _, err := getIgnoredInfo_2(task.AppId, task.Platform)
		if err != nil {
			logs.Warn("%s Failed to retrieve negligible information about APP ID: %v, Platform: %v", msgHeader, task.AppId, task.Platform)
		}
		if methodInfo == nil {
			logs.Warn("%s There are no negligible methods", msgHeader)
		}
		if strInfos == nil {
			logs.Warn("%s There are no negligible strings", msgHeader)
		}

		//敏感method解析----先外层去重
		mRepeat := make(map[string]int)
		newMethods := make([]dal.MethodInfo, 0) //第一层去重后的敏感方法集
		for _, method := range result.MethodInfos {
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
			detailContent.TaskId = int(task.ID)
			detailContent.ToolId = toolID
			//新增兼容下标
			detailContent.SubIndex = index
			allMethods = append(allMethods, *MethodAnalysis(newMethod, &detailContent, extraInfo)) //内层去重，并放入写库信息数组
		}
		if err := dal.InsertDetectDetailBatch(&allMethods); err != nil {
			logs.Error("%s insert detect detail failed: %v", msgHeader, err)
			return err, 0
		}

		//敏感方法解析
		allStrs := make([]dal.DetectContentDetail, 0)
		for _, strInfo := range result.StrInfos {
			var detailContent dal.DetectContentDetail
			detailContent.TaskId = int(task.ID)
			detailContent.ToolId = toolID
			detailContent.SubIndex = index
			allStrs = append(allStrs, *StrAnalysis(strInfo, &detailContent, strInfos))
		}
		if err := dal.InsertDetectDetailBatch(&allStrs); err != nil {
			logs.Error("%s insert detect detail failed: %v", msgHeader, err)
			return err, 0
		}
	}

	//任务状态更新----该app无需要特别确认的敏感方法、字符串或权限
	errTaskUpdate, unConfirms := taskStatusUpdate(task.ID, toolID, task, false, 0)
	if errTaskUpdate != "" {
		logs.Error("%s update task status failed", msgHeader)
		return fmt.Errorf(errTaskUpdate), 0
	}
	return nil, unConfirms
}

/**
appInfo解析，并写入数据库,此处包含权限的处理-------fj
新增了index下标，兼容.aab结果中新增sub_index，默认为0
*/
func AppInfoAnalysis_2(task *dal.DetectStruct, info dal.AppInfoStruct, toolID int, index int) error {

	msgHeader := fmt.Sprintf("task id: %v", task.ID)

	//判断appInfo信息是否为主要信息，只有主要信息--primary为1才会修改任务的appName和Version,或者primary为nil---只有一个信息
	if info.Primary == nil || fmt.Sprint(info.Primary) == "1" {
		task.AppName = info.ApkName
		task.AppVersion = info.ApkVersionName
		task.InnerVersion = info.Meta.InnerVersion
		if err := database.UpdateDBRecord(database.DB(), task); err != nil {
			logs.Error("%s update detect task failed: %v", msgHeader, err)
			return err
		}
	}
	//更新任务的权限信息
	permAppInfos, err := permUpdate(task, info.PermsInAppInfo)
	if err != nil {
		logs.Error("%s update permission failed: %v", msgHeader, err)
		return err
	}
	appID, err := strconv.Atoi(task.AppId)
	if err != nil {
		logs.Error("%s atoi error: %v", msgHeader, err)
		return err
	}
	if err := database.InsertDBRecord(database.DB(), &dal.PermAppRelation{
		TaskId:     int(task.ID),
		AppId:      appID,
		AppVersion: info.ApkVersionName,
		SubIndex:   index,
		PermInfos:  permAppInfos,
	}); err != nil {
		logs.Error("%s insert tb_perm_apprelation failed: %v", msgHeader, err)
		return err
	}

	if err := database.InsertDBRecord(database.DB(), &dal.DetectInfo{
		TaskId:   int(task.ID),
		ApkName:  info.ApkName,
		Version:  info.ApkVersionName,
		Channel:  info.Channel,
		ToolId:   toolID,
		SubIndex: index,
	}); err != nil {
		logs.Error("%s insert tb_detect_info_apk failed: %v", msgHeader, err)
		return err
	}

	return nil
}

/**
处理权限信息，包括（初次引入写入配置表，历史表，lark通知）
*/
func permUpdate(task *dal.DetectStruct, permissions []string) (string, error) {

	msgHeader := fmt.Sprintf("task id: %v", task.ID)

	var larkPerms string //lark消息通知的权限内容
	var first_history []dal.PermHistory
	//获取app的权限操作历史map
	impMap := getAPPPermissionHistory(task.AppId)
	//判断是否属于初次引入
	var fhflag bool
	//权限去重map
	permRepeatMap := make(map[string]int)
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
	for _, pers := range permissions {
		//权限去重
		if v, ok := permRepeatMap[pers]; ok && v == 1 {
			// Continue the loop if the permission has been handled.
			continue
		}
		permRepeatMap[pers] = 1

		//写app和perm对应关系
		queryResult := dal.QueryDetectConfig(map[string]interface{}{
			"key_info":   pers,
			"platform":   Android,
			"check_type": Permission})
		fhflag = false
		permInfo := make(map[string]interface{})
		if queryResult == nil || len(*queryResult) == 0 {
			// Cannot find any matched pemisstion in the safe
			// permission list which means this is a new permission.
			var conf dal.DetectConfigStruct
			conf.KeyInfo = pers
			conf.Priority = RiskLevelHigh
			conf.Creator = "itc"
			conf.Platform = Android
			if _, ok := _const.DetectBlackList[task.Creator]; !ok {
				if err := dal.InsertDetectConfig(&conf); err != nil {
					logs.Error("%s insert detect config failed: %v", msgHeader, err)
					return "", err
				}
				permInfo["perm_id"] = int(conf.ID)
			} else {
				logs.Notice("task id: %v creator: %v DO NOT INSERT THE NEW DETECTION", task.ID, task.Creator)
				// The permission will not be inserted into the official ITC configures.
				permInfo["perm_id"] = -1
			}
			fhflag = true
			larkPerms += "权限名为：" + pers + "\n"
			permInfo["key"] = pers
			permInfo["ability"] = ""
			//优先级默认为3---高危
			permInfo["priority"] = 3
			//此处state表明该权限是自动添加，信息不全，后面query时需要重新读取相关信息
			permInfo["state"] = 0
			permInfo["status"] = 0
			permInfo["first_version"] = task.AppVersion
		} else {
			permInfo["perm_id"] = int((*queryResult)[0].ID)
			permInfo["key"] = pers
			permInfo["ability"] = (*queryResult)[0].Ability
			permInfo["priority"] = (*queryResult)[0].Priority
			permInfo["state"] = 1

			//更新确认信息
			if v, ok := impMap[int((*queryResult)[0].ID)]; !ok {
				permInfo["status"] = 0
				permInfo["first_version"] = task.AppVersion
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
			appId, _ := strconv.Atoi(task.AppId)
			hist.AppId = appId
			hist.AppVersion = task.AppVersion
			hist.PermId = permInfo["perm_id"].(int)
			hist.Confirmer = task.Creator
			hist.Remarks = "包检测引入该权限"
			hist.TaskId = int(task.ID)
			first_history = append(first_history, hist)
		}
		permInfos = append(permInfos, permInfo)
	}
	bytePerms, _ := json.Marshal(permInfos)
	//若存在初次引入权限，批量写入引入信息
	if len(first_history) > 0 {
		if err := BatchInsertPermHistory(&first_history); err != nil {
			logs.Error("%s insert permission history failed: %v", msgHeader, err)
			return "", err
		}
	}
	//lark通知创建人完善权限信息-----只发一条消息
	if larkPerms != "" {
		message := "你好，安卓二进制静态包检测出未知权限，请去权限配置页面完善权限信息,需要完善的权限信息有：\n"
		message += larkPerms
		message += "修改链接：http://rocket.bytedance.net/rocket/itc/permission"
		for i := range _const.PermLarkPeople {
			utils.LarkDingOneInner(_const.PermLarkPeople[i], message)
		}
	}

	return string(bytePerms), nil
}

/**
批量插入群仙操作历史
*/
func BatchInsertPermHistory(infos *[]dal.PermHistory) error {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("connect to db failed,%v", err)
		return err
	}
	defer connection.Close()
	db := connection.Table(dal.PermHistory{}.TableName()).LogMode(_const.DB_LOG_MODE)
	db.Begin()
	for _, info := range *infos {
		info.CreatedAt = time.Now()
		info.UpdatedAt = time.Now()
		if err := db.Create(&info).Error; err != nil {
			logs.Error("insert perm history failed,%v", err)
			db.Rollback()
			return err
		}
	}
	db.Commit()
	return nil
}

/**
获取权限引入历史
*/
func getAPPPermissionHistory(appID interface{}) map[int]interface{} {

	history, err := dal.QueryPermHistory(map[string]interface{}{"app_id": appID})
	if err != nil || history == nil || len(*history) == 0 {
		logs.Error("Cannot find any permission about app id: %v", appID)
		return nil
	}

	result := make(map[int]interface{})
	for _, infoP := range *history {
		_, ok := result[infoP.PermId]
		if !ok {
			result[infoP.PermId] = map[string]interface{}{
				"version": infoP.AppVersion,
				"status":  infoP.Status}
		} else if ok && infoP.Status == 0 {
			// TODO
			v := result[infoP.PermId].(map[string]interface{})
			v["version"] = infoP.AppVersion
			result[infoP.PermId] = v
		}
	}

	return result
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
func handleDetectTaskError(detect *dal.DetectStruct,
	errCode interface{}, errInfo interface{}) error {

	errBytes, err := json.Marshal(map[string]interface{}{
		"errCode": errCode,
		"errInfo": errInfo})
	if err != nil {
		logs.Error("Marshal error: %v", err)
		return err
	}

	var errString = string(errBytes)
	detect.ErrInfo = &errString
	return dal.UpdateDetectModelNew(*detect)
}
