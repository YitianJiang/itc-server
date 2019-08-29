package developerconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"code.byted.org/yuyilei/bot-api/form"
	"code.byted.org/yuyilei/bot-api/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

func QueryDeviceInfo(c *gin.Context) {
	logs.Info("HttpRequest：查询设备信息")
	var requestData devconnmanager.GetDevicesInfoRequest
	bindJsonError := c.ShouldBindQuery(&requestData)
	if bindJsonError != nil {
		logs.Error("绑定post请求body出错：%v", bindJsonError)
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	condition := map[string]interface{}{
		"team_id": requestData.TeamId,
	}
	devicesInfo,queryResult:=devconnmanager.QueryDevicesInfo(condition)
	if !queryResult{
		logs.Error("从数据库中查询设备信息失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "从数据库中查询设备信息失败", "failed")
		return
	}
	AssembleJsonResponse(c, _const.SUCCESS, "success",devicesInfo)
	return
}

func generateCardOfDeviceUpdate(deviceId string, deviceName string, deviceStatus string,teamId string) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm

	//插入提示信息
	messageText := utils.UpdateDeviceMessage
	messageForm := form.GenerateTextTag(&messageText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})

	//插入设备名称、UDID、平台
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.TeamIdHeader, teamId))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.DeviceIdHeader, deviceId))
	if deviceName!=""&&deviceName!=_const.UNDEFINED {
		cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.DeviceNameUpdateHeader, deviceName))
	}
	if deviceStatus!=""&&deviceName!=_const.UNDEFINED {
		cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.DeviceStatusHeader, deviceStatus))
	}

	return &cardFormArray
}

//生成更新工单通知卡片---action信息
func generateAction(url string,param *map[string]interface{}) *[]form.CardActionForm {
	var cardActions []form.CardActionForm
	var cardAction form.CardActionForm
	var buttons []form.CardButtonForm
	var text = utils.UpdateDeviceButtonText
	var hideOther = false
	button, err := form.GenerateButtonForm(&text, nil, nil, nil, "get", url, false, false, nil, nil, &hideOther)
	if err != nil {
		utils.RecordError("生成卡片button失败，", err)
	}
	buttons = append(buttons, *button)
	cardAction.Buttons = buttons
	cardActions = append(cardActions, cardAction)
	return &cardActions
}

