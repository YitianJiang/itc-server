package detect

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)
/*
 *新增lark消息提醒
 */
func InsertLarkMsgCall(c *gin.Context) {
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.LarkMsgTimer
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("参数不合法!, ", err)
		c.JSON(http.StatusOK, gin.H{
			"message" : "json 参数不合法",
			"errorCode" : -1,
			"data" : "json 参数不合法",
		})
		return
	}
	name, f := c.Get("username")
	if !f {
		c.JSON(http.StatusOK, gin.H{
			"message" : "暂无权限！",
			"errorCode" : -2,
			"data" : "暂无权限！",
		})
		return
	}
	interval := t.MsgInterval
	if f, _ := regexp.MatchString("^\\d+$", strconv.Itoa(interval)); !f {
		logs.Error("时间间隔参数不合法！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "请填写正确的时间间隔！",
			"errorCode" : -3,
			"data" : "请填写正确的时间间隔！",
		})
		return
	}
	t.Operator = name.(string)
	logs.Info("%+v", t)
	flag := dal.InsertLarkMsgTimer(t)
	if !flag {
		logs.Error("Lark提醒配置新增失败！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "Lark提醒配置新增失败！",
			"errorCode" : -4,
			"data" : "Lark提醒配置新增失败！",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
}
/*
 *查询lark消息提醒
 */
func QueryLarkMsgCall(c *gin.Context) {
	appId := c.DefaultQuery("appId", "")
	if appId == "" {
		logs.Error("缺少appId参数！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少appId参数！",
			"errorCode" : -1,
			"data" : "缺少appId参数！",
		})
		return
	}
	appIdInt, err := strconv.Atoi(appId)
	if err != nil {
		logs.Error("appId参数不合法！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "appId参数不合法！",
			"errorCode" : -2,
			"data" : "appId参数不合法！",
		})
		return
	}
	var larkMsgConfig *dal.LarkMsgTimer
	larkMsgConfig = dal.QueryLarkMsgTimerByAppId(appIdInt)
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : larkMsgConfig,
	})
}
/*
 *新增lark群组配置信息
 */
func AddLarkGroup(c *gin.Context){
	name, f := c.Get("username")
	if !f {
		c.JSON(http.StatusOK, gin.H{
			"message" : "暂无权限！",
			"errorCode" : -1,
			"data" : "暂无权限！",
		})
		return
	}
	groupName := c.DefaultPostForm("groupName", "")
	if groupName == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少groupName参数",
			"errorCode" : -2,
			"data" : "缺少groupName参数",
		})
		return
	}
	groupId := c.DefaultPostForm("groupId", "")
	if groupId == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少groupId参数",
			"errorCode" : -3,
			"data" : "缺少groupId参数",
		})
		return
	}
	timerId := c.DefaultPostForm("timerId", "")
	if timerId == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少timerId参数",
			"errorCode" : -4,
			"data" : "缺少timerId参数",
		})
		return
	}
	platform := c.DefaultPostForm("platform", "")
	if platform == "" || (platform != "0" && platform != "1") {
		c.JSON(http.StatusOK, gin.H{
			"message" : "platform参数不合法",
			"errorCode" : -5,
			"data" : "platform参数不合法",
		})
		return
	}
	condition := "group_id='" + groupId + "'"
	groups := dal.QueryLarkGroupByCondition(condition)
	if groups != nil && len(*groups) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message" : "已存在该groupid对应的群组！",
			"errorCode" : -6,
			"data" : "已存在该groupid对应的群组！",
		})
		return
	}
	var larkGroup dal.LarkGroupMsg
	larkGroup.Operator = name.(string)
	larkGroup.GroupName = groupName
	larkGroup.GroupId = groupId
	larkGroup.TimerId, _ = strconv.Atoi(timerId)
	larkGroup.Platform, _ = strconv.Atoi(platform)
	larkGroup.CreatedAt = time.Now()
	larkGroup.UpdatedAt = time.Now()
	flag := dal.InsertLarkGroup(larkGroup)
	if !flag {
		c.JSON(http.StatusOK, gin.H{
			"message" : "lark群设置失败",
			"errorCode" : -5,
			"data" : "lark群设置失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
}
/*
 *更新lark群组配置信息
 */
func UpdateLarkGroup(c *gin.Context){
	name, f := c.Get("username")
	if !f {
		c.JSON(http.StatusOK, gin.H{
			"message" : "暂无权限！",
			"errorCode" : -1,
			"data" : "暂无权限！",
		})
		return
	}
	id := c.DefaultPostForm("id", "")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少id参数！",
			"errorCode" : -2,
			"data" : "缺少id参数！",
		})
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message" : "id参数格式不正确！",
			"errorCode" : -2,
			"data" : "id参数格式不正确！",
		})
		return
	}
	groupName := c.DefaultPostForm("groupName", "")
	if groupName == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少groupName参数",
			"errorCode" : -2,
			"data" : "缺少groupName参数",
		})
		return
	}
	groupId := c.DefaultPostForm("groupId", "")
	if groupId == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少groupId参数",
			"errorCode" : -3,
			"data" : "缺少groupId参数",
		})
		return
	}
	var larkGroup dal.LarkGroupMsg
	larkGroup.ID = uint(idInt)
	larkGroup.GroupName = groupName
	larkGroup.GroupId = groupId
	larkGroup.Operator = name.(string)
	flag := dal.UpdateLarkGroupById(larkGroup)
	if !flag {
		c.JSON(http.StatusOK, gin.H{
			"message" : "lark群组信息更新失败！",
			"errorCode" : -4,
			"data" : "lark群组信息更新失败！",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
}
/*
 *通过设置的lark信息id查询所对应的群组消息
 */
func QueryGroupInfosByTimerId(c *gin.Context){
	timerId := c.DefaultQuery("timerId", "")
	if timerId == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少timerId参数！",
			"errorCode" : -1,
			"data" : "缺少timerId参数！",
		})
		return
	}
	condition := "timer_id='" + timerId + "'"
	groups := dal.QueryLarkGroupByCondition(condition)
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : *groups,
	})
}