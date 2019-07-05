package detect

import (
	"bytes"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

/**
	aar新建检测任务接口
 */
func NewOtherDetect(c *gin.Context){

	url := ""
	nameI, f := c.Get("username")
	if !f {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未获取到用户信息！",
			"errorCode": -1,
			"data":      "未获取到用户信息！",
		})
		return
	}
	file, header, err := c.Request.FormFile("uploadFile")
	if file == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未选择上传的文件！",
			"errorCode": -2,
		})
		logs.Error("未选择上传的文件！")
		return
	}
	defer file.Close()
	filename := header.Filename
	//发送lark消息到个人
	toLarker := c.DefaultPostForm("toLarker", "")
	var name string
	if toLarker == "" {
		name = nameI.(string)
	} else {
		name = nameI.(string) + "," + toLarker
	}
	//发送lark消息到群
	toGroup := c.DefaultPostForm("toLarkGroupId", "")

	platform := c.DefaultPostForm("platform", "")
	if platform == "" {
		platform = "0"
	}

	appId := c.DefaultPostForm("appId", "")

	checkItem := c.DefaultPostForm("checkItem", "")
	logs.Info("checkItem: ", checkItem)

	//检验文件格式是否是aar
	flag := strings.HasSuffix(filename, ".aar")
	if !flag {
		errorFormatFile(c)
		return
	}
	url  =  "http://" + DETECT_URL_PRO + "/apk_post/v2"
	//url  = "http://" + Local_URL_PRO + "/apk_post/v2"
	_tmpDir := "./tmp"
	exist, err := PathExists(_tmpDir)
	if !exist {
		os.Mkdir(_tmpDir, os.ModePerm)
	}
	out, err := os.Create(_tmpDir + "/" + filename)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "安装包文件处理失败，请联系相关人员！",
			"errorCode": -6,
		})
		logs.Fatal("临时文件保存失败")
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   "安装包文件处理失败，请联系相关人员！",
			"errorCode": -6,
		})
		logs.Fatal("临时文件保存失败")
		return
	}
	//调试，暂时注释
	recipients := name
	filepath := _tmpDir + "/" + filename
	//1、上传至tos,测试暂时注释
	//tosUrl, err := upload2Tos(filepath)
	//2、将相关信息保存至数据库
	var dbDetectModel dal.OtherDetectModel
	dbDetectModel.Creator = nameI.(string)
	dbDetectModel.ToLarker = name
	dbDetectModel.ToGroup = toGroup
	dbDetectModel.Platform, _ = strconv.Atoi(platform)
	dbDetectModel.AppId = appId
	dbDetectModel.FileType = "aar"
	dbDetectModel.Status = 0//增加状态字段，0---未完全确认；1---已完全确认
	dbDetectModelId := dal.InsertOtherDetect(dbDetectModel)
	//3、调用检测接口，进行二进制检测 && 删掉本地临时文件
	if checkItem == "" {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未选择检测工具",
			"errorCode": -1,
			"data":      "未选择检测工具",
		})
		return
	}
	if dbDetectModelId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":   "数据库插入记录失败，请确认数据库字段与结构体定义一致",
			"errorCode": -1,
			"data":      "数据库插入记录失败，请确认数据库字段与结构体定义一致",
		})
		return
	}
	//go upload2Tos(filepath, dbDetectModelId)
	go func() {
		callBackUrl := "https://itc.bytedance.net/updateOtherDetectInfos"
		//callBackUrl := "http://10.224.13.149:6789/updateOtherDetectInfos"
		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)
		bodyWriter.WriteField("recipients", recipients)
		bodyWriter.WriteField("callback", callBackUrl)
		bodyWriter.WriteField("taskID", fmt.Sprint(dbDetectModelId))
		bodyWriter.WriteField("toolIds", checkItem)
		fileWriter, err := bodyWriter.CreateFormFile("file", filepath)
		if err != nil {
			logs.Error("%s", "error writing to buffer: "+err.Error())
			c.JSON(http.StatusOK, gin.H{
				"message":   "上传文件处理错误，请联系相关人员！",
				"errorCode": -1,
				"data":      "上传文件处理错误，请联系相关人员！",
			})
			return
		}
		fileHandler, err := os.Open(filepath)
		defer fileHandler.Close()
		_, err = io.Copy(fileWriter, fileHandler)
		contentType := bodyWriter.FormDataContentType()
		err = bodyWriter.Close()
		logs.Info("url: ", url)
		tr := http.Transport{DisableKeepAlives: true}
		toolHttp := &http.Client{
			Timeout: 300 * time.Second,
			Transport: &tr,
		}
		response, err := toolHttp.Post(url, contentType, bodyBuffer)
		if err != nil {
			logs.Error("taskId:"+fmt.Sprint(dbDetectModelId)+",aar上传文件处理错误", err)
			//及时报警
			utils.LarkDingOneInner("kanghuaisong", "二进制包检测服务无响应，请及时进行检查！aar任务ID："+fmt.Sprint(dbDetectModelId)+",创建人："+dbDetectModel.Creator)
			utils.LarkDingOneInner("yinzhihong", "二进制包检测服务无响应，请及时进行检查！aar任务ID："+fmt.Sprint(dbDetectModelId)+",创建人："+dbDetectModel.Creator)
			utils.LarkDingOneInner("fanjuan.xqp", "二进制包检测服务无响应，请及时进行检查！aar任务ID："+fmt.Sprint(dbDetectModelId)+",创建人："+dbDetectModel.Creator)
		}
		resBody := &bytes.Buffer{}
		if response != nil {
			defer response.Body.Close()
			_, err = resBody.ReadFrom(response.Body)
			var data map[string]interface{}
			data = make(map[string]interface{})
			json.Unmarshal(resBody.Bytes(), &data)
			logs.Info("upload detect url's response: %+v", data)
			//删掉临时文件
			os.Remove(filepath)
		}
	}()
	c.JSON(http.StatusOK, gin.H{
		"message":   "文件上传成功，请等待检测结果通知",
		"errorCode": 0,
		"data": map[string]interface{}{
			"taskId": dbDetectModelId,
		},
	})

}
/**
	aar检测成功结果回调接口
 */
