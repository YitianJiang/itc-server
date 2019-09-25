package detect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

//权限排序
type PermSlice []dal.Permissions

func (a PermSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a PermSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a PermSlice) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].Priority < a[i].Priority
}

//方法排序
type MethodSlice []dal.SMethod

func (a MethodSlice) Len() int { // 重写 Len() 方法
	return len(a)
}
func (a MethodSlice) Swap(i, j int) { // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a MethodSlice) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return a[j].RiskLevel < a[i].RiskLevel
}

/**
获取权限的确认历史信息------fj
*/

func GetIgnoredPermission(appId int) map[int]interface{} {
	result := make(map[int]interface{})
	queryResult, err := dal.QueryPermHistory(map[string]interface{}{
		"app_id": appId,
	})
	if err != nil || queryResult == nil || len(*queryResult) == 0 {
		logs.Error("该app暂时没有确认信息")
	} else {
		for _, infoP := range *queryResult {
			if _, ok := result[infoP.PermId]; !ok {
				if infoP.Status > 0 { //增加引入历史后，将此类信息过滤
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
func GetPermList() map[int]interface{} {
	result := make(map[int]interface{})
	queryResult := dal.QueryDetectConfig(map[string]interface{}{
		"platform":   0,
		"check_type": 0,
	})
	if queryResult == nil || len(*queryResult) == 0 {
		logs.Error("权限信息表为空")
	} else {
		for _, infoP := range *queryResult {
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
func GetImportedPermission(appId int) map[int]interface{} {
	result := make(map[int]interface{})
	queryResult, err := dal.QueryPermHistory(map[string]interface{}{
		"app_id": appId,
	})
	if err != nil || queryResult == nil || len(*queryResult) == 0 {
		logs.Error("该app暂时没有确认信息")
	} else {
		for _, infoP := range *queryResult {
			_, ok := result[infoP.PermId]
			if !ok {
				info := make(map[string]interface{})
				info["version"] = infoP.AppVersion
				info["status"] = infoP.Status
				result[infoP.PermId] = info
			} else if ok && infoP.Status == 0 {
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
func getIgnoredInfo_2(appID interface{}, platform interface{}) (map[string]interface{}, map[string]interface{}, map[string]interface{}, error) {

	result, err := QueryIgnoredInfo(map[string]interface{}{
		"app_id":   appID,
		"platform": platform})

	//此处如果条件1没有命中，但是23命中了，返回的err其实是nil
	if err != nil || result == nil || len(*result) == 0 {
		return nil, nil, nil, err
	}

	methodMap := make(map[string]interface{})
	strMap := make(map[string]interface{})
	perMap := make(map[string]interface{})
	for i := 0; i < len(*result); i++ {
		if (*result)[i].SensiType == 1 {
			if _, ok := methodMap[(*result)[i].KeysInfo]; !ok {
				info := make(map[string]interface{})
				info["status"] = (*result)[i].Status
				info["remarks"] = (*result)[i].Remarks
				info["confirmer"] = (*result)[i].Confirmer
				info["version"] = (*result)[i].Version
				methodMap[(*result)[i].KeysInfo] = info
			}
		} else if (*result)[i].SensiType == 2 {
			if _, ok := strMap[(*result)[i].KeysInfo]; !ok {
				info := make(map[string]interface{})
				info["status"] = (*result)[i].Status
				info["remarks"] = (*result)[i].Remarks
				info["confirmer"] = (*result)[i].Confirmer
				info["version"] = (*result)[i].Version
				//info["updateTime"] = (*result)[i].UpdatedAt
				strMap[(*result)[i].KeysInfo] = info
			}
		} else {

		}
	}
	return methodMap, strMap, perMap, nil
}

/**
获取敏感方法和字符串的确认历史
*/

func QueryIgnoredHistory_2(c *gin.Context) {
	type queryData struct {
		AppId    int    `json:"appId"`
		Platform int    `json:"platform"`
		Key      string `json:"key"`
	}
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t queryData
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("参数不合法 ，%v", err)
		c.JSON(http.StatusOK, gin.H{
			"message":   "参数不合法！",
			"errorCode": -1,
			"data":      "参数不合法！",
		})
		return
	}
	logs.Info(t.Key)
	queryDatas := make(map[string]interface{})
	queryDatas["condition"] = "app_id='" + strconv.Itoa(t.AppId) + "' and platform='" + strconv.Itoa(t.Platform) + "' and keys_info='" + t.Key + "'"
	result, err := QueryIgnoredInfo(queryDatas)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "查询确认历史失败",
			"errorCode": -1,
			"data":      "查询确认历史失败",
		})
		return
	}

	data := make([]map[string]interface{}, 0)

	for _, res := range *result {
		dd := map[string]interface{}{
			"key":        res.KeysInfo,
			"updateTime": res.UpdatedAt,
			"remark":     res.Remarks,
			"confirmer":  res.Confirmer,
			"version":    res.Version,
		}
		data = append(data, dd)
	}
	logs.Info("查询确认历史成功")
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      data,
	})
	return

}

// QueryIgnoredInfo retrieves task information which can be ignored
// from table tb_ignored_info.
func QueryIgnoredInfo(sieve map[string]interface{}) (*[]dal.IgnoreInfoStruct, error) {
	db, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, err
	}
	defer db.Close()

	var result []dal.IgnoreInfoStruct
	if err := db.Debug().Where(sieve).Order("updated_at DESC").
		Find(&result).Error; err != nil {
		logs.Error("Database error: %v", err)
		return nil, err
	}

	return &result, nil
}

/*
	获取权限的确认历史，为了和iOS兼容，此处的内容key其实传ID就可以了-----fj
*/
func QueryIgnoredHistory(c *gin.Context) {
	type queryData struct {
		AppId    int         `json:"appId"`
		Platform int         `json:"platform"`
		Key      interface{} `json:"key"`
	}
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t queryData
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("参数不合法 ，%v", err)
		c.JSON(http.StatusOK, gin.H{
			"message":   "参数不合法！",
			"errorCode": -1,
			"data":      "参数不合法！",
		})
		return
	}
	if t.Platform == 0 {
		perm_id := int(t.Key.(float64))

		result, err := dal.QueryPermHistory(map[string]interface{}{
			"perm_id": perm_id,
			"app_id":  t.AppId,
		})
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message":   "查询确认历史失败",
				"errorCode": -1,
				"data":      "查询确认历史失败",
			})
			return
		}

		data := make([]map[string]interface{}, 0)

		for _, res := range *result {
			dd := map[string]interface{}{
				"key":        res.PermId,
				"updateTime": res.UpdatedAt,
				"remark":     res.Remarks,
				"confirmer":  res.Confirmer,
				"version":    res.AppVersion,
				"status":     res.Status,
			}
			data = append(data, dd)
		}
		logs.Info("查询确认历史成功")
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data":      data,
		})
		return
	} else {
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
func QueryTaskApkBinaryCheckContentWithIgnorance_3(c *gin.Context) {
	taskID, taskExist := c.GetQuery("taskId")
	toolID, toolExist := c.GetQuery("toolId")
	if !taskExist || !toolExist {
		logs.Error("Miss taskId or toolId")
		errorReturn(c, "Miss taskId or toolId")
		return
	}

	//切换到旧版本
	if toolID != "6" {
		QueryTaskBinaryCheckContent(c)
		return
	}

	result := getDetectResult(c, taskID, toolID)
	if result == nil {
		ReturnMsg(c, FAILURE, fmt.Sprintf("Failed to get binary detect result about task id: %v", taskID))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": SUCCESS,
		"data":      result})

	logs.Debug("Get task ID %v binary detect result success", taskID)
	return
}