func UpdateDeviceInfo(c *gin.Context) {
	logs.Info("HttpRequest：更新设备信息")
	var requestData devconnmanager.UpdateDeviceInfoRequest
	bindJsonError := c.ShouldBindJSON(&requestData)
	if bindJsonError != nil {
		logs.Error("绑定post请求body出错：%v", bindJsonError)
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	var resPerms devconnmanager.GetPermsResponse
	url := _const.USER_ALL_RESOURCES_PERMS_URL + "userName=" + requestData.OpUser
	queryPermsResult := queryPerms(url, &resPerms)
	if !queryPermsResult{
		logs.Error("查询权限失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "查询权限失败", "failed")
		return
	}
	perms := resPerms.Data[strings.ToLower(requestData.TeamId)+"_space_account"]
	checkResult := devconnmanager.CheckAllCertManagerAndAdmin(perms)
	if !checkResult{
		logs.Error("没有all_cert_manager及以上权限")
		AssembleJsonResponse(c, http.StatusForbidden, "没有all_cert_manager及以上权限", "failed")
		return
	}
	condition:=map[string]interface{}{
		"team_id":requestData.TeamId,
		"ud_id":requestData.UdId,
	}
	updateInfo := map[string]interface{}{
		"device_name": requestData.DeviceName,
		"device_status":requestData.DeviceStatus,
	}
	if requestData.AccountType == "Enterprise" {
		abot := service.BotService{}
		abot.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
		cardElementForms := generateCardOfDeviceUpdate(requestData.UdId,requestData.DeviceName, requestData.DeviceStatus,requestData.TeamId)
		param := map[string]interface{}{
			"team_id": requestData.TeamId,
			"ud_id": requestData.UdId,
			"op_user": requestData.DevicePrincipal,
		}
		cardActionForm:=generateAction(_const.UPDATE_DEVICE_FEEDBACK_URL,&param)
		err := sendIOSCertLarkMessage(cardElementForms,cardActionForm , requestData.DevicePrincipal, &abot)
		if err != nil {
			logs.Error("发送lark消息通知负责人往苹果后台更新设备信息失败：%v", err)
			AssembleJsonResponse(c, http.StatusInternalServerError, "发送lark消息通知负责人往苹果后台更新设备信息失败", "failed")
			return
		}
		if updateResult:=devconnmanager.UpdateDeviceInfoDB(condition,updateInfo);!updateResult{
			return
		}
		AssembleJsonResponse(c, _const.SUCCESS, "success", "")
		return
	}
	tokenString:=GetTokenStringByTeamId(requestData.TeamId)
	var appUpdDevInfoReq devconnmanager.AppUpdDevInfoReq
	appUpdDevInfoReq.Data.Id=requestData.DeviceId
	appUpdDevInfoReq.Data.Type=_const.APPLE_DEVICE_TYPE
	appUpdDevInfoReq.Data.Attributes.Status=requestData.DeviceStatus
	appUpdDevInfoReq.Data.Attributes.Name=requestData.DeviceName
	sendResult:=ReqToAppleHasObjMethod("PATCH",_const.APPLE_UPDATE_DEVICE_INFO_URL+requestData.DeviceId,tokenString,&appUpdDevInfoReq, &struct {}{})
	if !sendResult{
		logs.Error("请求苹果接口更新设备信息失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "更新设备信息失败", "failed")
		return
	}
	updateResult:=devconnmanager.UpdateDeviceInfoDB(condition,updateInfo)
	if !updateResult{
		logs.Error("更新设备信息失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "更新设备信息失败", "failed")
		return
	}
	AssembleJsonResponse(c, 0, "success","更新设备信息成功")
	return
}

func TransferDevicesResObj2DeviceInfo(teamId string,deviceResObj *devconnmanager.DevicesDataObjRes,deviceInfo *devconnmanager.DeviceInfo){
	deviceInfo.TeamId=teamId
	deviceInfo.DeviceId=deviceResObj.Id
	deviceInfo.UdId=deviceResObj.Attributes.Udid
	deviceInfo.DeviceStatus=deviceResObj.Attributes.Status
	deviceInfo.DeviceName=deviceResObj.Attributes.Name
	deviceInfo.DevicePlatform=deviceResObj.Attributes.Platform
	deviceInfo.DeviceModel=deviceResObj.Attributes.Model
	deviceInfo.DeviceClass=deviceResObj.Attributes.DeviceClass
	deviceInfo.DeviceAddedDate=deviceResObj.Attributes.AddedDate
}

func SynchronizeDeviceInfo(c *gin.Context){
	logs.Info("HttpRequest：同步苹果设备信息")
	var requestData devconnmanager.GetDevicesInfoRequest
	bindJsonError := c.ShouldBind(&requestData)
	if bindJsonError != nil {
		logs.Error("绑定请求body出错：%v", bindJsonError)
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	var devicesResObj devconnmanager.DevicesDataRes
	tokenString := GetTokenStringByTeamId(requestData.TeamId)
	deviceResult := GetAllEnableDevicesObj("ALL", "ALL", tokenString, &devicesResObj)
	if !deviceResult {
		logs.Error("从苹果获取设备信息失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "从苹果获取设备信息失败", "failed")
		return
	}
	for _, deviceResObj := range devicesResObj.Data {
		var deviceInfo devconnmanager.DeviceInfo
		TransferDevicesResObj2DeviceInfo(requestData.TeamId,&deviceResObj, &deviceInfo)
		condition := map[string]interface{}{
			"team_id": deviceInfo.TeamId,
			"ud_id":   deviceInfo.UdId,
		}
		result := devconnmanager.AddOrUpdateDeviceInfo(condition, &deviceInfo)
		if !result {
			logs.Error("同步设备信息失败,设备udid: "+deviceInfo.UdId)
		}
	}
	AssembleJsonResponse(c, 0, "success","同步设备信息成功")
}

func AsynUpdateDeviceFeedback(c *gin.Context){
	var feedbackInfo devconnmanager.UpdateDeviceFeedback
	err := c.ShouldBindJSON(&feedbackInfo)
	if err != nil {
		logs.Error("请求参数绑定失败！", err)
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "")
		return
	}
	condition := map[string]interface{}{
		"team_id": feedbackInfo.FeedBackJson.TeamId,
		"ud_id": feedbackInfo.FeedBackJson.UdId,
	}
	updateInfo := map[string]interface{}{
		"op_user": feedbackInfo.FeedBackJson.OpUser,
	}
	updateResult := devconnmanager.UpdateDeviceInfoDB(condition, updateInfo)
	if !updateResult {
		logs.Error("异步更新设备信息失败，Device ID："+feedbackInfo.FeedBackJson.UdId)
		AssembleJsonResponse(c, http.StatusInternalServerError, "数据库异步更新设备信息失败", "")
		return
	}
	AssembleJsonResponse(c, 0, "success", "")
	return
}

func TransferAppleRet2DeviceInfo(addDevInfoAppRet *devconnmanager.AddDevInfoAppRet,deviceInfo *devconnmanager.DeviceInfo){
	deviceInfo.DevicePlatform=addDevInfoAppRet.Data.Attributes.Platform
	deviceInfo.DeviceName=addDevInfoAppRet.Data.Attributes.Name
	deviceInfo.DeviceId=addDevInfoAppRet.Data.Id
	deviceInfo.DeviceAddedDate=addDevInfoAppRet.Data.Attributes.AddedDate
	deviceInfo.DeviceClass=addDevInfoAppRet.Data.Attributes.DeviceClass
	deviceInfo.DeviceModel=addDevInfoAppRet.Data.Attributes.Model
	deviceInfo.DeviceStatus=addDevInfoAppRet.Data.Attributes.Status
	deviceInfo.UdId=addDevInfoAppRet.Data.Attributes.Udid
	return
}

func generateCardOfDeviceAdd(deviceName string, udid string, platform string,teamId string) *[][]form.CardElementForm {
	var cardFormArray [][]form.CardElementForm

	//插入提示信息
	messageText := utils.AddDeviceMessage
	messageForm := form.GenerateTextTag(&messageText, false, nil)
	cardFormArray = append(cardFormArray, []form.CardElementForm{*messageForm})

	//插入设备名称、UDID、平台
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.TeamIdHeader,teamId))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.UDIDHeader, udid))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.DeviceNameAddHeader, deviceName))
	cardFormArray = append(cardFormArray, *generateInfoLineOfCard(utils.PlatformHeader, platform))

	return &cardFormArray
}

