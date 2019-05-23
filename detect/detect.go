package detect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"


	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

const (
	DETECT_URL_DEV = "10.2.209.202:9527"
	DETECT_URL_PRO = "10.2.9.226:9527"

	//目前apk检测接口
	//TEST_DETECT_URL = "http://10.2.9.226:9527/apk_post/v2"
	//调试，暂时换成本机的ip地址和端口
	//DETECT_URL_PRO = "10.2.196.119:9527"
)


/**
 *安卓检测数据查询返回结构
 */
type DetectQueryStruct struct {
	ApkName				string							`json:"apkName"`
	Version				string							`json:"version"`
	Channel             string							`json:"channel"`
	Permissions			string 							`json:"permissions"`
	SMethods		    []SMethod						`json:"sMethods"`
	SStrs				[]SStr							`json:"sStrs"`
}

type SMethod struct {
	Id					uint 				`json:"id"`
	Status				int					`json:"status"`
	Remark				string 				`json:"remark"`
	Confirmer			string				`json:"confirmer"`
	MethodName			string				`json:"methodName"`
	ClassName			string				`json:"className"`
	Desc				string				`json:"desc"`
	CallLoc				[]MethodCallJson	`json:"callLoc"`
}
type MethodCallJson struct {
	MethodName			string				`json:"method_name"`
	ClassName			string				`json:"class_name"`
	LineNumber			interface{}			`json:"line_number"`
}

type SStr struct {
	Id					uint 				`json:"id"`
	Status				int					`json:"status"`
	Remark				string 				`json:"remark"`
	Confirmer			string				`json:"confirmer"`
	Keys				string				`json:"keys"`
	Desc				string				`json:"desc"`
	CallLoc				[]StrCallJson		`json:"callLoc"`
}

type StrCallJson struct {
	Key					string				`json:"key"`
	MethodName			string				`json:"method_name"`
	ClassName			string				`json:"class_name"`
	LineNumber			interface{}			`json:"line_number"`
}



var LARK_MSG_CALL_MAP = make(map[string]interface{})


/**
 *新建检测任务更新---------fj
 */