func getExactDetectTask(db *gorm.DB, sieve map[string]interface{}) (
	*dal.DetectStruct, error) {

	var task dal.DetectStruct
	if err := db.Debug().Where(sieve).First(&task).Error; err != nil {
		logs.Error("Database error: %v", err)
		return nil, err
	}

	return &task, nil
}

func getDetectResult(c *gin.Context, taskId string, toolId string) *[]dal.DetectQueryStruct {

	db, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer db.Close()

	task, err := getExactDetectTask(db, map[string]interface{}{"id": taskId})
	if err != nil {
		logs.Error("Task id: %v Failed to get the task information", taskId)
		return nil
	}

	//查询增量信息
	methodIgs, strIgs, _, errIg := getIgnoredInfo_2(task.AppId, task.Platform)
	if errIg != nil {
		// It's acceptable if failed to get the  negligible information.
		logs.Error("Task id: %v Failed to retrieve negligible information", taskId)
	}

	//查询基础信息和敏感信息
	contents, err := retrieveTaskAPP(db, map[string]interface{}{
		"task_id": taskId,
		"tool_id": toolId})
	if err != nil {
		logs.Error("Task id: %v Failed to retrieve APK information", taskId)
		return nil
	}
	if len(*contents) <= 0 {
		logs.Error("Task id: %v Cannot find any matched APK information", taskId)
		return nil
	}

	details, err := QueryDetectContentDetail(db, map[string]interface{}{
		"task_id": taskId,
		"tool_id": toolId})
	if err != nil {
		logs.Error("Task id: %v Failed to retrieve detect content detail", taskId)
		return nil
	}
	if len(*details) <= 0 {
		logs.Error("Task id: %v Cannot find any matched detect content detail", taskId)
		return nil
	}

	info, hasPermListFlag := getPermAPPReltion(taskId)
	detailMap := make(map[int][]dal.DetectContentDetail)
	permsMap := make(map[int]dal.PermAppRelation)
	var midResult []dal.DetectQueryStruct
	var firstResult dal.DetectQueryStruct //主要包检测结果
	for num, content := range *contents {
		var queryResult dal.DetectQueryStruct
		queryResult.Channel = content.Channel
		queryResult.ApkName = content.ApkName
		queryResult.Version = content.Version

		permission := ""
		//增量的时候，此处一般为""
		perms := strings.Split(content.Permissions, ";")
		for _, perm := range perms[0:(len(perms) - 1)] {
			permission += perm + "\n"
		}
		queryResult.Permissions = permission
		queryResult.Index = num + 1
		detailListIndex := make([]dal.DetectContentDetail, 0)
		for i := 0; i < len(*details); i++ {
			if (*details)[i].SubIndex == num {
				detailListIndex = append(detailListIndex, (*details)[i])
			}
		}
		detailMap[num+1] = detailListIndex

		if hasPermListFlag {
			var permInfo dal.PermAppRelation
			for i := 0; i < len(*info); i++ {
				if (*info)[i].SubIndex == num {
					permInfo = (*info)[i]
					break
				}
			}
			permsMap[num+1] = permInfo
		}
		if queryResult.ApkName == task.AppName {
			firstResult = queryResult
		} else {
			midResult = append(midResult, queryResult)
		}
	}
	fmt.Println(permsMap)
	finalResult := make([]dal.DetectQueryStruct, 0)
	finalResult = append(finalResult, firstResult)
	finalResult = append(finalResult, midResult...)

	appID, err := strconv.Atoi(task.AppId)
	if err != nil {
		logs.Error("Task id: %v atoi error: %v", taskId, err)
		return nil
	}
	perIgs := GetIgnoredPermission(appID)
	//任务检测结果组输出重组
	allPermList := GetPermList()
	for i := 0; i < len(finalResult); i++ {
		details := detailMap[finalResult[i].Index]
		permissions := make([]dal.Permissions, 0)

		//获取敏感信息输出结果
		methods_un, strs_un := GetDetectDetailOutInfo(details, c, methodIgs, strIgs)
		if methods_un == nil && strs_un == nil {
			return nil
		}
		finalResult[i].SMethods = methods_un
		finalResult[i].SStrs = make([]dal.SStr, 0)
		finalResult[i].SStrs_new = strs_un

		//权限结果重组
		if hasPermListFlag {
			thePerm := permsMap[finalResult[i].Index]
			permissionsP, errP := GetTaskPermissions_2(thePerm, perIgs, allPermList)
			if errP != nil || permissionsP == nil || len(*permissionsP) == 0 {
				finalResult[i].Permissions_2 = permissions
			} else {
				finalResult[i].Permissions_2 = (*permissionsP)
			}
		} else {
			finalResult[i].Permissions_2 = permissions
		}

	}

	return &finalResult
}

