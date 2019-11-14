package detect

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

type confirmParams struct {
	TaskID uint   `json:"taskId"`
	ToolID int    `json:"toolId"`
	Status int    `json:"status"`
	Remark string `json:"remark"`

	// Only used in Android.
	// ID is the id of table tb_detect_content_detail if the type is method or string.
	// ID is the array index of table tb_perm_app_relation's field perm_infos
	// if the type is permission. And must use ID-1 because it starts from 1.
	ID int `json:"id"`
	// Index corresponds to the table tb_perm_app_relation's field perm_infos
	// because .aab package may have more than one record in single task.
	Index int `json:"index"`
	// 0-->method/string 1-->permission
	TypeAndroid int `json:"type"`

	// Only used in iOS.
	// 1-->blacklist(string) 2-->method 3-->privacy(permission)
	TypeiOS int `json:"confirmType"`
	// Name=methodName+className if the type is method.
	Name string `json:"confirmContent"`

	// The field will be filled in the code.
	APPID      string
	Platform   int
	APPVersion string
	Item       *Item
	Confirmer  string
}

// ConfirmAndroid supports .apk and .aab format.
func ConfirmAndroid(c *gin.Context) {

	username, exist := c.Get("username")
	if !exist {
		utils.ReturnMsg(c, http.StatusUnauthorized, utils.FAILURE, "unauthorized user")
		return
	}

	var p confirmParams
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid parameter: %v", err))
		return
	}

	task, err := getExactDetectTask(database.DB(), map[string]interface{}{"id": p.TaskID})
	if err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid task id (%v): %v", err, p.TaskID))
		return
	}
	msgHeader := fmt.Sprintf("task id: %v", task.ID)

	p.Confirmer = username.(string)
	p.APPID = task.AppId
	p.Platform = task.Platform
	p.APPVersion = task.AppVersion
	if p.TypeAndroid == 0 { //敏感方法和字符串确认
		detection, err := readExactDetectContentDetail(database.DB(),
			map[string]interface{}{"id": p.ID})
		if err != nil {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("%s read detection failed: %v", msgHeader, err))
			return
		}
		if detection == nil {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("%s invalid detection id (%v)", msgHeader, p.ID))
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
		p.Item = &Item{Name: itemName, Type: itemType}
		p.Index = detection.SubIndex
	} else { //获取该任务的权限信息
		record, err := readExactPermAPPRelation(database.DB(), map[string]interface{}{
			"task_id": p.TaskID, "sub_index": p.Index - 1})
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
		if p.ID > len(permissionList) {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid id: %v", p.ID))
			return
		}
		m, ok := permissionList[p.ID-1].(map[string]interface{})
		if !ok {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("cannot assert to map[string]interface{}: %v", permissionList[p.ID-1]))
			return
		}
		p.Item = &Item{Name: m["key"].(string), Type: &TypePermission}
		p.Index--
	}

	if err := preAutoConfirmTask(&p); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("confirm Android detection failed: %v", err))
		return
	}

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success")
	return
}

// ConfirmiOS supports .ipa format.
func ConfirmiOS(c *gin.Context) {

	username, exist := c.Get("username")
	if !exist {
		utils.ReturnMsg(c, http.StatusUnauthorized, utils.FAILURE, "unauthorized user")
		return
	}
	var p confirmParams
	if err := c.ShouldBindJSON(&p); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid parameter: %v", err))
		return
	}

	task, err := getExactDetectTask(database.DB(), map[string]interface{}{"id": p.TaskID})
	if err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("read tb_binary_detect failed: %v", err))
		return
	}
	p.Confirmer = username.(string)
	p.APPID = task.AppId
	p.Platform = task.Platform
	p.APPVersion = task.AppVersion

	itemName := p.Name
	var itemType *string
	switch p.TypeiOS {
	case 1:
		itemType = &TypeString
	case 2:
		itemType = &TypeMethod
		i := strings.Index(itemName, "+")
		itemName = itemName[i+1:] + delimiter + itemName[:i]
	case 3:
		itemType = &TypePermission
	}
	p.Item = &Item{Name: itemName, Type: itemType}
	if err := preAutoConfirmTask(&p); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("confirm Android detection failed: %v", err))
		return
	}

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success")
	return
}