func UploadFile(c *gin.Context) {

	url := ""
	nameI, f := c.Get("username")
	if !f {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未获取到用户信息！",
			"errorCode": -1,
			"data":      "未获取到用户信息！",
		})
		return
	}

	file, header, err := c.Request.FormFile("uploadFile")
	if file == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未选择上传的文件！",
			"errorCode": -2,
		})
		logs.Error("未选择上传的文件！")
		return
	}
	defer file.Close()
	filename := header.Filename

	toLarker := c.DefaultPostForm("toLarker","")
	var name string
	if toLarker == "" {
		name = nameI.(string)
	}else {
		name = toLarker
	}

	platform := c.DefaultPostForm("platform", "")
	if platform == "" {
		logs.Error("缺少platform参数！")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少platform参数！",
			"errorCode": -3,
		})
		return
	}
	appId := c.DefaultPostForm("appId", "")
	if appId == "" {
		logs.Error("缺少appId参数！")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少appId参数！",
			"errorCode": -4,
		})
		return
	}
	checkItem := c.DefaultPostForm("checkItem", "")
	logs.Info("checkItem: ", checkItem)


	//检验文件格式是否是apk或者ipa
	flag := strings.HasSuffix(filename, ".apk") || strings.HasSuffix(filename, ".ipa") ||
		strings.HasSuffix(filename, ".aab")
	if !flag {
		errorFormatFile(c)
		return
	}
	//检验文件与platform是否匹配，0-安卓apk，1-iOS ipa
	if platform == "0" {
		flag := strings.HasSuffix(filename, ".apk") || strings.HasSuffix(filename, ".aab")
		if !flag {
			errorFormatFile(c)
			return
		}
		if checkItem != "6" {
			//旧服务url
			url = "http://" + DETECT_URL_PRO + "/apk_post"
		}else {
			//新服务url
			url = "http://" + DETECT_URL_PRO +"/apk_post/v2"
		}
		//url = "http://" + DETECT_URL_PRO + "/apk_post"

	} else if platform == "1"{
		flag := strings.HasSuffix(filename, ".ipa")
		if !flag {
			errorFormatFile(c)
			return
		}
		url = "http://" + DETECT_URL_PRO + "/ipa_post/v2"
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "platform参数不合法！",
			"errorCode": -5,
		})
		logs.Error("platform参数不合法！")
		return
	}
	_tmpDir := "./tmp"
	exist, err := PathExists(_tmpDir)
	if !exist {
		os.Mkdir(_tmpDir, os.ModePerm)
	}
	out, err := os.Create(_tmpDir + "/" + filename)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "安装包文件处理失败，请联系相关人员！",
			"errorCode": -6,
		})
		logs.Fatal("临时文件保存失败")
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "安装包文件处理失败，请联系相关人员！",
			"errorCode": -6,
		})
		logs.Fatal("临时文件保存失败")
		return
	}
	//调试，暂时注释
	recipients := name + "@bytedance.com"
	filepath := _tmpDir + "/" + filename
	//1、上传至tos,测试暂时注释
	//tosUrl, err := upload2Tos(filepath)
	//2、将相关信息保存至数据库
	var dbDetectModel dal.DetectStruct
	dbDetectModel.Creator = nameI.(string)
	dbDetectModel.ToLarker = name
	dbDetectModel.SelfCheckStatus = 0
	dbDetectModel.CreatedAt = time.Now()
	dbDetectModel.UpdatedAt = time.Now()
	dbDetectModel.Platform, _ = strconv.Atoi(platform)
	dbDetectModel.AppId = appId
	//增加状态字段，0---未完全确认；1---已完全确认
	dbDetectModel.Status = 0
	dbDetectModelId := dal.InsertDetectModel(dbDetectModel)
	//3、调用检测接口，进行二进制检测 && 删掉本地临时文件
	if checkItem == "" {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未选择二进制检测工具，请直接进行自查",
			"errorCode": 0,
			"data":      "未选择二进制检测工具，请直接进行自查",
		})
		return
	}
	//go upload2Tos(filepath, dbDetectModelId)
	go func() {
		callBackUrl := "https://itc.bytedance.net/updateDetectInfos"
		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)
		bodyWriter.WriteField("recipients", recipients)
		bodyWriter.WriteField("callback", callBackUrl)
		bodyWriter.WriteField("taskID", fmt.Sprint(dbDetectModelId))
		bodyWriter.WriteField("toolIds", checkItem)
		fileWriter, err := bodyWriter.CreateFormFile("file", filepath)
		if err != nil {
			logs.Error("%s", "error writing to buffer: "+err.Error())
			c.JSON(http.StatusOK, gin.H{
				"message":   "二进制包处理错误，请联系相关人员！",
				"errorCode": -1,
				"data":      "二进制包处理错误，请联系相关人员！",
			})
			return
		}
		fileHandler, err := os.Open(filepath)
		defer fileHandler.Close()
		_, err = io.Copy(fileWriter, fileHandler)
		contentType := bodyWriter.FormDataContentType()
		err = bodyWriter.Close()
		logs.Info("url: ", url)
		toolHttp := &http.Client{
			Timeout: 300 * time.Second,
		}
		response, err := toolHttp.Post(url, contentType, bodyBuffer)
		if err != nil {
			logs.Error("上传二进制包出错，将重试一次: ", err)
			response, err = toolHttp.Post(url, contentType, bodyBuffer)
		}
		if err != nil {
			logs.Error("上传二进制包出错，重试一次也失败", err)
			//及时报警
			utils.LarkDingOneInner("kanghuaisong", "二进制包检测服务无响应，请及时进行检查！")
		}
		resBody := &bytes.Buffer{}
		if response != nil {
			defer response.Body.Close()
			_, err = resBody.ReadFrom(response.Body)
			var data map[string]interface{}
			data = make(map[string]interface{})
			json.Unmarshal(resBody.Bytes(), &data)
			logs.Info("upload detect url's response: %+v", data)
			//删掉临时文件
			os.Remove(filepath)
		}
	}()
	c.JSON(http.StatusOK, gin.H{
		"message":   "文件上传成功，请等待检测结果通知",
		"errorCode": 0,
		"data":      "文件上传成功，请等待检测结果通知",
	})
}

/**
 *更新检测包的检测信息_v2——----------fj
 */
