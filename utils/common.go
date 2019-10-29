package utils

import (
	"net/http"

	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

// Error code
const (
	FAILURE = -1
	SUCCESS = 0
)

// ReturnMsg shows necessary information for requestor.
func ReturnMsg(c *gin.Context, code int, msg string) {

	switch code {
	case FAILURE:
		logs.Error(msg)
	case SUCCESS:
		logs.Debug(msg)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": msg,
	})

	return
}
