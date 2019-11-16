package detect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"

	"github.com/gin-gonic/gin"

	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
)

//0--一般，1--低危，2--中危，3--高危
var iosPrivacyPriority = map[string]int{
	"NSContactsUsageDescription":          3, //通讯录
	"NSLocationUsageDescription":          1, //位置
	"NSMicrophoneUsageDescription":        3, //麦克风
	"NSPhotoLibraryUsageDescription":      2, //相册
	"NSPhotoLibraryAddUsageDescription":   2, //保存到相册
	"NSLocationAlwaysUsageDescription":    1, //始终访问位置
	"NSCameraUsageDescription":            2, //相机
	"NSLocationWhenInUseUsageDescription": 1, //在使用期间访问位置
	"NSAppleMusicUsageDescription":        3, //媒体资料库
}

//ios检测结果黑名单描述说明
var iosBlackListDescription = map[string]interface{}{
	"alipay":           map[string]interface{}{"blackType": "支付", "description": "苹果对支付的严格，控制，需要参考相关苹果的政策确认是否符合支付条件"},
	"jspatch":          map[string]interface{}{"blackType": "热更新", "description": "苹果平台不愿意支持热更新"},
	"JPEngine":         map[string]interface{}{"blackType": "热更新", "description": "苹果平台不愿意支持热更新"},
	"JPLoader":         map[string]interface{}{"blackType": "热更新", "description": "苹果平台不愿意支持热更新"},
	"PayPal":           map[string]interface{}{"blackType": "支付", "description": "苹果对支付的严格，控制，需要参考相关苹果的政策确认是否符合支付条件"},
	"PayPalpay":        map[string]interface{}{"blackType": "支付", "description": "苹果对支付的严格，控制，需要参考相关苹果的政策确认是否符合支付条件"},
	"AVAudioRecorder":  map[string]interface{}{"blackType": "录音", "description": "隐私权限问题，需要确认是否符合调用场景"},
	"AVAudioSession":   map[string]interface{}{"blackType": "音频", "description": "隐私权限问题，需要确认是否符合调用场景"},
	"AVCaptureSession": map[string]interface{}{"blackType": "视频", "description": "隐私权限问题，需要确认是否符合调用场景"},
	"items-searvices":  map[string]interface{}{"blackType": "协议", "description": "通过苹果的items-searvices协议，可以进行ipa的安装，存在风险，苹果对渠道的控制"},
}

/**
 *iOS 检测结果jsonContent处理
 */