func UpdateDetectInfos(c *gin.Context) {
	taskId := c.Request.FormValue("task_ID")
	if taskId == "" {
		logs.Error("缺少task_ID参数")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少task_ID参数",
			"errorCode": -1,
			"data":      "缺少task_ID参数",
		})
		return
	}
	toolId := c.Request.FormValue("tool_ID")
	jsonContent := c.Request.FormValue("jsonContent")

	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id": taskId,
	})
	if (*detect) == nil {
		logs.Error("未查询到该taskid对应的检测任务，%v", taskId)
		c.JSON(http.StatusOK, gin.H{
			"message":   "未查询到该taskid对应的检测任务",
			"errorCode": -2,
			"data":      "未查询到该taskid对应的检测任务",
		})
		return
	}
	toolIdInt,_ := strconv.Atoi(toolId)

	if (*detect)[0].Platform == 0 {
		if toolIdInt == 6 {//安卓兼容新版本
			//安卓检测信息分析，并将检测信息写库-----fj
			mapInfo := make(map[string]int)
			mapInfo["taskId"],_ = strconv.Atoi(taskId)
			mapInfo["toolId"],_ = strconv.Atoi(toolId)
			jsonInfoAnalysis(jsonContent,mapInfo)
		}else{
			appName := c.Request.FormValue("appName")
			appVersion := c.Request.FormValue("appVersion")
			htmlContent := c.Request.FormValue("content")
			var detectContent dal.DetectContent
			detectContent.TaskId, _ = strconv.Atoi(taskId)
			detectContent.ToolId, _ = strconv.Atoi(toolId)
			detectContent.HtmlContent = htmlContent
			detectContent.JsonContent = jsonContent
			detectContent.CreatedAt = time.Now()
			detectContent.UpdatedAt = time.Now()
			(*detect)[0].AppName = appName
			(*detect)[0].AppVersion = appVersion
			(*detect)[0].UpdatedAt = time.Now()
			if err := dal.UpdateDetectModel((*detect)[0], detectContent); err != nil {
				c.JSON(http.StatusOK, gin.H{
					"message" : "数据库更新检测信息失败",
					"errorCode" : -3,
					"data" : "数据库更新检测信息失败",
				})
				return
			}
		}
	}
	//ios新检测内容存储
	if (*detect)[0].Platform == 1 {
		htmlContent := c.Request.FormValue("content")
		task_id, _ := strconv.Atoi(taskId)
		tool_id, _ := strconv.Atoi(toolId)
		condition := map[string]interface{}{
			"taskId":      task_id,
			"toolId":      tool_id,
			"htmlContent": htmlContent,
		}
		if iOSResultClassify(condition, jsonContent) == false {
			logs.Error("iOS 新增new detect content失败！！！") //防止影响现有用户，出错后暂不return
		}
	}
	//进行lark消息提醒
	var message string
	creator := (*detect)[0].ToLarker
	message = creator + "你好，" + (*detect)[0].AppName + " " + (*detect)[0].AppVersion
	platform := (*detect)[0].Platform
	if platform == 0 {
		message += "安卓包"
	} else {
		message += "iOS包"
	}
	message += "完成二进制检测，请及时进行确认！可在结果展示页面底部进行确认，确认后不会再有消息提醒！\n"
	message += "如果安卓选择了GooglePlay检测和隐私检测，两个检测结果都需要进行确认，请不要遗漏！！！\n"
	appId := (*detect)[0].AppId
	appIdInt, _ := strconv.Atoi(appId)
	appVersion := (*detect)[0].AppVersion
	var config *dal.LarkMsgTimer
	config = dal.QueryLarkMsgTimerByAppId(appIdInt)
	alterType := 0
	var interval int
	if config == nil { //如果未进行消息提醒设置，则默认10分钟提醒一次
		logs.Info("采用默认10分钟频率进行提醒")
		alterType = 1
		interval = 10
	} else {
		logs.Info("采用设置的频率进行提醒")
		alterType = config.Type
		interval = config.MsgInterval
	}
	var ticker *time.Ticker
	var duration time.Duration
	switch alterType {
	case 0:
		logs.Info("提醒方式为秒")
		duration = time.Duration(interval) * time.Second
	case 1:
		logs.Info("提醒方式为分钟")
		duration = time.Duration(interval) * time.Minute
	case 2:
		logs.Info("提醒方式为小时")
		duration = time.Duration(interval) * time.Hour
	case 3:
		logs.Info("提醒方式为天")
		duration = time.Duration(interval) * time.Duration(24) * time.Hour
	default:
		logs.Info("提醒方式为分钟")
		duration = 10 * time.Minute
	}
	ticker = time.NewTicker(duration)
	var key string
	key = taskId + "_" + appId + "_" + appVersion + "_" + toolId
	LARK_MSG_CALL_MAP[key] = ticker
	//此处测试时注释掉
	larkUrl := "http://rocket.bytedance.net/rocket/itc/task?biz="+ appId + "&showItcDetail=1&itcTaskId=" + taskId
	message += "地址链接：" + larkUrl
	utils.LarkDingOneInner(creator, message)
	if config != nil {
		timerId := config.ID
		condition := "timer_id='" + fmt.Sprint(timerId) + "' and platform='" + strconv.Itoa(platform) + "'"
		groups := dal.QueryLarkGroupByCondition(condition)
		if groups != nil && len(*groups) > 0 {
			for i := 0; i < len(*groups); i++ {
				g := (*groups)[i]
				utils.LarkGroup(message, g.GroupId)
			}
		}
	}
	//go alertLarkMsgCronNew(*ticker, creator, message, taskId, toolId)
}


/**
json检测信息分析------fj
 */
func jsonInfoAnalysis(info string,mapInfo map[string]int){
	var infoMap = make(map[string]interface{})
	json.Unmarshal([]byte(info),&infoMap)
	appInfos := infoMap["app_info"].(map[string]interface{})
	methodsInfo := infoMap["method_sensitive_infos"].([]interface{})
	strsInfo := infoMap["str_sensitive_infos"].([]interface{})
	var detectInfo dal.DetectInfo
	detectInfo.TaskId = mapInfo["taskId"]
	detectInfo.ToolId = mapInfo["toolId"]
	appInfoAnalysis(appInfos,&detectInfo)

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
	for _,newMethod := range newMethods {
		var detailContent dal.DetectContentDetail
		detailContent.TaskId = mapInfo["taskId"]
		detailContent.ToolId = mapInfo["toolId"]
		methodAnalysis(newMethod,&detailContent)
	}


	for _, strInfoi := range strsInfo {
		strInfo := strInfoi.(map[string]interface{})
		var detailContent dal.DetectContentDetail
		detailContent.TaskId = mapInfo["taskId"]
		detailContent.ToolId = mapInfo["toolId"]
		strAnalysis(strInfo,&detailContent)
	}
	return
}

