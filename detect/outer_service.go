package detect

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/gorm"
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
		ReturnMsg(c, FAILURE, "Miss APP id or version")
		return
	}

	db, err := database.GetDBConnection()
	if err != nil {
		ReturnMsg(c, FAILURE, fmt.Sprintf("Connect to DB failed: %v", err))
		return
	}
	defer db.Close()

	task, err := getLatestDetectResult(db, map[string]interface{}{
		"app_id":      appID,
		"app_version": appVersion,
		"platform":    0})
	if err != nil {
		ReturnMsg(c, FAILURE, "Failed to get binary detect result")
		return
	}

	data := getDetectResult(c, strconv.Itoa(int(task.ID)), "6")
	if data == nil {
		logs.Error("Failed to get task ID %v binary detect result", task.ID)
		return
	}

	extra, err := getExtraConfirmedDetection(db, map[string]interface{}{
		"app_id":   appID,
		"platform": 0})
	if err != nil {
		ReturnMsg(c, FAILURE, "Failed to get extra confirmed detections")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errorCode":    SUCCESS,
		"message":      "success",
		"data":         *data,
		"extraConfirm": extra})

	logs.Info("Get task ID %v binary detect result success", task.ID)
	return
}

// The status code of detection.
const (
	Unconfirmed   = 0
	ConfirmedPass = 1
	ConfirmedFail = 2
)

func getLatestDetectResult(db *gorm.DB, condition map[string]interface{}) (
	*dal.DetectStruct, error) {

	condition["status"] = ConfirmedPass
	result, err := retrieveLatestDetectResult(db, condition)
	if err != nil {
		// Return the lastest binary detect result if the binary detect
		// result of specific version was not found.
		delete(condition, "app_version")
		result, err = retrieveLatestDetectResult(db, condition)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// retrieveLatestDetectResult returns the first eligible record on success.
func retrieveLatestDetectResult(db *gorm.DB, condition map[string]interface{}) (
	*dal.DetectStruct, error) {

	var detect dal.DetectStruct
	if err := db.Debug().Where(condition).Order("created_at desc").
		First(&detect).Error; err != nil {
		logs.Error("Database error: %v", err)
		return nil, err
	}

	return &detect, nil
}

// The type of detection
const (
	TypePermission = "权限"
	TypeMethod     = "敏感方法"
	TypeString     = "敏感词汇"
)

func getExtraConfirmedDetection(db *gorm.DB, condition map[string]interface{}) (
	map[string]interface{}, error) {

	condition["confirmed"] = true
	detections, err := RetrieveDetection(db, condition)
	if err != nil {
		logs.Error("Failed to retrieve detection")
		return nil, err
	}

	result := make(map[string]interface{})
	var permissions []map[string]interface{}
	var methods []map[string]interface{}
	var strs []map[string]interface{}
	for i := range detections {
		m := make(map[string]interface{})
		m["configid"] = detections[i].DetectConfigID
		m["status"] = ConfirmedPass
		m["remark"] = ""
		m["confirmer"] = ""
		m["desc"] = detections[i].Description
		m["gpFlag"] = detections[i].GPFlag
		m["riskLevel"] = detections[i].RiskLevel
		if err := do(m, &detections[i]); err != nil {
			return nil, err
		}
		switch detections[i].Type {
		case TypePermission:
			m["key"] = detections[i].Key
			permissions = append(permissions, m)
		case TypeMethod:
			methods = append(methods, m)
		case TypeString:
			m["keys"] = detections[i].Key
			strs = append(strs, m)
		}
	}

	result["sMethods"] = methods
	result["newStrs"] = strs
	result["permissionList"] = permissions

	return result, nil
}

func do(m map[string]interface{}, detection *NewDetection) error {

	if detection.Type == TypeMethod || detection.Type == TypeString {
		var location []callLocation
		if err := json.Unmarshal(
			[]byte(detection.CallLocations), &location); err != nil {
			logs.Error("Unmarshal error: %v", err)
			return err
		}
		m["callLoc"] = location
	}

	if detection.Type == TypeMethod {
		k := strings.LastIndexByte(detection.Key, '.')
		m["className"] = detection.Key[:k]
		m["methodName"] = detection.Key[k+1:]
	}

	return nil
}
