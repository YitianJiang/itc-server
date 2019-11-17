package detect

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/settings"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

// UpdateDetectTask parses the response from the detect tool and stores it
// into database, then notifies the person or group who cares about the task.
// It makes no sense to send response to the detect tool
// because the detect tool doesn't care about the it at all.
func UpdateDetectTask(c *gin.Context) {

	msgHeader := fmt.Sprintf("update detect task (task id: %v)", c.Request.FormValue("task_ID"))
	task, err := getExactDetectTask(database.DB(),
		map[string]interface{}{"id": c.Request.FormValue("task_ID")})
	if err != nil {
		logs.Error("%s get detect task failed: %v", msgHeader, err)
		return
	}
	task.Status = TaskStatusUnconfirm
	if err := updateDetectTaskStatus(database.DB(), task); err != nil {
		logs.Error("%s update status failed: %v", msgHeader, err)
		return
	}

	toolID, err := strconv.Atoi(c.Request.FormValue("tool_ID"))
	if err != nil {
		logs.Error("%s tool id atoi error: %v", msgHeader, err)
		return
	}

	jsonContent := c.Request.FormValue("jsonContent")
	htmlContent := c.Request.FormValue("content")
	//消息通知条数--检测项+自查项
	var unConfirms int
	var unSelfCheck int
	if task.Platform == platformAndorid {
		var errApk error
		errApk, unConfirms = ParseResultAndroid(task, jsonContent, toolID)
		if errApk != nil {
			logs.Error("%s update apk detect result failed: %v", msgHeader, err)
			return
		}
	}
	//ios新检测内容存储
	if task.Platform == platformiOS {
		if err := database.InsertDBRecord(database.DB(), &dal.DetectContent{
			TaskId:      int(task.ID),
			ToolId:      toolID,
			HtmlContent: htmlContent,
			JsonContent: jsonContent,
		}); err != nil {
			logs.Error("%s store content error: %v", msgHeader, err)
			return
		}
		//新表jsonContent分类存储
		appId, _ := strconv.Atoi(task.AppId)
		res, warnFlag, detectNo := iOSResultClassify(task, toolID, &jsonContent) //检测结果处理
		unConfirms = detectNo
		if res == false {
			logs.Error("iOS 新增new detect content失败！！！") //防止影响现有用户，出错后暂不return
		}
		//iOS付费相关黑名单及时报警
		if res && warnFlag {
			tips := "Notice: " + task.AppName + " " + task.AppVersion + " iOS包完成二进制检测，检测黑名单中itms-services不为空，请及时关注！！！！\n"
			for _, lark_people := range _const.MiddleLarkPeople {
				utils.LarkDingOneInnerWithUrl(lark_people, tips, "点击跳转检测详情", fmt.Sprintf(settings.Get().Detect.TaskURL, task.AppId, task.ID))
			}
		}

		//获取未确认自查项数目
		var extra dal.ExtraStruct
		json.Unmarshal([]byte(task.ExtraInfo), &extra)
		skip := extra.SkipSelfFlag
		if !skip {
			isRight, selfNum := GetIOSSelfNum(appId, int(task.ID))
			if isRight {
				unSelfCheck = selfNum
			}
		}
	}

	go notifyDeteckTaskResult(task, &msgHeader, unConfirms, unSelfCheck)
}

// WARNING when update with struct, GORM will only update those fields
// that with non blank value. So it is nessary to update the status of
// detect task with a new function.
func updateDetectTaskStatus(db *gorm.DB, task *dal.DetectStruct) error {

	if err := db.Debug().Model(task).Update("status", task.Status).Error; err != nil {
		logs.Error("database error: %v", err)
		return err
	}

	return nil
}

func updateDetectTask(db *gorm.DB, task *dal.DetectStruct) error {

	if err := db.Debug().Model(task).Updates(task).Error; err != nil {
		logs.Error("database error: %v", err)
		return err
	}

	return nil
}