func AddDeviceInfo(c *gin.Context) {
	logs.Info("HttpRequest：添加设备信息")
	var requestData devconnmanager.AddDeviceInfoRequest
	bindJsonError := c.ShouldBindJSON(&requestData)
	if bindJsonError != nil {
		logs.Error("绑定post请求body出错：%v", bindJsonError)
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	var resPerms devconnmanager.GetPermsResponse
	url := _const.USER_ALL_RESOURCES_PERMS_URL + "userName=" + requestData.OpUser
	queryPermsResult := queryPerms(url, &resPerms)
	if !queryPermsResult{
		logs.Error("查询权限失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "查询权限失败", "failed")
		return
	}
	perms := resPerms.Data[strings.ToLower(requestData.TeamId)+"_space_account"]
	checkResult := devconnmanager.CheckAllCertManagerAndAdmin(perms)
	if !checkResult{
		logs.Error("没有all_cert_manager及以上权限")
		AssembleJsonResponse(c, http.StatusForbidden, "没有all_cert_manager及以上权限", "failed")
		return
	}
	if requestData.AccountType == "Enterprise" {
		abot := service.BotService{}
		abot.SetAppIdAndAppSecret(utils.IOSCertificateBotAppId, utils.IOSCertificateBotAppSecret)
		cardElementForms := generateCardOfDeviceAdd(requestData.DeviceName, requestData.Udid, requestData.DevicePlatform,requestData.TeamId)
		err := sendIOSCertLarkMessage(cardElementForms, nil, requestData.DevicePrincipal, &abot)
		if err != nil {
			logs.Error("发送lark消息通知负责人往苹果后台添加设备信息失败：%v", err)
			AssembleJsonResponse(c, http.StatusInternalServerError, "发送lark消息通知负责人往苹果后台添加设备信息失败", "failed")
		}
		AssembleJsonResponse(c, _const.SUCCESS, "success", "")
		return
	}
	tokenString:=GetTokenStringByTeamId(requestData.TeamId)
	var appleAddDeviceRequest devconnmanager.AppAddDevInfoReq
	appleAddDeviceRequest.Data.Type=_const.APPLE_DEVICE_TYPE
	appleAddDeviceRequest.Data.Attributes.Name=requestData.DeviceName
	appleAddDeviceRequest.Data.Attributes.Udid=requestData.Udid
	appleAddDeviceRequest.Data.Attributes.Platform=requestData.DevicePlatform
	var addDevInfoAppRet devconnmanager.AddDevInfoAppRet
	sendResult:=ReqToAppleHasObjMethod("POST",_const.APPLE_ADD_DEVICE_INFO_URL,tokenString,&appleAddDeviceRequest,&addDevInfoAppRet)
	if !sendResult{
		logs.Error("请求苹果接口添加设备信息失败")
		AssembleJsonResponse(c, http.StatusInternalServerError, "添加设备信息失败", "failed")
		return
	}
	var deviceInfo devconnmanager.DeviceInfo
	TransferAppleRet2DeviceInfo(&addDevInfoAppRet,&deviceInfo)
	deviceInfo.TeamId=requestData.TeamId
	addResult:=devconnmanager.AddDeviceInfoDB(&deviceInfo)
	if !addResult{
		logs.Error("往数据库中添加设备信息出错")
		AssembleJsonResponse(c, http.StatusInternalServerError, "添加设备信息失败", "failed")
		return
	}
	AssembleJsonResponse(c, _const.SUCCESS, "success","添加设备信息成功")
	return
}

func AsynAddDeviceFeedback(c *gin.Context){
	logs.Info("HttpRequest：添加设备信息反馈")
	var requestData devconnmanager.DeviceInfo
	bindJsonError := c.ShouldBindJSON(&requestData)
	if bindJsonError != nil {
		logs.Error("绑定post请求body出错：%v", bindJsonError)
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	requestData.DeviceStatus=_const.APPLE_DEVICE_ENABLED_STATUS
	requestData.DeviceAddedDate=time.Now().Format("2006-01-02 15:04:05")
	addResult:=devconnmanager.AddDeviceInfoDB(&requestData)
	if !addResult{
		logs.Error("往数据库中添加设备信息出错")
		AssembleJsonResponse(c, http.StatusInternalServerError, "添加设备信息失败", "failed")
		return
	}
	AssembleJsonResponse(c, _const.SUCCESS, "success","添加设备信息成功")
	return
}