package settings

import (
	"fmt"
	"net/http"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
	"github.com/gin-gonic/gin"
)

// Refresh will update the current settings
// while programming is still running.
func Refresh(c *gin.Context) {

	if err := Load(database.DB()); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("failed to refresh settings: %v", err))
	}

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "refresh settings success")
}