func UpdateOtherDetectInfos(c *gin.Context)  {
	logs.Info("回调开始，更新aar检测信息～～～")
	taskId := c.Request.FormValue("task_ID")
	if taskId == "" {
		logs.Error("缺少task_ID参数")
		c.JSON(http.StatusOK, gin.H{
			"message":   "缺少task_ID参数",
			"errorCode": -1,
			"data":      "缺少task_ID参数",
		})
		return
	}
	toolId := c.Request.FormValue("tool_ID")
	jsonContent := c.Request.FormValue("jsonContent")

	detect := dal.QueryOtherDetectModelsByMap(map[string]interface{}{
		"id": taskId,
	})
	if detect == nil || len(*detect) == 0 {
		logs.Error("未查询到该taskid对应的aar检测任务，%v", taskId)
		c.JSON(http.StatusOK, gin.H{
			"message":   "未查询到该taskid对应的aar检测任务",
			"errorCode": -2,
			"data":      "未查询到该taskid对应的aar检测任务",
		})
		return
	}

	if (*detect)[0].Platform == 0 {
		//检测信息分析，并将检测信息写库-----fj
		mapInfo := make(map[string]int)
		mapInfo["taskId"], _ = strconv.Atoi(taskId)
		mapInfo["toolId"], _ = strconv.Atoi(toolId)
		errApk := OtherJsonAnalysis_2(jsonContent, mapInfo)
		if errApk != nil {
			return
		}
	}
	//ios新检测内容存储
	if (*detect)[0].Platform == 1 {

	}

	//进行lark消息提醒
	var message string
	detect = dal.QueryOtherDetectModelsByMap(map[string]interface{}{
		"id": taskId,
	})
	creators := (*detect)[0].ToLarker
	larkList := strings.Split(creators, ",")
	//for _,creator := range larkList {
	message = "你好，" + (*detect)[0].OtherName + " " + (*detect)[0].OtherVersion
	message += " aar包完成二进制检测！\n"

	//此处测试时注释掉
	larkUrl := "http://rocket.bytedance.net/rocket/itc/basic?showAarDetail=1&aarTaskId=" + taskId
	message += "地址链接：" + larkUrl
	for _, creator := range larkList {
		utils.LarkDingOneInner(creator, message)
	}
	//给群ID对应群发送消息
	toGroupID := (*detect)[0].ToGroup
	if toGroupID != "" {
		group := strings.Replace(toGroupID, "，", ",", -1) //中文逗号切换成英文逗号
		groupArr := strings.Split(group, ",")
		for _, group_id := range groupArr {
			to_lark_group := strings.Trim(group_id, " ")
			utils.LarkGroup(message, to_lark_group)
		}
	}
}