func getPermAPPReltion(taskID string) (*[]dal.PermAppRelation, bool) {
	//查询权限信息，此处为空--代表旧版无权限确认部分
	info, err := dal.QueryPermAppRelation(map[string]interface{}{
		"task_id": taskID})
	// var hasPermListFlag = true //标识是否进行权限分条确认的结果标识
	if err != nil || info == nil || len(*info) == 0 {
		logs.Error("taskId:" + taskID + ",未查询到该任务的权限确认信息")
		// hasPermListFlag = false
		return nil, false
	}

	return info, true
}

/**
兼容.aab查询内容
*/
// retrieveTaskAPP returns the information of APP in the binary detect task.
func retrieveTaskAPP(db *gorm.DB, sieve map[string]interface{}) (
	*[]dal.DetectInfo, error) {

	var detectInfo []dal.DetectInfo
	if err := db.Debug().Where(sieve).Find(&detectInfo).Error; err != nil {
		logs.Error("Database error: %v", err)
		return nil, err
	}

	return &detectInfo, nil
}

/**
查询apk敏感信息----fj
*/
func QueryDetectContentDetail(db *gorm.DB, sieve map[string]interface{}) (
	*[]dal.DetectContentDetail, error) {

	var result []dal.DetectContentDetail
	if err := db.Debug().Where(sieve).Order("status ASC").
		Find(&result).Error; err != nil {
		logs.Error("Database error: %v", err)
		return nil, err
	}

	return &result, nil
}