func iOSResultClassify(taskId, toolId, appId int, jsonContent string) (bool, bool, int) {
	warnFlag := false
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &dat); err != nil {
		logs.Error("json转map出错！", err.Error())
		return false, warnFlag, 0
	}
	appName := dat["name"].(string)
	version := dat["version"].(string)
	bundleId := dat["bundle_id"].(string)
	minVersion := dat["min_version"].(string)
	sdkVersion := dat["tar_version"].(string)

	lastTaskId := dal.QueryLastTaskId(taskId) //获取相同app上次检测taskId
	//获取上次黑名单检测结果
	var blacklist []interface{}
	var method []interface{}
	lastDetectContent, err := dal.QueryNewIOSDetectModel(database.DB(), map[string]interface{}{
		"taskId": lastTaskId})
	if err != nil {
		logs.Warn("read iOS detect content failed: %v", err)
	}
	for _, lastDetect := range lastDetectContent {
		if lastDetect.DetectType == "blacklist" {
			b := make(map[string]interface{})
			err := json.Unmarshal([]byte(lastDetect.DetectContent), &b)
			if err != nil {
				logs.Error("Umarshal failed:", err.Error())
			}
			blacklist = b["blackList"].([]interface{})
		}
		if lastDetect.DetectType == "method" {
			m := make(map[string]interface{})
			err := json.Unmarshal([]byte(lastDetect.DetectContent), &m)
			if err != nil {
				logs.Error("Umarshal failed:", err.Error())
			}
			method = m["method"].([]interface{})
		}
	}
	//黑名单处理
	var blackDetect dal.IOSNewDetectContent
	blackContent := dat["blacklist_in_app"]
	if blackContent != nil {
		var blackList []map[string]interface{}
		for k, v := range blackContent.(map[string]interface{}) {
			blackMap := make(map[string]interface{})
			if len(v.([]interface{})) == 0 {
				continue //黑名单内容为空，跳过
			}
			blackMap["name"] = k
			blackMap["content"] = v
			if blacklist != nil || len(blacklist) != 0 { //diff处理
				status, confirmer, remark := iosDetectDiff(blackMap, blacklist)
				if status == 1 {
					blackMap["status"] = status
					blackMap["confirmer"] = confirmer
					blackMap["remark"] = remark
				} else {
					blackMap["status"] = 0
					blackMap["confirmer"] = ""
					blackMap["remark"] = ""
				}
			} else {
				blackMap["status"] = 0
				blackMap["confirmer"] = ""
				blackMap["remark"] = ""
			}
			blackList = append(blackList, blackMap)
			if k == "itms-services" {
				warnFlag = true
			}
		}
		BlackContentValue, err := json.Marshal(map[string]interface{}{
			"blackList": blackList,
		})
		if err != nil {
			logs.Error("map转json出错！")
			return false, warnFlag, 0
		}
		blackDetect.DetectContent = string(BlackContentValue)
		blackDetect.DetectType = "blacklist"
		blackDetect.ToolId = toolId
		blackDetect.TaskId = taskId
		blackDetect.AppId = appId
		blackDetect.AppName = appName
		blackDetect.Version = version
		blackDetect.BundleId = bundleId
		blackDetect.MinVersion = minVersion
		blackDetect.SdkVersion = sdkVersion
		blackDetect.JsonContent = jsonContent
	}
	//可疑方法名处理
	var methodDetect dal.IOSNewDetectContent
	methodContent := dat["methods_in_app"]
	if methodContent != nil {
		var methodList []map[string]interface{}
		//异常处理
		var newMethodContent []interface{}
		switch methodContent.(type) {
		case map[string]interface{}:
			newMethodContent = []interface{}{methodContent}
		case []interface{}:
			newMethodContent = methodContent.([]interface{})
		}
		for _, temMethod := range newMethodContent {
			methodMap := make(map[string]interface{})
			susApi := temMethod.(map[string]interface{})["api_name"].(string)
			susClass := temMethod.(map[string]interface{})["class_name"].(string)
			methodMap["name"] = susApi
			methodMap["content"] = susClass
			if method != nil || len(method) != 0 { //diff 处理
				status, confirmer, remark := iosDetectDiff(methodMap, method)
				if status == 1 {
					methodMap["status"] = status
					methodMap["confirmer"] = confirmer
					methodMap["remark"] = remark
				} else {
					methodMap["status"] = 0
					methodMap["confirmer"] = ""
					methodMap["remark"] = ""
				}
			} else {
				methodMap["status"] = 0
				methodMap["confirmer"] = ""
				methodMap["remark"] = ""
			}
			methodList = append(methodList, methodMap)
		}
		methodContentValue, err := json.Marshal(map[string]interface{}{
			"method": methodList,
		})
		if err != nil {
			logs.Error("map转json出错！")
			return false, warnFlag, 0
		}
		methodDetect.DetectContent = string(methodContentValue)
		methodDetect.DetectType = "method"
		methodDetect.ToolId = toolId
		methodDetect.TaskId = taskId
		methodDetect.AppId = appId
		methodDetect.AppName = appName
		methodDetect.Version = version
		methodDetect.BundleId = bundleId
		methodDetect.MinVersion = minVersion
		methodDetect.SdkVersion = sdkVersion
		methodDetect.JsonContent = jsonContent
	}
	//权限处理
	var privacyDetect dal.IOSNewDetectContent
	privacyContent := dat["privacy_keys"]
	if privacyContent != nil {
		var privacyList []map[string]interface{}
		for e, c := range privacyContent.(map[string]interface{}) {
			privacyMap := make(map[string]interface{})
			privacyMap["permission"] = e
			privacyMap["permission_C"] = c
			if priority, ok := iosPrivacyPriority[e]; ok {
				privacyMap["priority"] = priority
			} else {
				privacyMap["priority"] = 3
			}
			//找到权限确认信息
			confirmHistory := dal.QueryPrivacyHistoryModel(map[string]interface{}{
				"app_id":     appId,
				"appname":    appName,
				"permission": e,
				"platform":   1,
			})
			if confirmHistory == nil || len(*confirmHistory) == 0 {
				privacyMap["confirmer"] = ""
				privacyMap["confirmVersion"] = ""
				privacyMap["confirmReason"] = ""
				privacyMap["status"] = 0
			} else {
				privacyMap["confirmer"] = (*confirmHistory)[0].Confirmer
				privacyMap["confirmReason"] = (*confirmHistory)[0].ConfirmReason
				privacyMap["confirmVersion"] = (*confirmHistory)[len(*confirmHistory)-1].ConfirmVersion
				privacyMap["status"] = 1
			}
			privacyList = append(privacyList, privacyMap)
		}
		privacyContentValue, err := json.Marshal(map[string]interface{}{
			"privacy": privacyList,
		})
		if err != nil {
			logs.Error("map转json出错！")
			return false, warnFlag, 0
		}
		privacyDetect.DetectContent = string(privacyContentValue)
		privacyDetect.DetectType = "privacy"
		privacyDetect.ToolId = toolId
		privacyDetect.TaskId = taskId
		privacyDetect.AppId = appId
		privacyDetect.AppName = appName
		privacyDetect.Version = version
		privacyDetect.BundleId = bundleId
		privacyDetect.MinVersion = minVersion
		privacyDetect.SdkVersion = sdkVersion
		privacyDetect.JsonContent = jsonContent
	}
	insertFlag := dal.InsertNewIOSDetect(blackDetect, methodDetect, privacyDetect)
	//更新tb_binary_detect中status值
	err, unRes := changeTotalStatus(taskId, toolId, 0)
	if err != nil {
		logs.Error("判断总的total status出错！", err.Error())
	}
	return insertFlag, warnFlag, unRes
}

