package detect

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
)

const delimiter = "."

// MR confirm
func preAutoConfirmMR(appID string, platform int, version string,
	item *Item, status int, who string, remark string) error {

	freshman, err := preAutoConfirm(appID, platform, version, item, who, status, remark)
	if err != nil {
		logs.Error("pre auto confirm failed: %v", err)
		return err
	}

	if !freshman {
		// TODO
	}

	return nil
}

// Binary detect task confirm
func preAutoConfirmTask(task *dal.DetectStruct, item *Item, status int, who string, remark string) error {

	freshman, err := preAutoConfirm(task.AppId, task.Platform, task.AppVersion, item, who, status, remark)
	if err != nil {
		logs.Error("pre auto confirm failed: %v", err)
		return err
	}

	if !freshman {
		// TODO
		switch task.Platform {
		case platformAndorid:
		case platformiOS:

		default:
			return fmt.Errorf("unsupport platform: %v", task.Platform)
		}
	}

	return nil
}

// The bool in return value named freshman which only valid if no error.
func preAutoConfirm(appID string, platform int, version string,
	item *Item, who string, status int, remark string) (bool, error) {

	records, err := readAPPAttention(database.DB(), map[string]interface{}{
		"app_id": appID, "platform": platform, "version": version})
	if err != nil {
		logs.Error("read app attention failed: %v", err)
		return false, err
	}

	m := make(map[string]*Attention)
	if len(records) > 0 {
		if err := json.Unmarshal([]byte(records[0].Attention), &m); err != nil {
			logs.Error("unmarshal error: %v", err)
			return false, err
		}
	}
	m[item.Name] = &Attention{
		Type:        *item.Type,
		Status:      status,
		ConfirmedAt: time.Now(),
		Confirmer:   who,
		Remark:      remark}
	attention, err := json.Marshal(&m)
	if err != nil {
		logs.Error("marshal error: %v", err)
		return false, err
	}
	if len(records) > 0 {
		records[0].Attention = string(attention)
		if err := database.UpdateDBRecord(database.DB(), &records[0]); err != nil {
			logs.Error("update version attetion failed: %v", err)
			return false, err
		}
		return false, nil
	}

	if err := database.InsertDBRecord(database.DB(), &VersionDiff{
		APPID: appID, Platform: platform, Version: version,
		Attention: string(attention)}); err != nil {
		logs.Error("insert version record failed: %v", err)
		return false, err
	}
	return true, nil
}

func autoConfirm(task *dal.DetectStruct, permissions []string,
	sensitiveMethods []dal.MethodInfo, sensitiveStrings []dal.StrInfo) {

	_, freshman, err := preAutoConfirmCallback(task, permissions, sensitiveMethods, sensitiveStrings)
	if err != nil {
		logs.Error("task id: %v pre auto confirm failed: %v", task.ID, err)
	}
	if !freshman {
		// TODO
	}
}

func preAutoConfirmCallback(task *dal.DetectStruct, permissions []string,
	sensitiveMethods []dal.MethodInfo, sensitiveStrings []dal.StrInfo) (
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

func transformRawData(permissions []string, sensitiveMethods []dal.MethodInfo,
	sensitiveStrings []dal.StrInfo) []Item {

	var items []Item
	for i := range permissions {
		items = append(items, Item{
			Name: permissions[i],
			Type: &TypePermission})
	}
	for i := range sensitiveMethods {
		items = append(items, Item{Name: sensitiveMethods[i].ClassName +
			delimiter + sensitiveMethods[i].MethodName,
			Type: &TypeMethod})
	}
	for i := range sensitiveStrings {
		for j := range sensitiveStrings[i].Keys {
			items = append(items, Item{
				Name: sensitiveStrings[i].Keys[j],
				Type: &TypeString})
		}
	}

	return items
}

// The bool in return value named freshman which only valid is no error.
func firstTime(appID string, platform int, version string, items []Item) (
	map[string]*Attention, bool, error) {

	m := make(map[string]*Attention)
	createVersionRecord(m, items)

	previous, err := previousVersion(appID, platform, version)
	if err != nil {
		logs.Error("get previous version failed: %v", err)
		return nil, false, err
	}

	if previous != nil {
		if err := autoConfirmWithPreviousVersion(m, previous); err != nil {
			logs.Error("uto confirm with previous version failed: %v", err)
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
		result += addLeadingZero(v, 10)
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
	if err := json.Unmarshal([]byte(prev.Attention), &previous); err != nil {
		logs.Error("unmarshal error: %v", err)
		return err
	}

	for k := range current {
		if _, ok := previous[k]; ok {
			// It shows up in previous version!
			if previous[k].Status != Unconfirmed {
				current[k] = previous[k]
			}
		}
	}

	return nil
}
