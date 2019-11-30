package detect

import (
	"encoding/json"
	"strconv"
	"strings"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
)

const delimiter = "."
const toolIDiOS = 5
const toolIDAndroid = 6

// MR confirm
// func preAutoConfirmMR(appID string, platform int, version string,
// 	item *Item, status int, who string, remark string) error {

// 	freshman, err := preAutoConfirm(appID, platform, version, item, who, status, remark)
// 	if err != nil {
// 		logs.Error("pre auto confirm failed: %v", err)
// 		return err
// 	}

// 	// Distribute the confirm result.
// 	if !freshman {
// 		// TODO
// 	}

// 	return nil
}

// // Confirm binary package detect task.
// func autoConfirmTask(p *confirmParams) error {
// 	freshman, err := preAutoConfirm(p.APPID, p.Platform, p.APPVersion,
// 		p.Item, p.Confirmer, p.Status, p.Remark)
// 	if err != nil {
// 		logs.Error("pre auto confirm failed: %v", err)
// 		return err
// 	}

// 	// Distribute the confirm result.
// 	if !freshman {
// 		switch p.Platform {
// 		case platformAndorid:
// 			return autoConfirmAndroid(p)
// 		case platformiOS:
// 			return autoConfirmiOS(p)
// 		default:
// 			return fmt.Errorf("unsupported platform: %v", p.Platform)
// 		}
// 	}

// 	return nil
// }

// func autoConfirmAndroid(p *confirmParams) error {

// 	switch *p.Item.Type {
// 	case TypePermission:
// 		return autoConfirmAndroidEx(p, true)
// 	case TypeString:
// 		fallthrough
// 	case TypeMethod:
// 		return autoConfirmAndroidEx(p, false)
// 	default:
// 		return fmt.Errorf("unsupported platform: %v", p.Platform)
// 	}
// }

// func autoConfirmAndroidEx(p *confirmParams, tag bool) error {

// 	tasks, err := readDetectTask(database.DB(),
// 		map[string]interface{}{
// 			"app_id":      p.APPID,
// 			"platform":    p.Platform,
// 			"app_version": p.APPVersion})
// 	if err != nil {
// 		logs.Error("read tb_binary_detect failed: %v", err)
// 		return err
// 	}

// 	sync := make(chan error, 1)
// 	for i := range tasks {
// 		go func(task *dal.DetectStruct) {
// 			sieve := make(map[string]interface{})
// 			sieve["task_id"] = task.ID
// 			var updated bool
// 			var err error
// 			if tag {
// 				// Permission should be confirmed in each package
// 				// even they are same because it's very sensitive,
// 				// but method and list don't have to.
// 				sieve["sub_index"] = p.Index
// 				updated, err = autoConfirmAndroidPermission(p, sieve)
// 			} else {
// 				updated, err = autoConfirmAndroidStringMethod(p, sieve)
// 			}
// 			if err != nil {
// 				logs.Error("task id: %v auto confirm failed: %v", task.ID, err)
// 				if task.ID == p.TaskID {
// 					sync <- err
// 				}
// 				return
// 			}
// 			if !updated {
// 				if task.ID == p.TaskID {
// 					sync <- nil
// 				}
// 				return
// 			}
// 			_, err = updateTaskStatus(p.TaskID, p.ToolID, platformAndorid, 1)
// 			if err != nil {
// 				logs.Error("task id: %v update task status failed: %v", task.ID, err)
// 				if task.ID == p.TaskID {
// 					sync <- fmt.Errorf("update task status failed: %v", err)
// 				}
// 				return
// 			}
// 			if task.ID == p.TaskID {
// 				sync <- nil
// 			}
// 		}(&tasks[i])
// 	}

// 	return <-sync
// }

// func autoConfirmAndroidStringMethod(p *confirmParams, sieve map[string]interface{}) (bool, error) {