//检测结果与上一次检测结果比较
func iosDetectDiff(newDetect map[string]interface{}, lastDetect []interface{}) (status int, confirmer, remark string) {
	name := newDetect["name"]
	content := newDetect["content"]
	for _, last := range lastDetect {
		if last.(map[string]interface{})["name"].(string) == name.(string) {
			if _, ok := last.(map[string]interface{})["status"]; !ok { //bug导致异常修复后的兼容
				return 0, "", ""
			}
			var status int
			switch last.(map[string]interface{})["status"].(type) {
			case float64:
				status = int(last.(map[string]interface{})["status"].(float64))
			case int:
				status = last.(map[string]interface{})["status"].(int)
			}
			confirmer := last.(map[string]interface{})["confirmer"].(string)
			remark := last.(map[string]interface{})["remark"].(string)

			last_content := last.(map[string]interface{})["content"]
			switch content.(type) {
			//method 比较
			case string:
				if last_content.(string) == content.(string) {
					return status, confirmer, remark
				}
			// blacklist比较
			case []interface{}:
				if compareSlice(last_content.([]interface{}), content.([]interface{})) {
					return status, confirmer, remark
				}
			}
		}
	}
	return 0, "", ""
}

//比较两个数组是否deep相等，与list的顺序无关
func compareSlice(a, b []interface{}) bool {
	m := make(map[string]bool)
	for i := 0; i < len(a); i++ {
		m[a[i].(string)] = true
	}
	if len(a) != len(b) {
		return false
	}
	for _, value := range b {
		if _, ok := m[value.(string)]; !ok {
			return false
		}
	}
	return true
}

/**
 *查询iOS静态检测结果
 */