/**
	查询aar检测任务列表接口
 */
func GetOtherDetectTaskList(c *gin.Context)  {
	type queryStruct struct {
		PageSize 		int			`json:"pageSize"`
		Page 			int			`json:"page"`
		Info    		string		`json:"info"`
		Creator 		string		`json:"creator"`
	}
	var t queryStruct
	param,_ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(param,&t)
	if err != nil {
		logs.Error("query detectConfig 传入参数不合法！%v",err)
		errorReturn(c,"传入参数不合法！")
		return
	}
	if t.Page <= 0 || t.PageSize <= 0 {
		logs.Error("分页参数不合法!")
		errorReturn(c,"分页参数不合法")
		return
	}
	pageInfo := make(map[string]int)
	pageInfo["pageSize"] = t.PageSize
	pageInfo["page"]= t.Page

	condition := "1=1"
	if t.Creator != "" {
		condition += " and `creator` ='"+t.Creator+"'"
	}
	if t.Info != "" {
		condition += " and `other_name` like '%"+t.Info+"%'"
	}
	list,count,errL := dal.QueryOtherDetctModelOfList(pageInfo,condition)
	if errL != nil {
		logs.Error("查询aar任务列表数据库操作失败，查询信息：%v,错误信息：%v",t,errL)
		errorReturn(c,"查询aar任务列表数据库操作失败")
		return
	}
	if list == nil || len(*list)== 0 {
		logs.Error("无对应aar任务，查询信息：%v",t)
		c.JSON(http.StatusOK,gin.H{
			"errorCode":0,
			"message":"success",
			"data":map[string]interface{}{
				"count":count,
				"result":make([]dal.OtherDetectModel,0),
			},
		})
	}else {
		logs.Info("查询aar列表成功")
		c.JSON(http.StatusOK,gin.H{
			"errorCode":0,
			"message":"success",
			"data":map[string]interface{}{
				"count":count,
				"result":(*list),
			},
		})
	}
	return
}


/**
	查询aar检测详情接口
 */
func QueryAarBinaryDetectResult(c *gin.Context)  {
	taskId := c.DefaultQuery("taskId", "")
	if taskId == "" {
		logs.Error("缺少taskId参数")
		errorReturn(c,"缺少taskId参数")
		return
	}
	toolId := c.DefaultQuery("toolId", "")
	if toolId == "" {
		logs.Error("缺少toolId参数")
		errorReturn(c,"缺少toolId参数")
		return
	}
	//获取任务信息
	detect := dal.QueryOtherDetectModelsByMap(map[string]interface{}{
		"id" : taskId,
	})
	if detect == nil || len(*detect)==0{
		logs.Error("未查询到该taskid对应的检测任务，%v", taskId)
		errorReturn(c,"未查询到该taskid对应的检测任务")
		return
	}

	infos := dal.QueryOtherDetectDetail(map[string]interface{}{
		"task_id":taskId,
	})
	if infos == nil || len(*infos) == 0 {
		logs.Error("未查询到aar检测结果，任务ID："+taskId)
		errorReturn(c,"未查询到aar检测结果")
		return
	}
	//获取检测结果
	finalResult := getOtherDetectDetail(c,infos,&(*detect)[0])

	logs.Info("查询aar检测结果成果！")
	c.JSON(http.StatusOK,gin.H{
		"message":"success",
		"errorCode":0,
		"data":*finalResult,
	})
	return
}
/**
	确认aar信息接口---无diff
 */
