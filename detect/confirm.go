package detect

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

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
			t.Status, username.(string), t.Remark, t.Index); err != nil {
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
		m["status"] = t.Status
		m["confirmer"] = username.(string)
		m["remark"] = t.Remark
		data, err := json.Marshal(permissionList)
		if err != nil {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("marshal error: %v", err))
			return
		}
		record.PermInfos = string(data)
		if err := database.UpdateDBRecord(database.DB(), record); err != nil {
			utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("confirm Android detection failed: %v", err))
			return
		}

		// //写入操作历史
		// var history dal.PermHistory
		// history.Status = t.Status
		// history.AppVersion = (*perms)[t.Index-1].AppVersion
		// history.AppId = (*perms)[t.Index-1].AppId
		// history.PermId = permId
		// history.Remarks = t.Remark
		// history.Confirmer = username.(string)
		// history.TaskId = t.TaskId
		// if err := dal.InsertPermOperationHistory(history); err != nil {
		// 	logs.Error("taskId:" + fmt.Sprint(t.TaskId) + ",权限操作历史写入失败！")
		// 	errorReturn(c, "权限操作历史写入失败！")
		// 	return
		// }
	}

	//是否更新任务表中detect_no_pass字段的标志
	var notPassFlag = false
	if t.Status == 2 {
		notPassFlag = true
	}

	//任务状态更新
	updateInfo, _ := taskStatusUpdate(t.TaskId, t.ToolId, task, notPassFlag, 1)
	if updateInfo != "" {
		errorReturn(c, updateInfo)
		return
	}

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success")
	return
}
