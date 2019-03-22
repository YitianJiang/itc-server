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
)

func InsertLarkMsgCall(c *gin.Context) {
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.LarkMsgTimer
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("参数不合法!, ", err)
		c.JSON(http.StatusOK, gin.H{
			"message" : "json 参数不合法",
			"errorCode" : -5,
			"data" : "json 参数不合法",
		})
		return
	}
	name, f := c.Get("username")
	if !f {
		c.JSON(http.StatusOK, gin.H{
			"message" : "未获取到用户信息！",
			"errorCode" : -1,
			"data" : "未获取到用户信息！",
		})
		return
	}
	if name == "" {
		name = "kanghuaisong"
	}

	interval := t.MsgInterval
	if f, _ := regexp.MatchString("^\\d+$", strconv.Itoa(interval)); !f {
		logs.Error("时间间隔参数不合法！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "请填写正确的时间间隔！",
			"errorCode" : -4,
			"data" : "请填写正确的时间间隔！",
		})
		return
	}
	flag := dal.InsertLarkMsgTimer(t)
	if !flag {
		logs.Error("Lark提醒配置新增失败！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "Lark提醒配置新增失败！",
			"errorCode" : -6,
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
	if larkMsgConfig == nil {
		logs.Error("未查询到配置信息！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "未查询到配置信息！",
			"errorCode" : -3,
			"data" : "未查询到配置信息！",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : larkMsgConfig,
	})
}