func ConfirmAarDetectResult(c *gin.Context)  {
	param, _ := ioutil.ReadAll(c.Request.Body)
	var t dal.PostConfirm//t.Type为0敏感方法，1敏感字符串，2权限
	err := json.Unmarshal(param, &t)
	if err != nil {
		logs.Error("参数不合法 ，%v",err)
		errorReturn(c,"参数不合法")
		return
	}
	//获取确认人信息
	username, _ := c.Get("username")
	usernameStr := username.(string)

	//确认信息写入
	detail := dal.QueryOtherDetectDetail(map[string]interface{}{
		"task_id":t.TaskId,
		"detail_type":t.Type,
		"sub_index":t.Index-1,
		"tool_id":t.ToolId,

	})
	if detail == nil || len(*detail) == 0 {
		logs.Error("无该检测结果")
		errorReturn(c,"无该检测结果")
		return
	}
	var newInfos []byte
	if t.Type == 0 {
		var p []dal.SMethod
		if err := json.Unmarshal([]byte((*detail)[0].DetectInfos),&p);err != nil {
			logs.Error("aar检测结果信息存储格式错误，task ID："+fmt.Sprint(t.TaskId))
			errorReturn(c,"aar检测结果权限信息存储格式错误")
			return
		}
		if !judgeOutOfRange(c,t.Id,len(p)){
			return
		}
		p[t.Id-1].Status = t.Status
		p[t.Id-1].Confirmer = usernameStr
		p[t.Id-1].Remark = t.Remark
		newInfos,_ = json.Marshal(p)
	}else if t.Type == 1 {
		var p []dal.SStr
		if err := json.Unmarshal([]byte((*detail)[0].DetectInfos),&p);err != nil {
			logs.Error("aar检测结果信息存储格式错误，task ID："+fmt.Sprint(t.TaskId))
			errorReturn(c,"aar检测结果权限信息存储格式错误")
			return
		}
		if !judgeOutOfRange(c,t.Id,len(p)){
			return
		}
		p[t.Id-1].Status = t.Status
		p[t.Id-1].Confirmer = usernameStr
		p[t.Id-1].Remark = t.Remark
		//每个字符串的确认信息
		for i := 0; i<len(p[t.Id-1].ConfirmInfos);i++ {
			p[t.Id-1].ConfirmInfos[i].Status = t.Status
			p[t.Id-1].ConfirmInfos[i].Confirmer = usernameStr
			p[t.Id-1].ConfirmInfos[i].Remark = t.Remark
		}
		newInfos,_ = json.Marshal(p)
	}else if t.Type == 2 {
		var p []dal.Permissions
		if err := json.Unmarshal([]byte((*detail)[0].DetectInfos),&p);err != nil {
			logs.Error("aar检测结果信息存储格式错误，task ID："+fmt.Sprint(t.TaskId))
			errorReturn(c,"aar检测结果权限信息存储格式错误")
			return
		}
		if !judgeOutOfRange(c,t.Id,len(p)){
			return
		}
		p[t.Id-1].Status = t.Status
		p[t.Id-1].Confirmer = usernameStr
		p[t.Id-1].Remark = t.Remark
		newInfos,_ = json.Marshal(p)
	}
	condition := "id = '"+fmt.Sprint((*detail)[0].ID)+"'"
	if err := dal.UpdateOtherDetailInfoByMap(condition, map[string]interface{}{
		"detect_infos":string(newInfos),
	}); err != nil {
		logs.Error("更新aar确认信息失败，确认信息："+fmt.Sprint(t))
		errorReturn(c,"更新aar确认信息失败")
		return
	}
	//任务状态改变
	if !updateAarTaskStatus(t.TaskId) {
		errorReturn(c,"aar任务状态更改失败！")
		return
	}
	logs.Info("确认AAR信息成功")
	c.JSON(http.StatusOK,gin.H{
		"message":"success",
		"errorCode":0,
	})
}

