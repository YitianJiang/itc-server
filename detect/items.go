package detect

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
	"gopkg.in/cas.v2"
	"net/http"
	"regexp"
	"strconv"
)

//增加检查项
func AddDetectItem(c *gin.Context){

	questionType := c.GetInt("questionType")
	keyWord := c.GetInt("keyWord")
	fixWay := c.GetInt("fixWay")
	checkContent := c.DefaultQuery("checkContent", "")
	resolution := c.DefaultQuery("resolution", "")
	regulation := c.DefaultQuery("regulation", "")
	regulationUrl := c.DefaultQuery("regulationUrl", "")
	ggFlag := c.GetInt("isGg")
	platform := c.GetInt("platform")
	appId := c.DefaultQuery("appId", "")
	var itemModel dal.ItemStruct
	//platform
	if platform != 0 && platform != 1 {
		logs.Error("platform参数不合法！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "platform参数不合法！",
			"errorCode" : -1,
			"data" : "platform参数不合法！",
		})
		return
	}
	//是否公共,0-否；1-是
	if ggFlag != 0 && ggFlag != 1 {
		ggFlag = 1
	}
	if ggFlag == 0 && appId == "" {
		logs.Error("缺失参数appId！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺失参数appId！",
			"errorCode" : -2,
			"data" : "缺失参数appId！",
		})
		return
	}
	//校验
	if f, _ := regexp.MatchString("^http(s?)://*", regulationUrl); !f {
		logs.Error("条例链接格式不正确！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "条例链接格式不正确！",
			"errorCode" : -3,
			"data" : "条例链接格式不正确！",
		})
		return
	}
	itemModel.QuestionType = questionType
	itemModel.KeyWord = keyWord
	itemModel.FixWay = fixWay
	itemModel.CheckContent = checkContent
	itemModel.Resolution = resolution
	itemModel.Regulation = regulation
	itemModel.RegulationUrl = regulationUrl
	itemModel.Platform = platform
	itemModel.IsGG = ggFlag
	if appId != "" {
		itemModel.AppId, _ = strconv.Atoi(appId)
	}
	itemModel.Status = 0
	itemModelId := dal.InsertItemModel(itemModel)
	if itemModelId == 0 {
		logs.Error("新增检查项失败")
		c.JSON(http.StatusOK, gin.H{
			"message" : "新增检查项失败，请联系相关人员！",
			"errorCode" : -4,
			"data" : "新增检查项失败，请联系相关人员！",
		})
		return
	} else {
		logs.Error("新增检查项成功")
		c.JSON(http.StatusOK, gin.H{
			"message" : "success",
			"errorCode" : 0,
			"data" : "新增检查项成功！",
		})
		return
	}
}
//查询检查项
func GetSelfCheckItems(c *gin.Context){

	appIdParam, ok := c.GetQuery("appId")
	if !ok {
		logs.Error("缺少appId参数！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少appId参数！",
			"errorCode" : -1,
			"data" : "缺少appId参数！",
		})
		return
	}
	taskId, bool := c.GetQuery("taskId")
	condition := "1=1"
	if bool {
		condition += " id='" + taskId + "'"
	}
	var param map[string]interface{}
	var data map[string]interface{}
	itemCondition := ""
	if bool {
		tasks, _ := dal.QueryTasksByCondition(param)
		if len(*tasks) > 0 {
			appId := (*tasks)[0].AppId
			platform := (*tasks)[0].Platform
			itemCondition = "(platform='" + strconv.Itoa(platform) + "')" + "and ((is_gg='1') or (is_gg='0' and app_id='" + appId + "'))"
		}
	} else {
		itemCondition = "((is_gg='1') or (is_gg='0' and app_id='" + appIdParam + "'))"
	}
	data["condition"] = itemCondition
	items := dal.QueryItemsByCondition(data)
	if items==nil || len(*items)==0 {
		logs.Info("未查询到自查项信息！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "未查询到自查项信息！",
			"errorCode" : -3,
			"data" : "未查询到自查项信息！",
		})
		return
	}
	if !bool {
		c.JSON(http.StatusOK, gin.H{
			"message" : "success",
			"errorCode" : 0,
			"data" : *items,
		})
		return
	}
	tj := "task_id='" + taskId + "'"
	itemMap := dal.GetSelfCheckByTaskId(tj)
	if itemMap == nil {
		c.JSON(http.StatusOK, gin.H{
			"message" : "success",
			"errorCode" : 0,
			"data" : *items,
		})
	} else {
		for i := 0; i < len(*items); i++ {
			item := (*items)[i]
			status := itemMap[item.ID]
			item.Status = status
		}
		c.JSON(http.StatusOK, gin.H{
			"message" : "success",
			"errorCode" : 0,
			"data" : *items,
		})
	}
}
//完成自查
func ConfirmCheck(c *gin.Context){
	taskId, bool := c.GetQuery("taskId")
	if !bool {
		logs.Error("缺少taskId参数")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少taskId参数！",
			"errorCode" : -1,
			"data" : "缺少taskId参数！",
		})
		return
	}
	name := cas.Username(c.Request)
	if name == ""{
		c.JSON(http.StatusOK, gin.H{
			"message" : "用户未登录！",
			"errorCode" : -2,
			"data" : "用户未登录！",
		})
		return
	}
	data := c.PostFormArray("data")
	var param map[string]interface{}
	param["taskId"] = taskId
	param["data"] = data
	param["operator"] = name
	bool = dal.ConfirmSelfCheck(param)
	if !bool {
		c.JSON(http.StatusOK, gin.H{
			"message" : "自查确认失败，请联系相关人员！",
			"errorCode" : -3,
			"data" : "自查确认失败，请联系相关人员！",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "success",
	})
}
//更新自查项