package detect

import (
	"bytes"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gopkg.in/cas.v2"
	"io"
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
	DETECT_URL_PRO = "10.2.9.226:9527"
)
func UploadFile(c *gin.Context){

	url := ""
	//get user info from cas
	name := cas.Username(c.Request)
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
		url = "http://" + DETECT_URL_DEV + "/apk_post"
	} else if platform == "1"{
		flag := strings.HasSuffix(filename, ".ipa")
		if !flag{
			errorFormatFile(c)
			return
		}
		url = "http://" + DETECT_URL_DEV + "/ipa_post/v2"
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
	var recipients = "ttqaall@bytedance.com,tt_ios@bytedance.com,"
	recipients += name + "@bytedance.com"
	filepath := _tmpDir + "/" + filename
	//1、上传至tos
	tosUrl, err := upload2Tos(filepath)
	//2、将相关信息保存至数据库
	var dbDetectModel dal.DetectStruct
	dbDetectModel.Creator = name
	dbDetectModel.SelfCheckStatus = 0
	dbDetectModel.CreatedAt = time.Now()
	dbDetectModel.UpdatedAt = time.Now()
	dbDetectModel.TosUrl = tosUrl
	dbDetectModelId := dal.InsertDetectModel(dbDetectModel)
	//3、调用检测接口，进行二进制检测 && 删掉本地临时文件
	callBackUrl := "http://itc.byted.org/updateDetectInfos" + "?taskID=" + string(dbDetectModelId)
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)
	bodyWriter.WriteField("recipients", recipients)
	bodyWriter.WriteField("callback", callBackUrl)
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
	json.Unmarshal(resBody.Bytes(), &data)
	//删掉临时文件
	os.Remove(filepath)
	c.JSON(http.StatusOK, gin.H{
		"message" : data["msg"],
		"errorCode" : data["success"],
		"data" : data["msg"],
	})
	return
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
	detect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id" : taskId,
	})
	(*detect)[0].AppName = appName
	(*detect)[0].AppVersion = appVersion
	//(*detect)[0].CheckContent = htmlContent
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
	pageNo := c.DefaultQuery("pageNo", "")
	//如果缺少pageSize参数，则选用默认每页显示10条数据
	pageSize := c.DefaultQuery("pageSize", "10")
	//参数校验
	if pageNo == "" {
		logs.Error("缺少pageNo参数！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少pageNo参数！",
			"errorCode" : -1,
			"data" : "缺少pageNo参数！",
		})
		return
	}
	condition := "1=1"
	if appId != "" {
		condition += " and appId=" + appId
	}
	if version != "" {
		condition += " and version=" + version
	}
	if creator != "" {
		condition += " and creator=" + creator
	}
	var param map[string]interface{}
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

	/*var tool1 *dal.DetectTool = new(dal.DetectTool)
	tool1.Id = 1
	tool1.Platform = 0
	tool1.Name = "Android GooglePlay及图片检测"
	tool1.Desc = "进行GooglePlay及图片信息的检测，是否影响上架"
	var tool2 *dal.DetectTool = new(dal.DetectTool)
	tool2.Id = 2
	tool2.Platform = 0
	tool2.Name = "Android隐私检测"
	tool2.Desc = "进行隐私信息的检测"
	data := []dal.DetectTool{*tool1, *tool2}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : data,
	})*/
	platform := c.DefaultQuery("platform", "")
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
	condition := "platform=" + platform
	tools := dal.QueryBinaryToolsByCondition(condition)
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
	condition := "task_id=" + taskId + " and platform=" + strconv.Itoa(platform)
	tools := dal.QueryBinaryToolsByCondition(condition)
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : tools,
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
	condition := "task_id=" + taskId + " and tool_id=" + toolId
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