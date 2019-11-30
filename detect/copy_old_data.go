package detect

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

// ImportOldData copy confirmed items from history.
func ImportOldData(c *gin.Context) {

	if err := importOldData(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("import old data failed: %v", err),
			"code":    -1})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"code":    0})
}

func importOldData() error {

	db, err := database.GetDBConnection()
	if err != nil {
		logs.Error("connect to DB failed: %v", err)
		return err
	}
	defer db.Close()
	tasks, err := getDetectTask(db, nil)
	if err != nil {
		logs.Error("get tb_binary_detect failed: %v", err)
		return err
	}
	// Android method and string
	for i := range tasks {
		if tasks[i].Status != 0 && tasks[i].Status != 1 && tasks[i].Status != 2 {
			continue
		}
		fmt.Printf("%v\t%v\t%v\t%v\n", tasks[i].ID, tasks[i].AppId, tasks[i].Platform, tasks[i].AppVersion)
		details, err := getDetectContentDetail(db, map[string]interface{}{
			"task_id": tasks[i].ID})
		if err != nil {
			logs.Error("get tb_detect_content_detail failed: %v", err)
			return err
		}
		m := make(map[string]*Attention)
		for j := range details {
			var name string
			var _type string
			switch details[j].SensiType {
			case 1:
				name = details[j].ClassName + "." + details[j].KeyInfo
				_type = "敏感方法"
			case 2:
				name = details[j].KeyInfo
				_type = "敏感词汇"
			default:
				continue
			}
			fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", name, _type, details[j].Status, details[j].UpdatedAt, details[j].Confirmer, details[j].Remark)
			if details[j].Status == 1 && details[j].Confirmer != "" && details[j].Remark != "" {
				m[name] = &Attention{
					Type:          _type,
					OriginVersion: tasks[i].AppVersion,
					Status:        details[j].Status,
					ConfirmedAt:   details[j].UpdatedAt,
					Confirmer:     details[j].Confirmer,
					Remark:        details[j].Remark}
			} else {
				m[name] = &Attention{
					Type:          _type,
					OriginVersion: tasks[i].AppVersion,
					Status:        0,
				}
			}
		}
		autoImport(tasks[i].AppId, tasks[i].Platform, tasks[i].AppVersion, m)
	}
	// Android permission
	permissions, err := getPermissionAPPRelation(db, nil)
	if err != nil {
		logs.Error("get tb_perm_app_relation failed: %v", err)
		return err
	}
	for i := range permissions {
		fmt.Printf("%v\t%v\t%v\n", permissions[i].AppId, 0, permissions[i].AppVersion)
		var list []map[string]interface{}
		if err := json.Unmarshal([]byte(permissions[i].PermInfos), &list); err != nil {
			logs.Error("unmarshal error: %v", err)
			return err
		}
		m := make(map[string]*Attention)
		for j := range list {
			fmt.Printf(">>>>> %v\n", list[j])
			if v, ok := list[j]["status"]; ok {
				if fmt.Sprint(v) == "1" {
					c, cok := list[j]["confirmer"]
					r, rok := list[j]["remark"]
					if cok && rok {
						fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", list[j]["key"], "权限", v, permissions[i].UpdatedAt, c, r)
						m[fmt.Sprint(list[j]["key"])] = &Attention{
							Type:          "权限",
							OriginVersion: permissions[i].AppVersion,
							Status:        1,
							ConfirmedAt:   permissions[i].UpdatedAt,
							Confirmer:     fmt.Sprint(c),
							Remark:        fmt.Sprint(r),
						}
					} else {
						historys, err := getPermissionHistroy(db, map[string]interface{}{
							"perm_id":     list[j]["perm_id"],
							"app_id":      permissions[i].AppId,
							"app_version": permissions[i].AppVersion,
						})
						if err != nil {
							logs.Error("get tb_perm_history failed: %v", err)
							return err
						}
						if len(historys) > 0 {
							if historys[0].Status == 1 && historys[0].Confirmer != "" && historys[0].Remarks != "" {
								m[fmt.Sprint(list[j]["key"])] = &Attention{
									Type:          "权限",
									OriginVersion: permissions[i].AppVersion,
									Status:        1,
									ConfirmedAt:   historys[0].UpdatedAt,
									Confirmer:     historys[0].Confirmer,
									Remark:        historys[0].Remarks,
								}
							} else {
								m[fmt.Sprint(list[j]["key"])] = &Attention{
									Type:          "权限",
									OriginVersion: permissions[i].AppVersion,
									Status:        0,
								}
							}
							fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", list[j]["key"], "权限", v, historys[0].UpdatedAt, historys[0].Confirmer, historys[0].Remarks)
						} else {
							// Cannot find the confirmed record, so treat it like unconfirmed.
							fmt.Printf("%v\t%v\t%v\n", list[j]["key"], "权限", 0)
							m[fmt.Sprint(list[j]["key"])] = &Attention{
								Type:          "权限",
								OriginVersion: permissions[i].AppVersion,
								Status:        0,
							}
						}
					}
				} else {
					fmt.Printf("%v\t%v\t%v\n", list[j]["key"], "权限", v)
					m[fmt.Sprint(list[j]["key"])] = &Attention{
						Type:          "权限",
						OriginVersion: permissions[i].AppVersion,
						Status:        0,
					}
				}
			}
		}
		autoImport(fmt.Sprint(permissions[i].AppId), 0, permissions[i].AppVersion, m)
	}

	return nil
}