/**
敏感方法和字符串的结果输出解析---新版
*/
func GetDetectDetailOutInfo(details []dal.DetectContentDetail, c *gin.Context, methodIgs map[string]interface{}, strIgs map[string]interface{}) ([]dal.SMethod, []dal.SStr) {
	methods_un := make(MethodSlice, 0)
	methods_con := make(MethodSlice, 0)
	strs_un := make([]dal.SStr, 0)
	strs_con := make([]dal.SStr, 0)
	//旧版本不带危险等级更新标识，及更新内容
	var updateFlag = false
	var updateIds = make([]string, 0)
	var updateLevels = make([]string, 0)
	var updateConfigIds = make([]dal.DetailExtraInfo, 0)
	//配置表中信息获取
	var apiMap *map[string]interface{}
	if len(details) > 0 {
		if details[0].RiskLevel == "" {
			updateFlag = true
			//logs.Notice("需要更新旧的敏感方法")
			apiMap = GetAllAPIConfigs()
		}
	}
	//敏感方法和字符串增量形式检测结果重组
	for _, detail := range details {
		if detail.SensiType == 1 { //敏感方法
			var t dal.DetailExtraInfo
			json.Unmarshal([]byte(detail.ExtraInfo), &t)

			var method dal.SMethod
			method.Status = detail.Status
			method.Confirmer = detail.Confirmer
			method.Remark = detail.Remark
			method.ClassName = detail.ClassName
			method.Desc = detail.DescInfo
			method.Status = detail.Status
			method.Id = detail.ID
			method.MethodName = detail.KeyInfo
			method.GPFlag = t.GPFlag
			method.RiskLevel = detail.RiskLevel
			method.ConfigId = t.ConfigId
			//若为旧版本，更新内容收集
			if updateFlag {
				updateIds = append(updateIds, fmt.Sprint(method.Id))
				if v, ok := (*apiMap)[method.ClassName+"."+method.MethodName]; ok {
					info := v.(map[string]int)
					method.RiskLevel = fmt.Sprint(info["priority"])
					method.ConfigId = info["id"]
					updateLevels = append(updateLevels, method.RiskLevel)
					t.ConfigId = info["id"]
				} else {
					method.RiskLevel = "3"
					method.ConfigId = 0
					updateLevels = append(updateLevels, "3")
					t.ConfigId = 0
				}
				updateConfigIds = append(updateConfigIds, t)
			}
			if methodIgs != nil {
				if v, ok := methodIgs[detail.ClassName+"."+detail.KeyInfo]; ok {
					info := v.(map[string]interface{})
					if info["status"] != 0 && method.Status != 0 {
						method.Status = info["status"].(int)
						method.Confirmer = info["confirmer"].(string)
						method.Remark = info["remarks"].(string)
						method.OtherVersion = info["version"].(string)
					}
				}
			}
			callLocs := strings.Split(detail.CallLoc, ";")
			callLoc := make([]dal.MethodCallJson, 0)
			for _, call_loc := range callLocs[0:(len(callLocs) - 1)] {
				var call_loc_json dal.MethodCallJson
				err := json.Unmarshal([]byte(call_loc), &call_loc_json)
				if err != nil {
					logs.Error("taskId:"+fmt.Sprint(detail.TaskId)+",callLoc数据不符合要求，%v===========%s", err, call_loc)
					errorReturn(c, "callLoc数据不符合要求")
					return nil, nil
				}
				callLoc = append(callLoc, call_loc_json)
			}
			method.CallLoc = callLoc
			if method.Status == 0 {
				methods_un = append(methods_un, method)
			} else {
				methods_con = append(methods_con, method)
			}
		} else if detail.SensiType == 2 { //敏感字符串
			keys2 := make(map[string]int)
			//var keys3 = detail.KeyInfo
			var str dal.SStr
			str.Status = detail.Status
			confirmInfos := make([]dal.ConfirmInfo, 0)
			keys := strings.Split(detail.KeyInfo, ";")
			keys3 := ""
			for _, keyInfo := range keys[0 : len(keys)-1] {
				var confirmInfo dal.ConfirmInfo
				if v, ok := strIgs[keyInfo]; ok && str.Status != 0 {
					info := v.(map[string]interface{})
					if info["status"] != 0 {
						keys2[keyInfo] = 1
						confirmInfo.Key = keyInfo
						confirmInfo.Status = info["status"].(int)
						confirmInfo.Remark = info["remarks"].(string)
						confirmInfo.Status = info["status"].(int)
						confirmInfo.Confirmer = info["confirmer"].(string)
						confirmInfo.OtherVersion = info["version"].(string)
						confirmInfos = append(confirmInfos, confirmInfo)
					}
				} else {
					keys3 += keyInfo + ";"
					confirmInfo.Key = keyInfo
					confirmInfos = append(confirmInfos, confirmInfo)
				}
			}
			if keys3 == "" {
				str.Status = 1
			}
			str.Keys = detail.KeyInfo
			str.Remark = detail.Remark
			str.Confirmer = detail.Confirmer
			str.Desc = detail.DescInfo
			str.Id = detail.ID
			if detail.ExtraInfo != "" {
				var t dal.DetailExtraInfo
				json.Unmarshal([]byte(detail.ExtraInfo), &t)
				str.GPFlag = int(t.GPFlag)
			}
			callLocs := strings.Split(detail.CallLoc, ";")
			callLoc := make([]dal.StrCallJson, 0)
			for _, call_loc := range callLocs[0:(len(callLocs) - 1)] {
				var callLoc_json dal.StrCallJson
				err := json.Unmarshal([]byte(call_loc), &callLoc_json)
				if err != nil {
					logs.Error("taskId:"+fmt.Sprint(detail.TaskId)+",callLoc数据不符合要求，%v========%s", err, call_loc)
					errorReturn(c, "callLoc数据不符合要求")
					return nil, nil
				}
				callLoc = append(callLoc, callLoc_json)
			}
			str.CallLoc = callLoc
			str.ConfirmInfos = confirmInfos
			if str.Status == 0 {
				strs_un = append(strs_un, str)
			} else {
				strs_con = append(strs_con, str)
			}
		}
	}
	//保证结果未确认结果在前
	sort.Sort(MethodSlice(methods_un))
	sort.Sort(MethodSlice(methods_con))
	for _, m := range methods_con {
		methods_un = append(methods_un, m)
	}
	for _, str := range strs_con {
		strs_un = append(strs_un, str)
	}
	//异步更新
	if updateFlag {
		//logs.Notice("go 协程")
		go methodRiskLevelUpdate(&updateIds, &updateLevels, &updateConfigIds)
	}
	//logs.Notice("原线程返回")
	return methods_un, strs_un
}

