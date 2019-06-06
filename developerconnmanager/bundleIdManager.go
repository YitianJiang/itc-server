package developerconnmanager

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"github.com/gin-gonic/gin"
	"net/http"
	"code.byted.org/gopkg/logs"
)

func TestAskBundleId(c *gin.Context)  {
	logs.Info("开始测试Bundle ID逻辑")
	nameBundle := c.Query("nameBundle")
	var bundleIdObj dal.BundleIdManager
	bundleIdObj.BundleId = nameBundle

	if dal.InsertBundleId(bundleIdObj) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data": "插入BundleID正常",
			"name": nameBundle,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode": -1,
			"message":   "数据库存储错误，新建BundleID失败！",
		})
	}
}

func GetBunldIdsObj(c *gin.Context){
	logs.Info("返回数据库中所有的Bundle ID")
	BundleIdsObjResponse,boolResult := dal.SearchBundleIds()
	if boolResult{
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"errorCode": 0,
			"data": BundleIdsObjResponse,
		})
	}else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "fail DB",
			"errorCode": 1,
		})
	}
}