package detect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

const (
	platformAndorid = 0
	platformiOS     = 1
)

// Status of Detect Task
const (
	TaskStatusError       = -2
	TaskStatusRunning     = -1
	TaskStatusUnconfirm   = 0
	TaskStatusConfirmPass = 1
	TaskStatusConfirmFail = 2
)

var LARK_MSG_CALL_MAP = make(map[string]interface{})

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
	if items == nil {
		ReturnMsg(c, SUCCESS, "Cannot find any matched task")
		return
	}
	if uint(page*size) >= total {
		more = 0
	} else {
		more = 1
	}
	data.GetMore = more
	data.Total = total
	data.NowPage = uint(page)
	data.Tasks = *items
	if appId == "1319" {
		for i := 0; i < len(*items); i++ {
			(*items)[i].AppName = "皮皮虾"
		}
	}
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
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, "miss taskId")
		return
	}
	task := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id": taskId,
	})
	if task == nil || len(*task) == 0 {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid taskId: %v", taskId))
		return
	}
	platform := (*task)[0].Platform
	condition := "task_id='" + taskId + "'"
	toolsContent := dal.QueryTaskBinaryCheckContent(condition)
	//scanner检测工具查询内容
	toolsContent_2, _ := QueryDetectInfo(condition)
	if (toolsContent == nil || len(*toolsContent) == 0) && (toolsContent_2 == nil) {
		logs.Error("未查询到该检测任务对应的二进制检测结果")
		var res [0]dal.DetectContent
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"platform":  platform,
			"errorCode": 0,
			"appId":     (*task)[0].AppId,
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
		"platform":  platform,
		"errorCode": 0,
		"appId":     (*task)[0].AppId,
		"data":      *selected,
	})
}

/**
查询apk检测info-----fj
*/
func QueryDetectInfo(condition string) (*dal.DetectInfo, error) {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, err
	}
	defer connection.Close()

	db := connection.Table(dal.DetectInfo{}.TableName()).LogMode(_const.DB_LOG_MODE)

	var detectInfo dal.DetectInfo
	if err = db.Where(condition).Find(&detectInfo).Error; err != nil {
		logs.Error("Database error: %v", err)
		return nil, err
	}

	return &detectInfo, nil
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
	//larkList := strings.Split("kanghuaisong,yinzhihong,fanjuan.xqp", ",")
	for _, creator := range _const.LowLarkPeople {
		utils.LarkDingOneInner(creator, "检测服务异常，请立即关注！"+message)
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
	})
}

/**
预审发送post信息
*/
func PostInfos(url string, data map[string]string) error {
	taskId := data["task_id"]
	bytesData, err1 := json.Marshal(data)
	if err1 != nil {
		logs.Error("任务ID：" + fmt.Sprint(taskId) + ",CI回调信息转换失败" + fmt.Sprint(err1))
		for _, lark_people := range _const.LowLarkPeople {
			utils.LarkDingOneInner(lark_people, "CI回调信息转换失败，请及时进行检查！任务ID："+fmt.Sprint(taskId))
		}
		return err1
	}
	reader := bytes.NewReader(bytesData)

	request, err2 := http.NewRequest("POST", url, reader)
	if err2 != nil {
		logs.Error("任务ID：" + fmt.Sprint(taskId) + ",CI回调请求Create失败" + fmt.Sprint(err2))
		for _, lark_people := range _const.LowLarkPeople {
			utils.LarkDingOneInner(lark_people, "CI回调请求Create失败，请及时进行检查！任务ID："+fmt.Sprint(taskId))
		}
		return err2
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err3 := client.Do(request)
	if err3 != nil {
		logs.Error("任务ID："+taskId+",回调CI接口失败,%v", err3)
		//及时报警
		//utils.LarkDingOneInner("kanghuaisong", "二进制包检测服务无响应，请及时进行检查！任务ID："+fmt.Sprint(task.ID))
		for _, lark_people := range _const.LowLarkPeople {
			utils.LarkDingOneInner(lark_people, "CI回调请求发送失败，请及时进行检查！任务ID："+fmt.Sprint(taskId))
		}
		return err3
	}
	logs.Info("任务ID：" + fmt.Sprint(taskId) + "回调成功,回调信息：" + fmt.Sprint(data) + ",回调地址：" + url)
	if resp != nil {
		defer resp.Body.Close()
		respBytes, _ := ioutil.ReadAll(resp.Body)
		var data map[string]interface{}
		data = make(map[string]interface{})
		json.Unmarshal(respBytes, &data)
		logs.Info("taskId :"+taskId+",CI detect url's response: %+v", data)
	}
	return nil
}