//旧版本信息更新
func methodRiskLevelUpdate(ids *[]string, levels *[]string, configIds *[]dal.DetailExtraInfo) {
	err := dal.UpdateOldApkDetectDetailLevel(ids, levels, configIds)
	if err != nil {
		logs.Error("更新旧任务敏感方法危险等级失败，id信息：" + fmt.Sprint((*ids)[0]))
		utils.LarkDingOneInner("fanjuan.xqp", "更新旧任务敏感方法危险等级失败")
	}
	//logs.Notice("协程done")
}

/**
权限结果输出解析
*/
func GetTaskPermissions_2(info dal.PermAppRelation, perIgs map[int]interface{}, allPermList map[int]interface{}) (*PermSlice, error) {
	var infos []interface{}
	if err := json.Unmarshal([]byte(info.PermInfos), &infos); err != nil {
		logs.Error("taskId:" + fmt.Sprint(info.TaskId) + ",该任务的权限信息存储格式出错")
		return nil, err
	}

	var result PermSlice
	var reulst_con PermSlice
	for v, permInfo := range infos {
		var permOut dal.Permissions
		permMap := permInfo.(map[string]interface{})
		//更新权限信息
		permMap["priority"] = int(permMap["priority"].(float64))
		if v, ok := allPermList[int(permMap["perm_id"].(float64))]; ok {
			info := v.(map[string]interface{})
			permMap["priority"] = info["priority"].(int)
			permMap["ability"] = info["ability"].(string)
		}
		permOut.Id = uint(v) + 1
		permOut.Priority = permMap["priority"].(int)
		permOut.Status = int(permMap["status"].(float64))
		permOut.Key = permMap["key"].(string)
		permOut.PermId = int(permMap["perm_id"].(float64))
		permOut.Desc = permMap["ability"].(string)
		permOut.OtherVersion = permMap["first_version"].(string)

		if v, ok := perIgs[int(permMap["perm_id"].(float64))]; ok && permOut.Status != 0 {
			perm := v.(map[string]interface{})
			permOut.Status = perm["status"].(int)
			permOut.Remark = perm["remarks"].(string)
			permOut.Confirmer = perm["confirmer"].(string)
		}

		if permOut.Status == 0 {
			result = append(result, permOut)
		} else {
			reulst_con = append(reulst_con, permOut)
		}
	}
	sort.Sort(PermSlice(result))
	sort.Sort(PermSlice(reulst_con))
	for _, outInfo := range reulst_con {
		result = append(result, outInfo)
	}
	return &result, nil
}