/**
appInfo解析，并写入数据库-------fj
 */
func appInfoAnalysis(info map[string]interface{},detectInfo *dal.DetectInfo)  {
	taskId := detectInfo.TaskId
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : taskId,
	})
	(*detect)[0].AppName = info["apk_name"].(string)
	(*detect)[0].AppVersion = info["apk_version_name"].(string)
	(*detect)[0].UpdatedAt = time.Now()
	if err := dal.UpdateDetectModelNew((*detect)[0]); err != nil {
		logs.Error("任务id:%s信息更新失败，%v",taskId,err)
		return
	}
	detectInfo.ApkName = info["apk_name"].(string)
	detectInfo.Version = info["apk_version_name"].(string)
	detectInfo.Channel = info["channel"].(string)
	permissionArr := info["permissions"].([]interface{})
	//permissionArr := permissionArr1.([]string)

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
func methodAnalysis(method map[string]interface{},detail *dal.DetectContentDetail)  {
	detail.SensiType = 1
	detail.Status = 0

	detail.Key = method["method_name"].(string)
	detail.Desc = method["desc"].(string)
	detail.ClassName = method["method_class_name"].(string)
	call := method["call_location"].([]interface{})

	callLocation := methodRmRepeat(call)
	//for _,loc1 := range call {
	//	//var loc MethodCallJson
	//	loc := loc1.(map[string]interface{})
	//	mapLoc,_ := json.Marshal(loc)
	//	callLocation += string(mapLoc)+";"
	//}
	detail.CallLoc = callLocation

	err := dal.InsertDetectDetail(*detail)
	if err != nil {
		//及时报警
		message := "敏感method写入数据库失败，请解决;"+fmt.Sprint(err)+"\n敏感方法名："+fmt.Sprint(detail.Key)
		utils.LarkDingOneInner("fanjuan.xqp", message)
	}
	return
}