func QueryIOSTaskBinaryCheckContent(c *gin.Context) {
	taskId := c.DefaultQuery("taskId", "")
	if taskId == "" {
		logs.Error("缺少taskId参数")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少taskId参数",
			"errorCode": -1,
			"data":      "缺少taskId参数",
		})
		return
	}
	toolId := c.DefaultQuery("toolId", "")
	if toolId == "" {
		logs.Error("缺少toolId参数")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少toolId参数",
			"errorCode": -2,
			"data":      "缺少toolId参数",
		})
		return
	}
	//首先查询新库
	iosTaskBinaryCheckContent, err := dal.QueryNewIOSDetectModel(database.DB(), map[string]interface{}{
		"taskId": taskId,
		"toolId": toolId,
	})
	//如果数据在新库中查不到就去访问老库
	if err != nil || len(iosTaskBinaryCheckContent) == 0 {
		//中间数据处理
		detect := dal.QueryDetectModelsByMap(map[string]interface{}{
			"id": taskId,
		})
		aId, _ := strconv.Atoi((*detect)[0].AppId)
		sId, _ := strconv.Atoi(taskId)
		tId, _ := strconv.Atoi(toolId)
		middleFlag, dealFlag := middleDataDeal(sId, tId, aId)
		if middleFlag {
			//如果中间数据处理失败
			if dealFlag == false {
				c.JSON(http.StatusOK, gin.H{
					"message":   "中间数据处理失败！",
					"errorCode": 0,
					"data":      "中间数据处理失败！",
				})
				return
			} else {
				//中间数据处理成功后，重新读取一次新库
				iosTaskBinaryCheckContent, _ = dal.QueryNewIOSDetectModel(database.DB(), map[string]interface{}{
					"taskId": taskId,
					"toolId": toolId,
				})
			}
		} else {
			//不再中间库，去老库中寻找
			conditionOld := "task_id='" + taskId + "' and tool_id='" + toolId + "'"
			content := dal.QueryTaskBinaryCheckContent(conditionOld)
			if content == nil || len(*content) == 0 {
				logs.Info("未查询到检测内容")
				c.JSON(http.StatusOK, gin.H{
					"message":   "未查询到检测内容",
					"errorCode": -3,
					"data":      "未查询到检测内容",
				})
				return
			} else {
				c.JSON(http.StatusOK, gin.H{
					"message":   "success",
					"errorCode": 0,
					"data":      (*content)[0],
				})
				return
			}
		}
	}
	//返回检测结果给前端
	data := map[string]interface{}{
		"appName":    iosTaskBinaryCheckContent[0].AppName,
		"appVersion": iosTaskBinaryCheckContent[0].Version,
		"bundleId":   iosTaskBinaryCheckContent[0].BundleId,
		"sdkVersion": iosTaskBinaryCheckContent[0].SdkVersion,
		"minVersion": iosTaskBinaryCheckContent[0].MinVersion,
	}
	for _, iosContent := range iosTaskBinaryCheckContent {
		var m map[string]interface{}
		err := json.Unmarshal([]byte(iosContent.DetectContent), &m)
		if err != nil {
			logs.Error("数据库内容转map出错!", err.Error())
			c.JSON(http.StatusOK, gin.H{
				"message":   "数据库内容读取错误！",
				"errorCode": -1,
				"data":      []interface{}{},
			})
			return
		}
		for k, v := range m {
			var highRisk []interface{}
			var middleRisk []interface{}
			var lowRisk []interface{}
			var notice []interface{}
			var sortedPrivacy []interface{}
			//权限兼容处理
			if k == "privacy" {
				for _, pp := range v.([]interface{}) {
					//添加权限优先级
					if _, ok := pp.(map[string]interface{})["priority"]; !ok {
						if priority, ok := iosPrivacyPriority[pp.(map[string]interface{})["permission"].(string)]; ok {
							pp.(map[string]interface{})["priority"] = priority
						} else {
							pp.(map[string]interface{})["priority"] = 3
						}
					}
					//添加权限确认信息
					if pp.(map[string]interface{})["confirmer"] != "" {
						pp.(map[string]interface{})["status"] = 1
					} else {
						pp.(map[string]interface{})["status"] = 0
					}
					//按照优先级排列
					temPriority := pp.(map[string]interface{})["priority"]
					switch temPriority.(type) {
					case int:
						temPriority = pp.(map[string]interface{})["priority"].(int)
					case float64:
						temPriority = int(pp.(map[string]interface{})["priority"].(float64))
					}
					if temPriority == 3 {
						highRisk = append(highRisk, pp)
					} else if temPriority == 2 {
						middleRisk = append(middleRisk, pp)
					} else if temPriority == 1 {
						lowRisk = append(lowRisk, pp)
					} else {
						notice = append(notice, pp)
					}
				}
				for _, high := range highRisk {
					sortedPrivacy = append(sortedPrivacy, high)
				}
				for _, middle := range middleRisk {
					sortedPrivacy = append(sortedPrivacy, middle)
				}
				for _, low := range lowRisk {
					sortedPrivacy = append(sortedPrivacy, low)
				}
				for _, noti := range notice {
					sortedPrivacy = append(sortedPrivacy, noti)
				}
				v = sortedPrivacy
			}
			//黑名单增加描述
			var addDescriptionBlackList []interface{}
			if k == "blackList" {
				for _, black := range v.([]interface{}) {
					blackName := black.(map[string]interface{})["name"]
					if blackDescription, ok := iosBlackListDescription[blackName.(string)]; ok {
						black.(map[string]interface{})["blackType"] = blackDescription.(map[string]interface{})["blackType"]
						black.(map[string]interface{})["description"] = blackDescription.(map[string]interface{})["description"]
					}
					addDescriptionBlackList = append(addDescriptionBlackList, black)
				}
				v = addDescriptionBlackList
			}
			data[k] = v
		}
		continue
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      data,
	})
}

