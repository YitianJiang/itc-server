package detect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

// GetTMPConfig returns detecte configures provided by shiyanlong@bytedance.com
func GetTMPConfig(c *gin.Context) {

	data, err := ioutil.ReadFile("tmp_rules.json")
	if err != nil {
		ReturnMsg(c, FAILURE, fmt.Sprintf("io read file error: %v", err))
		return
	}

	detections := make(map[string]interface{})
	if err := json.Unmarshal(data, &detections); err != nil {
		ReturnMsg(c, FAILURE, fmt.Sprintf("unmarshal error: %v", err))
		return
	}

	c.JSON(http.StatusOK, detections)
	logs.Info("get tmp rules success")
	return
}
