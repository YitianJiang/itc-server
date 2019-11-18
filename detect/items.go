package detect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
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
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, "miss appId")
		return
	}
	taskId, bool := c.GetQuery("taskId")
	if !bool { //检查项管理中返回和appID相关检查项，
		var ItemList []interface{}
		appSelfItem_A := dal.QueryAppSelfItem(map[string]interface{}{
			"appId":    appIdParam,
			"platform": 0,
		})
		appSelfItem_O := dal.QueryAppSelfItem(map[string]interface{}{
			"appId":    appIdParam,
			"platform": 1,
		})
		var androidItemList []interface{}
		if appSelfItem_A == nil || len(*appSelfItem_A) == 0 {
			androidItemList = getGGItem(map[string]interface{}{
				"is_gg":    1,
				"platform": 0,
			})
		} else {
			androidMap := make(map[string]interface{})
			if err := json.Unmarshal([]byte((*appSelfItem_A)[0].SelfItems), &androidMap); err != nil {
				utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("unmarshal error: %v content: %v", err, (*appSelfItem_A)[0].SelfItems))
				return
			}
			for _, i := range androidMap["item"].([]interface{}) {
				androidItemList = append(androidItemList, i)
			}
		}
		var iOSItemList []interface{}
		if appSelfItem_O == nil || len(*appSelfItem_O) == 0 {
			iOSItemList = getGGItem(map[string]interface{}{
				"is_gg":    1,
				"platform": 1,
			})
		} else {
			iosMap := make(map[string]interface{})
			if err := json.Unmarshal([]byte((*appSelfItem_O)[0].SelfItems), &iosMap); err != nil {
				utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("unmarshal error: %v content: %v", err, (*appSelfItem_O)[0].SelfItems))
				return
			}
			for _, i := range iosMap["item"].([]interface{}) {
				iOSItemList = append(iOSItemList, i)
			}
		}
		//Android和iOS list合并
		if len(androidItemList) == 0 && len(iOSItemList) == 0 {
			ItemList = []interface{}{}
		} else if len(androidItemList) == 0 && len(iOSItemList) != 0 {
			ItemList = iOSItemList
		} else if len(androidItemList) != 0 && len(iOSItemList) == 0 {
			ItemList = androidItemList
		} else {
			for _, a := range androidItemList {
				ItemList = append(ItemList, a)
			}
			for _, o := range iOSItemList {
				ItemList = append(ItemList, o)
			}
		}
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, "success", ItemList)
		return
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
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("get self-check items failed"))
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
				"is_gg":    1,
				"platform": platform,
			})
			if len(ggList) == 0 { //公共检查项为空
				utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success")
				return
			}
			taskSelf = ggList //公共项不为空
			//插入appItem
			var appGGSelf dal.AppSelfItem
			appGGSelf.Platform = platform
			appGGSelf.AppId, _ = strconv.Atoi(appIdParam)
			ggJson, _ := json.Marshal(map[string]interface{}{
				"item": ggList,
			})
			appGGSelf.SelfItems = string(ggJson)
			dal.InsertAppSelfItem(appGGSelf)
		} else {
			//appItem检查项不为空
			returnSelf := (*appSelf)[0].SelfItems
			var temp = make(map[string]interface{})
			json.Unmarshal([]byte(returnSelf), &temp)
			taskSelf = temp["item"].([]interface{})
		}
		//兼容之前的taskID
		itemMap, remarkMap, confirmerMap := dal.GetSelfCheckByTaskId("task_id='" + taskId + "'")
		//task插入检查项
		for _, self := range taskSelf {
			var item_id uint
			switch self.(map[string]interface{})["id"].(type) {
			case uint:
				item_id = self.(map[string]interface{})["id"].(uint)
			case float64:
				item_id = uint(self.(map[string]interface{})["id"].(float64))
			case int:
				item_id = uint(self.(map[string]interface{})["id"].(int))
			}
			if status, ok := itemMap[item_id]; ok {
				self.(map[string]interface{})["status"] = status
			} else {
				self.(map[string]interface{})["status"] = 0
			}
			if confirmer, ok := confirmerMap[item_id]; ok {
				self.(map[string]interface{})["confirmer"] = confirmer
			} else {
				self.(map[string]interface{})["confirmer"] = ""
			}
			if remark, ok := remarkMap[item_id]; ok {
				self.(map[string]interface{})["remark"] = remark
			} else {
				self.(map[string]interface{})["remark"] = ""
			}
		}
		task_self, err := json.Marshal(map[string]interface{}{
			"item": taskSelf,
		})
		if err != nil {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("marshal error: %v", err))
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
			utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success", taskSelf)
			return
		}
	}
	if flag && data != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success", data)
	}
}

func getGGItem(condition map[string]interface{}) []interface{} {
	ggItem := dal.QueryItem(condition)
	//获取配置项
	var configMap map[int]string
	configMap = make(map[int]string)
	configs := dal.QueryConfigByCondition("1=1")
	if configs != nil && len(*configs) > 0 {
		for i := 0; i < len(*configs); i++ {
			config := (*configs)[i]
			configMap[int(config.ID)] = config.Name
		}
	}
	if ggItem == nil || len(*ggItem) == 0 {
		return []interface{}{}
	}
	var ggItemList []interface{}
	for _, gg_item := range *ggItem {
		itemJson, _ := json.Marshal(gg_item)
		m := make(map[string]interface{})
		json.Unmarshal(itemJson, &m)
		delete(m, "ID")
		delete(m, "CreatedAt")
		delete(m, "DeletedAt")
		delete(m, "UpdatedAt")
		m["id"] = gg_item.ID
		keyWord := m["keyWord"].(float64)
		fixWay := m["fixWay"].(float64)
		questionType := m["questionType"].(float64)
		m["keyWordName"] = configMap[int(keyWord)]
		m["fixWayName"] = configMap[int(fixWay)]
		m["questionTypeName"] = configMap[int(questionType)]
		ggItemList = append(ggItemList, m)
	}
	return ggItemList
}

/*
 *完成自查
 */
func ConfirmSelfCheckItems(c *gin.Context) {

	userName, exist := c.Get("username")
	if !exist {
		utils.ReturnMsg(c, http.StatusUnauthorized, utils.FAILURE, "unauthrized user")
		return
	}

	var t dal.Confirm
	if err := c.ShouldBindJSON(&t); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid parameter: %v", err))
		return
	}

	var passItems []dal.Self
	for i := range t.Data {
		if t.Data[i].Status == 1 {
			passItems = append(passItems, t.Data[i])
		}
	}

	detect, err := dal.ConfirmSelfCheck(map[string]interface{}{
		"taskId":   t.TaskId,
		"data":     passItems,
		"operator": userName})
	if err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("self-check failed: %v", err))
		return
	}
	if detect != nil && detect.Status != 0 && detect.SelfCheckStatus != 0 {
		StatusDeal(*detect, 2)
		// TODO: auto update self item?
	}

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success")
	return
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