/**
 *更新iOS静态检测结果
 */
type IOSConfirm struct {
	TaskId         int    `json:"taskId"           form:"taskId"`
	ToolId         int    `json:"toolId"           form:"toolId"`
	Status         int    `json:"status"           form:"status"`
	Remark         string `json:"remark"           form:"remark"`
	ConfirmType    int    `json:"confirmType"      form:"confirmType"` //0是旧样式黑名单，1是新样式黑名单，2是可疑方法，3是权限
	ConfirmContent string `json:"confirmContent"   form:"confirmContent"`
}

func ConfirmIOSBinaryResult(c *gin.Context) {
	//参数校验
	var ios IOSConfirm
	if err := c.BindJSON(&ios); err != nil {
		logs.Error("确认二进制检测结果传参出错！", err.Error())
		c.JSON(http.StatusOK, gin.H{
			"message":   "参数不合法！",
			"errorCode": -1,
			"data":      "参数不合法！",
		})
		return
	}
	//参数异常处理
	if ios.ConfirmType < 0 || ios.ConfirmType > 3 {
		c.JSON(http.StatusOK, gin.H{
			"message":   "参数错误！",
			"errorCode": -1,
			"data":      "参数错误，id和permission不能同时传入！",
		})
		return
	}
	//获取确认人信息
	username, _ := c.Get("username")
	if confirmIOSBinaryResult(ios, username.(string)) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data":      "success",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "update failed！",
			"errorCode": -1,
			"data":      "update failed！",
		})
	}
}
func confirmIOSBinaryResult(ios IOSConfirm, confirmer string) bool {
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id": ios.TaskId,
	})
	//兼容旧接口内容
	if ios.ConfirmType == 0 {
		data := make(map[string]string)
		data["task_id"] = strconv.Itoa(ios.TaskId)
		data["tool_id"] = strconv.Itoa(ios.ToolId)
		data["confirmer"] = confirmer
		data["remark"] = ios.Remark
		data["status"] = strconv.Itoa(ios.Status)
		flag := dal.ConfirmBinaryResult(data)
		if !flag {
			logs.Error("二进制检测内容确认失败")
			return false
		}
		//更新旧接口任务状态
		condition := "task_id = '" + fmt.Sprint(ios.TaskId) + "'"
		detectContent := dal.QueryTaskBinaryCheckContent(condition)
		if detectContent == nil || len(*detectContent) == 0 {
			logs.Error("未查询到相关二进制检测内容,更新任务状态失败")
			return false
		} else {
			changeFlag := true
			for _, detectCon := range *detectContent {
				if detectCon.Status == 0 {
					changeFlag = false
					break
				}
			}
			if changeFlag {
				(*detect)[0].Status = 1
				err := dal.UpdateDetectModelNew((*detect)[0])
				if err != nil {
					logs.Error("更新任务状态失败，任务ID："+fmt.Sprint(ios.TaskId)+",错误原因:%v", err)
					return false
				}
			}
		}
		return true
	}
	//新接口内容
	var queryKey string
	switch ios.ConfirmType {
	case 1:
		queryKey = "blackList"
	case 2:
		queryKey = "method"
	case 3:
		queryKey = "privacy"
	}
	//数据库读取检测内容转为map
	iosDetect, err := dal.QueryNewIOSDetectModel(database.DB(), map[string]interface{}{
		"taskId":      ios.TaskId,
		"toolId":      ios.ToolId,
		"detect_type": queryKey,
	})
	if err != nil || len(iosDetect) == 0 {
		logs.Error("查询iOS检测内容数据库出错！！！")
		return false
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(iosDetect[0].DetectContent), &m); err != nil {
		logs.Error("数据库内容转map出错!", err.Error())
		return false
	}
	//更新blasklist
	if ios.ConfirmType == 1 {
		isExit := false
		for _, oneBlack := range m[queryKey].([]interface{}) {
			needConfirm := oneBlack.(map[string]interface{})
			if needConfirm["name"] == ios.ConfirmContent {
				needConfirm["status"] = ios.Status
				needConfirm["confirmer"] = confirmer
				needConfirm["remark"] = ios.Remark
				isExit = true
			}
		}
		if isExit == false {
			return false
		}
	} else if ios.ConfirmType == 2 { //更新method，解决不同class下方法名相同
		isExit := false
		for _, oneBlack := range m[queryKey].([]interface{}) {
			needConfirm := oneBlack.(map[string]interface{})
			var comfirmApi, confirmClass string
			arr := strings.Split(ios.ConfirmContent, "+")
			if arr != nil && len(arr) != 0 {
				comfirmApi = arr[0]
				confirmClass = arr[1]
			}
			if needConfirm["name"] == comfirmApi && needConfirm["content"] == confirmClass {
				needConfirm["status"] = ios.Status
				needConfirm["confirmer"] = confirmer
				needConfirm["remark"] = ios.Remark
				isExit = true
			}
		}
		if isExit == false {
			return false
		}
	} else {
		//权限确认历史表中增加一条记录
		var newHistory dal.PrivacyHistory
		newHistory.AppId = iosDetect[0].AppId
		newHistory.AppName = iosDetect[0].AppName
		newHistory.ConfirmVersion = iosDetect[0].Version
		newHistory.Status = ios.Status
		newHistory.Confirmer = confirmer
		newHistory.ConfirmReason = ios.Remark
		newHistory.Platform = 1
		newHistory.Permission = ios.ConfirmContent
		if err := dal.CreatePrivacyHistoryModel(newHistory); err != nil {
			logs.Error("更新数据库出错！")
			return false
		}
		confirmHistory := dal.QueryPrivacyHistoryModel(map[string]interface{}{
			"app_id":     iosDetect[0].AppId,
			"appname":    iosDetect[0].AppName,
			"permission": ios.ConfirmContent,
			"platform":   1,
		})
		confirmer := (*confirmHistory)[0].Confirmer
		confirmReason := (*confirmHistory)[0].ConfirmReason
		confirmVersion := (*confirmHistory)[len(*confirmHistory)-1].ConfirmVersion
		for _, pp := range m["privacy"].([]interface{}) {
			confirmprivacy := pp.(map[string]interface{})
			if confirmprivacy["permission"] == ios.ConfirmContent {
				confirmprivacy["confirmVersion"] = confirmVersion
				confirmprivacy["confirmReason"] = confirmReason
				confirmprivacy["confirmer"] = confirmer
				confirmprivacy["status"] = ios.Status
			}
		}
	}
	//更新后的内容重新转成json存储到数据库
	confirmedContent, _ := json.Marshal(m)
	flag := dal.UpdateNewIOSDetectModel(iosDetect[0], map[string]interface{}{
		"detect_content": string(confirmedContent),
	})
	if !flag {
		logs.Error("iOS黑名单检测内容确认失败")
		return false
	}
	//取消Lark通知逻辑暂时不改，增量ready后修改
	appId := (*detect)[0].AppId
	appVersion := (*detect)[0].AppVersion
	key := strconv.Itoa(ios.TaskId) + "_" + appId + "_" + appVersion + "_" + strconv.Itoa(ios.ToolId)
	ticker := LARK_MSG_CALL_MAP[key]
	if ticker != nil {
		ticker.(*time.Ticker).Stop()
		delete(LARK_MSG_CALL_MAP, key)
	}
	//更新tb_binary_detect中status值
	if err, _ := changeTotalStatus(ios.TaskId, ios.ToolId, 1); err != nil {
		logs.Error("总任务状态更改出错！", err)
		return false
	}
	return true
}

