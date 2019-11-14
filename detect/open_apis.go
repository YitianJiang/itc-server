package detect

import (
	"fmt"
	"net/http"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

// Constant
const (
	latestDetectTool = "6"

	ErrTaskDone     = 0
	ErrTaskRunning  = -1
	ErrTaskIDMiss   = -2
	ErrTaskNotFound = -3
	ErrTaskOther    = -4
)

// GetDetectTaskResult returns the result of binary detect task for
// the given task id if it is valid.
func GetDetectTaskResult(c *gin.Context) {

	taskID, exist := c.GetQuery("task_id")
	if !exist {
		ReturnMsg(c, ErrTaskIDMiss, "Miss task id")
		return
	}

	task, err := getExactDetectTask(database.DB(), map[string]interface{}{"id": taskID})
	if err != nil {
		ReturnMsg(c, ErrTaskNotFound, fmt.Sprintf("Task id: %v Failed to get the task information", taskID))
		return
	}
	if task.Status == -1 {
		ReturnMsg(c, ErrTaskRunning, fmt.Sprintf("Task id: %v Still running...", taskID))
		return
	}

	data := getDetectResult(taskID, latestDetectTool)
	if data == nil {
		ReturnMsg(c, ErrTaskOther, fmt.Sprintf("Task id: %v Failed to get binary detect result", taskID))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errorCode": ErrTaskDone,
		"message":   "success",
		"data":      *data,
		"summary":   getRiskSummary(*data)})

	logs.Info("Task id: %v Get binary detect result success", taskID)
	return
}

// RiskSummary stores the summary of binary detect task.
type RiskSummary struct {
	Harmless   uint `json:"harmless"`
	Ordinary   uint `json:"ordinary"`
	LowRisk    uint `json:"low_risk"`
	MiddleRisk uint `json:"middle_risk"`
	HighRisk   uint `json:"high_risk"`
	Unknown    uint `json:"unknown"`
}

func getRiskSummary(data []dal.DetectQueryStruct) *RiskSummary {

	var summary RiskSummary
	for i := range data {
		for j := range data[i].SMethods {
			switch data[i].SMethods[j].RiskLevel {
			case "-1":
				summary.Harmless++
			case "0":
				summary.Ordinary++
			case "1":
				summary.LowRisk++
			case "2":
				summary.MiddleRisk++
			case "3":
				summary.HighRisk++
			default:
				summary.Unknown++
			}
		}
		for j := range data[i].SStrs_new {
			switch data[i].SStrs_new[j].RiskLevel {
			case "-1":
				summary.Harmless++
			case "0":
				summary.Ordinary++
			case "1":
				summary.LowRisk++
			case "2":
				summary.MiddleRisk++
			case "3":
				summary.HighRisk++
			default:
				summary.Unknown++
			}
		}
		for j := range data[i].Permissions_2 {
			switch data[i].Permissions_2[j].RiskLevel {
			case "-1":
				summary.Harmless++
			case "0":
				summary.Ordinary++
			case "1":
				summary.LowRisk++
			case "2":
				summary.MiddleRisk++
			case "3":
				summary.HighRisk++
			default:
				summary.Unknown++
			}
		}
	}

	return &summary
}