/**
安卓确认检测结果----兼容.aab结果
*/
func ConfirmApkBinaryResultv_5(c *gin.Context) {
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.PostConfirm
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("参数不合法 ，%v", err)
		errorReturn(c, "参数不合法")
		return
	}
	//获取确认人信息
	username, _ := c.Get("username")
	usernameStr := username.(string)

	//获取任务信息
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id": t.TaskId,
	})

	//获取该任务的权限信息
	perms, errPerm := dal.QueryPermAppRelation(map[string]interface{}{
		"task_id": t.TaskId,
	})

	//是否更新任务表中detect_no_pass字段的标志
	var notPassFlag = false
	if t.Status == 2 {
		notPassFlag = true
	}

	db, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return
	}
	defer db.Close()

	if t.Type == 0 { //敏感方法和字符串确认
		detailInfo, err := QueryDetectContentDetail(db,
			map[string]interface{}{"id": t.Id})
		if err != nil || detailInfo == nil || len(*detailInfo) == 0 {
			logs.Error("taskId:"+fmt.Sprint(t.TaskId)+",不存在该检测结果，ID：%d", t.Id)
			errorReturn(c, "不存在该检测结果")
			return
		}
		if detect == nil || len(*detect) == 0 {
			logs.Error("未查询到该taskid对应的检测任务，%v", t.TaskId)
			errorReturn(c, "未查询到该taskid对应的检测任务")
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
			logs.Error("taskId:" + fmt.Sprint(t.TaskId) + ",二进制检测内容确认失败")
			errorReturn(c, "二进制检测内容确认失败")
			return
		}

		//增量忽略结果录入
		if t.Status != 0 {
			senType := (*detailInfo)[0].SensiType
			if senType == 1 { //敏感方法
				keyInfo := (*detailInfo)[0].ClassName + "." + (*detailInfo)[0].KeyInfo
				//结果写入
				if err := createIgnoreInfo(c, &t, &(*detect)[0], usernameStr, keyInfo, 1); err != nil {
					return
				}
			} else { //敏感字符串
				keys := strings.Split((*detailInfo)[0].KeyInfo, ";")
				var strIgnoreList []dal.IgnoreInfoStruct
				for _, key := range keys[0 : len(keys)-1] {
					//结果写入
					strIgnoreList = append(strIgnoreList, *createIgnoreInfoBatch(&t, &(*detect)[0], usernameStr, key, 2))
				}
				if err := dal.InsertIgnoredInfoBatch(&strIgnoreList); err != nil {
					errorReturn(c, "增量信息更新失败！")
					return
				}
			}
		}
	} else {
		if errPerm != nil || perms == nil || len(*perms) == 0 {
			logs.Error("taskId:" + fmt.Sprint(t.TaskId) + ",未查询到该任务的检测信息")
			errorReturn(c, "未查询到该任务的检测信息")
			return
		}
		if t.Index > len(*perms) {
			logs.Error("权限结果数组下标越界")
			errorReturn(c, "权限结果数组下标越界")
			return
		}
		permsInfoDB := (*perms)[t.Index-1].PermInfos
		var permList []interface{}
		if err := json.Unmarshal([]byte(permsInfoDB), &permList); err != nil {
			logs.Error("taskId:" + fmt.Sprint(t.TaskId) + ",该任务的权限存储信息格式出错")
			errorReturn(c, "该任务的权限存储信息格式出错")
			return
		}
		//增加数组越界维护
		if t.Id > len(permList) {
			logs.Error("权限ID越界")
			errorReturn(c, "权限ID越界")
			return
		}
		permMap := permList[t.Id-1].(map[string]interface{})
		permMap["status"] = t.Status

		permId := int(permMap["perm_id"].(float64))
		newPerms, _ := json.Marshal(permList)
		(*perms)[t.Index-1].PermInfos = string(newPerms)
		if err := dal.UpdataPermAppRelation(&(*perms)[t.Index-1]); err != nil {
			logs.Error("taskId:" + fmt.Sprint(t.TaskId) + ",更新任务权限确认情况失败")
			errorReturn(c, "更新任务权限确认情况失败")
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
			logs.Error("taskId:" + fmt.Sprint(t.TaskId) + ",权限操作历史写入失败！")
			errorReturn(c, "权限操作历史写入失败！")
			return
		}
	}

	//任务状态更新
	updateInfo, _ := taskStatusUpdate(t.TaskId, t.ToolId, &(*detect)[0], notPassFlag, 1)
	if updateInfo != "" {
		errorReturn(c, updateInfo)
	}
	logs.Info("confirm success +id :%d", t.Id)
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      "success",
	})
	return
}

