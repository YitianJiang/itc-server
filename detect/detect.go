package detect

import (
	"bytes"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const(
	DETECT_URL_DEV = "10.2.209.202:9527"
	//DETECT_URL_PRO = "10.2.9.226:9527"
	//调试，暂时换成本机的ip地址和端口
	DETECT_URL_PRO = "10.2.196.119:9527"
)
var LARK_MSG_CALL_MAP map[string]interface{} = make(map[string]interface{})
func UploadFile(c *gin.Context){

	url := ""
	//get user info from cas
	//name := cas.Username(c.Request)
	name := "kanghuaisong"
	if name == ""{
		c.JSON(http.StatusOK, gin.H{
			"message":"用户未登录！",
			"errorCode":-1,
		})
		return
	}
	file, header, err := c.Request.FormFile("uploadFile")
	if file == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":"未选择上传的文件！",
			"errorCode":-2,
		})
		logs.Error("未选择上传的文件！")
		return
	}
	defer file.Close()
	filename := header.Filename
	platform := c.DefaultPostForm("platform", "")
	if platform == ""{
		c.JSON(http.StatusOK, gin.H{
			"message":"缺少platform参数！",
			"errorCode":-3,
		})
		logs.Error("缺少platform参数！")
		return
	}
	//检验文件格式是否是apk或者ipa
	flag := strings.HasSuffix(filename, ".apk") || strings.HasSuffix(filename, ".ipa")
	if !flag{
		errorFormatFile(c)
		return
	}
	//检验文件与platform是否匹配，0-安卓apk，1-iOS ipa
	if platform == "0"{
		flag := strings.HasSuffix(filename, ".apk")
		if !flag{
			errorFormatFile(c)
			return
		}
		url = "http://" + DETECT_URL_PRO + "/apk_post"
	} else if platform == "1"{
		flag := strings.HasSuffix(filename, ".ipa")
		if !flag{
			errorFormatFile(c)
			return
		}
		url = "http://" + DETECT_URL_PRO + "/ipa_post/v2"
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":"platform参数不合法！",
			"errorCode":-5,
		})
		logs.Error("platform参数不合法！")
		return
	}
	_tmpDir := "./tmp"
	exist, err := PathExists(_tmpDir)
	if !exist{
		os.Mkdir(_tmpDir, os.ModePerm)
	}
	out, err := os.Create(_tmpDir + "/" + filename)
	if err != nil{
		c.JSON(http.StatusOK, gin.H{
			"message":"安装包文件处理失败，请联系相关人员！",
			"errorCode":-6,
		})
		logs.Fatal("临时文件保存失败")
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":"安装包文件处理失败，请联系相关人员！",
			"errorCode":-6,
		})
		logs.Fatal("临时文件保存失败")
		return
	}
	//调试，暂时注释
	//var recipients = "ttqaall@bytedance.com,tt_ios@bytedance.com,"
	var recipients = "kanghuaisong@bytedance.com"
	//recipients += name + "@bytedance.com"
	filepath := _tmpDir + "/" + filename
	//1、上传至tos,测试暂时注释
	//tosUrl, err := upload2Tos(filepath)
	//2、将相关信息保存至数据库
	var dbDetectModel dal.DetectStruct
	dbDetectModel.Creator = name
	dbDetectModel.SelfCheckStatus = 0
	dbDetectModel.CreatedAt = time.Now()
	dbDetectModel.UpdatedAt = time.Now()
	dbDetectModel.Platform, _ = strconv.Atoi(platform)
	//dbDetectModel.TosUrl = tosUrl
	dbDetectModelId := dal.InsertDetectModel(dbDetectModel)
	//3、调用检测接口，进行二进制检测 && 删掉本地临时文件
	go func() {
		callBackUrl := "http://10.224.10.61:6789/updateDetectInfos"
		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)
		bodyWriter.WriteField("recipients", recipients)
		bodyWriter.WriteField("callback", callBackUrl)
		bodyWriter.WriteField("taskID", fmt.Sprint(dbDetectModelId))
		fileWriter, err := bodyWriter.CreateFormFile("file", filepath)
		if err != nil {
			logs.Error("%s", "error writing to buffer: " + err.Error())
			c.JSON(http.StatusOK, gin.H{
				"message" : "二进制包处理错误，请联系相关人员！",
				"errorCode" : -1,
				"data" : "二进制包处理错误，请联系相关人员！",
			})
			return
		}
		filehandler, err := os.Open(filepath)
		defer filehandler.Close()
		_, err = io.Copy(fileWriter, filehandler)
		contentType := bodyWriter.FormDataContentType()
		bodyWriter.Close()
		response, err := http.Post(url, contentType, bodyBuffer)
		defer response.Body.Close()
		resBody := &bytes.Buffer{}
		_, err = resBody.ReadFrom(response.Body)
		var data map[string]interface{}
		data = make(map[string]interface{})
		json.Unmarshal(resBody.Bytes(), &data)
		//删掉临时文件
		os.Remove(filepath)
		c.JSON(http.StatusOK, gin.H{
			"message" : data["msg"],
			"errorCode" : data["success"],
			"data" : data["msg"],
		})
	}()
	c.JSON(http.StatusOK, gin.H{
		"message" : "文件上传成功，请等待检测结果通知",
		"errorCode" : 0,
		"data" : "文件上传成功，请等待检测结果通知",
	})
}