//判断是否需要更新total status状态值
func changeTotalStatus(taskId, toolId, confirmLark int) (error, int) {

	header := fmt.Sprintf("task id: %v", taskId)
	unconfirmed, _, fail, err := taskDetailiOS(taskId, toolId)
	if err != nil {
		logs.Error("%s get iOS task detail failed: %v", header, err)
		return err, unconfirmed
	}
	if fail <= 0 && unconfirmed > 0 {
		return nil, unconfirmed
	}

	task, err := getExactDetectTask(database.DB(), map[string]interface{}{"id": taskId})
	if err != nil {
		logs.Error("%s get detect task failed: %v", header, err)
		return err, unconfirmed
	}
	if unconfirmed <= 0 {
		task.Status = ConfirmedPass
	}
	if fail > 0 {
		task.Status = ConfirmedFail
		task.DetectNoPass = fail
	}
	if err := dal.UpdateDetectModelNew(*task); err != nil {
		logs.Error("%s update detect task failed: %v", header, err)
		return err, unconfirmed
	}
	StatusDeal(*task, confirmLark) //ci回调和不通过block处理

	return nil, unconfirmed
}

func taskDetailiOS(taskID interface{}, toolID interface{}) (int, int, int, error) {

	header := fmt.Sprintf("task id: %v", taskID)
	unconfirmed := 0
	pass := 0
	fail := 0
	content, err := dal.QueryNewIOSDetectModel(database.DB(), map[string]interface{}{
		"taskId": taskID,
		"toolId": toolID,
	})
	if err != nil {
		logs.Error("%s read iOS detect content failed: %v", header, err)
		return unconfirmed, pass, fail, err
	}
	for i := range content {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(content[i].DetectContent), &m); err != nil {
			logs.Error("%s unmarshal error: %v", header, err)
			return unconfirmed, pass, fail, err
		}
		keyword := content[i].DetectType
		if keyword == "blacklist" {
			keyword = "blackList"
		}
		list, ok := m[keyword].([]interface{})
		if !ok {
			logs.Error("%s cannot assert to []interface{}: %v", header, m[keyword])
			return unconfirmed, pass, fail, fmt.Errorf("%s cannot assert to []interface{}: %v", header, m[keyword])
		}
		for j := range list {
			t, ok := list[j].(map[string]interface{})
			if !ok {
				logs.Error("%s cannot assert to map[string]interface{}: %v", header, list[j])
				return unconfirmed, pass, fail, fmt.Errorf("%s cannot assert to map[string]interface{}: %v", header, list[j])
			}
			switch int(t["status"].(float64)) {
			case ConfirmedPass:
				pass++
			case ConfirmedFail:
				fail++
			default:
				unconfirmed++
			}
		}
	}

	return unconfirmed, pass, fail, nil
}

