package detect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

type outerConfig struct {
	ConfigID            int       `json:"configid"`
	CreatedAt           time.Time `json:"CreatedAt"`
	MethodClass         string    `json:"method_class"`
	MethodName          string    `json:"method_name"`
	SearchKeys          []string  `json:"search_keys"`
	Keys                string    `json:"keys"`
	StartSearchMethod   int       `json:"start_search_method"`
	ContainSearchMethod int       `json:"contain_search_method"`
	Description         string    `json:"desc"`
	GPFlag              string    `json:"gpflag"`
	Ability             string    `json:"ability"`
	Suggestion          string    `json:"suggestion"`
	Level               int       `json:"level"`
	Creator             string    `json:"creator"`
	DetectType          string    `json:"detectType"`
	Reference           string    `json:"refer"`
	SensitiveFlag       int       `json:"sensiFlag"`
	Permissions         []string  `json:"permissions"`
}

type outerData struct {
	MethodSearchKey     []outerConfig `json:"methodSearchKey"`
	StrSearchKey        []outerConfig `json:"strSearchKey"`
	PermissionSearchKey []outerConfig `json:"permissionSearchKey"`
	// TODO
	ComposeSearchInOneMethod []outerConfig `json:"composeSearchInOneMethod"`
}

// GetDetectConfig retrieves all eligible configures from tb_detect_config table
// and returns them to the requestor.
func GetDetectConfig(c *gin.Context) {
	platform, exist := c.GetQuery("platform")
	if !exist {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   "platform was not found",
			"data":      "platform was not found"})

		return
	}

	switch {
	case platform == "0": // Android
		detectConfig(c, 0)
	// case platform == "1": // iOS
	// 	detectConfig(c, 1)
	default:
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   "platform doesn't support",
			"data":      "platform doesn't support"})
	}

	return
}

func detectConfig(c *gin.Context, platform int) {
	permissions, methods, strs, composites := getConfigList(platform)
	if permissions == nil || methods == nil ||
		strs == nil || composites == nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   "database error",
			"data":      "database error"})

		return
	}

	data := outerData{pack(methods, 1), pack(strs, 2),
		pack(permissions, 0), nil}
	c.JSON(http.StatusOK, gin.H{
		"errorCode": 0,
		"message":   "success",
		"data":      data})

	return
}

func getConfigList(platform int) (
	permissions, methods, strs, composites *[]dal.DetectConfigStruct) {
	permissions = dal.QueryDetectConfig(map[string]interface{}{
		"platform":   platform,
		"check_type": 0})
	methods = dal.QueryDetectConfig(map[string]interface{}{
		"platform":   platform,
		"check_type": 1})
	strs = dal.QueryDetectConfig(map[string]interface{}{
		"platform":   platform,
		"check_type": 2})
	composites = dal.QueryDetectConfig(map[string]interface{}{
		"platform":   platform,
		"check_type": 3})

	return
}

func pack(origin *[]dal.DetectConfigStruct, checkType int) []outerConfig {
	var packed []outerConfig

	for _, o := range *origin {
		var t outerConfig
		switch checkType {
		case 0:
			t.Keys = o.KeyInfo
		case 1:
			k := strings.LastIndex(o.KeyInfo, ".")
			if k != -1 {
				t.MethodClass = o.KeyInfo[:k]
				t.MethodName = o.KeyInfo[k+1:]
			}
			if o.Permissions != "" {
				t.Permissions = strings.Split(o.Permissions, ";")
			}
		case 2:
			t.SearchKeys = strings.Split(o.KeyInfo, ";")
			if o.Permissions != "" {
				t.Permissions = strings.Split(o.Permissions, ";")
			}
			// TODO
			// case 3:
		}
		t.Description = o.DescInfo
		t.GPFlag = strconv.Itoa(o.GpFlag)
		t.Ability = o.Ability
		t.Suggestion = o.Suggestion
		t.Level = o.Priority
		t.ConfigID = int(o.ID)
		t.CreatedAt = o.CreatedAt
		t.Creator = o.Creator
		t.DetectType = o.DetectType
		t.Reference = o.Reference
		t.SensitiveFlag = o.SensiFlag

		packed = append(packed, t)
	}

	return packed
}

func AddConfig(c *gin.Context) {
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.ItemConfig
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("json unmarshal failed!, ", err)
		c.JSON(http.StatusOK, gin.H{
			"message":   "json unmarshal failed",
			"errorCode": -5,
			"data":      "json unmarshal failed",
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
				"message":   "配置更新失败！",
				"errorCode": -5,
				"data":      "配置更新失败！",
			})
			return
		}
	} else {
		condition := "config_type='" + fmt.Sprint(t.ConfigType) + "' and name='" + t.Name + "'" + " and platform='" + fmt.Sprint(t.Platform) + "'"
		var config *[]dal.ItemConfig
		config = dal.QueryConfigByCondition(condition)
		if config != nil && len(*config) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"message":   "该配置参数已经存在！",
				"errorCode": -3,
				"data":      "该配置参数已经存在！",
			})
			return
		}
		flag := dal.InsertItemConfig(t)
		if !flag {
			c.JSON(http.StatusOK, gin.H{
				"message":   "新增配置失败！",
				"errorCode": -4,
				"data":      "新增配置失败！",
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      "success",
	})
}

/**
 *查询配置项
 */
func QueryConfigs(c *gin.Context) {

	condition := "1=1"
	var config *[]dal.ItemConfig
	config = dal.QueryConfigByCondition(condition)
	var config0 []dal.ItemConfig
	var config1 []dal.ItemConfig
	var config2 []dal.ItemConfig
	if config == nil || len(*config) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未查询到配置信息！",
			"errorCode": -1,
			"data":      "未查询到配置信息！",
		})
		return
	}
	for i := 0; i < len(*config); i++ {
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
		"message":   "success",
		"errorCode": 0,
		"data":      data,
	})
}

/**
 *查询平台所配置的问题类型数据
 */
func QueryProblemConfigs(c *gin.Context) {
	platform := c.DefaultQuery("platform", "")
	if platform == "" {
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少platform参数！",
			"errorCode": -1,
			"data":      "缺少platform参数！",
		})
		return
	}
	if platform != "0" && platform != "1" {
		c.JSON(http.StatusOK, gin.H{
			"message":   "platform参数不合法！",
			"errorCode": -2,
			"data":      "platform参数不合法！",
		})
		return
	}
	condition := "platform='" + platform + "' and config_type='0'"
	var config *[]dal.ItemConfig
	config = dal.QueryConfigByCondition(condition)
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      config,
	})
}