/**
任务确认状态更新
*/
func taskStatusUpdate(taskId int, toolId int, detect *dal.DetectStruct, notPassFlag bool, confirmLark int) (string, int) {
	condition := "deleted_at IS NULL and task_id='" + strconv.Itoa(taskId) + "' and tool_id='" + strconv.Itoa(toolId) + "'"
	counts, countsUn := dal.QueryUnConfirmDetectContent(condition)

	perms, _ := dal.QueryPermAppRelation(map[string]interface{}{
		"task_id": taskId,
	})
	var permFlag = true
	var updateFlag = false
	var permCounts = 0
	if perms == nil || len(*perms) == 0 {
		logs.Error("taskId:" + fmt.Sprint(taskId) + ",该任务无权限检测信息！")
	} else {
		for i := 0; i < len(*perms); i++ {
			permsInfoDB := (*perms)[i].PermInfos
			var permList []interface{}
			if err := json.Unmarshal([]byte(permsInfoDB), &permList); err != nil {
				return "该任务的权限存储信息格式出错", 0
			}
			for _, m := range permList {
				permInfo := m.(map[string]interface{})
				if permInfo["status"].(float64) == 0 {
					permFlag = false
					permCounts += 1
				}
			}
		}
	}

	logs.Notice("当前确认情况，字符串和方法剩余：" + fmt.Sprint(counts) + " ,权限是否全部确认：" + fmt.Sprint(permFlag) + "权限剩余数：" + fmt.Sprint(permCounts))
	if counts == 0 && permFlag {
		updateFlag = true
		if countsUn == 0 {
			detect.Status = 1
		} else {
			detect.Status = 2
		}
	}
	if notPassFlag {
		detect.DetectNoPass = countsUn - counts
		updateFlag = true
	}
	if updateFlag {
		err := dal.UpdateDetectModelNew(*detect)
		if err != nil {
			logs.Error("taskId:"+fmt.Sprint(taskId)+",任务确认状态更新失败！%v", err)
			utils.LarkDingOneInner("fanjuan.xqp", "任务ID："+fmt.Sprint(taskId)+"确认状态更新失败，失败原因："+err.Error())
			return "任务确认状态更新失败！", 0
		}
		if err := StatusDeal(*detect, confirmLark); err != nil {
			return err.Error(), 0
		}
	}
	//if countsUn == 0 && permFlag {
	//	if err := StatusDeal(*detect); err != nil {
	//		return err.Error(), 0
	//	}
	//	//logs.Notice("回调了CI接口")
	//}
	return "", counts + permCounts

}