//返回两个bool值，第一个代表是否是middl数据，第二个代表处理是否成功
func middleDataDeal(taskId, toolId, aId int) (bool, bool) {
	middleData := dal.QueryIOSDetectContent(map[string]interface{}{
		"taskId": taskId,
		"toolId": toolId,
	})
	if middleData == nil || len(*middleData) == 0 {
		logs.Error("没有查询到中间数据！")
		return false, false
	}
	insertFlag, _, _ := iOSResultClassify(taskId, toolId, aId, (*middleData)[0].JsonContent) //插入数据
	//已经确认记得在map中更新数据
	if insertFlag {
		for _, m := range *middleData {
			var updateConfirm IOSConfirm
			var detectType int
			var detectContent string
			switch m.Category {
			case "blacklist":
				detectType = 1
				detectContent = m.CategoryName
			case "method":
				detectType = 2
				detectContent = m.CategoryName + "+" + m.CategoryContent
			}
			updateConfirm.TaskId = taskId
			updateConfirm.ToolId = toolId
			updateConfirm.ConfirmType = detectType
			updateConfirm.ConfirmContent = detectContent
			updateConfirm.Status = m.Status
			updateConfirm.Remark = m.Remark
			if confirmIOSBinaryResult(updateConfirm, m.Confirmer) == false {
				logs.Error("兼容中间数据更新出错！")
				return true, false
			}
		}
	} else {
		logs.Error("兼容旧数据插入出错！")
		return true, false
	}
	return true, true
}