// 	switch *p.Item.Type {
// 	case TypeString:
// 		sieve["sensi_type"] = String
// 		sieve["key_info"] = p.Item.Name
// 	case TypeMethod:
// 		k := strings.LastIndex(p.Item.Name, delimiter)
// 		sieve["sensi_type"] = Method
// 		sieve["class_name"] = p.Item.Name[:k]
// 		sieve["key_info"] = p.Item.Name[k+1:]
// 	}

// 	records, err := readDetectContentDetail(database.DB(), sieve)
// 	if err != nil {
// 		logs.Error("read tb_detect_content_detail failed: %v", err)
// 		return false, err
// 	}
// 	for i := range records {
// 		records[i].Status = p.Status
// 		records[i].Confirmer = p.Confirmer
// 		records[i].Remark = p.Remark
// 		if err := database.UpdateDBRecord(database.DB(), &records[i]); err != nil {
// 			logs.Error("update tb_detect_content_detail failed: %v", err)
// 			return false, err
// 		}
// 	}

// 	return true, nil
// }

// func autoConfirmAndroidPermission(p *confirmParams, sieve map[string]interface{}) (bool, error) {

// 	record, err := readExactPermAPPRelation(database.DB(), sieve)
// 	if err != nil {
// 		logs.Error("read perm_app_relation failed: %v", err)
// 		return false, err
// 	}
// 	if record == nil {
// 		// It's ok because the tasks selected contain fail and detecting state.
// 		logs.Warn("cannot find any matched record with sieve: %v", sieve)
// 		return false, nil
// 	}
// 	var permissionList []interface{}
// 	if err := json.Unmarshal([]byte(record.PermInfos), &permissionList); err != nil {
// 		logs.Error("unmarshal error: %v", err)
// 		return false, err
// 	}
// 	var updated bool
// 	for j := range permissionList {
// 		t, ok := permissionList[j].(map[string]interface{})
// 		if !ok {
// 			logs.Error("cannot assert to map[string]interface{}: %v", permissionList[j])
// 			return false, err
// 		}
// 		if t["key"] == p.Item.Name {
// 			t["status"] = p.Status
// 			t["confirmer"] = p.Confirmer
// 			t["remark"] = p.Remark
// 			updated = true
// 		}
// 	}
// 	if updated {
// 		data, err := json.Marshal(permissionList)
// 		if err != nil {
// 			logs.Error("marshal error: %v", err)
// 			return false, err
// 		}
// 		record.PermInfos = string(data)
// 		if err := database.UpdateDBRecord(database.DB(), record); err != nil {
// 			logs.Error("update tb_perm_app_relation failed: %v", err)
// 			return false, err
// 		}
// 	}

// 	return updated, nil
// }

// func autoConfirmiOS(p *confirmParams) error {

