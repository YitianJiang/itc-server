package detect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
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
	var lastDetectContent []dal.IOSNewDetectContent
	if lastTaskId >= 0 {
		var err error
		lastDetectContent, err = readDetectContentiOS(database.DB(),
			map[string]interface{}{"taskId": lastTaskId})
		if err != nil {
			logs.Error("read tb_ios_new_detect_content failed: %v", err)
			return false, warnFlag, 0
		}
	}
	if lastDetectContent == nil || len(lastDetectContent) == 0 {
		logs.Error(strconv.Itoa(lastTaskId) + "没有存储在检测结果中！原因可能为：上一次检测任务没有检测结果，检测工具回调出错！")
	} else {
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
	unRes, err := updateTaskStatusiOS(taskId, toolId, 0)
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
	iosTaskBinaryCheckContent, err := readDetectContentiOS(database.DB(),
		map[string]interface{}{"taskId": taskId, "toolId": toolId})
	if err != nil {
		logs.Warn("read tb_ios_new_detect_content failed: %v", err)
	}
	//如果数据在新库中查不到就去访问老库
	// if iosTaskBinaryCheckContent == nil || len(iosTaskBinaryCheckContent) == 0 {
	// 	//中间数据处理
	// 	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
	// 		"id": taskId,
	// 	})
	// 	aId, _ := strconv.Atoi((*detect)[0].AppId)
	// 	sId, _ := strconv.Atoi(taskId)
	// 	tId, _ := strconv.Atoi(toolId)
	// 	middleFlag, dealFlag := middleDataDeal(sId, tId, aId)
	// 	if middleFlag {
	// 		//如果中间数据处理失败
	// 		if dealFlag == false {
	// 			c.JSON(http.StatusOK, gin.H{
	// 				"message":   "中间数据处理失败！",
	// 				"errorCode": 0,
	// 				"data":      "中间数据处理失败！",
	// 			})
	// 			return
	// 		} else {
	// 			//中间数据处理成功后，重新读取一次新库
	// 			var err error
	// 			iosTaskBinaryCheckContent, err = readDetectContentiOS(database.DB(),
	// 				map[string]interface{}{"taskId": taskId, "toolId": toolId})
	// 			if err != nil {
	// 				logs.Error("read tb_ios_new_detect_content failed: %v", err)
	// 				return
	// 			}
	// 		}
	// 	} else {
	// 		//不再中间库，去老库中寻找
	// 		conditionOld := "task_id='" + taskId + "' and tool_id='" + toolId + "'"
	// 		content := dal.QueryTaskBinaryCheckContent(conditionOld)
	// 		if content == nil || len(*content) == 0 {
	// 			logs.Info("未查询到检测内容")
	// 			c.JSON(http.StatusOK, gin.H{
	// 				"message":   "未查询到检测内容",
	// 				"errorCode": -3,
	// 				"data":      "未查询到检测内容",
	// 			})
	// 			return
	// 		} else {
	// 			c.JSON(http.StatusOK, gin.H{
	// 				"message":   "success",
	// 				"errorCode": 0,
	// 				"data":      (*content)[0],
	// 			})
	// 			return
	// 		}
	// 	}
	// }
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

	username, exist := c.Get("username")
	if !exist {
		utils.ReturnMsg(c, http.StatusUnauthorized, utils.FAILURE, "unauthorized user")
		return
	}
	var ios IOSConfirm
	if err := c.ShouldBindJSON(&ios); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid user: %v", err))
		return
	}

	task, err := getExactDetectTask(database.DB(), map[string]interface{}{"id": ios.TaskId})
	if err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("read tb_binary_detect failed: %v", err))
		return
	}
	itemName := ios.ConfirmContent
	var itemType *string
	switch ios.ConfirmType {
	case 1:
		itemType = &TypeString
	case 2:
		itemType = &TypeMethod
		i := strings.Index(itemName, "+")
		itemName = itemName[i+1:] + delimiter + itemName[:i]
	case 3:
		itemType = &TypePermission
	}
	if err := preAutoConfirmTask(task, &Item{
		Name: itemName,
		Type: itemType},
		ios.Status, username.(string), ios.Remark, 0, ios.ToolId); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("confirm iOS detection failed: %v", err))
		return
	}

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success")
}

