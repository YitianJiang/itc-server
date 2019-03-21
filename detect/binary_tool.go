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
