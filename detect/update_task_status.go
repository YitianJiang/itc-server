package detect

import (
	"encoding/json"
	"fmt"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
)

func updateTaskStatus(taskID, toolID interface{}, platform int, confirmLark int) (int, error) {

	header := fmt.Sprintf("task id: %v", taskID)
	var unconfirmed int
	var fail int
	var err error
	switch platform {
	case platformAndorid:
		unconfirmed, _, fail, err = taskDetailAndroid(taskID, toolID)
	case platformiOS:
		unconfirmed, _, fail, err = taskDetailiOS(taskID, toolID)
	default:
		return unconfirmed, fmt.Errorf("%s unsupported platform: %v", header, platform)
	}
	if err != nil {
		logs.Error("%s get iOS task detail failed: %v", header, err)
		return unconfirmed, err
	}
	if fail <= 0 && unconfirmed > 0 {
		return unconfirmed, nil
	}

	task, err := getExactDetectTask(database.DB(), map[string]interface{}{"id": taskID})
	if err != nil {
		logs.Error("%s get detect task failed: %v", header, err)
		return unconfirmed, err
	}
	if unconfirmed <= 0 {
		task.Status = ConfirmedPass
	}
	if fail > 0 {
		task.Status = ConfirmedFail
		task.DetectNoPass = fail
	}
	if err := dal.UpdateDetectModelNew(*task); err != nil {
		logs.Error("%s update detect task failed: %v", header, err)
		return unconfirmed, err
	}
	StatusDeal(*task, confirmLark) //ci回调和不通过block处理

	return unconfirmed, nil
}

func taskDetailAndroid(taskID interface{}, toolID interface{}) (int, int, int, error) {

	unconfirmed := 0
	pass := 0
	fail := 0
	header := fmt.Sprintf("task id: %v", taskID)
	details, err := readDetectContentDetail(database.DB(), map[string]interface{}{
		"task_id": taskID, "tool_id": toolID})
	if err != nil {
		logs.Error("%s read tb_detect_content_detail failed: %v", header, err)
		return unconfirmed, pass, fail, err
	}
	for i := range details {
		switch details[i].Status {
		case ConfirmedPass:
			pass++
		case ConfirmedFail:
			fail++
		default:
			unconfirmed++
		}
	}

	permissions, err := readPermAPPRelation(database.DB(), map[string]interface{}{
		"task_id": taskID})
	if err != nil {
		logs.Error("%s read tb_perm_app_relation failed: %v", header, err)
		return unconfirmed, pass, fail, err
	}
	for i := range permissions {
		var list []interface{}
		if err := json.Unmarshal([]byte(permissions[i].PermInfos), &list); err != nil {
			logs.Error("%s unmarshal error: %v content: %s", header, err, permissions[i].PermInfos)
			return unconfirmed, pass, fail, err
		}
		for j := range list {
			m, ok := list[j].(map[string]interface{})
			if !ok {
				logs.Error("%s cannot assert to map[string]interface{}: %v", header, list[j])
				return unconfirmed, pass, fail, fmt.Errorf("%s cannot assert to map[string]interface{}: %v", header, list[j])

			}
			switch fmt.Sprint(m["status"]) {
			case Pass:
				pass++
			case Fail:
				fail++
			default:
				unconfirmed++
			}
		}
	}

	return unconfirmed, pass, fail, nil
}

func taskDetailiOS(taskID interface{}, toolID interface{}) (int, int, int, error) {

	header := fmt.Sprintf("task id: %v", taskID)
	unconfirmed := 0
	pass := 0
	fail := 0
	content, err := dal.QueryNewIOSDetectModel(database.DB(), map[string]interface{}{
		"taskId": taskID,
		"toolId": toolID,
	})
	if err != nil {
		logs.Error("%s read iOS detect content failed: %v", header, err)
		return unconfirmed, pass, fail, err
	}
	for i := range content {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(content[i].DetectContent), &m); err != nil {
			logs.Error("%s unmarshal error: %v", header, err)
		}
		keyword := content[i].DetectType
		if keyword == "blacklist" {
			keyword = "blackList"
		}
		list, ok := m[keyword].([]interface{})
		if !ok {
			logs.Error("%s cannot assert to []interface{}: %v", header, m[keyword])
			return unconfirmed, pass, fail, fmt.Errorf("%s cannot assert to []interface{}: %v", header, m[keyword])
		}
		for j := range list {
			t, ok := list[j].(map[string]interface{})
			if !ok {
				logs.Error("%s cannot assert to map[string]interface{}: %v", header, list[j])
				return unconfirmed, pass, fail, fmt.Errorf("%s cannot assert to map[string]interface{}: %v", header, list[j])
			}
			switch fmt.Sprint(t["status"]) {
			case "1":
				pass++
			case "2":
				fail++
			default:
				unconfirmed++
			}
		}
	}

	return unconfirmed, pass, fail, nil
}
