package detect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	//test----fj
	Local_URL_PRO = "10.2.221.213:9527"

	//目前apk检测接口
	//TEST_DETECT_URL = "http://10.2.9.226:9527/apk_post/v2"
	//调试，暂时换成本机的ip地址和端口
	//DETECT_URL_PRO = "10.2.196.119:9527"
)

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
	//发送lark消息到个人
	toLarker := c.DefaultPostForm("toLarker", "")
	var name string
	if toLarker == "" {
		name = nameI.(string)
	} else {
		name = nameI.(string) + "," + toLarker
	}
	//发送lark消息到群
	toGroup := c.DefaultPostForm("toLarkGroupId", "")

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

	//增加任务来源判断
	//sourceStr := c.DefaultPostForm("source","")
	//var source int
	//if sourceStr == "" {
	//	source = 0
	//}else {
	//	source = 1
	//}

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
		} else {
			//新服务url
			//url = "http://"+Local_URL_PRO +"/apk_post/v2"
			url = "http://" + DETECT_URL_PRO + "/apk_post/v2"
		}
		//url = "http://" + DETECT_URL_PRO + "/apk_post"

	} else if platform == "1" {
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
	recipients := name
	filepath := _tmpDir + "/" + filename
	//1、上传至tos,测试暂时注释
	//tosUrl, err := upload2Tos(filepath)
	//2、将相关信息保存至数据库
	var dbDetectModel dal.DetectStruct
	dbDetectModel.Creator = nameI.(string)
	dbDetectModel.ToLarker = name
	dbDetectModel.ToGroup = toGroup
	dbDetectModel.SelfCheckStatus = 0
	dbDetectModel.CreatedAt = time.Now()
	dbDetectModel.UpdatedAt = time.Now()
	dbDetectModel.Platform, _ = strconv.Atoi(platform)
	dbDetectModel.AppId = appId
	//增加状态字段，0---未完全确认；1---已完全确认
	dbDetectModel.Status = 0
	//dbDetectModel.Source = source
	dbDetectModelId := dal.InsertDetectModel(dbDetectModel)
	//3、调用检测接口，进行二进制检测 && 删掉本地临时文件
	if checkItem == "" {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未选择二进制检测工具，请直接进行自查",
			"errorCode": -1,
			"data":      "未选择二进制检测工具，请直接进行自查",
		})
		return
	}
	if dbDetectModelId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":   "数据库插入记录失败，请确认数据库字段与结构体定义一致",
			"errorCode": -1,
			"data":      "数据库插入记录失败，请确认数据库字段与结构体定义一致",
		})
		return
	}
	//go upload2Tos(filepath, dbDetectModelId)
	go func() {
		callBackUrl := "https://itc.bytedance.net/updateDetectInfos"
		//callBackUrl := "http://10.224.14.220:6789/updateDetectInfos"
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
			utils.LarkDingOneInner("kanghuaisong", "二进制包检测服务无响应，请及时进行检查！任务ID："+fmt.Sprint(dbDetectModelId)+",创建人："+dbDetectModel.Creator)
			utils.LarkDingOneInner("yinzhihong", "二进制包检测服务无响应，请及时进行检查！任务ID："+fmt.Sprint(dbDetectModelId)+",创建人："+dbDetectModel.Creator)
			utils.LarkDingOneInner("fanjuan.xqp", "二进制包检测服务无响应，请及时进行检查！任务ID："+fmt.Sprint(dbDetectModelId)+",创建人："+dbDetectModel.Creator)
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
		"data": map[string]interface{}{
			"taskId": dbDetectModelId,
		},
	})
}

/**
 *更新检测包的检测信息_v2——----------fj
 */
