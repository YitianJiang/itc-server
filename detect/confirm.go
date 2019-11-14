package detect

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

type confirmaParams struct {
	TaskID int    `json:"taskId"`
	ToolID int    `json:"toolId"`
	Status int    `json:"status"`
	Remark string `json:"remark"`

	// Only used in Android.
	// ID is the id of table tb_detect_content_detail if the type is method or string.
	ID int `json:"id"`
	// Index is the array index of table tb_perm_app_relation's field perm_infos
	// if the type is permission. And must use Index-1 because it starts from one.
	Index int `json:"index"`
	// 0-->method/string 1-->permission
	TypeAndroid int `json:"type"`

	// Only used in iOS.
	// 1-->blacklist(string) 2-->method 3-->privacy(permission)
	TypeiOS int `json:"confirmType"`
	// Name=methodName+className if the type is method.
	Name string `json:"confirmContent"`
}

// ConfirmAndroid supports .apk and .aab format.
func ConfirmAndroid(c *gin.Context) {

	username, exist := c.Get("username")
	if !exist {
		utils.ReturnMsg(c, http.StatusUnauthorized, utils.FAILURE, "unauthorized user")
		return
	}

	var t dal.PostConfirm
	if err := c.ShouldBindJSON(&t); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid parameter: %v", err))
		return
	}

	task, err := getExactDetectTask(database.DB(), map[string]interface{}{"id": t.TaskId})
	if err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid task id (%v): %v", err, t.TaskId))
		return
	}
	msgHeader := fmt.Sprintf("task id: %v", task.ID)

	if t.Type == 0 { //敏感方法和字符串确认
		detection, err := readExactDetectContentDetail(database.DB(),
			map[string]interface{}{"id": t.Id})
		if err != nil {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("%s read detection failed: %v", msgHeader, err))
			return
		}
		if detection == nil {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("%s invalid detection id (%v)", msgHeader, t.Id))
			return
		}

		itemName := detection.KeyInfo
		var itemType *string
		switch detection.SensiType {
		case Permission:
			itemType = &TypePermission
		case Method:
			itemType = &TypeMethod
			itemName = detection.ClassName + delimiter + detection.KeyInfo
		case String:
			itemType = &TypeString
		}
		if err := preAutoConfirmTask(task,
			&Item{Name: itemName, Type: itemType},
			t.Status, username.(string), t.Remark, detection.SubIndex, t.ToolId); err != nil {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("confirm Android detection failed: %v", err))
			return
		}
	} else { //获取该任务的权限信息
		record, err := readExactPermAPPRelation(database.DB(), map[string]interface{}{
			"task_id": t.TaskId, "sub_index": t.Index - 1})
		if err != nil {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("%s read perm_app_relation error: %v", msgHeader, err))
			return
		}
		var permissionList []interface{}
		if err := json.Unmarshal([]byte(record.PermInfos), &permissionList); err != nil {
			logs.Error("unmarshal error: %v", err)
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("unmarshal error: %v", err))
			return
		}
		if t.Id > len(permissionList) {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid id: %v", t.Id))
			return
		}
		m, ok := permissionList[t.Id-1].(map[string]interface{})
		if !ok {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("cannot assert to map[string]interface{}: %v", permissionList[t.Id-1]))
			return
		}

		if err := preAutoConfirmTask(task,
			&Item{Name: m["key"].(string), Type: &TypePermission},
			t.Status, username.(string), t.Remark, t.Index-1, t.ToolId); err != nil {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("confirm Android detection failed:%v", err))
			return
		}
	}

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success")
	return
}

type IOSConfirm struct {
	TaskId         int    `json:"taskId"           form:"taskId"`
	ToolId         int    `json:"toolId"           form:"toolId"`
	Status         int    `json:"status"           form:"status"`
	Remark         string `json:"remark"           form:"remark"`
	ConfirmType    int    `json:"confirmType"      form:"confirmType"` //0是旧样式黑名单，1是新样式黑名单，2是可疑方法，3是权限
	ConfirmContent string `json:"confirmContent"   form:"confirmContent"`
}

// ConfirmiOS supports .ipa format.
func ConfirmiOS(c *gin.Context) {

	username, exist := c.Get("username")
	if !exist {
		utils.ReturnMsg(c, http.StatusUnauthorized, utils.FAILURE, "unauthorized user")
		return
	}
	var p confirmaParams
	if err := c.ShouldBindJSON(&p); err != nil {
		// var ios IOSConfirm
		// if err := c.ShouldBindJSON(&ios); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid user: %v", err))
		return
	}

	task, err := getExactDetectTask(database.DB(), map[string]interface{}{"id": p.TaskID})
	// task, err := getExactDetectTask(database.DB(), map[string]interface{}{"id": ios.TaskId})
	if err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("read tb_binary_detect failed: %v", err))
		return
	}
	itemName := p.Name
	// itemName := ios.ConfirmContent
	var itemType *string
	switch p.TypeiOS {
	// switch ios.ConfirmType {
	case 1:
		itemType = &TypeString
	case 2:
		itemType = &TypeMethod
		i := strings.Index(itemName, "+")
		itemName = itemName[i+1:] + delimiter + itemName[:i]
	case 3:
		itemType = &TypePermission
	}
	if err := preAutoConfirmTask(task, &Item{
		Name: itemName,
		Type: itemType},
		p.Status, username.(string), p.Remark, 0, p.ToolID); err != nil {
		// ios.Status, username.(string), ios.Remark, 0, ios.ToolId); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("confirm iOS detection failed: %v", err))
		return
	}

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success")
}