/**
apk敏感方法内容去重--------fj
*/
func methodRmRepeat(callInfo []interface{}) string  {
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
func strAnalysis(str map[string]interface{},detail *dal.DetectContentDetail)  {
	detail.SensiType = 2
	detail.Status = 0

	keys := str["keys"].([]interface{})
	key := ""
	for _,ks1 := range keys {
		ks := ks1.(string)
		key += ks +";"
	}
	detail.Key = key
	detail.Desc = str["desc"].(string)

	callInfo := str["call_location"].([]interface{})
	//敏感字段信息去重
	call_location := strRmRepeat(callInfo)
	detail.CallLoc = call_location

	err := dal.InsertDetectDetail(*detail)
	if err != nil {
		//及时报警
		message := "敏感str写入数据库失败，请解决;"+fmt.Sprint(err)+"\n敏感方法名："+fmt.Sprint(key)
		utils.LarkDingOneInner("fanjuan.xqp",message)
	}
	return

}

/**
apk敏感字符串去重--------fj
 */
func strRmRepeat(callInfo []interface{}) string {
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
 *iOS 检测结果jsonContent处理
 */
func iOSResultClassify(condition map[string]interface{}, jsonContent string) bool {
	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &dat); err != nil {
		logs.Error("json转map出错！", err.Error())
		return false
	}
	//黑名单处理
	blackContent := dat["blacklist_in_app"]
	condition["category"] = "blacklist"
	for k, v := range blackContent.(map[string]interface{}) {
		if len(v.([]interface{})) != 0 {
			temRes := ""
			for _, s := range v.([]interface{}) {
				temRes += s.(string) + "###"
			}
			condition["categoryName"] = k
			condition["categoryContent"] = temRes
			fmt.Println(condition)
			var newDetectContent dal.IOSDetectContent
			if err := mapstructure.Decode(condition, &newDetectContent); err != nil {
				logs.Error("map转struct出错！", err.Error())
				return false
			}
			if err := dal.CreateIOSDetectModel(newDetectContent); err != nil {
				logs.Error("数据库中新增new detect content 出错！！！", err.Error())
				return false
			}
		}
	}
	//可疑方法名处理
	methodContent := dat["methods_in_app"]
	condition["category"] = "method"
	for _, temMethod := range methodContent.([]interface{}) {
		susApi := temMethod.(map[string]interface{})["api_name"].(string)
		susClass := temMethod.(map[string]interface{})["class_name"].(string)
		condition["categoryName"] = susApi
		condition["categoryContent"] = susClass
		var newDetectContent dal.IOSDetectContent
		if err := mapstructure.Decode(condition, &newDetectContent); err != nil {
			logs.Error("map转struct出错！", err.Error())
			return false
		}
		if err := dal.CreateIOSDetectModel(newDetectContent); err != nil {
			logs.Error("数据库中新增new detect content 出错！！！", err.Error())
			return false
		}
	}
	return true
}

/**
 *lark消息定时提醒
 */
func alertLarkMsgCron(ticker time.Ticker, receiver string, msg string, taskId string, toolId string) {
	flag := false
	logs.Info("taskId: " + taskId + ", toolId: " + toolId)
	if taskId == "" || toolId == "" {
		return
	}
	for _ = range ticker.C {
		condition := "task_id='" + taskId + "' and tool_id='" + toolId + "'"
		binaryTool := dal.QueryTaskBinaryCheckContent(condition)
		logs.Info("每次提醒前进行提醒检查")
		if *binaryTool != nil && len(*binaryTool) > 0 {
			dc := (*binaryTool)[0]
			status := dc.Status
			if status == 0 {
				utils.LarkDingOneInner(receiver, msg)
				logs.Info("调试，先以打印输出代替lark通知 ")
			} else {
				flag = true
				break
			}
		}
	}
	if flag {
		logs.Info("stop interval lark call")
		ticker.Stop()
	}
}


/**
 *lark消息定时提醒,根据任务中的status提醒--------fj
 */
func alertLarkMsgCronNew(ticker time.Ticker, receiver string, msg string, taskId string, toolId string){
	flag := false
	logs.Info("taskId: " + taskId + ", toolId: " + toolId)
	if taskId=="" || toolId=="" {
		return
	}
	for _ = range ticker.C {
		detect := dal.QueryDetectModelsByMap(map[string]interface{}{
			"id" : taskId,
		})
		logs.Info("每次提醒前进行提醒检查")
		if *detect != nil && len(*detect)>0{
			if (*detect)[0].Status == 0{
				utils.LarkDingOneInner(receiver, msg)
				logs.Info("调试，先以打印输出代替lark通知 ")
			} else {
				flag = true
				break
			}
		}
	}
	if flag {
		logs.Info("stop interval lark call")
		ticker.Stop()
	}
}

/**
 *确认二进制包检测结果，更新数据库，并停止lark消息
 */
func ConfirmBinaryResult(c *gin.Context) {
	type confirm struct {
		TaskId  int		`json:"taskId"`
		ToolId	int		`json:"toolId"`
		Remark  string	`json:"remark"`
		Status  int		`json:"status"`
	}//测试兼容性增加
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t confirm
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("wrong params %v",err)
		c.JSON(http.StatusOK, gin.H{
			"message":   "参数不合法！",
			"errorCode": -1,
			"data":      "参数不合法！",
		})
		return
	}
	//获取确认人信息
	username, _ := c.Get("username")
	var data map[string]string
	data = make(map[string]string)
	data["task_id"] = strconv.Itoa(t.TaskId)
	data["tool_id"] = strconv.Itoa(t.ToolId)
	data["confirmer"] = username.(string)
	data["remark"] = t.Remark
	data["status"] = strconv.Itoa(t.Status)
	flag := dal.ConfirmBinaryResult(data)
	if !flag {
		logs.Error("二进制检测内容确认失败")
		c.JSON(http.StatusOK, gin.H{
			"message":   "二进制检测内容确认失败！",
			"errorCode": -1,
			"data":      "二进制检测内容确认失败！",
		})
		return
	}
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id": t.TaskId,
	})
	appId := (*detect)[0].AppId
	appVersion := (*detect)[0].AppVersion
	key := strconv.Itoa(t.TaskId) + "_" + appId + "_" + appVersion + "_" + strconv.Itoa(t.ToolId)
	ticker := LARK_MSG_CALL_MAP[key]
	if ticker != nil {
		ticker.(*time.Ticker).Stop()
		delete(LARK_MSG_CALL_MAP, key)
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      "success",
	})


}


/**
 *确认安卓二进制包检测结果，更新数据库，并判断是否停止lark消息--------fj
 */
func ConfirmApkBinaryResult(c *gin.Context){
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

	//切换到旧版本
	//if (t.ToolId != 6){
	//	c.Request.Body.Close()
	//	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(param))
	//	ConfirmBinaryResult(c)
	//	return
	//}
	//获取确认人信息
	username, _ := c.Get("username")
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

	//改变任务确认状态
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
 *将安装包上传至tos
 */
func upload2Tos(path string, taskId uint) (string, error) {

	var tosBucket = tos.WithAuth(_const.TOS_BUCKET_NAME, _const.TOS_BUCKET_KEY)
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tosPutClient, err := tos.NewTos(tosBucket)
	fileName := filepath.Base(path)
	byte, err := ioutil.ReadFile(path)
	if err != nil {
		logs.Error("%s", "打开文件失败"+err.Error())
	}
	key := fmt.Sprint(time.Now().UnixNano()) + "_" + fileName
	logs.Info("key: " + key)
	err = tosPutClient.PutObject(context, key, int64(len(byte)), bytes.NewBuffer(byte))
	if err != nil {
		logs.Error("%s", "上传tos失败："+err.Error())
	}
	domains := tos.GetDomainsForLargeFile("TT", path)
	domain := domains[rand.Intn(len(domains)-1)]
	domain = "tosv.byted.org/obj/" + _const.TOS_BUCKET_NAME
	var returnUrl string
	returnUrl = "https://" + domain + "/" + key
	dal.UpdateDetectTosUrl(returnUrl, taskId)
	return returnUrl, nil
}

/**
 * test upload tos
 */
func UploadTos(c *gin.Context) {
	data := ""
	path := "/home/kanghuaisong/test.py"
	var tosBucket = tos.WithAuth(_const.TOS_BUCKET_NAME, _const.TOS_BUCKET_KEY)
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tosPutClient, err := tos.NewTos(tosBucket)
	fileName := filepath.Base(path)
	byte, err := ioutil.ReadFile(path)
	if err != nil {
		logs.Error("%s", "打开文件失败"+err.Error())
		data = "打开文件失败"
	}
	key := fmt.Sprint(time.Now().UnixNano()) + "_" + fileName
	logs.Info("key: " + key)
	err = tosPutClient.PutObject(context, key, int64(len(byte)), bytes.NewBuffer(byte))
	if err != nil {
		logs.Error("%s", "上传tos失败："+err.Error())
		data = "上传tos失败：" + err.Error()
	}
	domains := tos.GetDomainsForLargeFile("TT", path)
	domain := domains[rand.Intn(len(domains)-1)]
	domain = "tosv.byted.org/obj/" + _const.TOS_BUCKET_NAME
	var returnUrl string
	returnUrl = "https://" + domain + "/" + key
	logs.Info("returnUrl: " + returnUrl)
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      data,
	})
}

