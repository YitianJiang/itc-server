package main

import (
	"code.byted.org/clientQA/itc-server/casemanage"
	"code.byted.org/clientQA/itc-server/controllers"
	"code.byted.org/clientQA/itc-server/detect"
	"code.byted.org/clientQA/itc-server/developerconnmanager"
	"code.byted.org/clientQA/itc-server/middleware"
	"code.byted.org/clientQA/itc-server/settings"
	"code.byted.org/gin/ginex"
)

func InitRouter(r *ginex.Engine) {

	r.POST("/import-old-data-android", detect.ImportOldDataAndroid) // TEMP
	r.POST("/import-old-data-ios", detect.ImportOldDataiOS)         // TEMP

	r.POST("/settings", settings.Insert)
	r.PUT("/settings", settings.Refresh)
	r.GET("/settings", settings.Show)
	api := r.GroupEX("/api")
	//二进制包检测回调接口
	r.POST("/updateDetectInfos", detect.UpdateDetectTask)
	r.POST("/updateOtherDetectInfos", detect.UpdateOtherDetectInfos)

	//获取鉴权接口
	r.GET("/t/generateToken", detect.GetToken)

	// 获取检测项配置信息
	r.GET("/api/detectConfig/getItcDetectConfig", detect.GetDetectConfig)
	r.GET("/api/detect/config/tmp", detect.GetTMPConfig)
	// 获取检测任务结果
	r.GET("/api/detect/getTaskDetailNotItc", detect.GetSpecificAppVersionDetectResults)

	r.GET("/open-apis/detect/task/result", detect.GetDetectTaskResult)
	detectapi := r.GroupEX("/detect")
	detectapi.POST("/new/uploadUnconfirmedDetections", detect.UploadUnconfirmedDetections)
	detectapi.Use(middleware.JWTCheck())
	{
		detectapi.POST("/new/list", detect.List)
		detectapi.GET("/new/detail", detect.Detail)
		detectapi.PUT("/new/confirmation", detect.Confirm)
		detectapi.DELETE("/new", detect.Delete)
	}

	//检测服务检测异常报警接口
	api.POST("/check_server/alarm", detect.Alram)
	//获取IOS隐私调用API信息
	api.GET("/perm/getPerms", detect.GetPermsInfo)

	api.Use(middleware.JWTCheck())
	{
		//上传ipa和apk
		api.POST("/uploadFile", detect.UploadFile)
		//上传aar------fj
		api.POST("/uploadFileOther", detect.NewOtherDetect)
		//增加检查项
		api.POST("/addDetectItem", detect.AddDetectItem)
		//二进制任务查询
		api.GET("/queryTasks", detect.QueryDetectTasks)
		//获取任务对应的自查项
		api.GET("/getSelfCheckItems", detect.GetSelfCheckItems)
		//删除检查项
		api.POST("/deleteDetectItem", detect.DropDetectItem)
		//完成自查
		api.POST("/confirmCheck", detect.ConfirmSelfCheckItems)
		//获取检测列表
		api.GET("/queryDetectTools", detect.QueryDetectTools)
		//获取当前任务已选择的检测工具列表
		api.GET("/task/queryTools", detect.QueryTaskQueryTools)
		//获取当前任务的二进制工具检测内容
		api.GET("/task/queryBinaryContent", detect.QueryTaskBinaryCheckContent)
		//获取当前任务的apk二进制工具检测内容
		api.GET("/task/queryApkBinaryContent", detect.QueryTaskApkBinaryCheckContentWithIgnorance_3)
		//新增二进制检测工具
		api.POST("/tool/insert", detect.InsertBinaryTool)
		//查询二进制检测工具列表
		api.GET("/tool/queryTools", detect.QueryBinaryTools)
		//新增lark消息提醒配置
		api.POST("/config/larkMsgCall", detect.InsertLarkMsgCall)
		//查询lark消息提醒配置
		api.GET("/config/queryLarkMsgCall", detect.QueryLarkMsgCall)
		//确认二进制包检测信息
		api.POST("/detect/confirmResult", detect.ConfirmBinaryResult)
		//确认apk二进制包检测信息
		api.POST("/detect/confirmApkResult", detect.ConfirmAndroid)
		//根据platform获取配置的问题类型
		api.GET("/config/queryProblemConfigs", detect.QueryProblemConfigs)
		//增加配置项
		api.POST("/config/addConfig", detect.AddConfig)
		//查询配置项
		api.GET("/config/queryConfigs", detect.QueryConfigs)
		//lark消息提醒
		api.POST("/lark", detect.LarkMsg)
		//新建ios源码隐私检测任务
		api.POST("/ios/scNewTask", detect.ScNewTask)
		//增加lark群配置
		api.POST("/lark/addGroup", detect.AddLarkGroup)
		//查询lark群配置
		api.GET("/lark/queryGroups", detect.QueryGroupInfosByTimerId)
		//更新lark群配置
		api.POST("/lark/updateGroup", detect.UpdateLarkGroup)
		//删除lark群配置
		api.DELETE("/lark/deleteGroup", detect.DeleteGroupInfoById)
		//添加证书
		api.POST("/certificate", controllers.AddCertificate)
		//查询展示证书
		api.GET("/certificates", controllers.GetCertificates)
		//过期证书提醒
		api.GET("/certificates/controller", controllers.CertificateController)
		//证书删除
		api.POST("/certificate/delete", controllers.DeteleCertificate)
		//获取iOS当前任务的二进制工具检测内容
		api.GET("/task/queryIOSBinaryContent", detect.QueryIOSTaskBinaryCheckContent)
		//确认iOS二进制检测信息
		api.POST("/detect/confirmIOSResult", detect.ConfirmiOS)
		//查询权限确认历史
		api.POST("detect/queryIgnoreHistory", detect.QueryIgnoredHistory)
		//新增权限
		api.POST("/perm/addPermission", detect.AddDetectConfig)
		//删除权限
		//api.GET("/perm/deletePermission",detect.DeleteDetectConfig)
		//修改权限
		api.POST("/perm/editPermission", detect.EditDectecConfig)
		//查询权限
		api.POST("/perm/queryPermission", detect.QueryDectecConfig)
		//根据权限查询信息
		api.POST("/perm/queryWithPermission", detect.GetRelationsWithPermission)
		//根据App查询权限信息
		api.POST("/perm/queryPermissionsOfApp", detect.QueryPermissionsWithApp)
		//查询权限详情
		api.GET("/perm/getpermDetails", detect.GetPermDetails)
		//获取app的版本号---权限关联查询使用
		api.GET("/perm/getAppVesions", detect.GetAppVersions)
		//aar检测结果查询
		api.GET("/detect/getAarDetectResults", detect.QueryAarBinaryDetectResult)
		//aar检测结果确认
		api.POST("/detect/confirmAarResult", detect.ConfirmAarDetectResult)
		//aar任务列表查询
		api.POST("/detect/getAarTaskList", detect.GetOtherDetectTaskList)
		//查询被拒案例
		api.GET("/casemanage/queryRejCases", casemanage.GetRejCasesByConditions)
		//新增被拒案例
		api.POST("/casemanage/addRejCase", casemanage.AddRejCase)
		//删除被拒案例
		api.POST("/casemanage/deleteRejCase", casemanage.DeleteRejCase)
		//更新被拒案例
		api.POST("/casemanage/updateRejCase", casemanage.EditRejCaseofSolution)
		//组件平台结果接口
		api.GET("/aar/getAarTaskDetail", detect.GetAARInfoNotITC)

	}
	//todo 巩锐开始开发证书体系监管后台API
	connapi := r.Group("/v1/devConnManage")
	{
		connapi.GET("/getBundleIdsList", developerconnmanager.GetBunldIdsObj)
		connapi.GET("/testPrivatePrint", developerconnmanager.ParsePrivateKey)
		connapi.GET("/createProvProfile", developerconnmanager.Test64DecodeToString)
		connapi.POST("/createP8DBInfo", developerconnmanager.CreateP8DBInfoToTable)

	}
	accountapi := r.Group("/v1/accountManage")
	{
		accountapi.POST("/accountInfoUpdate", developerconnmanager.UpdateAccount)
		accountapi.GET("/accountInfoGet", developerconnmanager.QueryAccount)
		accountapi.POST("/accountInfoWriter", developerconnmanager.InsertAccount)
		accountapi.DELETE("/accountInfoDelete", developerconnmanager.DeleteByTeamId)
	}
	certificateapi := r.Group("/v1/appleCertManage/")
	{
		certificateapi.GET("/certificateInfoGet", developerconnmanager.QueryCertificatesInfo)
		certificateapi.POST("/certificateCreate", developerconnmanager.InsertCertificate)
		certificateapi.DELETE("/certificateDelete", developerconnmanager.DeleteCertificate)
		certificateapi.GET("/certExpireDateCheck", developerconnmanager.CheckCertExpireDate)
		certificateapi.POST("/privUploadForCert", developerconnmanager.UploadPrivKey)
		certificateapi.POST("/certificateUpload", developerconnmanager.UploadCertificate)
		certificateapi.POST("/asynDeleteFeedback", developerconnmanager.AsynDeleteCertFeedback)
	}
	usermanagerapi := r.Group("/v1/userManage")
	{
		//账号人员管理相关开发
		usermanagerapi.GET("/userRolesGet", developerconnmanager.UserRolesGetInfo)
		usermanagerapi.GET("/userInfoGet", developerconnmanager.UserDetailInfoGet)
		usermanagerapi.GET("/userInvitedInfoGet", developerconnmanager.UserInvitedDetailInfoGet)
		usermanagerapi.GET("/visibleAppsFromAccount", developerconnmanager.VisibleAppsInfoGet)
		usermanagerapi.GET("/visibleAppsOfUser", developerconnmanager.VisibleAppsOfUserGet)
		usermanagerapi.POST("/editPermOfUser", developerconnmanager.EditPermOfUserFunc)
		usermanagerapi.POST("/userInvitations", developerconnmanager.UserInvitedFunc)
		usermanagerapi.POST("/userDelete", developerconnmanager.UserDeleteFunc)
		usermanagerapi.POST("/userInvitationsDelete", developerconnmanager.UserInvitedDeleteFunc)
	}

	appSignManager := r.Group("/v1/appleConnManage")
	{
		appSignManager.GET("/appDetailInfoGet", developerconnmanager.GetAppSignListDetailInfo)
		appSignManager.GET("/getCapabilitiesInfo", developerconnmanager.GetBundleIdCapabilitiesInfo)
		appSignManager.POST("/createAppBindAccount", developerconnmanager.CreateAppBindAccount)
		appSignManager.POST("/approveAppBindAccountFeedback", developerconnmanager.ApproveAppBindAccountFeedback)
		//接口删除app相关所有信息
		appSignManager.DELETE("/deleteAppAllInfo", developerconnmanager.DeleteAppAllInfoFromDB)

		//接口绑定\换绑签名证书接口
		appSignManager.POST("/appBindCert", developerconnmanager.AppBindCert)

		//接口生成Profile描述文件
		appSignManager.POST("/createOrUpdateProfile", developerconnmanager.CreateOrUpdateProfile)
		//接口上传Profile描述文件
		appSignManager.POST("/profileUpload", developerconnmanager.ProfileUploadFunc)
		appSignManager.DELETE("/deleteProfile", developerconnmanager.DeleteProfile)
		appSignManager.POST("/asynDeleteProfileFeedback", developerconnmanager.AsynProfileDeleteFeedback)
		appSignManager.DELETE("/deleteBundleid", developerconnmanager.DeleteBundleid)
		appSignManager.POST("/asynDeleteBundleFeedback", developerconnmanager.AsynBundleDeleteFeedback)
		appSignManager.POST("/createBundleProfile", developerconnmanager.CreateOrUpdateOrRestoreBundleId)
		appSignManager.POST("/updateBundleIdId", developerconnmanager.UpdateBundleIdIdOfBundleId)
		appSignManager.POST("/finishTicket", developerconnmanager.FinishTicketFeedback)
		appSignManager.GET("/primaryBundleIds", developerconnmanager.QueryPrimaryBundleIds)
	}

	authorizationManager := r.Group("/v1/authorization")
	{
		authorizationManager.POST("/applyForAuthorization", developerconnmanager.ApplyForAuthorization)
		authorizationManager.POST("/approveAuthorizationApplication", developerconnmanager.ApproveAuthorizationApplication)
	}

	deviceapi := r.Group("/v1/deviceManage")
	{
		deviceapi.POST("/deviceInfoAdd", developerconnmanager.AddDeviceInfo)
		deviceapi.GET("/deviceInfoGet", developerconnmanager.QueryDeviceInfo)
		deviceapi.POST("/deviceInfoUpdate", developerconnmanager.UpdateDeviceInfo)
		deviceapi.POST("/asynAddDeviceFeedback", developerconnmanager.AsynAddDeviceFeedback)
		deviceapi.POST("/asynUpdateDeviceFeedback", developerconnmanager.AsynUpdateDeviceFeedback)
		deviceapi.GET("/deviceInfoSynchronize", developerconnmanager.SynchronizeDeviceInfo)
	}

	timedtaskapi := r.Group("/v1/timedTask")
	{
		timedtaskapi.GET("/CertExpiredNotify", developerconnmanager.NotifyCertExpired)
		timedtaskapi.GET("/ProfileExpiredNotify", developerconnmanager.NotifyProfileExpired)
	}

	binaryCheckApi := r.Group("/v1/checkBinaryResult")
	{
		binaryCheckApi.POST("/checkIsBinary", developerconnmanager.CheckIsBinary)
	}

	testflightapi := r.Group("/v1/testflightserver")
	{
		testflightapi.GET("/getRecentVersionReviewInfo", developerconnmanager.GetRecentVersionReviewInfo)
		testflightapi.POST("/createGroupAddVesion", developerconnmanager.CreateGroupAddVesion)
		testflightapi.POST("/deleteGroupTester", developerconnmanager.DeleteGroupTester)
		testflightapi.POST("/uploadFileToTos", developerconnmanager.UploadFileToTos)
	}

	//哭泣，因为DBA下掉了所有的gui操作数据库的功能，固有此变更数据的接口出现
	operationDBApi := r.Group("/v1/deleteBundleTable")
	{
		operationDBApi.POST("deleteBundleIdCap", developerconnmanager.DeleteBundleIdCap)
		operationDBApi.POST("deleteCertPrivKey", developerconnmanager.DeleteCertPrivKey)
		operationDBApi.POST("insertCertInfoTest", developerconnmanager.InsertCertInfoTest)
	}
}