func getDetectTask(db *gorm.DB, sieve map[string]interface{}) (
	[]dal.DetectStruct, error) {

	var tasks []dal.DetectStruct
	if err := db.Where(sieve).Find(&tasks).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return tasks, nil
}

func getDetectContentDetail(db *gorm.DB, sieve map[string]interface{}) (
	[]dal.DetectContentDetail, error) {

	var details []dal.DetectContentDetail
	if err := db.Where(sieve).Find(&details).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return details, nil
}

func getPermissionAPPRelation(db *gorm.DB, sieve map[string]interface{}) (
	[]dal.PermAppRelation, error) {

	var permssions []dal.PermAppRelation
	if err := db.Where(sieve).Find(&permssions).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return permssions, nil
}

func getPermissionHistroy(db *gorm.DB, sieve map[string]interface{}) (
	[]dal.PermHistory, error) {

	var historys []dal.PermHistory
	if err := db.Where(sieve).Find(&historys).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return historys, nil
}

func autoImport(appID string, platform int, version string, m map[string]*Attention) error {

	if len(m) <= 0 {
		return nil
	}

	records, err := readAPPAttention(database.DB(), map[string]interface{}{
		"app_id": appID, "platform": platform, "version": version})
	if err != nil {
		logs.Error("read app attention failed: %v", err)
		return err
	}

	if len(records) <= 0 {
		previous, err := previousVersion(appID, platform, version)
		if err != nil {
			logs.Error("get previous version failed: %v", err)
			return err
		}

		if previous != nil {
			logs.Notice("current version: %v previous version: %v", version, previous.Version)
			if err := autoConfirmWithPreviousVersion(m, previous); err != nil {
				logs.Error("auto confirm with previous version failed: %v", err)
				return err
			}
		}

		record, err := json.Marshal(&m)
		if err != nil {
			logs.Error("marshal error: %v", err)
			return err
		}
		if err := database.InsertDBRecord(database.DB(), &VersionDiff{
			APPID: appID, Platform: platform, Version: version,
			Attention: string(record)}); err != nil {
			logs.Error("insert version record failed: %v", err)
			return err
		}
	} else {
		t := make(map[string]*Attention)
		if err := json.Unmarshal([]byte(records[0].Attention), &t); err != nil {
			logs.Error("unmarshal error: %v", err)
			return err
		}

		for k := range t {
			if v, ok := m[k]; ok {
				if t[k].Status != 1 && v.Status == 1 {
					t[k].Status = v.Status
					t[k].ConfirmedAt = v.ConfirmedAt
					t[k].Confirmer = v.Confirmer
					t[k].Remark = v.Remark
				}
			}
		}
		// Something new was added into the version.
		attention, err := json.Marshal(&t)
		if err != nil {
			logs.Error("marshal error: %v", err)
			return err
		}
		records[0].Attention = string(attention)
		if err := database.UpdateDBRecord(database.DB(), records[0]); err != nil {
			logs.Error("update version attetion failed: %v", err)
			return err
		}
	}

	return nil
}
