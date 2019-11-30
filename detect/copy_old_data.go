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
		}
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
		for j := range list {
			fmt.Printf(">>>>> %v\n", list[j])
			if v, ok := list[j]["status"]; ok {
				if fmt.Sprint(v) == "1" {
					c, cok := list[j]["confirmer"]
					r, rok := list[j]["remark"]
					if cok && rok {
						fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", list[j]["key"], "权限", v, permissions[i].UpdatedAt, c, r)
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
							fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", list[j]["key"], "权限", v, historys[0].UpdatedAt, historys[0].Confirmer, historys[0].Remarks)
						} else {
							// Cannot find the confirmed record, so treat it like unconfirmed.
							fmt.Printf("%v\t%v\t%v\n", list[j]["key"], "权限", 0)
						}
					}
				} else {
					fmt.Printf("%v\t%v\t%v\n", list[j]["key"], "权限", v)
				}
			}
		}
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
