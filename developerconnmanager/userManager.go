package developerconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/utils"
	"github.com/astaxie/beego/logs"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GetCapabilitiesInfoReq struct {
	AppType     string  `form:"app_type" json:"app_type" binding:"required"`
}

//type GetPlatformCapabilitiesInfoRes struct {
//	IOSCapabilitiesInfo CapabilitiesInfo `json:"ios_capabilities,omitempty"`
//	MacCapabilitiesInfo CapabilitiesInfo `json:"mac_capabilities,omitempty"`
//}

type CapabilitiesInfo struct {
	SelectCapabilitiesInfo []string `json:"select_capabilities"`
	SettingsCapabilitiesInfo map[string][]string `json:"settings_capabilities"`
}

func AssembleJsonResponse(c *gin.Context, errorNo int, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"error_code": errorNo,
		"error_info": message,
		"data":    data,
	})
}

func GetBundleIdCapabilitiesInfo(c *gin.Context){
	logs.Info("返回BundleID的能力给前端做展示")
	var requestData GetCapabilitiesInfoReq
	bindQueryError := c.ShouldBindQuery(&requestData)
	utils.RecordError("请求参数绑定错误: ", bindQueryError)
	var responseData CapabilitiesInfo
	if bindQueryError != nil {
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", responseData)
		return
	}
	responseData.SettingsCapabilitiesInfo = make(map[string][]string)
	responseData.SettingsCapabilitiesInfo[_const.ICLOUD] = _const.CloudSettings
	if requestData.AppType == "iOS"{
		responseData.SelectCapabilitiesInfo = _const.IOSSelectCapabilities
		responseData.SettingsCapabilitiesInfo[_const.DATA_PROTECTION] = _const.ProtectionSettings
	}else {
		responseData.SelectCapabilitiesInfo = _const.MacSelectCapabilities
	}
	c.JSON(http.StatusOK, gin.H{
				"message":   "success",
				"errorCode": 0,
				"data": responseData,
			})
}