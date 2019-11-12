package settings

import (
	"fmt"
	"net/http"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
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

	Refresh(c)
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
