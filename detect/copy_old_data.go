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
	}

	return nil
}

func getDetectTask(db *gorm.DB, sieve map[string]interface{}) (
	[]dal.DetectStruct, error) {

	var tasks []dal.DetectStruct
	if err := db.Debug().Where(sieve).Find(&tasks).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return tasks, nil
}