func readDetectContentiOS(db *gorm.DB, sieve map[string]interface{}) ([]dal.IOSNewDetectContent, error) {

	var result []dal.IOSNewDetectContent
	if err := db.Debug().Where(sieve).Find(&result).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return result, nil
}

//判断是否需要更新total status状态值
func updateTaskStatusiOS(taskId, toolId interface{}, confirmLark int) (int, error) {
	var newChangeFlag = true
	var unconfirmedCount = 0
	var confirmedFailCount = 0 //确认不通过数目
	iosDetectAll, err := readDetectContentiOS(database.DB(), map[string]interface{}{
		"taskId": taskId,
		"toolId": toolId,
	})
	if err != nil {
		logs.Error("read tb_ios_new_detect_content failed: %v", err)
		return unconfirmedCount, err
	}
	for _, oneDetect := range iosDetectAll {
		var im map[string]interface{}
		if err := json.Unmarshal([]byte(oneDetect.DetectContent), &im); err != nil {
			logs.Error("unmarshal error: %v", err)
			return unconfirmedCount, err
		}
		newQueryKey := oneDetect.DetectType
		if newQueryKey == "blacklist" {
			newQueryKey = "blackList"
		}
		a := im[newQueryKey].([]interface{})
		for _, oneBlack := range a {
			needConfirm := oneBlack.(map[string]interface{})
			if int(needConfirm["status"].(float64)) == Unconfirmed {
				newChangeFlag = false
				unconfirmedCount++
			} else if int(needConfirm["status"].(float64)) == ConfirmedFail {
				confirmedFailCount++
			}
		}
	}
	//检测项全部确认，更改任务状态
	if newChangeFlag {
		detect := dal.QueryDetectModelsByMap(map[string]interface{}{
			"id": taskId,
		})
		if confirmedFailCount == 0 {
			(*detect)[0].Status = 1 //1代表全部确认且确认通过
		} else {
			(*detect)[0].Status = 2 //2代表全部确认且有确认不通过
		}
		(*detect)[0].DetectNoPass = confirmedFailCount //不通过总数
		err := dal.UpdateDetectModelNew((*detect)[0])
		if err != nil {
			logs.Error("task id: %v update tb_binary_detect failed: %v", taskId, err)
			return unconfirmedCount, err
		}
		StatusDeal((*detect)[0], confirmLark) //ci回调和不通过block处理
		sameConfirm((*detect)[0])             //相同包检测结果确认
	}

	return unconfirmedCount, nil
}

// //返回两个bool值，第一个代表是否是middl数据，第二个代表处理是否成功
// func middleDataDeal(taskId, toolId, aId int) (bool, bool) {
// 	middleData := dal.QueryIOSDetectContent(map[string]interface{}{
// 		"taskId": taskId,
// 		"toolId": toolId,
// 	})
// 	if middleData == nil || len(*middleData) == 0 {
// 		logs.Error("没有查询到中间数据！")
// 		return false, false
// 	}
// 	insertFlag, _, _ := iOSResultClassify(taskId, toolId, aId, (*middleData)[0].JsonContent) //插入数据
// 	//已经确认记得在map中更新数据
// 	if insertFlag {
// 		for _, m := range *middleData {
// 			var updateConfirm IOSConfirm
// 			var detectType int
// 			var detectContent string
// 			switch m.Category {
// 			case "blacklist":
// 				detectType = 1
// 				detectContent = m.CategoryName
// 			case "method":
// 				detectType = 2
// 				detectContent = m.CategoryName + "+" + m.CategoryContent
// 			}
// 			updateConfirm.TaskId = taskId
// 			updateConfirm.ToolId = toolId
// 			updateConfirm.ConfirmType = detectType
// 			updateConfirm.ConfirmContent = detectContent
// 			updateConfirm.Status = m.Status
// 			updateConfirm.Remark = m.Remark
// 			// if confirmIOSBinaryResult(updateConfirm, m.Confirmer) == false {
// 			if confirmIOSBinaryResult(&updateConfirm, m.Confirmer) == false {
// 				logs.Error("兼容中间数据更新出错！")
// 				return true, false
// 			}
// 		}
// 	} else {
// 		logs.Error("兼容旧数据插入出错！")
// 		return true, false
// 	}
// 	return true, true
// }

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