/**
	aar结果解析
*/
func OtherJsonAnalysis_2(info string,mapInfo map[string]int) error  {
	logs.Info("aar安卓解析开始～～～～")
	var fisrtResult dal.JSONResultStruct
	err_f := json.Unmarshal([]byte(info),&fisrtResult)
	if err_f != nil {
		logs.Error("taskId:"+fmt.Sprint(mapInfo["taskId"])+",arr包检测返回信息格式错误！,%v",err_f)
		message := "taskId:"+fmt.Sprint(mapInfo["taskId"])+",arr包检测返回信息格式错误，请解决;"+fmt.Sprint(err_f)
		utils.LarkDingOneInner("fanjuan.xqp",message)
		return err_f
	}

	infos := make([]dal.OtherDetailInfoStruct,0)
	for index,result := range fisrtResult.Result {
		appInfos := result.AppInfo
		methodsInfo := result.MethodInfos
		strsInfo := result.StrInfos
		//missCalls := result.MissSearchInfos
		basicInfos,errB := otherBasicInfoAna(appInfos,mapInfo,index)
		if errB != nil {
			return errB
		}
		infos = append(infos,basicInfos)

		//敏感method解析----先外层去重
		mRepeat := make(map[string]int)
		newMethods := make([]dal.MethodInfo,0)//第一层去重后的敏感方法集
		for _, method := range methodsInfo {
			var keystr = method.MethodName+method.ClassName
			if v,ok := mRepeat[keystr]; (!ok||ok&&v==0){
				newMethods = append(newMethods, method)
				mRepeat[keystr]=1
			}
		}
		infos = append(infos,otherMethodAna(&newMethods,mapInfo,index))
		infos = append(infos,otherStrAna(&strsInfo,mapInfo,index))

		otherPerms, errP := otherPermAna(&appInfos.PermsInAppInfo,mapInfo,index)
		if errP != nil {
			return errP
		}
		infos = append(infos,otherPerms)
	}
	err := dal.InsertOtherDetectDetailBatch(&infos)
	if err != nil {
		message := "taskId:"+fmt.Sprint(mapInfo["taskId"])+",arr包检测入库格式错误！,"+fmt.Sprint(err)
		logs.Error(message)
		utils.LarkDingOneInner("fanjuan.xqp",message)
		return err
	}
	if !updateAarTaskStatus(mapInfo["taskId"]) {
		utils.LarkDingOneInner("fanjuan.xqp","更新任务状态失败，任务ID："+fmt.Sprint(mapInfo["taskId"]))
		return fmt.Errorf("更新任务状态失败，任务ID："+fmt.Sprint(mapInfo["taskId"]))
	}
	return nil
}

/**
	确认信息下标是否越界
 */
func judgeOutOfRange(c *gin.Context,id int,length int) bool {
	if id>length {
		logs.Error("确认信息ID越界"+fmt.Sprint(id))
		errorReturn(c,"确认信息ID越界")
		return false
	}
	return true
}

/**
	更新任务状态
 */
func updateAarTaskStatus(taskId int) bool {
	details := dal.QueryOtherDetectDetail(map[string]interface{}{
		"task_id":taskId,
	})
	var updateFlag = true
	P:
	for _,detail := range (*details) {
		if detail.DetailType != 4 {
			if detail.DetailType == 0 {
				var infos []dal.SMethod
				json.Unmarshal([]byte(detail.DetectInfos),&infos)
				for _,info := range infos {
					if info.Status == 0{
						updateFlag = false
						break P
					}
				}
			}else if detail.DetailType == 1 {
				var infos []dal.SStr
				json.Unmarshal([]byte(detail.DetectInfos),&infos)
				for _,info := range infos {
					if info.Status == 0{
						updateFlag = false
						break P
					}
				}
			}else if detail.DetailType == 2 {
				var infos []dal.Permissions
				json.Unmarshal([]byte(detail.DetectInfos),&infos)
				for _,info := range infos {
					if info.Status == 0{
						updateFlag = false
						break P
					}
				}
			}
		}
	}
	if updateFlag {
		condition := "id = '"+fmt.Sprint(taskId)+"'"
		if err := dal.UpdateOtherDetectModelByMap(condition, map[string]interface{}{
			"status":1,
		});err != nil {
			logs.Error("aar任务确认状态更新失败，task ID："+fmt.Sprint(taskId)+",原因：%v",err)
			return false
		}
	}
	return true
}

