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

const logHeader = "[task id %v]"

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

func GetIgnoredPermission(appId interface{}) map[int]interface{} {
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
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, "miss task id or tool id")
		return
	}

	//切换到旧版本
	if toolID != "6" {
		QueryTaskBinaryCheckContent(c)
		return
	}

	result := getDetectResult(taskID, toolID)
	if result == nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("get binary detect result (task id: %v) failed", taskID))
		return
	}

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success", result)
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

func getDetectResult(taskId string, toolId string) *[]dal.DetectQueryStruct {

	header := fmt.Sprintf(logHeader, taskId)
	task, err := getExactDetectTask(database.DB(), map[string]interface{}{"id": taskId})
	if err != nil {
		logs.Error("%s get detect task failed: %v", header, err)
		return nil
	}

	//查询基础信息和敏感信息
	contents, err := retrieveTaskAPP(database.DB(), map[string]interface{}{
		"task_id": taskId,
		"tool_id": toolId})
	if err != nil {
		logs.Error("%s read app information failed: %v", header, err)
		return nil
	}
	if len(contents) <= 0 {
		logs.Error("%s cannot find any matched app information", header)
		return nil
	}

	details, err := readDetectContentDetail(database.DB(), map[string]interface{}{
		"task_id": taskId,
		"tool_id": toolId})
	if err != nil {
		logs.Error("%s read tb_detect_content_detail failed: %v", header, err)
		return nil
	}
	if len(details) <= 0 {
		logs.Error("%s cannot find any matched detect content detail", header)
		return nil
	}

	info, hasPermListFlag := getPermAPPReltion(taskId)
	detailMap := make(map[int][]dal.DetectContentDetail)
	permsMap := make(map[int]dal.PermAppRelation)
	var midResult []dal.DetectQueryStruct
	var firstResult dal.DetectQueryStruct //主要包检测结果
	for num, content := range contents {
		var queryResult dal.DetectQueryStruct
		queryResult.Channel = content.Channel
		queryResult.ApkName = content.ApkName
		queryResult.Version = content.Version
		queryResult.Index = num + 1
		var detailListIndex []dal.DetectContentDetail
		for i := 0; i < len(details); i++ {
			if details[i].SubIndex == num {
				detailListIndex = append(detailListIndex, details[i])
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
	finalResult := []dal.DetectQueryStruct{firstResult}
	finalResult = append(finalResult, midResult...)

	//任务检测结果组输出重组
	for i := 0; i < len(finalResult); i++ {
		details := detailMap[finalResult[i].Index]
		//获取敏感信息输出结果
		methods_un, strs_un := GetDetectDetailOutInfo(details)
		if methods_un == nil && strs_un == nil {
			return nil
		}
		finalResult[i].SMethods = methods_un
		finalResult[i].SStrs_new = strs_un

		if hasPermListFlag { //权限结果重组
			permissionsP, err := packPermissionListAndroid(permsMap[finalResult[i].Index].PermInfos)
			if err != nil {
				logs.Error("%s construct permission list failed: %v", header, err)
				return nil
			}
			finalResult[i].Permissions_2 = (*permissionsP)
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
	[]dal.DetectInfo, error) {

	var detectInfo []dal.DetectInfo
	if err := db.Debug().Where(sieve).Find(&detectInfo).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return detectInfo, nil
}

func readExactDetectContentDetail(db *gorm.DB, sieve map[string]interface{}) (
	*dal.DetectContentDetail, error) {

	data, err := readDetectContentDetail(db, sieve)
	if err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	if len(data) <= 0 {
		return nil, nil
	}

	return &data[0], nil
}

/**
查询apk敏感信息----fj
*/
func readDetectContentDetail(db *gorm.DB, sieve map[string]interface{}) (
	[]dal.DetectContentDetail, error) {

	var result []dal.DetectContentDetail
	if err := db.Debug().Where(sieve).Order("status ASC").
		Find(&result).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return result, nil
}

/**
敏感方法和字符串的结果输出解析---新版
*/
func GetDetectDetailOutInfo(details []dal.DetectContentDetail) ([]dal.SMethod, []dal.SStr) {
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
			callLocs := strings.Split(detail.CallLoc, ";")
			callLoc := make([]dal.MethodCallJson, 0)
			for _, call_loc := range callLocs[0:(len(callLocs) - 1)] {
				var call_loc_json dal.MethodCallJson
				if err := json.Unmarshal([]byte(call_loc), &call_loc_json); err != nil {
					logs.Error("taskId:"+fmt.Sprint(detail.TaskId)+",callLoc数据不符合要求，%v===========%s", err, call_loc)
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
			var str dal.SStr
			str.Status = detail.Status
			confirmInfos := make([]dal.ConfirmInfo, 0)
			keys := strings.Split(detail.KeyInfo, ";")

			for _, keyInfo := range keys[0 : len(keys)-1] {

				confirmInfos = append(confirmInfos, dal.ConfirmInfo{
					Key:       keyInfo,
					Status:    detail.Status,
					Confirmer: detail.Confirmer,
					Remark:    detail.Remark})
			}
			str.Keys = detail.KeyInfo
			str.Remark = detail.Remark
			str.Confirmer = detail.Confirmer
			str.Desc = detail.DescInfo
			str.Id = detail.ID
			str.RiskLevel = detail.RiskLevel
			if detection, err := retrieveDetection(database.DB(),
				map[string]interface{}{"key_info": strings.Split(str.Keys, ";")[0]}); err != nil {
				str.RiskLevel = "unknown"
			} else {
				str.RiskLevel = fmt.Sprint(detection.Priority)
			}
			if detail.ExtraInfo != "" {
				var t dal.DetailExtraInfo
				json.Unmarshal([]byte(detail.ExtraInfo), &t)
				str.GPFlag = int(t.GPFlag)
			}
			callLocs := strings.Split(detail.CallLoc, ";")
			callLoc := make([]dal.StrCallJson, 0)
			for _, call_loc := range callLocs[0:(len(callLocs) - 1)] {
				var callLoc_json dal.StrCallJson
				if err := json.Unmarshal([]byte(call_loc), &callLoc_json); err != nil {
					logs.Error("taskId:"+fmt.Sprint(detail.TaskId)+",callLoc数据不符合要求，%v========%s", err, call_loc)
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

func retrieveDetection(db *gorm.DB, sieve map[string]interface{}) (
	*dal.DetectConfigStruct, error) {

	var detection dal.DetectConfigStruct
	if err := db.Debug().Where(sieve).
		First(&detection).Error; err != nil {
		logs.Error("Database error: %v", err)
		return nil, err
	}

	return &detection, nil
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
func packPermissionListAndroid(permission string) (*PermSlice, error) {
	var permissionList []interface{}
	if err := json.Unmarshal([]byte(permission), &permissionList); err != nil {
		logs.Error("unmarshal error: %v content: %v", err, permission)
		return nil, err
	}

	var result PermSlice
	var reulst_con PermSlice
	for v, permInfo := range permissionList {
		permMap := permInfo.(map[string]interface{})
		var permOut dal.Permissions
		permOut.Id = uint(v) + 1
		permOut.Priority = int(permMap["priority"].(float64))
		permOut.RiskLevel = fmt.Sprint(permMap["priority"])
		status, _ := strconv.Atoi(fmt.Sprint(permMap["status"]))
		permOut.Status = status
		permOut.Key = fmt.Sprint(permMap["key"])
		permOut.PermId = int(permMap["perm_id"].(float64))
		permOut.Desc = fmt.Sprint(permMap["ability"])
		// permOut.OtherVersion = permMap["first_version"].(string)
		permOut.Confirmer = fmt.Sprint(permMap["confirmer"])
		permOut.Remark = fmt.Sprint(permMap["remark"])
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
任务确认状态更新
*/
func taskStatusUpdate(taskId uint, toolId int, detect *dal.DetectStruct, notPassFlag bool, confirmLark int) (string, int) {

	condition := fmt.Sprintf("deleted_at IS NULL and task_id='%v' and tool_id='%v'", taskId, toolId)
	counts, countsUn := QueryUnConfirmDetectContent(database.DB(), condition)

	perms, _ := dal.QueryPermAppRelation(map[string]interface{}{"task_id": taskId})
	var permFlag = true
	var updateFlag = false
	var permCounts = 0
	if perms == nil || len(*perms) == 0 {
		logs.Warn("taskId:" + fmt.Sprint(taskId) + ",该任务无权限检测信息！")
	} else {
		for i := 0; i < len(*perms); i++ {
			permsInfoDB := (*perms)[i].PermInfos
			var permList []interface{}
			if err := json.Unmarshal([]byte(permsInfoDB), &permList); err != nil {
				return "该任务的权限存储信息格式出错", 0
			}
			for _, m := range permList {
				permInfo := m.(map[string]interface{})
				if fmt.Sprint(permInfo["status"]) == "0" {
					// if permInfo["status"].(float64) == 0 {
					permFlag = false
					permCounts++
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

	return "", counts + permCounts
}
func readExactPermAPPRelation(db *gorm.DB, sieve map[string]interface{}) (*dal.PermAppRelation, error) {

	var result dal.PermAppRelation
	if err := db.Debug().Where(sieve).Last(&result).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return &result, nil
}

// func getLastestPermAPPRelation(db *gorm.DB, sieve map[string]interface{}) (*dal.PermAppRelation, error) {

// 	data, err := readPermAPPRelation(db, sieve)
// 	if err != nil {
// 		logs.Error("read tb_perm_app_relation failed: %v", err)
// 		return nil, err
// 	}

// 	if len(data) <= 0 {
// 		return nil, nil
// 	}

// 	// The default order is created_at asc, so return the final element.
// 	return &data[len(data)-1], nil
// }

func readPermAPPRelation(db *gorm.DB, sieve map[string]interface{}) ([]dal.PermAppRelation, error) {

	var result []dal.PermAppRelation
	if err := db.Debug().Where(sieve).Find(&result).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return result, nil
}

/**
未确认敏感信息数据量查询-----fj
*/
func QueryUnConfirmDetectContent(db *gorm.DB, condition string) (int, int) {

	var total dal.RecordTotal
	if err := db.Debug().Table(dal.DetectContentDetail{}.TableName()).Select("sum(case when status = '0' then 1 else 0 end) as total, sum(case when status <> '1' then 1 else 0 end) as total_un").Where(condition).Find(&total).Error; err != nil {
		logs.Error("Database error: %v", err)
		return -1, -1
	}

	return int(total.Total), int(total.TotalUn)
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