/**
 *判断路径是否存在
 */
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, nil
}

/*
 *错误格式化信息
 */
func errorFormatFile(c *gin.Context) {
	logs.Infof("文件格式不正确，请上传正确的文件！")
	c.JSON(http.StatusOK, gin.H{
		"message":   "文件格式不正确，请上传正确的文件！",
		"errorCode": -4,
		"data":      "文件格式不正确，请上传正确的文件！",
	})
}

/*
 *查询检测任务列表
 */
func QueryDetectTasks(c *gin.Context) {

	appId := c.DefaultQuery("appId", "")
	version := c.DefaultQuery("version", "")
	creator := c.DefaultQuery("user", "")
	pageNo := c.DefaultQuery("page", "")
	//如果缺少pageSize参数，则选用默认每页显示10条数据
	pageSize := c.DefaultQuery("pageSize", "10")
	//参数校验
	if pageNo == "" {
		logs.Error("缺少page参数！")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少page参数！",
			"errorCode": -1,
			"data":      "缺少page参数！",
		})
		return
	}
	condition := "1=1"
	_, err := strconv.Atoi(appId)
	if err != nil {
		logs.Error("appId参数不合法！")
		c.JSON(http.StatusOK, gin.H{
			"message":   "appId参数不合法！",
			"errorCode": -2,
			"data":      "appId参数不合法！",
		})
		return
	}
	if appId != "" {
		condition += " and app_id='" + appId + "'"
	}
	if version != "" {
		condition += " and app_version='" + version + "'"
	}
	if creator != "" {
		condition += " and creator='" + creator + "'"
	}
	var param map[string]interface{}
	param = make(map[string]interface{})
	if condition != "" {
		param["condition"] = condition
	}
	page, _ := strconv.Atoi(pageNo)
	size, _ := strconv.Atoi(pageSize)
	param["pageNo"] = page
	param["pageSize"] = size
	var data dal.RetDetectTasks
	var more uint
	items, total := dal.QueryTasksByCondition(param)
	if uint(page*size) >= total {
		more = 0
	} else {
		more = 1
	}
	data.GetMore = more
	data.Total = total
	data.NowPage = uint(page)
	data.Tasks = *items
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      data,
	})
}

/**
 *根据二进制工具列表
 */
func QueryDetectTools(c *gin.Context) {

	name := c.DefaultQuery("name", "")
	condition := "1=1"
	if name != "" {
		condition += " and name like '%" + name + "%'"
	}
	tools := dal.QueryBinaryToolsByCondition(condition)
	if tools == nil {
		logs.Error("二进制检测工具列表查询失败")
		c.JSON(http.StatusOK, gin.H{
			"message":   "二进制检测工具列表查询失败",
			"errorCode": -1,
			"data":      "二进制检测工具列表查询失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      tools,
	})
}

/**
 *查询检测任务选择的二进制检查工具
 */
func QueryTaskQueryTools(c *gin.Context) {

	taskId := c.DefaultQuery("taskId", "")
	if taskId == "" {
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少taskId参数",
			"errorCode": -1,
			"data":      "缺少taskId参数",
		})
		return
	}
	task := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id": taskId,
	})
	if task == nil || len(*task) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未查询到该taskId对应的检测任务",
			"errorCode": -2,
			"data":      "未查询到该taskId对应的检测任务",
		})
		return
	}
	platform := (*task)[0].Platform
	condition := "task_id='" + taskId + "'"
	toolsContent := dal.QueryTaskBinaryCheckContent(condition)
	//scanner检测工具查询内容
	toolsContent_2,_ := dal.QueryDetectInfo(condition)
	if (toolsContent == nil || len(*toolsContent) == 0 )&& (toolsContent_2 == nil ) {
		logs.Error("未查询到该检测任务对应的二进制检测结果")
		var res [0]dal.DetectContent
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data":      res,
		})
		return
	}
	toolCondition := "id in("
	if toolsContent_2 != nil{
		toolCondition += "'"+fmt.Sprint((*toolsContent_2).ToolId)+"')"
	}else{
		for i:=0; i<len(*toolsContent); i++ {
			content := (*toolsContent)[i]
			if i==len(*toolsContent)-1 {
				toolCondition += "'" + fmt.Sprint(content.ToolId) + "')"
			} else {
				toolCondition += "'" + fmt.Sprint(content.ToolId) + "',"
			}
		}
	}

	toolCondition += " and platform ='" + strconv.Itoa(platform) + "'"
	selected := dal.QueryBinaryToolsByCondition(toolCondition)
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      *selected,
	})
}

