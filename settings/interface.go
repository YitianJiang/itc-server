package settings

import (
	"fmt"
	"net/http"
	"os/exec"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

// Insert inserts a new settings to database, then update current settings.
func Insert(c *gin.Context) {

	var settings Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid parameter: %v", err))
		return
	}

	if err := Store(database.DB(), &settings); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("store new settings failed: %v", err))
		return
	}

	go Sync()

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success")
}

// Refresh will update the current settings
// while programming is still running.
func Refresh(c *gin.Context) {

	if err := Load(database.DB()); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("refresh settings failed: %v", err))
		return
	}

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success")
}

// Show returns the current configure.
func Show(c *gin.Context) {

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "success", *settings)
}

// Sync notifies all instances to update their settings.
func Sync() {

	file := "./refresh_online_settings.sh"
	cmd := exec.Command("/bin/bash", "-c", file)
	output, err := cmd.Output()
	if err != nil {
		// Wrning: Sync error means settings is different in different instance,
		// we must handle the error manually!
		logs.Error("sync error: %v", err)
		utils.LarkDingOneInner(Get().NightWatchman, fmt.Sprintf("sync error: %v", err))
		return
	}

	logs.Notice("sync success!\n%s", output)
	utils.LarkDingOneInner(Get().NightWatchman, fmt.Sprintf("sync success!\n%s", output))
}