/**
	获取检测详情结果
 */
func getOtherDetectDetail (c *gin.Context, infos *[]dal.OtherDetailInfoStruct,task *dal.OtherDetectModel) *[]dal.DetectQueryStruct{
	detailMap := make(map[int][]dal.OtherDetailInfoStruct)
	finalResult := make([]dal.DetectQueryStruct,0)
	var midResult []dal.DetectQueryStruct
	var firstResult dal.DetectQueryStruct//主要包检测结果
	for _,info := range  (*infos) {
		if info.DetailType == 4 {
			var t dal.OtherBasicInfo
			if err := json.Unmarshal([]byte(info.DetectInfos),&t);err != nil {
				logs.Error("arr基础信息存储格式错误，taskID："+fmt.Sprint(task.ID))
				errorReturn(c,"arr基础信息存储格式错误，taskID："+fmt.Sprint(task.ID))
				return nil
			}
			if t.Name == task.OtherName {
				firstResult.ApkName = t.Name
				firstResult.Version = t.Version
				firstResult.Channel = t.Channel
				firstResult.Index = info.SubIndex+1
			}else{
				var mid dal.DetectQueryStruct
				mid.ApkName = t.Name
				mid.Version = t.Version
				mid.Channel = t.Channel
				mid.Index = info.SubIndex+1
				midResult = append(midResult,mid)
			}
		}else {
			if detailMap[info.SubIndex+1] == nil {
				infoList := make([]dal.OtherDetailInfoStruct,0)
				infoList = append(infoList,info)
				detailMap[info.SubIndex+1] = infoList
			}else {
				detailMap[info.SubIndex+1] = append(detailMap[info.SubIndex+1],info)
			}
		}
	}
	finalResult = append(finalResult,firstResult)
	finalResult = append(finalResult,midResult...)

	for i := 0;i<len(finalResult);i++ {
		detailArray := detailMap[finalResult[i].Index]
		for _,detailOne := range detailArray {
			if detailOne.DetailType == 0 {
				var t []dal.SMethod
				if err := json.Unmarshal([]byte(detailOne.DetectInfos),&t);err != nil {
					logs.Error("arr敏感方法信息存储格式错误，taskID："+fmt.Sprint(task.ID))
					errorReturn(c,"arr敏感方法存储格式错误，taskID："+fmt.Sprint(task.ID))
					return nil
				}
				var result []dal.SMethod
				var result_con []dal.SMethod
				for _,method := range t {
					if method.Status == 0 {
						result = append(result,method)
					}else {
						result_con = append(result_con,method)
					}
				}
				result = append(result,result_con...)
				finalResult[i].SMethods = result
			}else if detailOne.DetailType == 1 {
				var t []dal.SStr
				if err := json.Unmarshal([]byte(detailOne.DetectInfos),&t);err != nil {
					logs.Error("arr敏感字符串信息存储格式错误，taskID："+fmt.Sprint(task.ID))
					errorReturn(c,"arr敏感字符串存储格式错误，taskID："+fmt.Sprint(task.ID))
					return nil
				}
				var result []dal.SStr
				var result_con []dal.SStr
				for _,str := range t {
					if str.Status == 0 {
						 result= append(result,str)
					}else {
						result_con = append(result_con,str)
					}
				}
				result = append(result,result_con...)
				finalResult[i].SStrs_new = result
				finalResult[i].SStrs = make([]dal.SStr,0)
			}else if detailOne.DetailType == 2 {
				var t []dal.Permissions
				if err := json.Unmarshal([]byte(detailOne.DetectInfos),&t);err != nil {
					logs.Error("arr权限信息存储格式错误，taskID："+fmt.Sprint(task.ID))
					errorReturn(c,"arr权限信息存储格式错误，taskID："+fmt.Sprint(task.ID))
					return nil
				}
				//权限排序
				var result PermSlice
				var result_con PermSlice
				for _,perm := range t {
					if perm.Status == 0 {
						result = append(result,perm)
					}else {
						result_con = append(result_con,perm)
					}
				}
				sort.Sort(PermSlice(result))
				sort.Sort(PermSlice(result_con))
				result = append(result,result_con...)
				finalResult[i].Permissions_2 = result
			}
		}
	}
	return &finalResult
}

