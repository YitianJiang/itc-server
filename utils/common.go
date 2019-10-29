package utils

import (
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

// Error code
const (
	FAILURE = -1
	SUCCESS = 0
)

// ReturnMsg return the response to requester.
// If the data is not empty, only the first data will be accept while the rest
// will be abandoned.
func ReturnMsg(c *gin.Context, httpCode int, code int, msg string,
	data ...interface{}) {

	switch code {
	case FAILURE:
		logs.Error(msg)
	case SUCCESS:
		logs.Debug(msg)
	}

	obj := gin.H{"code": code, "message": msg}
	if len(data) > 0 {
		obj["data"] = data[0]
	}

	c.JSON(httpCode, obj)

	return
}
