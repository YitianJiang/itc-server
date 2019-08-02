package developerconnmanager

import (
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetAppDetailInfo(c *gin.Context) {

}

func CreateAppBindAccount(c *gin.Context) {
	var requestData devconnmanager.CreateAppBindAccountRequest
	//获取请求参数
	bindJsonError := c.ShouldBindJSON(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindJsonError)
	if bindJsonError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	logs.Info("request:%v", requestData)

	/*inputs := map[string]interface{}{
		"app_id": requestData.AppId,
		"app_name": requestData.AppName,
	}*/
	//todo 根据app_id和app_name执行update，如果返回的操作行数为0，则插入数据
	//todo 等待kani提供根据资源和权限获取人员信息的接口，根据该接口获取需要发送审批消息的用户list
	//todo lark消息生成并批量发送
}