func GetIOSSelfNum(appid, taskId int) (bool, int) {
	url := "https://itc.bytedance.net/api/getSelfCheckItems?taskId=" + strconv.Itoa(taskId) + "&appId=" + strconv.Itoa(appid)
	//url := "http://10.224.14.220:6789/api/getSelfCheckItems?taskId=" + strconv.Itoa(taskId) + "&appId=" + strconv.Itoa(appid)
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error("自查个数获取中构造request出错！", err.Error())
		return false, 0
	}
	itc_token := utils.GetItcToken("yinzhihong")
	reqest.Header.Add("Authorization", itc_token)
	resp, err := client.Do(reqest)
	if err != nil {
		logs.Error("访问iOS自查项返回失败！", err.Error())
		return false, 0
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("获取iOS自查项失败！", err.Error())
		return false, 0
	}
	m := make(map[string]interface{})
	json.Unmarshal(body, &m)
	data := m["data"].([]interface{})
	var selfNum0 = 0
	for _, d := range data {
		if int(d.(map[string]interface{})["status"].(float64)) == 0 {
			selfNum0++
		}
	}
	return true, selfNum0
	//return true,0
}

//全部确认完成后处理
//confirmLark 0:检测完成diff时，1：确认检测结果，2：确认自查结果
func StatusDeal(detect dal.DetectStruct, confirmLark int) error {
	//ci回调
	if detect.Status == 1 && (detect.Platform == 0 || detect.SelfCheckStatus == 1) {
		if err := CICallBack(&detect); err != nil {
			logs.Error("回调ci出错！", err.Error())
			return err
		}
	}
	if detect.Status != 0 && (detect.Platform == 0 || detect.SelfCheckStatus != 0) {
		//diff时调用，不用发冗余消息提醒
		if confirmLark == 0 {
			return nil
		}
		//结果通知
		go func() {
			selfNoPass := detect.SelftNoPass
			detectNoPass := detect.DetectNoPass
			message := "你好，" + detect.AppName + " " + detect.AppVersion
			if detect.Platform == 0 {
				message += " Android包"
			} else {
				message += " iOS包"
			}
			message += "  已经确认完毕！"
			url := "http://rocket.bytedance.net/rocket/itc/task?biz=" + detect.AppId + "&showItcDetail=1&itcTaskId=" + strconv.Itoa(int(detect.ID))
			lark_people := detect.ToLarker
			peoples := strings.Replace(lark_people, "，", ",", -1)
			lark_people_arr := strings.Split(peoples, ",")
			for _, p := range lark_people_arr {
				utils.LarkConfirmResult(strings.TrimSpace(p), message, url, detectNoPass, selfNoPass, false)
			}
			lark_group := detect.ToGroup
			groups := strings.Replace(lark_group, "，", ",", -1)
			lark_group_arr := strings.Split(groups, ",")
			for _, g := range lark_group_arr {
				utils.LarkConfirmResult(strings.TrimSpace(g), message, url, detectNoPass, selfNoPass, true)
			}
		}()
	}
	return nil
}

func sameConfirm(detect dal.DetectStruct) {
	//相同appname、appversion和appid任务结果一致确认
	sameDetect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"app_name":    detect.AppName,
		"app_version": detect.AppVersion,
		"platform":    detect.Platform,
	})
	if len(*sameDetect) != 1 {
		for _, same := range *sameDetect {
			same.SelftNoPass = detect.SelftNoPass
			same.DetectNoPass = detect.DetectNoPass
			same.Status = detect.Status
			same.SelfCheckStatus = detect.SelfCheckStatus
			dal.UpdateDetectModelNew(same)
			StatusDeal(same, 0)
		}
	}
}