//更新检测包的检测信息
func UpdateDetectInfos(c *gin.Context){

	taskId := c.Request.FormValue("task_ID")
	if taskId == "" {
		logs.Error("缺少task_ID参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少task_ID参数",
			"errorCode" : -1,
			"data" : "缺少task_ID参数",
		})
		return
	}
	appName := c.Request.FormValue("appName")
	appVersion := c.Request.FormValue("appVersion")
	htmlContent := c.Request.FormValue("content")
	jsonContent := c.Request.FormValue("jsonContent")
	toolId := c.Request.FormValue("tool_ID")
	var detectContent dal.DetectContent
	detectContent.TaskId, _ = strconv.Atoi(taskId)
	detectContent.ToolId, _ = strconv.Atoi(toolId)
	detectContent.HtmlContent = htmlContent
	detectContent.JsonContent = jsonContent
	detectContent.CreatedAt = time.Now()
	detectContent.UpdatedAt = time.Now()
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
	//进行lark消息提醒
	var message string
	creator := (*detect)[0].Creator
	message = creator + "你好，" + (*detect)[0].AppName + " " + (*detect)[0].AppVersion
	platform := (*detect)[0].Platform
	if platform == 0 {
		message += "安卓包"
	} else {
		message += "iOS包"
	}
	message += "完成二进制检测，请及时进行确认！"
	appId := (*detect)[0].AppId
	appIdInt, _ := strconv.Atoi(appId)
	var config *dal.LarkMsgTimer
	config = dal.QueryLarkMsgTimerByAppId(appIdInt)
	alterType := 0
	var interval int
	if config == nil {//如果未进行消息提醒设置，则默认10分钟提醒一次
		alterType = 1
		interval = 10
	} else {
		alterType = config.Type
		interval = config.MsgInterval
	}
	var ticker *time.Ticker
	var duration time.Duration
	switch alterType {
	case 0:
		duration = time.Duration(interval) * time.Second
	case 1:
		duration = time.Duration(interval) * time.Minute
	case 2:
		duration = time.Duration(interval) * time.Hour
	case 3:
		duration = time.Duration(interval) * time.Duration(24) * time.Hour
	default:
		duration = 10 * time.Minute
	}
	ticker = time.NewTicker(duration)
	var key string
	key = taskId + "_" + appId + "_" + appVersion + "_" + toolId
	LARK_MSG_CALL_MAP[key] = ticker
	//utils.LarkDingOneInnerV2(creator, message)
	go alertLarkMsgCron(*ticker, creator, message)
}
func alertLarkMsgCron(ticker time.Ticker, receiver string, msg string){
	for _ = range ticker.C {
		//utils.LarkDingOneInnerV2(receiver, msg)
		logs.Info("调试，先以打印输出代替lark通知 ")
	}
}
//确认二进制包检测结果，更新数据库，并停止lark消息
func ConfirmBinaryResult(c *gin.Context){
	type confirm struct {
		TaskId  int		`json:"taskId"`
		toolId	int		`json:"toolId"`
	}
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t confirm
	err := json.Unmarshal(param, &t)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message" : "参数不合法！",
			"errorCode" : -1,
			"data" : "参数不合法！",
		})
		return
	}
	var data map[string]string
	data = make(map[string]string)
	data["task_id"] = strconv.Itoa(t.TaskId)
	data["tool_id"] = strconv.Itoa(t.toolId)
	flag := dal.ConfirmBinaryResult(data)
	if !flag {
		logs.Error("二进制检测内容确认失败")
		c.JSON(http.StatusOK, gin.H{
			"message" : "二进制检测内容确认失败！",
			"errorCode" : -1,
			"data" : "二进制检测内容确认失败！",
		})
		return
	}
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : t.TaskId,
	})
	appId := (*detect)[0].AppId
	appVersion := (*detect)[0].AppVersion
	key := strconv.Itoa(t.TaskId) + "_" + appId + "_" + appVersion + "_" + strconv.Itoa(t.toolId)
	ticker := LARK_MSG_CALL_MAP[key]
	if ticker != nil {
		ticker.(*time.Ticker).Stop()
		delete(LARK_MSG_CALL_MAP, key)
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
}
//将安装包上传至tos
func upload2Tos(path string) (string, error){

	var tosBucket = tos.WithAuth("", "")
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tosPutClient, err := tos.NewTos(tosBucket)
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		logs.Error("%s", "打开文件失败" + err.Error())
		return "打开文件失败", err
	}
	stat, err := file.Stat()
	if err != nil {
		logs.Error("%s", "获取文件大小失败：" + err.Error())
		return "获取文件大小失败", err
	}
	err = tosPutClient.PutObject(context, path, stat.Size(), file)
	if err != nil {
		logs.Error("%s", "上传tos失败：" + err.Error())
		return "上传tos失败", err
	}
	domains := tos.GetDomainsForLargeFile("TT", path)
	domain := domains[rand.Intn(len(domains)-1)]
	domain = "tosv.byted.org/obj/" + "itcserver"
	var returnUrl string
	returnUrl = "https://" + domain + "/" + path
	return returnUrl, nil
}