/**
 *查询二进制检查结果信息
 */
func QueryTaskBinaryCheckContent(c *gin.Context) {
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
	condition := "task_id='" + taskId + "' and tool_id='" + toolId + "'"
	content := dal.QueryTaskBinaryCheckContent(condition)
	if content == nil || len(*content) == 0 {
		logs.Info("未查询到检测内容")
		c.JSON(http.StatusOK, gin.H{
			"message":   "未查询到检测内容",
			"errorCode": -3,
			"data":      "未查询到检测内容",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      (*content)[0],
	})
	return
}


/**
 *安卓查询二进制检查结果信息-------fj
 */
func QueryTaskApkBinaryCheckContent(c *gin.Context){
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
	//if toolId != "6"{
	//	QueryTaskBinaryCheckContent(c)
	//	return
	//}
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
		logs.Info("未查询到对应任务的检测内容")
		c.JSON(http.StatusOK, gin.H{
			"message" : "未查询到对应任务的检测内容",
			"errorCode" : -3,
			"data" : "未查询到对应任务的检测内容",
		})
		return
	}

	var queryResult DetectQueryStruct
	//queryResult.TaskId = (*content).TaskId
	//queryResult.ToolId = (*content).ToolId
	queryResult.Channel = (*content).Channel
	queryResult.ApkName = (*content).ApkName
	queryResult.Version = (*content).Version

	permission := ""
	perms := strings.Split((*content).Permissions,";")
	for _,perm := range perms[0:(len(perms)-1)]{
		permission += perm +"\n"
	}
	queryResult.Permissions = permission

	methods := make([]SMethod,0)
	strs := make([]SStr,0)

	for _,detail := range (*details) {
		if detail.SensiType == 1 {
			var method SMethod
			method.ClassName = detail.ClassName
			method.Desc = detail.Desc
			method.Status = detail.Status
			method.Id = detail.ID
			method.Confirmer = detail.Confirmer
			method.Remark = detail.Remark
			method.MethodName = detail.Key
			callLocs := strings.Split(detail.CallLoc,";")
			callLoc :=make([]MethodCallJson,0)
			for _,call_loc := range callLocs[0:(len(callLocs)-1)] {
				var call_loc_json MethodCallJson
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
			methods = append(methods,method)
		}else{
			var str SStr
			str.Keys = detail.Key
			str.Remark = detail.Remark
			str.Confirmer = detail.Confirmer
			str.Status = detail.Status
			str.Desc = detail.Desc
			str.Id = detail.ID
			callLocs := strings.Split(detail.CallLoc,";")
			callLoc := make([]StrCallJson,0)
			for _,call_loc := range callLocs[0:(len(callLocs)-1)] {
				var callLoc_json StrCallJson
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
				callLoc = append(callLoc,callLoc_json)
			}
			str.CallLoc = callLoc
			strs = append(strs,str)
		}
	}
	queryResult.SMethods = methods
	queryResult.SStrs = strs

	logs.Info("query detect result success!")
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : queryResult,
	})
	return
}

/**
 *获取token接口
 */
func GetToken(c *gin.Context) {
	var jwtSecret = []byte("itc_jwt_secret")
	username := c.DefaultQuery("username", "")
	if username == "" {
		logs.Error("未获取到username")
		c.JSON(http.StatusOK, gin.H{
			"message":   "未获取到username",
			"errorCode": -1,
			"data":      "未获取到username",
		})
		return
	}
	claim := jwt.MapClaims{
		"name": username,
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	token, err := tokenClaims.SignedString(jwtSecret)
	if err != nil {
		logs.Error("生成token失败")
		c.JSON(http.StatusOK, gin.H{
			"message":   "生成token失败",
			"errorCode": -2,
			"data":      "生成token失败",
		})
	} else {
		logs.Error("生成token成功")
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data":      token,
		})
	}
}


/**
 *确认安卓二进制包检测结果，更新数据库（包括确认信息入库），并判断是否停止lark消息--------fj
 */
