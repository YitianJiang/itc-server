package detect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

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
func iOSResultClassify(task *dal.DetectStruct, toolId int, jsonContent *string) (bool, bool, int) {

	header := fmt.Sprintf("task id: %v", task.ID)
	warnFlag := false
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(*jsonContent), &dat); err != nil {
		logs.Error("json转map出错！", err.Error())
		return false, warnFlag, 0
	}
	appName := dat["name"].(string)
	version := dat["version"].(string)
	bundleId := dat["bundle_id"].(string)
	minVersion := dat["min_version"].(string)
	sdkVersion := dat["tar_version"].(string)

	task.AppName = appName
	task.AppVersion = version
	if err := database.UpdateDBRecord(database.DB(), task); err != nil {
		logs.Error("%s update detect task failed: %v", header, err)
		return false, warnFlag, 0
	}
	appID, _ := strconv.Atoi(task.AppId)
	var permissions []string
	var sensitiveMethods []string
	var sensitiveStrings []string
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
			blackMap["origin_version"] = task.AppVersion
			blackMap["content"] = v
			blackMap["status"] = 0
			blackMap["confirmer"] = ""
			blackMap["remark"] = ""
			// }
			sensitiveStrings = append(sensitiveStrings, k)
			blackList = append(blackList, blackMap)
			if k == "itms-services" {
				warnFlag = true
			}
		}
		BlackContentValue, err := json.Marshal(map[string]interface{}{
			"blackList": blackList,
		})
		if err != nil {
			logs.Error("%s marshal error: %v", header, err)
			return false, warnFlag, 0
		}
		blackDetect.DetectContent = string(BlackContentValue)
		blackDetect.DetectType = "blacklist"
		blackDetect.ToolId = toolId
		blackDetect.TaskId = int(task.ID)
		blackDetect.AppId = appID
		blackDetect.AppName = appName
		blackDetect.Version = version
		blackDetect.BundleId = bundleId
		blackDetect.MinVersion = minVersion
		blackDetect.SdkVersion = sdkVersion
		blackDetect.JsonContent = *jsonContent
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
			methodMap["origin_version"] = task.AppVersion
			methodMap["status"] = 0
			methodMap["confirmer"] = ""
			methodMap["remark"] = ""
			sensitiveMethods = append(sensitiveMethods, susClass+delimiter+susApi)
			methodList = append(methodList, methodMap)
		}
		methodContentValue, err := json.Marshal(map[string]interface{}{
			"method": methodList,
		})
		if err != nil {
			logs.Error("%s marshal error: %v", header, err)
			return false, warnFlag, 0
		}
		methodDetect.DetectContent = string(methodContentValue)
		methodDetect.DetectType = "method"
		methodDetect.ToolId = toolId
		methodDetect.TaskId = int(task.ID)
		methodDetect.AppId = appID
		methodDetect.AppName = appName
		methodDetect.Version = version
		methodDetect.BundleId = bundleId
		methodDetect.MinVersion = minVersion
		methodDetect.SdkVersion = sdkVersion
		methodDetect.JsonContent = *jsonContent
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
			privacyMap["origin_version"] = task.AppVersion
			privacyMap["confirmer"] = ""
			privacyMap["confirmReason"] = ""
			privacyMap["status"] = 0
			permissions = append(permissions, e)
			privacyList = append(privacyList, privacyMap)
		}
		privacyContentValue, err := json.Marshal(map[string]interface{}{
			"privacy": privacyList,
		})
		if err != nil {
			logs.Error("%s marshal error: %v", header, err)
			return false, warnFlag, 0
		}
		privacyDetect.DetectContent = string(privacyContentValue)
		privacyDetect.DetectType = "privacy"
		privacyDetect.ToolId = toolId
		privacyDetect.TaskId = int(task.ID)
		privacyDetect.AppId = appID
		privacyDetect.AppName = appName
		privacyDetect.Version = version
		privacyDetect.BundleId = bundleId
		privacyDetect.MinVersion = minVersion
		privacyDetect.SdkVersion = sdkVersion
		privacyDetect.JsonContent = *jsonContent
	}
	insertFlag := dal.InsertNewIOSDetect(blackDetect, methodDetect, privacyDetect)
	//更新tb_binary_detect中status值
	sync := make(chan struct{}, 1)
	go func() {
		autoConfirmCallBack(task, permissions, sensitiveMethods, sensitiveStrings)
		sync <- struct{}{}
	}()
	<-sync
	unRes, err := updateTaskStatus(int(task.ID), toolId, platformiOS, 0)
	if err != nil {
		logs.Error("update iOS task status failed: %v", err)
		return false, warnFlag, 0
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
	iosTaskBinaryCheckContent, err := readDetectContentiOS(database.DB(),
		map[string]interface{}{"taskId": taskId, "toolId": toolId})
	if err != nil {
		logs.Warn("read tb_ios_new_detect_content failed: %v", err)
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
			if v == nil {
				continue
			}
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

func readDetectContentiOS(db *gorm.DB, sieve map[string]interface{}) ([]dal.IOSNewDetectContent, error) {

	var result []dal.IOSNewDetectContent
	if err := db.Debug().Where(sieve).Find(&result).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return result, nil
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
