package detect

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

/*
 *增加检查项
 */
func AddDetectItem(c *gin.Context) {

	param, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.MutilitemStruct
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("参数格式错误!, ", err)
		c.JSON(http.StatusOK, gin.H{
			"message":   "参数格式错误",
			"errorCode": -5,
			"data":      "参数格式错误",
		})
		return
	}
	//输入参数判断处理
	//regulationUrl := t.RegulationUrl
	ggFlag := t.IsGG //是否公共,0-否；1-是
	platform := t.Platform
	appIdArr := strings.Split(t.AppId, ",")              //支持多个app
	QuestionArr := strings.Split(t.QuestionTypeArr, ",") //支持多种问题类型
	//参数判断
	if platform != 0 && platform != 1 {
		logs.Error("platform参数不合法！")
		c.JSON(http.StatusOK, gin.H{
			"message":   "platform参数不合法！",
			"errorCode": -1,
			"data":      "platform参数不合法！",
		})
		return
	}
	if ggFlag != 0 && ggFlag != 1 {
		ggFlag = 0 //没有勾选是否公共，以当前appId作为检查项归属app
	}
	if len(appIdArr) == 0 {
		logs.Error("缺失参数appId！")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺失参数appId！",
			"errorCode": -2,
			"data":      "缺失参数appId！",
		})
		return
	}
	if len(QuestionArr) == 0 {
		logs.Error("缺失参数QuestionType！")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺失参数QuestionType！",
			"errorCode": -2,
			"data":      "缺失参数QuestionType！",
		})
		return
	}
	//校验
	//暂时注释掉对链接的正则检测
	/*if f, _ := regexp.MatchString("^http(s?)://*", regulationUrl); !f {
		logs.Error("条例链接格式不正确！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "条例链接格式不正确！",
			"errorCode" : -3,
			"data" : "条例链接格式不正确！",
		})
		return
	}*/
	itemModelId := dal.InsertItemModel(t)
	if itemModelId == false {
		logs.Error("新增检查项失败")
		c.JSON(http.StatusOK, gin.H{
			"message":   "新增检查项失败，请联系相关人员！",
			"errorCode": -4,
			"data":      "新增检查项失败，请联系相关人员！",
		})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data":      "新增检查项成功！",
		})
		return
	}
}

/*
 *查询检查项
 */
func GetSelfCheckItems(c *gin.Context) {

	appIdParam, ok := c.GetQuery("appId")
	if !ok {
		logs.Error("缺少appId参数！")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少appId参数！",
			"errorCode": -1,
			"data":      "缺少appId参数！",
		})
		return
	}
	taskId, bool := c.GetQuery("taskId")
	if !bool { //检查项管理中返回和appID相关检查项，
		appSelfItem := dal.QueryAppSelfItem(map[string]interface{}{
			"appId": appIdParam,
		})
		if appSelfItem == nil || len(*appSelfItem) == 0 {
			ggItemList := getGGItem(map[string]interface{}{
				"is_gg" : 1,
			})
			c.JSON(http.StatusOK, gin.H{
				"message":   "success！",
				"errorCode": 0,
				"data":      ggItemList,
			})
			return
		} else {
			var appItemList []interface{}
			for _, appSelf := range *appSelfItem {
				appMap := make(map[string]interface{})
				if err := json.Unmarshal([]byte(appSelf.SelfItems), &appMap); err != nil {
					logs.Error(err.Error())
					c.JSON(http.StatusOK, gin.H{
						"message":   "解析json出错！",
						"errorCode": -1,
						"data":      []interface{}{},
					})
					return
				}
				for _, i := range appMap["item"].([]interface{}) {
					appItemList = append(appItemList, i)
				}
			}
			c.JSON(http.StatusOK, gin.H{
				"message":   "success！",
				"errorCode": 0,
				"data":      appItemList,
			})
			return
		}
	}
	//任务管理中展示自查项
	task_id, _ := strconv.Atoi(taskId)
	taskDetect := dal.QueryDetectModelsByMap(map[string]interface{}{
		"id": task_id,
	})
	var platform int
	if taskDetect != nil && len(*taskDetect) != 0 {
		platform = (*taskDetect)[0].Platform
	}
	flag, data := dal.QueryTaskSelfItemList(task_id)
	//查询出错
	if !flag {
		c.JSON(http.StatusOK, gin.H{
			"message":   "查询Task自查项失败！",
			"errorCode": -1,
			"data":      []interface{}{},
		})
		return
	}
	//查询taskItemList为空
	if data == nil {
		appSelf := dal.QueryAppSelfItem(map[string]interface{}{
			"appId":    appIdParam,
			"platform": platform,
		})
		var taskSelf []interface{} //task的自查项
		//appItem检查项为空
		if appSelf == nil || len(*appSelf) == 0 {
			ggList := getGGItem(map[string]interface{}{
				"is_gg":1,
				"platform":platform,
			})
			if len(ggList) == 0{ //公共检查项为空
				c.JSON(http.StatusOK, gin.H{
					"message":   "success,该APP没有自查项！",
					"errorCode": 0,
					"data":      []interface{}{},
				})
				return
			}
			taskSelf = ggList //公共项不为空
			//插入appItem
			var appGGSelf dal.AppSelfItem
			appGGSelf.Platform = platform
			appGGSelf.AppId, _ = strconv.Atoi(appIdParam)
			ggJson, _ := json.Marshal(map[string]interface{}{
				"item":ggList,
			})
			appGGSelf.SelfItems = string(ggJson)
			dal.InsertAppSelfItem(appGGSelf)
		}else{
			//appItem检查项不为空
			returnSelf := (*appSelf)[0].SelfItems
			var temp = make(map[string]interface{})
			json.Unmarshal([]byte(returnSelf), &temp)
			taskSelf = temp["item"].([]interface{})
		}
		//兼容之前的taskID

		//task插入检查项
		for _, self := range taskSelf {
			self.(map[string]interface{})["status"] = 0
			self.(map[string]interface{})["confirmer"] = ""
			self.(map[string]interface{})["remark"] = ""
		}
		task_self, err := json.Marshal(map[string]interface{}{
			"item":taskSelf,
		})
		if err != nil {
			logs.Error(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"message":   "map json转换出错！",
				"errorCode": -1,
				"data":      []interface{}{},
			})
			return
		}
		var taskSelfItem dal.TaskSelfItem
		taskSelfItem.SelfItems = string(task_self)
		taskSelfItem.TaskId = task_id
		isInsert := dal.InsertTaskSelfItem(taskSelfItem)
		if !isInsert {
			c.JSON(http.StatusOK, gin.H{
				"message":   "Task自查项插入数据库失败",
				"errorCode": -1,
				"data":      []interface{}{},
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message":   "success！",
				"errorCode": 0,
				"data":      taskSelf,
			})
			return
		}
	}
	if flag && data != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success！",
			"errorCode": 0,
			"data":      data,
		})
	}
}