func ConfirmApkBinaryResultv_2(c *gin.Context){
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

	//切换到旧版本
	if (t.ToolId != 6){
		c.Request.Body.Close()
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(param))
		ConfirmBinaryResult(c)
		return
	}

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

	//增量忽略结果录入
	if t.Status == 1 {
		senType := (*detailInfo)[0].SensiType
		if senType == 1 {
			var igInfo dal.IgnoreInfoStruct
			igInfo.Platform = (*detect)[0].Platform
			igInfo.AppId,_ = strconv.Atoi((*detect)[0].AppId)
			igInfo.SensiType = (*detailInfo)[0].SensiType
			igInfo.Keys = (*detailInfo)[0].ClassName+"."+(*detailInfo)[0].Key
			err := dal.InsertIgnoredInfo(igInfo)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"message" : "可忽略信息更新失败！",
					"errorCode" : -1,
					"data" : "可忽略信息更新失败！",
				})
				return
			}
		}else{
			keys := strings.Split((*detailInfo)[0].Key,";")
			for _,key := range keys[0:len(keys)-1] {
				var igInfos dal.IgnoreInfoStruct
				igInfos.SensiType = 2
				igInfos.Keys = key
				igInfos.AppId,_ = strconv.Atoi((*detect)[0].AppId)
				igInfos.Platform = (*detect)[0].Platform
				err := dal.InsertIgnoredInfo(igInfos)
				if err != nil {
					c.JSON(http.StatusOK, gin.H{
						"message" : "可忽略信息更新失败！",
						"errorCode" : -1,
						"data" : "可忽略信息更新失败！",
					})
					return
				}
			}
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
func QueryTaskApkBinaryCheckContentWithIgnorance(c *gin.Context){
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
	var methodIgs = make(map[string]int)
	var strIgs = make(map[string]int)
	var errIg error
	if queryType == ""{
		flag = true
		//获取任务信息
		detect := dal.QueryDetectModelsByMap(map[string]interface{}{
			"id" : taskId,
		})
		if (*detect) == nil {
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

		//此处的逻辑需要再看一下，如果可忽略信息没有的话如何处理
		methodIgs,strIgs,errIg = getIgnoredInfo(queryData)
		if errIg != nil {
			c.JSON(http.StatusOK, gin.H{
				"message" : "可忽略信息数据库查询失败",
				"errorCode" : -1,
				"data" : errIg,
			})
			return
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
	var queryResult DetectQueryStruct
	queryResult.Channel = (*content).Channel
	queryResult.ApkName = (*content).ApkName
	queryResult.Version = (*content).Version

	permission := ""
	perms := strings.Split((*content).Permissions,";")
	for _,perm := range perms[0:(len(perms)-1)]{
		permission += perm +"\n"
	}
	queryResult.Permissions = permission

	methods := make([]SMethod,0)
	strs := make([]SStr,0)

	for _,detail := range (*details) {
		if detail.SensiType == 1 {
			if flag && methodIgs != nil {
				if v,ok := methodIgs[detail.ClassName+"."+detail.Key]; ok&&v==1{
					continue
				}
			}
			var method SMethod
			method.ClassName = detail.ClassName
			method.Desc = detail.Desc
			method.Status = detail.Status
			method.Id = detail.ID
			method.Confirmer = detail.Confirmer
			method.Remark = detail.Remark
			method.MethodName = detail.Key
			callLocs := strings.Split(detail.CallLoc,";")
			callLoc :=make([]MethodCallJson,0)
			for _,call_loc := range callLocs[0:(len(callLocs)-1)] {
				var call_loc_json MethodCallJson
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
			methods = append(methods,method)
		}else{
			keys2 := make(map[string]int)
			var keys3 = detail.Key
			if flag && strIgs != nil{
				keys := strings.Split(detail.Key,";")

				keys3 = ""
				for _,keyInfo := range keys[0:len(keys)-1] {
					if v,ok := strIgs[keyInfo]; ok&&v==1{
						keys2[keyInfo] = 1
					}else{
						keys3 += keyInfo+";"
					}
				}
				if keys3 =="" {
					continue
				}
			}
			var str SStr
			str.Keys = keys3
			str.Remark = detail.Remark
			str.Confirmer = detail.Confirmer
			str.Status = detail.Status
			str.Desc = detail.Desc
			str.Id = detail.ID
			callLocs := strings.Split(detail.CallLoc,";")
			callLoc := make([]StrCallJson,0)
			for _,call_loc := range callLocs[0:(len(callLocs)-1)] {
				var callLoc_json StrCallJson
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
			strs = append(strs,str)
		}
	}
	queryResult.SMethods = methods
	queryResult.SStrs = strs

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
func getIgnoredInfo(data map[string]string) (map[string]int,map[string]int,error) {
	condition := "app_id ="+data["appId"]+" and platform = "+data["platform"]
	result,err := dal.QueryIgnoredInfo(condition)
	if err != nil {
		return nil,nil,err
	}
	if result == nil || len(*result)==0{
		return nil,nil,nil
	}
	methodMap := make(map[string]int)
	strMap := make(map[string]int)
	for i:=0;i<len(*result);i++{
		if (*result)[i].SensiType == 1{
			methodMap[(*result)[i].Keys]=1
		}else{
			strMap[(*result)[i].Keys]=1
		}
	}
	return methodMap,strMap,nil
}
