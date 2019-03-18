package detect

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strconv"
)

func InsertLarkMsgCall(c *gin.Context) {
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
	intervalType := c.DefaultQuery("type", "")
	if intervalType == "" {
		logs.Error("缺少type参数！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少type参数！",
			"errorCode" : -2,
			"data" : "缺少type参数！",
		})
		return
	}
	interval := c.DefaultQuery("interval", "")
	if interval == "" {
		logs.Error("缺少interval参数！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少interval参数！",
			"errorCode" : -3,
			"data" : "缺少interval参数！",
		})
		return
	}
	//校验
	if f, _ := regexp.MatchString("^\\d+$", interval); !f {
		logs.Error("时间间隔参数不合法！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "请填写正确的时间间隔！",
			"errorCode" : -4,
			"data" : "请填写正确的时间间隔！",
		})
		return
	}
	intervalInt, _ := strconv.Atoi(interval)
	if intervalInt == 0 {
		logs.Error("时间间隔应该大于0！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "时间间隔应该大于0！",
			"errorCode" : -5,
			"data" : "时间间隔应该大于0！",
		})
		return
	}
	var config dal.LarkMsgTimer
	config.AppId, _ = strconv.Atoi(appId)
	config.Type, _ = strconv.Atoi(intervalType)
	config.Interval = intervalInt
	flag := dal.InsertLarkMsgTimer(config)
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