func PathExists(path string)(bool, error){
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err){
		return false, nil
	}
	return false, nil
}

func errorFormatFile(c *gin.Context){
	logs.Infof("文件格式不正确，请上传正确的文件！")
	c.JSON(http.StatusOK, gin.H{
		"message" : "文件格式不正确，请上传正确的文件！",
		"errorCode" : -4,
		"data" : "文件格式不正确，请上传正确的文件！",
	})
}
//查询检测任务列表
func QueryDetectTasks(c *gin.Context){

	appId := c.DefaultQuery("appId", "")
	version := c.DefaultQuery("version", "")
	creator := c.DefaultQuery("creator", "")
	pageNo := c.DefaultQuery("page", "")
	//如果缺少pageSize参数，则选用默认每页显示10条数据
	pageSize := c.DefaultQuery("pageSize", "10")
	//参数校验
	if pageNo == "" {
		logs.Error("缺少page参数！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少page参数！",
			"errorCode" : -1,
			"data" : "缺少page参数！",
		})
		return
	}
	condition := "1=1"
	_, err := strconv.Atoi(appId)
	if err != nil {
		logs.Error("appId参数不合法！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "appId参数不合法！",
			"errorCode" : -2,
			"data" : "appId参数不合法！",
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
	}else{
		more = 1
	}
	data.GetMore = more
	data.Total = total
	data.NowPage = uint(page)
	data.Tasks = *items
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : data,
	})
}
func QueryDetectTools(c *gin.Context){

	/*platform := c.DefaultQuery("platform", "")
	if platform == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少platform参数",
			"errorCode" : -1,
			"data" : "缺少platform参数",
		})
		return
	}
	if platform != "0" && platform != "1" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "platform参数不合法",
			"errorCode" : -2,
			"data" : "platform参数不合法",
		})
		return
	}
	condition := "platform='" + platform + "'"*/
	name := c.DefaultQuery("name", "")
	condition := "1=1"
	if name != "" {
		condition += " and name like '%" + name + "%'"
	}
	tools := dal.QueryBinaryToolsByCondition(condition)
	if tools == nil {
		logs.Error("二进制检测工具列表查询失败")
		c.JSON(http.StatusOK, gin.H{
			"message" : "二进制检测工具列表查询失败",
			"errorCode" : -1,
			"data" : "二进制检测工具列表查询失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : tools,
	})
}

func QueryTaskQueryTools(c *gin.Context){

	taskId := c.DefaultQuery("taskId", "")
	if taskId == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少taskId参数",
			"errorCode" : -1,
			"data" : "缺少taskId参数",
		})
		return
	}
	task := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : taskId,
	})
	if task == nil {
		c.JSON(http.StatusOK, gin.H{
			"message" : "未查询到该taskId对应的检测任务",
			"errorCode" : -2,
			"data" : "未查询到该taskId对应的检测任务",
		})
		return
	}
	platform := (*task)[0].Platform
	condition := "task_id='" + taskId + "'"
	toolsContent := dal.QueryTaskBinaryCheckContent(condition)
	if toolsContent == nil || len(*toolsContent) == 0 {
		logs.Info("未查询到该检测任务对应的自查工具")
		c.JSON(http.StatusOK, gin.H{
			"message" : "未查询到该检测任务对应的自查工具",
			"errorCode" : -3,
			"data" : "未查询到该检测任务对应的自查工具",
		})
		return
	}
	toolCondition := "id in("
	for i:=0; i<len(*toolsContent); i++ {
		content := (*toolsContent)[i]
		if i==len(*toolsContent)-1 {
			toolCondition += "'" + fmt.Sprint(content.ToolId) + "')"
		} else {
			toolCondition += "'" + fmt.Sprint(content.ToolId) + "',"
		}
	}
	toolCondition += " and platform ='" + strconv.Itoa(platform) + "'"
	selected := dal.QueryBinaryToolsByCondition(toolCondition)
	if selected==nil || len(*selected)==0 {
		c.JSON(http.StatusOK, gin.H{
			"message" : "未查询到该检测任务对应的自查工具",
			"errorCode" : -3,
			"data" : "未查询到该检测任务对应的自查工具",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : *selected,
	})
}

func QueryTaskBinaryCheckContent(c *gin.Context){
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
	condition := "task_id='" + taskId + "' and tool_id='" + toolId + "'"
	content := dal.QueryTaskBinaryCheckContent(condition)
	if content == nil || len(*content)==0{
		logs.Info("未查询到检测内容")
		c.JSON(http.StatusOK, gin.H{
			"message" : "未查询到检测内容",
			"errorCode" : -3,
			"data" : "未查询到检测内容",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : (*content)[0],
	})
}