func notifyDeteckTaskResult(task *dal.DetectStruct, msgHeader *string, unConfirms int, unSelfCheck int) {

	var err error
	task, err = getExactDetectTask(database.DB(), map[string]interface{}{"id": task.ID})
	if err != nil {
		logs.Error("%s get detect task failed: %v", *msgHeader, err)
		return
	}

	message := "你好，包检测完成。\n应用名称：%s\n版本号：%s"
	message = fmt.Sprintf(message, task.AppName, task.AppVersion)
	var os string
	switch task.Platform {
	case platformAndorid:
		message += fmt.Sprintf("(%s)", task.InnerVersion)
		message += fmt.Sprintf("\n操作系统：%s", "Android")
		os = "1"
	case platformiOS:
		message += fmt.Sprintf("\n操作系统：%s", "iOS")
		os = "2"
	default:
		message += "unknow"
	}
	var qa_bm string
	var rd_bm string
	if project_id, ok := _const.AppVersionProject[task.AppId]; ok {
		rd, qa, err := GetVersionBMInfo(task.AppId, project_id, task.AppVersion, os)
		if err != nil {
			logs.Warn("%s get preject version failed: %v", *msgHeader, err)
		}
		rd_id := utils.GetUserOpenId(rd + "@bytedance.com")
		if rd_id != "" {
			rd_info := utils.GetUserAllInfo(rd_id)
			rd_map := make(map[string]interface{})
			json.Unmarshal([]byte(rd_info), &rd_map)
			rd_bm = rd_map["name"].(string)
		}
		qa_id := utils.GetUserOpenId(qa + "@bytedance.com")
		if qa_id != "" {
			qa_info := utils.GetUserAllInfo(qa_id)
			qa_map := make(map[string]interface{})
			json.Unmarshal([]byte(qa_info), &qa_map)
			qa_bm = qa_map["name"].(string)
		}
	}
	larkUrl := fmt.Sprintf(settings.Get().Detect.TaskURL, task.AppId, task.ID)
	for _, creator := range strings.Split(task.ToLarker, ",") {
		utils.UserInGroup(creator)                                                                                      //将用户拉入预审平台群
		res := utils.LarkDetectResult(task.ID, creator, rd_bm, qa_bm, message, larkUrl, unConfirms, unSelfCheck, false) //new lark卡片通知形式
		logs.Info("%s creator: %s lark message result: %v", *msgHeader, creator, res)
	}
	//发给群消息沿用旧的机器人，给群ID对应群发送消息
	toGroupID := task.ToGroup
	if toGroupID != "" {
		group := strings.Replace(toGroupID, "，", ",", -1) //中文逗号切换成英文逗号
		groupArr := strings.Split(group, ",")
		for _, group_id := range groupArr {
			to_lark_group := strings.Trim(group_id, " ")
			if utils.LarkDetectResult(task.ID, to_lark_group, rd_bm, qa_bm, message, larkUrl, unConfirms, unSelfCheck, true) == false {
				message += message + larkUrl
				utils.LarkGroup(message, to_lark_group)
			}
		}
	}
}

func GetVersionBMInfo(biz, project, version, os_type string) (rd string, qa string, err error) {
	version_arr := strings.Split(version, ".")
	//TikTok这类型版本号：122005 无法获取BM信息
	if len(version_arr) < 3 {
		return "", "", fmt.Errorf("unsupported version format: %v", version)
	}
	new_version := version_arr[0] + "." + version_arr[1] + "." + version_arr[2]
	body, err := utils.SendHTTPRequest("GET",
		settings.Get().RocketAPI.ProjectVersionURL,
		map[string]string{
			"project":      project,
			"biz":          biz,
			"achieve_type": os_type,
			"version_code": new_version,
			"nextpage":     "1"},
		map[string]string{
			"token": settings.Get().RocketAPI.Token}, nil)
	if err != nil {
		logs.Error("send http request error: %v", err)
		return "", "", err
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(body, &m); err != nil {
		logs.Error("unmarshal error: %v", err)
		return "", "", err
	}
	if fmt.Sprint(m["errorCode"]) != "0" {
		logs.Error("invalid response: %s", body)
		return "", "", fmt.Errorf("invalid response: %s", body)
	}
	versionInfo, ok := m["data"].(map[string]interface{})["VersionCards"].([]interface{})
	if !ok {
		logs.Error("cannot assert to []interface{}: %v", m["data"].(map[string]interface{})["VersionCards"])
		return "", "", fmt.Errorf("cannot assert to []interface{}: %v", m["data"].(map[string]interface{})["VersionCards"])
	}
	if len(versionInfo) == 0 {
		return "", "", fmt.Errorf("VersionCards is empty: %v", m["data"].(map[string]interface{})["VersionCards"])
	}
	versionParam, ok := versionInfo[0].(map[string]interface{})["Param_ext"].(string)
	if !ok {
		logs.Error("cannot assert to string: %v", versionInfo[0].(map[string]interface{})["Param_ext"])
		return "", "", fmt.Errorf("cannot assert to string: %v", versionInfo[0].(map[string]interface{})["Param_ext"])
	}
	var l []interface{}
	if err = json.Unmarshal([]byte(versionParam), &l); err != nil {
		logs.Error("unmarshal error: %v", err)
		return "", "", fmt.Errorf("unmarshal error: %v", err)
	}
	var rd_bm string
	var qa_bm string
	for _, bm := range l {
		if bm.(map[string]interface{})["Param_desc"].(string) == "RD BM" || bm.(map[string]interface{})["Param_desc"].(string) == "QA BM" {
			if bm.(map[string]interface{})["Param_desc"].(string) == "RD BM" {
				rd_bm = bm.(map[string]interface{})["Value"].(string)
			} else {
				qa_bm = bm.(map[string]interface{})["Value"].(string)
			}
			if rd_bm != "" && qa_bm != "" {
				break
			} else {
				continue
			}
		}
	}

	return rd_bm, qa_bm, nil
}