// 	var detectType string
// 	switch *p.Item.Type {
// 	case TypeString:
// 		detectType = "blackList"
// 	case TypeMethod:
// 		detectType = "method"
// 	case TypePermission:
// 		detectType = "privacy"
// 	default:
// 		return fmt.Errorf("invalid detect type: %v", *p.Item.Type)
// 	}
// 	content, err := readDetectContentiOS(database.DB(), map[string]interface{}{
// 		"taskId":      p.TaskID,
// 		"toolId":      p.ToolID,
// 		"detect_type": detectType})
// 	if err != nil {
// 		logs.Error("read tb_ios_new_detect_content failed: %v", err)
// 		return err
// 	}
// 	if len(content) == 0 {
// 		logs.Error("invalid task id: %v cannot find any matched record", p.TaskID)
// 		return fmt.Errorf("invalid task id: %v cannot find any matched record", p.TaskID)
// 	}
// 	m := make(map[string]interface{})
// 	if err := json.Unmarshal([]byte(content[0].DetectContent), &m); err != nil {
// 		logs.Error("unmarshal error: %v", err)
// 		return err
// 	}
// 	switch *p.Item.Type {
// 	case TypeString:
// 		stringList, ok := m[detectType].([]interface{})
// 		if !ok {
// 			return fmt.Errorf("%s cannot assert to []interface{}: %v", TypeString, m[TypeString])
// 		}
// 		for i := range stringList {
// 			t, ok := stringList[i].(map[string]interface{})
// 			if !ok {
// 				return fmt.Errorf("cannot assert to map[string]interface{}: %v", stringList[i])
// 			}
// 			if t["name"] == p.Item.Name {
// 				t["status"] = p.Status
// 				t["confirmer"] = p.Confirmer
// 				t["remark"] = p.Remark
// 			}
// 		}
// 	case TypeMethod:
// 		methodList, ok := m[detectType].([]interface{})
// 		if !ok {
// 			return fmt.Errorf("%s cannot assert to []interface{}: %v", TypeMethod, m[TypeMethod])
// 		}
// 		for i := range methodList {
// 			t, ok := methodList[i].(map[string]interface{})
// 			if !ok {
// 				return fmt.Errorf("cannot assert to map[string]interface{}: %v", methodList[i])
// 			}
// 			if fmt.Sprintf("%v%v%v", t["content"], delimiter, t["name"]) == p.Item.Name {
// 				t["status"] = p.Status
// 				t["confirmer"] = p.Confirmer
// 				t["remark"] = p.Remark
// 			}
// 		}
// 	case TypePermission:
// 		permissionList, ok := m[detectType].([]interface{})
// 		if !ok {
// 			return fmt.Errorf("%s cannot assert to []interface{}: %v", TypePermission, m[TypeMethod])
// 		}
// 		for i := range permissionList {
// 			t, ok := permissionList[i].(map[string]interface{})
// 			if !ok {
// 				return fmt.Errorf("cannot assert to map[string]interface{}: %v", permissionList[i])
// 			}
// 			if t["permission"] == p.Item.Name {
// 				t["status"] = p.Status
// 				t["confirmer"] = p.Confirmer
// 				t["confirmReason"] = p.Remark
// 			}
// 		}
// 	default:
// 		return fmt.Errorf("invalid detect type: %v", *p.Item.Type)
// 	}
// 	confirmedContent, err := json.Marshal(m)
// 	if err != nil {
// 		logs.Error("marshal error: %v", err)
// 		return err
// 	}
// 	content[0].DetectContent = string(confirmedContent)
// 	if err := database.UpdateDBRecord(database.DB(), &content[0]); err != nil {
// 		logs.Error("update tb_ios_new_detect_content failed: %v", err)
// 		return err
// 	}
// 	if _, err := updateTaskStatus(p.TaskID, p.ToolID, platformiOS, 1); err != nil {
// 		logs.Error("update iOS detect task failed: %v", err)
// 		return err
// 	}

// 	return nil
// }

// func readDetectTask(db *gorm.DB, sieve map[string]interface{}) (
// 	[]dal.DetectStruct, error) {

// 	var tasks []dal.DetectStruct
// 	if err := db.Debug().Where(sieve).Find(&tasks).Error; err != nil {
// 		logs.Error("database error: %v", err)
// 		return nil, err
// 	}

// 	return tasks, nil
// }

// // The returned bool named freshman which only valid if there is no error.
// func preAutoConfirm(appID string, platform int, version string,
// 	item *Item, who string, status int, remark string) (bool, error) {

// 	records, err := readAPPAttention(database.DB(), map[string]interface{}{
// 		"app_id": appID, "platform": platform, "version": version})
// 	if err != nil {
// 		logs.Error("read app attention failed: %v", err)
// 		return false, err
// 	}

