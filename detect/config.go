package detect

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func AddConfig(c *gin.Context) {
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.ItemConfig
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
	id := t.ID
	if id != 0 {
		condition := "id='" + fmt.Sprint(id) + "'"
		var config *[]dal.ItemConfig
		config = dal.QueryConfigByCondition(condition)
		var item = (*config)[0]
		item.Name = t.Name
		flag := dal.UpdateConfigByCondition(condition, item)
		if !flag {
			c.JSON(http.StatusOK, gin.H{
				"message" : "配置更新失败！",
				"errorCode" : -5,
				"data" : "配置更新失败！",
			})
			return
		}
	} else {
		condition := "config_type='" + fmt.Sprint(t.ConfigType) + "' and name='" + t.Name + "'"
		var config *[]dal.ItemConfig
		config = dal.QueryConfigByCondition(condition)
		if config != nil && len(*config)>0 {
			c.JSON(http.StatusOK, gin.H{
				"message" : "该配置参数已经存在！",
				"errorCode" : -3,
				"data" : "该配置参数已经存在！",
			})
			return
		}
		flag := dal.InsertItemConfig(t)
		if !flag {
			c.JSON(http.StatusOK, gin.H{
				"message" : "新增配置失败！",
				"errorCode" : -4,
				"data" : "新增配置失败！",
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
}
//查询配置项
func QueryConfigs(c *gin.Context) {

	condition := "1=1"
	var config *[]dal.ItemConfig
	config = dal.QueryConfigByCondition(condition)
	var config0 []dal.ItemConfig
	var config1 []dal.ItemConfig
	var config2 []dal.ItemConfig
	if config==nil || len(*config)==0 {
		c.JSON(http.StatusOK, gin.H{
			"message" : "未查询到配置信息！",
			"errorCode" : -1,
			"data" : "未查询到配置信息！",
		})
		return
	}
	for i:=0; i<len(*config); i++ {
		item := (*config)[i]
		if item.ConfigType == 0 {
			config0 = append(config0, item)
		} else if item.ConfigType == 1 {
			config1 = append(config1, item)
		} else if item.ConfigType == 2 {
			config2 = append(config2, item)
		}
	}
	var data map[int]interface{}
	data = make(map[int]interface{})
	data[0] = config0
	data[1] = config1
	data[2] = config2
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : data,
	})
}
