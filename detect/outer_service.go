package detect

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
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
		msg := "Miss platform"
		logs.Error(msg)
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   msg,
			"data":      msg})

		return
	}

	switch {
	case platform == "0": // Android
		detectConfig(c, 0)
	// case platform == "1": // iOS
	// 	detectConfig(c, 1)
	default:
		msg := "Do not support platform: " + platform
		logs.Error(msg)
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   msg,
			"data":      msg})
	}

	logs.Debug("Get detect config success")
	return
}

func detectConfig(c *gin.Context, platform int) {
	permissions, methods, strs, composites := getConfigList(platform)
	if permissions == nil || methods == nil ||
		strs == nil || composites == nil {
		msg := "Database error"
		logs.Error(msg)
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   msg,
			"data":      msg})

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

// GetSpecificAppVersionDetectResults retrieves lastest eligible record
// from database and returns it to the requestor.
func GetSpecificAppVersionDetectResults(c *gin.Context) {
	appID, idExist := c.GetQuery("appId")
	appVersion, versionExist := c.GetQuery("appVersion")
	if !idExist || !versionExist {
		msg := "Miss APP id or version"
		logs.Error(msg)
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   msg,
			"data":      msg})

		return
	}

	task := queryLastestDetectResult(map[string]interface{}{
		"app_id":      appID,
		"app_version": appVersion,
		"platform":    0})
	if task == nil {
		msg := "Failed to find binary detect result in database about" +
			" APP ID is " + appID + " and Version is " + appVersion
		logs.Error(msg)
		errorReturn(c, msg)
		return
	}

	result := getDetectResult(c, strconv.Itoa(int(task.ID)), "6")
	if result == nil {
		logs.Error("Failed to get task ID %v binary detect result", task.ID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errorCode":    0,
		"message":      "success",
		"data":         *result,
		"extraConfirm": nil})

	logs.Debug("Get task ID %v binary detect result success", task.ID)
	return
}

func queryLastestDetectResult(param map[string]interface{}) *dal.DetectStruct {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer connection.Close()

	var detect dal.DetectStruct
	db := connection.Table(dal.DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(param).Order("created_at desc").First(&detect).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}

	return &detect
}
