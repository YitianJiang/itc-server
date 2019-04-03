package detect

import (
	"bytes"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)
/*
 *新建任务
 */
func ScNewTask(c *gin.Context){
	username, exist := c.Get("username")
	if !exist {
		logs.Error("未获取到username")
		c.JSON(http.StatusOK, gin.H{
			"message" : "未获取到username",
			"errorCode" : -1,
			"data" : "未获取到username",
		})
		return
	}
	repoUrl := c.DefaultQuery("repoUrl", "")
	if repoUrl == "" {
		logs.Error("未填写正确的仓库地址")
		c.JSON(http.StatusOK, gin.H{
			"message" : "未填写正确的仓库地址",
			"errorCode" : -2,
			"data" : "未填写正确的仓库地址",
		})
		return
	}
	branch := c.DefaultQuery("branch", "")
	if branch == "" {
		logs.Error("未填写正确的仓库分支")
		c.JSON(http.StatusOK, gin.H{
			"message" : "未填写正确的仓库分支",
			"errorCode" : -3,
			"data" : "未填写正确的仓库分支",
		})
		return
	}
	appId := c.DefaultQuery("appId", "")
	if appId == "" {
		logs.Error("缺少appId参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少appId参数",
			"errorCode" : -4,
			"data" : "缺少appId参数",
		})
		return
	}
	var ict dal.IOSCheckTask
	ict.AppId, _ = strconv.Atoi(appId)
	ict.UserName = username.(string)
	ict.RepoUrl = repoUrl
	ict.Branch = branch
	id, flag := dal.InsertICT(ict)
	if !flag {
		logs.Error("新建任务失败")
		c.JSON(http.StatusOK, gin.H{
			"message" : "新建任务失败",
			"errorCode" : -5,
			"data" : "新建任务失败",
		})
		return
	}
	ciUrl := ""
	callBackUrl := ""
	go func() {
		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)
		bodyWriter.WriteField("callback", callBackUrl)
		bodyWriter.WriteField("taskID", fmt.Sprint(id))
		logs.Info("url: ", ciUrl)
		toolHttp := &http.Client{
			Timeout: 60 * time.Second,
		}
		contentType := bodyWriter.FormDataContentType()
		response, err := toolHttp.Post(ciUrl, contentType, bodyBuffer)
		if err != nil {
			logs.Error("触发ci检测ios任务出错，将重试一次: ", err)
			response, err = toolHttp.Post(ciUrl, contentType, bodyBuffer)
		}
		if err != nil {
			logs.Error("触发ci检测ios任务出错，重试一次也失败", err)
		}
		resBody := &bytes.Buffer{}
		if response != nil {
			defer response.Body.Close()
			_, err = resBody.ReadFrom(response.Body)
			var data map[string]interface{}
			data = make(map[string]interface{})
			json.Unmarshal(resBody.Bytes(), &data)
			logs.Info("upload detect url's response: %+v", data)
		}
	}()
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
}
/*
 *分页查询任务列表
 */
func QueryICTTasks(c *gin.Context){
	appId := c.DefaultQuery("appId", "")
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
	var param map[string]interface{}
	param = make(map[string]interface{})
	if condition != "" {
		param["condition"] = condition
	}
	page, _ := strconv.Atoi(pageNo)
	size, _ := strconv.Atoi(pageSize)
	param["pageNo"] = page
	param["pageSize"] = size
	var data dal.RetICTTasks
	var more uint
	items, total := dal.QueryICT(param)
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
/*
 *更新任务检测结果
 */
func UpdateTask(c *gin.Context){
	username, _ := c.Get("username")
	id := c.DefaultQuery("id", "")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少id参数",
			"errorCode" : -1,
			"data" : "缺少id参数",
		})
		return
	}
	result := c.DefaultQuery("result", "")
	var ict dal.IOSCheckTask
	intId, _ := strconv.Atoi(id)
	ict.ID = uint(intId)
	ict.Result = result
	flag := dal.UpdateICT(ict)
	if !flag {
		c.JSON(http.StatusOK, gin.H{
			"message" : "检测任务更新失败",
			"errorCode" : -2,
			"data" : "检测任务更新失败",
		})
		return
	}
	utils.LarkDingOneInner(username.(string), "ios源码隐私检测已完成，请及时查看结果")
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
}