// 	m := make(map[string]*Attention)
// 	if len(records) > 0 {
// 		if err := json.Unmarshal([]byte(records[0].Attention), &m); err != nil {
// 			logs.Error("unmarshal error: %v", err)
// 			return false, err
// 		}
// 	} else {
// 		m[item.Name].Type = *item.Type
// 		m[item.Name].OriginVersion = version
// 	}
// 	m[item.Name].ConfirmedAt = time.Now()
// 	m[item.Name].Status = status
// 	m[item.Name].Confirmer = who
// 	m[item.Name].Remark = remark
// 	attention, err := json.Marshal(&m)
// 	if err != nil {
// 		logs.Error("marshal error: %v", err)
// 		return false, err
// 	}
// 	if len(records) > 0 {
// 		records[0].Attention = string(attention)
// 		if err := database.UpdateDBRecord(database.DB(), &records[0]); err != nil {
// 			logs.Error("update version attetion failed: %v", err)
// 			return false, err
// 		}
// 		return false, nil
// 	}

// 	if err := database.InsertDBRecord(database.DB(), &VersionDiff{
// 		APPID: appID, Platform: platform, Version: version,
// 		Attention: string(attention)}); err != nil {
// 		logs.Error("insert version record failed: %v", err)
// 		return false, err
// 	}

// 	return true, nil
// }

// func autoConfirmCallBack(task *dal.DetectStruct, permissions []string,
// 	sensitiveMethods []string, sensitiveStrings []string) error {

// 	header := fmt.Sprintf("task id: %v", task.ID)
// 	m, freshman, err := preAutoConfirmCallback(task, permissions, sensitiveMethods, sensitiveStrings)
// 	if err != nil {
// 		logs.Error("%s pre auto confirm failed: %v", header, err)
// 	}
// 	if !freshman {
// 		switch task.Platform {
// 		case platformAndorid:
// 			if err := autoConfirmCallBackAndroidPermission(task.ID, m); err != nil {
// 				logs.Error("%s auto confirm Android permission failed: %v", header, err)
// 				return err
// 			}
// 			if err := autoConfirmCallBackAndroidMethodAndString(task.ID, m); err != nil {
// 				logs.Error("%s auto confirm Android method and string failed: %v", header, err)
// 				return err
// 			}
// 		case platformiOS:
// 			return autoConfirmCallBackiOS(task.ID, m)
// 		default:
// 			logs.Error("%s unsupport platform: %v", header, task.Platform)
// 			return fmt.Errorf("%s unsupport platform: %v", header, task.Platform)
// 		}
// 	}

// 	return nil
// }

// func autoConfirmCallBackAndroidPermission(taskID interface{}, m map[string]*Attention) error {

// 	header := fmt.Sprintf("task id: %v", taskID)
// 	permissions, err := readPermAPPRelation(database.DB(), map[string]interface{}{"task_id": taskID})
// 	if err != nil {
// 		logs.Error("%s read tb_perm_app_relation error: %v", header, err)
// 		return err
// 	}
// 	for i := range permissions {
// 		var permissionList []interface{}
// 		if err := json.Unmarshal([]byte(permissions[i].PermInfos), &permissionList); err != nil {
// 			logs.Error("%s unmarshal error: %v", header, err)
// 			return err
// 		}
// 		var updated bool
// 		for j := range permissionList {
// 			t, ok := permissionList[j].(map[string]interface{})
// 			if !ok {
// 				logs.Error("%s cannot assert to map[string]interface{}: %v", header, permissionList[j])
// 				return fmt.Errorf("%s cannot assert to map[string]interface{}: %v", header, permissionList[j])
// 			}
// 			if v, ok := m[fmt.Sprint(t["key"])]; ok {
// 				t["status"] = v.Status
// 				t["confirmer"] = v.Confirmer
// 				t["remark"] = v.Remark
// 				updated = true
// 			}
// 		}
// 		if updated {
// 			data, err := json.Marshal(permissionList)
// 			if err != nil {
// 				logs.Error("%s marshal error: %v", header, err)
// 				return err
// 			}
// 			permissions[i].PermInfos = string(data)
// 			if err := database.UpdateDBRecord(database.DB(), &permissions[i]); err != nil {
// 				logs.Error("%s update tb_perm_app_relation failed: %v", header, err)
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// func autoConfirmCallBackAndroidMethodAndString(taskID interface{}, m map[string]*Attention) error {

