package detect

import (
	"fmt"
	"net/http"

	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

// GetDetectTaskResult returns the result of binary detect task about
// the given task id if the task id is valid.
func GetDetectTaskResult(c *gin.Context) {

	taskID, exist := c.GetQuery("task_id")
	if !exist {
		ReturnMsg(c, FAILURE, "Miss task id")
		return
	}

	data := getDetectResult(c, taskID, "6")
	if data == nil {
		ReturnMsg(c, FAILURE, fmt.Sprintf("Task id: %v Failed to get binary detect result", taskID))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errorCode": SUCCESS,
		"message":   "success",
		"data":      *data})

	logs.Info("Task id: %v Get binary detect result success", taskID)
	return
}
