package detect

import (
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
		logs.Error("get detect task failed: %v", err)
		return err
	}
	for i := range tasks {
		if tasks[i].Status != 0 && tasks[i].Status != 1 && tasks[i].Status != 2 {
			continue
		}
		fmt.Printf("%v\t%v\t%v\t%v\n", tasks[i].ID, tasks[i].AppId, tasks[i].Platform, tasks[i].AppVersion)
		details, err := getDetectContentDetail(db, map[string]interface{}{
			"task_id": tasks[i].ID,
			"status":  1})
		if err != nil {
			logs.Error("get detect content detail failed: %v", err)
			return err
		}
		for j := range details {
			if details[j].Remark == "" || details[j].Confirmer == "" {
				continue
			}
			var name string
			var _type string
			switch details[j].SensiType {
			case 1:
				name = details[j].ClassName + "." + details[j].KeyInfo
				_type = "敏感方法"
			case 2:
				name = details[j].KeyInfo
				_type = "敏感字符"
			default:
				continue

			}
			fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", name, _type, details[j].Status, details[j].UpdatedAt, details[j].Confirmer, details[j].Remark)
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