/**
	解析检测基础信息
 */
func otherBasicInfoAna(info dal.AppInfoStruct,mapInfo map[string]int,index int) (dal.OtherDetailInfoStruct,error){
	var detailInfo dal.OtherDetailInfoStruct
	if info.Primary == nil || info.Primary.(float64) == 1 {
		condition := "id = '"+fmt.Sprint(mapInfo["taskId"])+"'"
		errU := dal.UpdateOtherDetectModelByMap(condition, map[string]interface{}{
			"other_name":info.ApkName,
			"other_version":info.ApkVersionName,
		})
		if errU != nil {
			message := "更新aar检测任务信息失败，aar检测taskID:"+fmt.Sprint(mapInfo["taskId"])
			utils.LarkDingOneInner("fanjuan.xqp",message)
			return detailInfo,errU
		}
	}
	detailInfo.TaskId = mapInfo["taskId"]
	detailInfo.ToolId = mapInfo["toolId"]
	detailInfo.SubIndex = index
	detailInfo.DetailType = 4

	var t dal.OtherBasicInfo
	t.Name = info.ApkName
	t.Version = info.ApkVersionName
	t.Channel = info.Channel
	param,_ := json.Marshal(t)
	detailInfo.DetectInfos = string(param)
	return detailInfo, nil
}

/**
	解析敏感方法
 */
func otherMethodAna(info *[]dal.MethodInfo,mapInfo map[string]int, index int) dal.OtherDetailInfoStruct{
	var detailInfo dal.OtherDetailInfoStruct
	detailInfo.TaskId = mapInfo["taskId"]
	detailInfo.ToolId = mapInfo["toolId"]
	detailInfo.SubIndex = index
	detailInfo.DetailType = 0
	methods := make([]dal.SMethod,0)
	for i,method := range (*info) {
		var t dal.SMethod
		t.Id = uint(i+1)
		t.Status = 0
		t.MethodName = method.MethodName
		t.ClassName = method.ClassName
		t.Desc = method.Desc
		t.GPFlag = method.Flag
		t.CallLoc = *methodRmRepeat_other(&method.CallLocation)
		methods = append(methods,t)
	}
	param,_ := json.Marshal(methods)
	detailInfo.DetectInfos = string(param)
	return detailInfo
}

func methodRmRepeat_other(callInfo *[]dal.CallLocInfo) *[]dal.MethodCallJson  {
	repeatMap := make(map[string]int)
	result := make([]dal.MethodCallJson,0)
	for _,info := range (*callInfo) {
		var keystr string
		keystr = info.ClassName+info.CallMethodName+fmt.Sprint(info.LineNum)
		if v,ok := repeatMap[keystr]; (!ok||(ok&&v==0)) {
			repeatMap[keystr] = 1
			var t dal.MethodCallJson
			t.ClassName = info.ClassName
			t.MethodName = info.CallMethodName
			t.LineNumber = info.LineNum
			result = append(result,t)
		}
	}
	return &result
}

/**
	解析敏感字符串
*/
func otherStrAna(info *[]dal.StrInfo,mapInfo map[string]int, index int) dal.OtherDetailInfoStruct{
	var detailInfo dal.OtherDetailInfoStruct
	detailInfo.TaskId = mapInfo["taskId"]
	detailInfo.ToolId = mapInfo["toolId"]
	detailInfo.SubIndex = index
	detailInfo.DetailType = 1
	methods := make([]dal.SStr,0)
	for i,str := range (*info) {
		var t dal.SStr
		t.Id = uint(i+1)
		t.Status = 0
		var keys = ""
		t.ConfirmInfos = make([]dal.ConfirmInfo,0)
		for _,keyInfo := range str.Keys {
			keys += keyInfo +";"
			var con dal.ConfirmInfo
			con.Key = keyInfo
			con.Status = 0
			t.ConfirmInfos = append(t.ConfirmInfos,con)
		}
		t.Keys = keys
		t.Desc = str.Desc
		t.GPFlag = str.Flag
		t.CallLoc = *strRmRepeat_other(&str.CallLocation)
		methods = append(methods,t)
	}
	param,_ := json.Marshal(methods)
	detailInfo.DetectInfos = string(param)
	return detailInfo
}