func getGGItem(condition map[string]interface{}) []interface{}{
	ggItem := dal.QueryItem(condition)
	if ggItem == nil || len(*ggItem) == 0{
		return []interface{}{}
	}
	var ggItemList []interface{}
	for _, gg_item := range *ggItem{
		itemJson, _ := json.Marshal(gg_item)
		m := make(map[string]interface{})
		json.Unmarshal(itemJson, &m)
		delete(m, "ID")
		delete(m, "CreatedAt")
		delete(m, "DeletedAt")
		delete(m, "UpdatedAt")
		m["id"] = gg_item.ID
		ggItemList = append(ggItemList, m)
	}
	return ggItemList
}
/*
 *完成自查
 */
func ConfirmCheck(c *gin.Context) {
	p, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.Confirm
	err := json.Unmarshal(p, &t)
	if err != nil {
		logs.Error("参数不合法!, ", err)
		c.JSON(http.StatusOK, gin.H{
			"message":   "参数不合法",
			"errorCode": -2,
			"data":      "参数不合法",
		})
		return
	}
	name, flag := c.Get("username")
	if !flag {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未获取到用户信息！",
			"errorCode": -1,
			"data":      "未获取到用户信息！",
		})
		return
	}
	var param map[string]interface{}
	param = make(map[string]interface{})
	param["taskId"] = t.TaskId
	param["data"] = t.Data
	param["operator"] = name
	bool, detect := dal.ConfirmSelfCheck(param)
	if !bool {
		c.JSON(http.StatusOK, gin.H{
			"message":   "自查确认失败，请联系相关人员！",
			"errorCode": -1,
			"data":      "自查确认失败，请联系相关人员！",
		})
		return
	}
	if detect != nil && detect.Status == 1 && detect.SelfCheckStatus == 1{
		CICallBack(detect)
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      "success",
	})
}

/*
 *删除检查项
 */
var whiteList = []string{"kanghuaisong", "zhangshuai.02", "liusiyu.lsy", "yinzhihong"} //检查项删除白名单
func DropDetectItem(c *gin.Context) {
	itemId := c.DefaultPostForm("itemId", "")
	if itemId == "" {
		c.JSON(http.StatusOK, gin.H{
			"message":   "参数不合法",
			"errorCode": -2,
			"data":      "缺少itemId！",
		})
		return
	}
	isAll := c.DefaultPostForm("isGG", "")
	if isAll == "" {
		c.JSON(http.StatusOK, gin.H{
			"message":   "参数不合法",
			"errorCode": -2,
			"data":      "缺少isGG！",
		})
		return
	}
	appId := c.DefaultPostForm("appId", "")
	if appId == "" {
		c.JSON(http.StatusOK, gin.H{
			"message":   "参数不合法",
			"errorCode": -2,
			"data":      "缺少appId！",
		})
		return
	}
	name, flag := c.Get("username")
	if !flag {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未获取到用户信息！",
			"errorCode": -2,
			"data":      "未获取到用户信息！",
		})
		return
	}
	//判断用户是否有权限
	isPrivacy := false
	for _, people := range whiteList {
		if people == name {
			isPrivacy = true
		}
	}
	is_all, _ := strconv.Atoi(isAll)
	if !isPrivacy && is_all == 1 {
		c.JSON(http.StatusOK, gin.H{
			"message":   "该用户没有删除权限",
			"errorCode": -3,
			"data":      "该用户没有删除权限",
		})
		return
	}
	if dal.DeleteItemsByCondition(map[string]interface{}{
		"id":    itemId,
		"isGG":  isAll,
		"appId": appId,
	}) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data":      "success",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "删除自查项失败！",
			"errorCode": -1,
			"data":      "删除自查项失败！",
		})
	}

}
