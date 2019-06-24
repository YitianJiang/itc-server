package detect

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

/*
 *增加检查项
 */
func AddDetectItem(c *gin.Context) {

	param, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.ItemStruct
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
	//regulationUrl := t.RegulationUrl
	ggFlag := t.IsGG
	platform := t.Platform
	appId := t.AppId
	//platform
	if platform != 0 && platform != 1 {
		logs.Error("platform参数不合法！")
		c.JSON(http.StatusOK, gin.H{
			"message":   "platform参数不合法！",
			"errorCode": -1,
			"data":      "platform参数不合法！",
		})
		return
	}
	//是否公共,0-否；1-是
	if ggFlag != 0 && ggFlag != 1 {
		ggFlag = 1
	}
	if ggFlag == 0 && appId == 0 {
		logs.Error("缺失参数appId！")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺失参数appId！",
			"errorCode": -2,
			"data":      "缺失参数appId！",
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
	t.Status = 0
	itemModelId := dal.InsertItemModel(t)
	if itemModelId == 0 {
		logs.Error("新增检查项失败")
		c.JSON(http.StatusOK, gin.H{
			"message":   "新增检查项失败，请联系相关人员！",
			"errorCode": -4,
			"data":      "新增检查项失败，请联系相关人员！",
		})
		return
	} else {
		logs.Error("新增检查项成功")
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
	condition := "1=1"
	if bool {
		condition += " and id='" + taskId + "'"
	}
	var param map[string]interface{}
	var data map[string]interface{}
	param = make(map[string]interface{})
	data = make(map[string]interface{})
	itemCondition := ""
	if bool {
		param["condition"] = condition
		tasks, _ := dal.QueryTasksByCondition(param)
		if len(*tasks) > 0 {
			appId := (*tasks)[0].AppId
			platform := (*tasks)[0].Platform
			itemCondition = "(platform='" + strconv.Itoa(platform) + "')" + "and ((is_gg='1') or (is_gg='0' and app_id='" + appId + "'))"
		}
	} else {
		itemCondition = "((is_gg='1') or (is_gg='0' and app_id='" + appIdParam + "'))"
	}
	data["condition"] = itemCondition
	items := dal.QueryItemsByCondition(data)
	if items == nil || len(*items) == 0 {
		logs.Error("未查询到自查项信息！")
		var res [0]dal.QueryItemStruct
		c.JSON(http.StatusOK, gin.H{
			"message":   "success！",
			"errorCode": 0,
			"data":      res,
		})
		return
	}
	if !bool {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data":      *items,
		})
		return
	}
	tj := "task_id='" + taskId + "'"
	itemMap, remarkMap, confirmerMap := dal.GetSelfCheckByTaskId(tj)
	var filterItem []dal.QueryItemStruct
	if itemMap == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data":      *items,
		})
	} else {
		for i := 0; i < len(*items); i++ {
			item := (*items)[i]
			if _, ok := itemMap[item.ID]; ok {
				status := itemMap[item.ID]
				item.Status = status
				item.Remark = remarkMap[item.ID]
				item.Confirmer = confirmerMap[item.ID]
			} else {
				item.Status = 0
				item.Remark = ""
				item.Confirmer = ""
			}
			(*items)[i] = item
			filterItem = append(filterItem, item)
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data":      filterItem,
		})
	}
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
			"errorCode": -5,
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
	bool := dal.ConfirmSelfCheck(param)
	if !bool {
		c.JSON(http.StatusOK, gin.H{
			"message":   "自查确认失败，请联系相关人员！",
			"errorCode": -3,
			"data":      "自查确认失败，请联系相关人员！",
		})
		return
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
			"data":      "参数不合法,请重新输入！",
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
	if isPrivacy {
		if dal.DeleteItemsByCondition(map[string]interface{}{
			"id": itemId,
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
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "该用户没有删除权限",
			"errorCode": -3,
			"data":      "该用户没有删除权限",
		})
	}
}
