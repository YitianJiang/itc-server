package main

import (
	"code.byted.org/clientQA/itc-server/detect"
	"code.byted.org/gin/ginex"
)

func InitRouter(r *ginex.Engine){

	//上传ipa和apk
	r.POST("/uploadFile", detect.UploadFile)
	//二进制包检测回调接口
	r.POST("/updateDetectInfos", detect.UpdateDetectInfos)
	//增加检查项
	r.POST("/addDetectItem", detect.AddDetectItem)
	//二进制任务查询
	r.GET("/queryTasks", detect.QueryDetectTasks)
	//获取任务对应的自查项
	r.GET("/getSelfCheckItems", detect.GetSelfCheckItems)
	//完成自查
	r.POST("/confirmCheck", detect.ConfirmCheck)
	//获取检测列表
	r.GET("/queryDetectTools", detect.QueryDetectTools)
	//获取当前任务已选择的检测工具列表
	r.GET("/task/queryTools", detect.QueryTaskQueryTools)
	//获取当前任务的二进制工具检测内容
	r.GET("/task/queryBinaryContent", detect.QueryTaskBinaryCheckContent)
	//新增二进制检测工具
	r.POST("/tool/insert", detect.InsertBinaryTool)
	//查询二进制检测工具列表
	r.GET("/tool/queryTools", detect.QueryBinaryTools)
	//新增lark消息提醒配置
	r.POST("/config/larkMsgCall", detect.InsertLarkMsgCall)
	//查询lark消息提醒配置
	r.GET("/config/queryLarkMsgCall", detect.QueryLarkMsgCall)
	//确认二进制包检测信息
	r.POST("/detect/confirmResult", detect.ConfirmBinaryResult)
	//更新自查项
	//r.POST("/detect/updateItem", detect.UpdateItem)
	//增加配置项
	r.POST("/config/addConfig", detect.AddConfig)
	//查询配置项
	r.GET("/config/queryConfigs", detect.QueryConfigs)
}
