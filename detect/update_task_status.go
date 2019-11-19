package detect

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
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
	StatusDeal(task, confirmLark) //ci回调和不通过block处理

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

//全部确认完成后处理
//confirmLark 0:检测完成diff时，1：确认检测结果，2：确认自查结果
func StatusDeal(task *dal.DetectStruct, confirmLark int) error {
	//ci回调
	if task.Status == 1 && (task.Platform == 0 || task.SelfCheckStatus == 1) {
		if err := callbackCI(task); err != nil {
			logs.Error("回调ci出错！", err.Error())
			return err
		}
	}
	if task.Status != 0 && (task.Platform == 0 || task.SelfCheckStatus != 0) {
		//diff时调用，不用发冗余消息提醒
		if confirmLark == 0 {
			return nil
		}
		//结果通知
		go func() {
			selfNoPass := task.SelftNoPass
			detectNoPass := task.DetectNoPass
			message := "你好，" + task.AppName + " " + task.AppVersion
			if task.Platform == 0 {
				message += " Android包"
			} else {
				message += " iOS包"
			}
			message += "  已经确认完毕！"
			url := "http://rocket.bytedance.net/rocket/itc/task?biz=" + task.AppId + "&showItcDetail=1&itcTaskId=" + strconv.Itoa(int(task.ID))
			lark_people := task.ToLarker
			peoples := strings.Replace(lark_people, "，", ",", -1)
			lark_people_arr := strings.Split(peoples, ",")
			for _, p := range lark_people_arr {
				utils.LarkConfirmResult(strings.TrimSpace(p), message, url, detectNoPass, selfNoPass, false)
			}
			lark_group := task.ToGroup
			groups := strings.Replace(lark_group, "，", ",", -1)
			lark_group_arr := strings.Split(groups, ",")
			for _, g := range lark_group_arr {
				utils.LarkConfirmResult(strings.TrimSpace(g), message, url, detectNoPass, selfNoPass, true)
			}
		}()
	}
	return nil
}

func callbackCI(task *dal.DetectStruct) error {
	if task.Platform == 1 && (task.SelfCheckStatus != 1 || task.Status != 1) {
		logs.Info("不满足callback条件")
		return nil
	}
	var t dal.ExtraStruct
	//兼容旧信息---无extra_info字段
	if task.ExtraInfo == "" {
		return nil
	}
	err := json.Unmarshal([]byte(task.ExtraInfo), &t)
	if err != nil {
		logs.Error("任务附加信息存储格式错误，任务ID：" + fmt.Sprint(task.ID))
		utils.LarkDingOneInner("fanjuan.xqp", "任务附加信息存储格式错误，任务ID："+fmt.Sprint(task.ID))
		return err
	}
	//无回调地址（页面上传），不需要进行回调
	if t.CallBackAddr == "" {
		return nil
	}
	urlInfos := strings.Split(t.CallBackAddr, "?")
	workflow_id := ""
	job_id := ""
	if len(urlInfos) > 1 {
		queryInfos := getUrlInfo(urlInfos[1])
		if v, ok := queryInfos["workflow_id"]; ok {
			workflow_id = v
		}
		if v, ok := queryInfos["job_id"]; ok {
			job_id = v
		}
	}

	//回调CI接口，发送post请求
	data := make(map[string]string)
	data["workflow_id"] = workflow_id
	data["job_id"] = job_id
	data["statsu"] = "2"
	data["task_id"] = fmt.Sprint(task.ID)
	url := urlInfos[0]
	return PostInfos(url, data)
}