// 	header := fmt.Sprintf("task id: %v", taskID)
// 	details, err := readDetectContentDetail(database.DB(), map[string]interface{}{"task_id": taskID})
// 	if err != nil {
// 		logs.Error("%s read tb_detect_content_deatil error: %v", header, err)
// 		return err
// 	}
// 	for i := range details {
// 		var updated bool
// 		switch details[i].SensiType {
// 		case Method:
// 			if v, ok := m[details[i].ClassName+delimiter+details[i].KeyInfo]; ok {
// 				details[i].Status = v.Status
// 				details[i].Confirmer = v.Confirmer
// 				details[i].Remark = v.Remark
// 				updated = true
// 			}
// 		case String:
// 			if v, ok := m[details[i].KeyInfo]; ok {
// 				details[i].Status = v.Status
// 				details[i].Confirmer = v.Confirmer
// 				details[i].Remark = v.Remark
// 				updated = true
// 			}
// 		default: // Do nothing
// 		}
// 		if updated {
// 			if err := database.UpdateDBRecord(database.DB(), &details[i]); err != nil {
// 				logs.Error("%s update tb_detect_content_detail failed: %v", header, err)
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// func autoConfirmCallBackiOS(taskID interface{}, m map[string]*Attention) error {

// 	header := fmt.Sprintf("task id: %v", taskID)
// 	content, err := readDetectContentiOS(database.DB(), map[string]interface{}{
// 		"taskId": taskID})
// 	if err != nil {
// 		logs.Error("%s read iOS detect content failed: %v", header, err)
// 		return err
// 	}
// 	for i := range content {
// 		t := make(map[string]interface{})
// 		if err := json.Unmarshal([]byte(content[i].DetectContent), &t); err != nil {
// 			logs.Error("%s unmarshal error: %v", header, err)
// 			return err
// 		}
// 		switch content[i].DetectType {
// 		case "privacy":
// 			list, ok := t["privacy"].([]interface{})
// 			if !ok {
// 				logs.Error("%s cannot assert to []interface{}: %v", header, t["privacy"])
// 				return fmt.Errorf("%s cannot assert to []interface{}: %v", header, t["privacy"])
// 			}
// 			for j := range list {
// 				u, ok := list[j].(map[string]interface{})
// 				if !ok {
// 					logs.Error("%s cannot assert to map[string]interface{}: %v", header, list[j])
// 					return fmt.Errorf("%s cannot assert to map[string]interface{}: %v", header, list[j])
// 				}
// 				if v, ok := m[fmt.Sprint(u["permission"])]; ok {
// 					u["origin_version"] = v.OriginVersion
// 					u["status"] = v.Status
// 					u["confirmer"] = v.Confirmer
// 					u["confirmReason"] = v.Remark
// 				}
// 			}
// 		case "method":
// 			list, ok := t["method"].([]interface{})
// 			if !ok {
// 				logs.Error("%s cannot assert to []interface{}: %v", header, t["method"])
// 				return fmt.Errorf("%s cannot assert to []interface{}: %v", header, t["method"])
// 			}
// 			for j := range list {
// 				u, ok := list[j].(map[string]interface{})
// 				if !ok {
// 					logs.Error("%s cannot assert to map[string]interface{}: %v", header, list[j])
// 					return fmt.Errorf("%s cannot assert to map[string]interface{}: %v", header, list[j])
// 				}
// 				if v, ok := m[fmt.Sprintf("%v%v%v", u["content"], delimiter, u["name"])]; ok {
// 					u["origin_version"] = v.OriginVersion
// 					u["status"] = v.Status
// 					u["confirmer"] = v.Confirmer
// 					u["remark"] = v.Remark
// 				}
// 			}
// 		case "blacklist":
// 			list, ok := t["blackList"].([]interface{})
// 			if !ok {
// 				logs.Error("%s cannot assert to []interface{}: %v", header, t["blackList"])
// 				return fmt.Errorf("%s cannot assert to []interface{}: %v", header, t["blackList"])
// 			}
// 			for j := range list {
// 				u, ok := list[j].(map[string]interface{})
// 				if !ok {
// 					logs.Error("%s cannot assert to map[string]interface{}: %v", header, list[j])
// 					return fmt.Errorf("%s cannot assert to map[string]interface{}: %v", header, list[j])
// 				}
// 				if v, ok := m[fmt.Sprint(u["name"])]; ok {
// 					u["origin_version"] = v.OriginVersion
// 					u["status"] = v.Status
// 					u["confirmer"] = v.Confirmer
// 					u["remark"] = v.Remark
// 				}
// 			}
// 		}
// 		data, err := json.Marshal(&t)
// 		if err != nil {
// 			logs.Error("%s unmarshal error: %v", header, err)
// 			return err
// 		}
// 		content[i].DetectContent = string(data)
// 		if err := database.UpdateDBRecord(database.DB(), &content[i]); err != nil {
// 			logs.Error("%s update tb_ios_new_detect_content: %v", header, err)
// 			return err
// 		}
// 	}

