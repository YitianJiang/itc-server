package developerconnmanager

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"code.byted.org/gopkg/logs"
)

func TestAskBundleId(c *gin.Context)  {
	logs.Info("开始测试Bundle ID逻辑")
	nameBundle := "NewsTargetName"
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data": "BundleID返回正常",
		"name": nameBundle,
	})
}