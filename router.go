package main

import (
	"code.byted.org/clientQA/itc-server/controllers"
	"code.byted.org/clientQA/itc-server/detect"
	"code.byted.org/clientQA/itc-server/middleware"
	"code.byted.org/gin/ginex"
)

func InitRouter(r *ginex.Engine){

	api := r.GroupEX("/api")
	//二进制包检测回调接口
	r.POST("/updateDetectInfos", detect.UpdateDetectInfos)
	//获取鉴权接口
	r.GET("/t/generateToken", detect.GetToken)

	api.Use(middleware.JWTCheck())
	{
		//上传ipa和apk
		api.POST("/uploadFile", detect.UploadFile)
		//增加检查项
		api.POST("/addDetectItem", detect.AddDetectItem)
		//二进制任务查询
		api.GET("/queryTasks", detect.QueryDetectTasks)
		//获取任务对应的自查项
		api.GET("/getSelfCheckItems", detect.GetSelfCheckItems)
		//完成自查
		api.POST("/confirmCheck", detect.ConfirmCheck)
		//获取检测列表
		api.GET("/queryDetectTools", detect.QueryDetectTools)
		//获取当前任务已选择的检测工具列表
		api.GET("/task/queryTools", detect.QueryTaskQueryTools)
		//获取当前任务的二进制工具检测内容
		api.GET("/task/queryBinaryContent", detect.QueryTaskBinaryCheckContent)
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


	}
}