// 	return nil
// }

func preAutoConfirmCallback(task *dal.DetectStruct, permissions []string,
	sensitiveMethods []string, sensitiveStrings []string) (
	map[string]*Attention, bool, error) {

	items := transformRawData(permissions, sensitiveMethods, sensitiveStrings)
	records, err := readAPPAttention(database.DB(), map[string]interface{}{
		"app_id": task.AppId, "platform": task.Platform, "version": task.AppVersion})
	if err != nil {
		logs.Error("read app attention failed: %v", err)
		return nil, false, err
	}

	if len(records) <= 0 {
		// The is the first time that the version was detected.
		return firstTime(task.AppId, task.Platform, task.AppVersion, items)
	}

	// The versinon was detected more than once.
	m, err := notfirstTime(&records[0], items)
	if err != nil {
		logs.Error("not first time error: %v", err)
		return nil, false, err
	}

	return m, false, nil
}

func transformRawData(permissions []string, sensitiveMethods []string,
	sensitiveStrings []string) []Item {

	var items []Item
	for i := range permissions {
		items = append(items, Item{Name: permissions[i],
			Type: &TypePermission})
	}
	for i := range sensitiveMethods {
		items = append(items, Item{Name: sensitiveMethods[i],
			Type: &TypeMethod})
	}
	for i := range sensitiveStrings {
		items = append(items, Item{Name: sensitiveStrings[i],
			Type: &TypeString})

	}

	return items
}

// The returned bool named freshman which only valid if there is no error.
func firstTime(appID string, platform int, version string, items []Item) (
	map[string]*Attention, bool, error) {

	m := make(map[string]*Attention)
	createVersionRecord(m, items)
	// Initialize the origin version since this is freshman.
	for key := range m {
		m[key].OriginVersion = version
	}

	previous, err := previousVersion(appID, platform, version)
	if err != nil {
		logs.Error("get previous version failed: %v", err)
		return nil, false, err
	}

	if previous != nil {
		logs.Notice("current version: %v previous version: %v", version, previous.Version)
		if err := autoConfirmWithPreviousVersion(m, previous); err != nil {
			logs.Error("auto confirm with previous version failed: %v", err)
			return nil, false, err
		}
	}

	record, err := json.Marshal(&m)
	if err != nil {
		logs.Error("marshal error: %v", err)
		return nil, false, err
	}
	if err := database.InsertDBRecord(database.DB(), &VersionDiff{
		APPID: appID, Platform: platform, Version: version,
		Attention: string(record)}); err != nil {
		logs.Error("insert version record failed: %v", err)
		return nil, false, err
	}

	if previous == nil {
		return m, true, nil // This is a absolute freshman.
	}

	return m, false, nil
}