func strRmRepeat_other(callInfo *[]dal.CallLocInfo) *[]dal.StrCallJson {
	repeatMap := make(map[string]int)
	result := make([]dal.StrCallJson,0)
	for _, info := range (*callInfo) {
		var keystr string
		keystr = info.ClassName + info.CallMethodName + info.Key + fmt.Sprint(info.LineNum)
		if v, ok := repeatMap[keystr]; (!ok || (ok && v == 0)) {
			repeatMap[keystr] = 1
			var t dal.StrCallJson
			t.LineNumber = info.LineNum
			t.Key = info.Key
			t.MethodName = info.CallMethodName
			t.ClassName = info.ClassName
			result = append(result,t)
		}
	}
	return &result
}

/**
	解析权限信息
 */
func otherPermAna(perms *[]string,mapInfo map[string]int,index int) (dal.OtherDetailInfoStruct,error)  {
	var detailInfo dal.OtherDetailInfoStruct
	detailInfo.TaskId = mapInfo["taskId"]
	detailInfo.ToolId = mapInfo["toolId"]
	detailInfo.SubIndex = index
	detailInfo.DetailType = 2

	larkPerms := ""//lark消息通知的权限内容

	//权限去重map
	permRepeatMap := make(map[string]int)
	newPermList := make([]string,0)
	for _,perss := range (*perms) {
		//权限去重
		if v, okp := permRepeatMap[perss]; okp && v == 1 {
			continue
		}
		permRepeatMap[perss] = 1
		newPermList = append(newPermList, perss)
	}
	permList := make([]dal.Permissions,0)
	for i,pers := range newPermList {
		var t dal.Permissions
		t.Id = uint(i+1)
		t.Status = 0
		t.Key = pers

		queryResult := dal.QueryDetectConfig(map[string]interface{}{
			"key_info":pers,
			"platform":0,
			"check_type":0,
		})
		if queryResult == nil || len(*queryResult)==0{
			var conf dal.DetectConfigStruct
			conf.KeyInfo = pers
			//将该权限的优先级定为--3高危
			conf.Priority = 3
			//暂时定为固定人选
			conf.Creator = "kanghuaisong"
			conf.Platform = 0
			perm_id,err := dal.InsertDetectConfig(conf)

			if err != nil {
				logs.Error("taskId:"+fmt.Sprint(mapInfo["taskId"])+",aar检测update回调时新增权限失败，%v",err)
				//及时报警
				utils.LarkDingOneInner("fanjuan.xqp","taskId:"+fmt.Sprint(mapInfo["taskId"])+",update回调新增权限失败")
				return detailInfo,err
			}else {
				larkPerms += "权限名为："+pers+"\n"
				t.PermId = int(perm_id)
				t.Priority = 3
			}
		}else{
			t.PermId = int((*queryResult)[0].ID)
			t.Desc = (*queryResult)[0].Ability
			t.Priority = (*queryResult)[0].Priority
		}
		permList = append(permList,t)
	}
	param,_ := json.Marshal(permList)
	detailInfo.DetectInfos = string(param)
	//lark通知创建人完善权限信息-----只发一条消息
	if larkPerms != ""{
		message := "你好，aar包检测出未知权限，请去权限配置页面完善权限信息,需要完善的权限信息有：\n"
		message += larkPerms
		message += "修改链接：http://cloud.bytedance.net/rocket/itc/permission?biz=13"
		utils.LarkDingOneInner("kanghuaisong",message)
		//测试时使用
		//utils.LarkDingOneInner("fanjuan.xqp",message)
		//上线时使用
		//utils.LarkDingOneInner("lirensheng",message)
	}
	return detailInfo,nil

}