/**
新增敏感字符or敏感方法确认历史
*/
func createIgnoreInfo(c *gin.Context, t *dal.PostConfirm, detect *dal.DetectStruct, usernameStr string, key string, senType int) error {
	var igInfo dal.IgnoreInfoStruct
	igInfo.Platform = detect.Platform
	igInfo.AppId, _ = strconv.Atoi(detect.AppId)
	igInfo.SensiType = senType
	igInfo.KeysInfo = key
	igInfo.Confirmer = usernameStr
	igInfo.Remarks = t.Remark
	igInfo.Version = detect.AppVersion
	igInfo.Status = t.Status
	igInfo.TaskId = t.TaskId
	err := dal.InsertIgnoredInfo(igInfo)
	if err != nil {
		errorReturn(c, "增量信息更新失败！")
		return err
	}
	return nil
}

func createIgnoreInfoBatch(t *dal.PostConfirm, detect *dal.DetectStruct, usernameStr string, key string, senType int) *dal.IgnoreInfoStruct {
	var igInfo dal.IgnoreInfoStruct
	igInfo.Platform = detect.Platform
	igInfo.AppId, _ = strconv.Atoi(detect.AppId)
	igInfo.SensiType = senType
	igInfo.KeysInfo = key
	igInfo.Confirmer = usernameStr
	igInfo.Remarks = t.Remark
	igInfo.Version = detect.AppVersion
	igInfo.Status = t.Status
	igInfo.TaskId = t.TaskId
	return &igInfo
}

/**
请求错误信息返回统一格式
*/
func errorReturn(c *gin.Context, message string, errCodes ...int) {
	var errCode = -1
	if len(errCodes) > 0 {
		errCode = errCodes[0]
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   message,
		"errorCode": errCode,
		"data":      message,
	})
	return
}