func notfirstTime(record *VersionDiff, items []Item) (map[string]*Attention, error) {

	m := make(map[string]*Attention)
	if err := json.Unmarshal([]byte(record.Attention), &m); err != nil {
		logs.Error("unmarshal error: %v", err)
		return nil, err
	}
	originLen := len(m)
	createVersionRecord(m, items)
	if len(m) > originLen {
		// Update the origin version for the new items.
		for _, v := range m {
			if v.OriginVersion == "" {
				v.OriginVersion = record.Version
			}
		}
		// Something new was added into the version.
		attention, err := json.Marshal(&m)
		if err != nil {
			logs.Error("marshal error: %v", err)
			return nil, err
		}
		record.Attention = string(attention)
		if err := database.UpdateDBRecord(database.DB(), record); err != nil {
			logs.Error("update version attetion failed: %v", err)
			return nil, err
		}
	}

	return m, nil
}

func createVersionRecord(m map[string]*Attention, items []Item) {

	for i := range items {
		if _, ok := m[items[i].Name]; !ok {
			m[items[i].Name] = &Attention{Type: *items[i].Type}
		}
	}
}

func previousVersion(appID interface{}, platform interface{}, version string) (*VersionDiff, error) {

	current, err := strconv.ParseInt(transformVersion(version), 10, 64)
	if err != nil {
		logs.Error("unsupported version format (%v) parse error: %v", version, err)
		return nil, err
	}

	// Load all versions of the app.
	records, err := readAPPAttention(database.DB(), map[string]interface{}{
		"app_id": appID, "platform": platform})
	if err != nil {
		logs.Error("read app_attention_history failed: %v", err)
		return nil, err
	}
	if len(records) <= 0 {
		return nil, nil
	}

	m := make(map[int64]*VersionDiff)
	for i := range records {
		code, err := strconv.ParseInt(transformVersion(records[i].Version), 10, 64)
		if err != nil {
			logs.Warn("parse version (%v) error: %v", records[i].Version, err)
			continue
		}
		m[code] = &records[i]
	}

	var closest int64 = -1
	var result *VersionDiff
	for k := range m {
		if k > current {
			continue
		}
		if k > closest {
			closest = k
			result = m[k]
		}
	}

	return result, nil
}

func transformVersion(s string) string {

	var t string
	for i := range s {
		if s[i] >= '0' && s[i] <= '9' {
			t += string(s[i])
		} else {
			t += " "
		}
	}

	var result string
	for _, v := range strings.Split(strings.TrimSpace(t), " ") {
		result += addLeadingZero(v, 5)
	}

	return result
}

// The length of s must less than or equal to fixedLen.
func addLeadingZero(s string, fixedLen int) string {

	var leading string
	for i := 0; i < fixedLen-len(s); i++ {
		leading += "0"
	}

	return leading + s
}

func autoConfirmWithPreviousVersion(current map[string]*Attention,
	prev *VersionDiff) error {

	// Unmarshal unmarshals the JSON into the value pointed at by the pointer.
	// If the pointer is nil, Unmarshal allocates a new value for it to point to.
	previous := make(map[string]*Attention)
	logs.Notice("content: %s", prev.Attention)
	if err := json.Unmarshal([]byte(prev.Attention), &previous); err != nil {
		logs.Error("unmarshal error: %v", err)
		return err
	}

	for k := range current {
		if v, ok := previous[k]; ok {
			// It shows up in previous version!
			current[k].OriginVersion = v.OriginVersion
			if previous[k].Status != Unconfirmed {
				current[k] = v
			}
		}
	}

	return nil
}