func UpdateDetectInfos(c *gin.Context) {
	logs.Info("回调开始，更新检测信息～～～")
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
	appName := c.Request.FormValue("appName")
	appVersion := c.Request.FormValue("appVersion")
	htmlContent := c.Request.FormValue("content")

	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id": taskId,
	})
	if detect == nil || len(*detect) == 0 {
		logs.Error("未查询到该taskid对应的检测任务，%v", taskId)
		c.JSON(http.StatusOK, gin.H{
			"message":   "未查询到该taskid对应的检测任务",
			"errorCode": -2,
			"data":      "未查询到该taskid对应的检测任务",
		})
		return
	}
	toolIdInt, _ := strconv.Atoi(toolId)

	if (*detect)[0].Platform == 0 {
		if toolIdInt == 6 { //安卓兼容新版本
			//安卓检测信息分析，并将检测信息写库-----fj
			mapInfo := make(map[string]int)
			mapInfo["taskId"], _ = strconv.Atoi(taskId)
			mapInfo["toolId"], _ = strconv.Atoi(toolId)
			ApkJsonAnalysis(jsonContent, mapInfo)
		} else {
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
					"message":   "数据库更新检测信息失败",
					"errorCode": -3,
					"data":      "数据库更新检测信息失败",
				})
				return
			}
		}
	}
	//ios新检测内容存储
	if (*detect)[0].Platform == 1 {
		task_id, _ := strconv.Atoi(taskId)
		tool_id, _ := strconv.Atoi(toolId)
		//旧表更新
		var detectContent dal.DetectContent
		detectContent.TaskId = task_id
		detectContent.ToolId = tool_id
		detectContent.HtmlContent = htmlContent
		detectContent.JsonContent = jsonContent
		detectContent.CreatedAt = time.Now()
		detectContent.UpdatedAt = time.Now()
		(*detect)[0].AppName = appName
		(*detect)[0].AppVersion = appVersion
		(*detect)[0].UpdatedAt = time.Now()
		if err := dal.UpdateDetectModel((*detect)[0], detectContent); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message":   "数据库更新检测信息失败",
				"errorCode": -3,
				"data":      "数据库更新检测信息失败",
			})
			return
		}
		//新表jsonContent分类存储
		taskId, _ := strconv.Atoi(taskId)
		toolId, _ := strconv.Atoi(toolId)
		appId, _ := strconv.Atoi((*detect)[0].AppId)
		res, warnFlag := iOSResultClassify(taskId, toolId, appId, jsonContent) //检测结果处理
		if res == false {
			logs.Error("iOS 新增new detect content失败！！！") //防止影响现有用户，出错后暂不return
		}
		//iOS付费相关黑名单及时报警
		if res && warnFlag {
			tips := "Notice: " + (*detect)[0].AppName + " " + (*detect)[0].AppVersion + " iOS包完成二进制检测，检测黑名单中itms-services不为空，请及时关注！！！！\n"
			larkUrl := "http://rocket.bytedance.net/rocket/itc/task?biz=" + strconv.Itoa(toolId) + "&showItcDetail=1&itcTaskId=" + strconv.Itoa(taskId)
			tips += "地址链接：" + larkUrl
			utils.LarkDingOneInner("zhangshuai.02", tips)
			utils.LarkDingOneInner("gongrui", tips)
			utils.LarkDingOneInner("kanghuaisong", tips)
			utils.LarkDingOneInner((*detect)[0].ToLarker, tips)
		}
	}

	//进行lark消息提醒
	var message string
	creators := (*detect)[0].ToLarker
	larkList := strings.Split(creators, ",")
	//for _,creator := range larkList {
	message = "你好，" + (*detect)[0].AppName + " " + (*detect)[0].AppVersion
	platform := (*detect)[0].Platform
	if platform == 0 {
		message += "安卓包"
	} else {
		message += "iOS包"
	}
	message += "完成二进制检测，请及时对每条未确认信息进行确认！\n"
	//message += "如果安卓选择了GooglePlay检测和隐私检测，两个检测结果都需要进行确认，请不要遗漏！！！\n"
	appId := (*detect)[0].AppId
	appIdInt, _ := strconv.Atoi(appId)
	//appVersion := (*detect)[0].AppVersion
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
	larkUrl := "http://rocket.bytedance.net/rocket/itc/task?biz=" + appId + "&showItcDetail=1&itcTaskId=" + taskId
	message += "地址链接：" + larkUrl
	for _, creator := range larkList {
		utils.LarkDingOneInner(creator, message)
	}
	//给群ID对应群发送消息
	toGroupID := (*detect)[0].ToGroup
	if toGroupID != "" {
		group := strings.Replace(toGroupID, "，", ",", -1) //中文逗号切换成英文逗号
		groupArr := strings.Split(group, ",")
		for _, group_id := range groupArr {
			to_lark_group := strings.Trim(group_id, " ")
			utils.LarkGroup(message, to_lark_group)
		}
	}

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
func alertLarkMsgCronNew(ticker time.Ticker, receiver string, msg string, taskId string, toolId string) {
	flag := false
	logs.Info("taskId: " + taskId + ", toolId: " + toolId)
	if taskId == "" || toolId == "" {
		return
	}
	for _ = range ticker.C {
		detect := dal.QueryDetectModelsByMap(map[string]interface{}{
			"id": taskId,
		})
		logs.Info("每次提醒前进行提醒检查")
		if *detect != nil && len(*detect) > 0 {
			if (*detect)[0].Status == 0 {
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
		TaskId int    `json:"taskId"`
		ToolId int    `json:"toolId"`
		Remark string `json:"remark"`
		Status int    `json:"status"`
	} //测试兼容性增加
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t confirm
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("wrong params %v", err)
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
	//更新旧接口任务状态
	condition := "task_id = '" + fmt.Sprint(t.TaskId) + "'"
	detectContent := dal.QueryTaskBinaryCheckContent(condition)
	if detectContent == nil || len(*detectContent) == 0 {
		logs.Error("未查询到相关二进制检测内容,更新任务状态失败")
		c.JSON(http.StatusOK, gin.H{
			"message":   "未查询到相关二进制检测内容,更新任务状态失败！",
			"errorCode": -1,
			"data":      "未查询到相关二进制检测内容,更新任务状态失败",
		})
		return
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
				logs.Error("更新任务状态失败，任务ID："+fmt.Sprint(t.TaskId)+",错误原因:%v", err)
				c.JSON(http.StatusOK, gin.H{
					"message":   "更新任务状态失败！",
					"errorCode": -1,
					"data":      "更新任务状态失败",
				})
				return
			}
		}
	}
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
	toolsContent_2, _ := dal.QueryDetectInfo(condition)
	if (toolsContent == nil || len(*toolsContent) == 0) && (toolsContent_2 == nil) {
		logs.Error("未查询到该检测任务对应的二进制检测结果")
		var res [0]dal.DetectContent
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"appId": (*task)[0].AppId,
			"data":      res,
		})
		return
	}
	toolCondition := "id in("
	if toolsContent_2 != nil {
		toolCondition += "'" + fmt.Sprint((*toolsContent_2).ToolId) + "')"
	} else {
		for i := 0; i < len(*toolsContent); i++ {
			content := (*toolsContent)[i]
			if i == len(*toolsContent)-1 {
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
		"appId": (*task)[0].AppId,
		"data":      *selected,
	})
}

/**
 *查询二进制检查结果信息(旧)
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
 * 检测服务报警接口
 */
func Alram(c *gin.Context) {
	message := c.Request.FormValue("errorMsg")
	larkList := strings.Split("kanghuaisong,yinzhihong,fanjuan.xqp", ",")
	for _, creator := range larkList {
		utils.LarkDingOneInner(creator, "检测服务异常，请立即关注！"+message)
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
	})
}
