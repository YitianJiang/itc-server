package detect

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func AddConfig(c *gin.Context) {
	configType := c.DefaultQuery("configType", "")
	if configType == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少configType参数",
			"errorCode" : -1,
			"data" : "缺少configType参数",
		})
		return
	}
	name := c.DefaultQuery("name", "")
	if name == "" {
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少name参数",
			"errorCode" : -2,
			"data" : "缺少name参数",
		})
		return
	}
	id := c.DefaultQuery("id", "")
	if id != "" {
		condition := "id=" + id
		var config *[]dal.ItemConfig
		config = dal.QueryConfigByCondition(condition)
		var item = (*config)[0]
		item.Name = name
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
		condition := "config_type=" + configType + " and name='" + name + "'"
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
		var configItem dal.ItemConfig
		configItem.ConfigType, _ = strconv.Atoi(configType)
		configItem.Name = name
		flag := dal.InsertItemConfig(configItem)
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
	data[0] = config0
	data[1] = config1
	data[2] = config2
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : data,
	})
}
