package detect

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func InsertBinaryTool(c *gin.Context) {
	/*platform := c.DefaultPostForm("platform", "")
	if platform == "" {
		logs.Error("缺少platform参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少platform参数",
			"errorCode" : -1,
			"data" : "缺少platform参数",
		})
		return
	}
	if platform != "0" && platform != "1" {
		logs.Error("platform参数不合法")
		c.JSON(http.StatusOK, gin.H{
			"message" : "platform参数不合法",
			"errorCode" : -2,
			"data" : "platform参数不合法",
		})
		return
	}
	name := c.DefaultPostForm("name", "")
	if name == "" {
		logs.Error("缺少检测工具name参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少检测工具name参数",
			"errorCode" : -3,
			"data" : "缺少检测工具name参数",
		})
		return
	}
	desc := c.DefaultPostForm("desc", "")
	if desc == "" {
		logs.Error("缺少检测工具desc参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少检测工具desc参数",
			"errorCode" : -4,
			"data" : "缺少检测工具desc参数",
		})
		return
	}*/
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.BinaryDetectTool
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("json unmarshal failed!, ", err)
		c.JSON(http.StatusOK, gin.H{
			"message" : "json unmarshal failed",
			"errorCode" : -5,
			"data" : "json unmarshal failed",
		})
		return
	}
	//按name进行检索，是否已经存在
	condition := "name='" + t.Name + "'"
	var exist *[]dal.BinaryDetectTool
	exist = dal.QueryBinaryToolsByCondition(condition)
	if exist!=nil && len(*exist)>0 {
		logs.Error("二进制检测工具新增失败，名称已经存在")
		c.JSON(http.StatusOK, gin.H{
			"message" : "该检测工具" + t.Name + "已经存在",
			"errorCode" : -5,
			"data" : "该检测工具" + t.Name + "已经存在",
		})
		return
	}
	/*var tool dal.BinaryDetectTool
	tool.Name = name
	tool.Description = desc
	tool.Platform, _ = strconv.Atoi(platform)*/
	bool := dal.InsertBinaryTool(t)
	if !bool {
		logs.Error("二进制检测工具新增失败")
		c.JSON(http.StatusOK, gin.H{
			"message" : "检测工具新增失败",
			"errorCode" : -5,
			"data" : "检测工具新增失败",
		})
		return
	}
	logs.Info("二进制检测工具新增成功")
	c.JSON(http.StatusOK, gin.H{
		"message" : "sccess",
		"errorCode" : 0,
		"data" : "sccess",
	})
}

func QueryBinaryTools(c *gin.Context) {
	platform := c.DefaultQuery("platform", "")
	condition := ""
	if platform != "" {
		condition = "platform='" + platform + "'"
	}
	var tools *[]dal.BinaryDetectTool
	tools = dal.QueryBinaryToolsByCondition(condition)
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
		"message" : "sccess",
		"errorCode" : 0,
		"data" : *tools,
	})
}
