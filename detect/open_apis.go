package detect

import (
	"fmt"
	"net/http"

	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

const latestDetectTool = "6"

// GetDetectTaskResult returns the result of binary detect task for
// the given task id if it is valid.
func GetDetectTaskResult(c *gin.Context) {

	taskID, exist := c.GetQuery("task_id")
	if !exist {
		ReturnMsg(c, FAILURE, "Miss task id")
		return
	}

	data := getDetectResult(c, taskID, latestDetectTool)
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
