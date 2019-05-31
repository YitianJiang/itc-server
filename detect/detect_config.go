package detect

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
	"strconv"
)

func AddDetectConfig(c *gin.Context)  {

	param,_ := ioutil.ReadAll(c.Request.Body)
	var t dal.DetectConfigInfo
	err := json.Unmarshal(param,&t)
	if err != nil {
		logs.Error("add detectConfig 传入参数不合法！%v",err)
		c.JSON(http.StatusOK,gin.H{
			"message":"传入参数不合法！",
			"errorCode":-1,
		})
		return
	}
	var data dal.DetectConfigStruct
	data.KeyInfo = t.KeyInfo
	data.DescInfo = t.DescInfo
	data.Ability = t.Ability
	data.AbilityGroup = t.AbilityGroup
	data.DetectType = t.DetectType
	data.Priority = t.Priority
	data.Reference = t.Reference
	//data.Type = t.Type
	data.Suggestion = t.Suggestion
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	if err := dal.InsertDetectConfig(data); err != nil {
		c.JSON(http.StatusOK,gin.H{
			"message":"新增权限设置失败！",
			"errorCode":-1,
		})
		return
	}
	logs.Info("add detectConfig success!")
	c.JSON(http.StatusOK,gin.H{
		"message":"success",
		"errorCode":0,
	})
	return
}

func QueryDectecConfig(c *gin.Context)  {
	param,_ := ioutil.ReadAll(c.Request.Body)
	var t map[string]interface{}
	err := json.Unmarshal(param,&t)
	if err != nil {
		logs.Error("check detectConfig 传入参数不合法！%v",err)
		c.JSON(http.StatusOK,gin.H{
			"message":"传入参数不合法！",
			"errorCode":-1,
			"data":err,
		})
		return
	}

	result := dal.QueryDetectConfig(t)
	if result == nil || len(*result)== 0{
		logs.Error("未查询到相关权限信息")
		c.JSON(http.StatusOK,gin.H{
			"message":"未查询到相关权限信息！",
			"errorCode":-1,
			"data":"未查询到相关权限信息！",
		})
		return
	}
	logs.Info("query detectConfig success!")
	c.JSON(http.StatusOK,gin.H{
		"message":"success",
		"errorCode":0,
		"data":(*result),
	})
	return
}

func EditDectecConfig(c *gin.Context){
	param,_ := ioutil.ReadAll(c.Request.Body)
	var t dal.DetectConfigInfo
	err := json.Unmarshal(param,&t)
	if err != nil {
		logs.Error("edit detectConfig 传入参数不合法！%v",err)
		c.JSON(http.StatusOK,gin.H{
			"message":"传入参数不合法！",
			"errorCode":-1,
		})
		return
	}
	id := t.ID
	if id == 0 {
		logs.Error("缺少ID参数！")
		c.JSON(http.StatusOK,gin.H{
			"message":"缺少ID参数！",
			"errorCode":-1,
		})
		return
	}
	data := make(map[string]interface{})
	flag := false
	priority := t.Priority
	if priority != 0 {
		flag = true
		data["priority"] = priority
	}
	ability := t.Ability
	if ability != "" {
		flag = true
		data["ability"] = ability
	}
	abilityG := t.AbilityGroup
	if abilityG != "" {
		flag = true
		data["ability_group"] = abilityG
	}
	desc := t.DescInfo
	if ability != "" {
		flag = true
		data["desc_info"] = desc
	}
	deType := t.DetectType
	if ability != "" {
		flag = true
		data["detect_type"] = deType
	}
	suggestion := t.Suggestion
	if ability != "" {
		flag = true
		data["suggestion"] = suggestion
	}
	refer := t.Reference
	if ability != "" {
		flag = true
		data["reference"] = refer
	}
	if !flag {
		logs.Error("无修改参数！")
		c.JSON(http.StatusOK,gin.H{
			"message":"无修改参数！",
			"errorCode":-1,
		})
		return
	}
	data["updated_at"] = time.Now()
	updateData := make(map[string]interface{})
	condition := "id = '"+strconv.Itoa(id)+"'"
	updateData["condition"] = condition
	updateData["update"] = data
	if err := dal.UpdataDetectConfig(updateData); err != nil {
		c.JSON(http.StatusOK,gin.H{
			"message":"数据库修改失败！",
			"errorCode":-1,
		})
		return
	}
	logs.Info("edit detectConfig success！")
	c.JSON(http.StatusOK,gin.H{
		"message":"success",
		"errorCode":0,
	})
	return
}

func DeleteDetectConfig(c *gin.Context)  {
	id,ok := c.GetQuery("id")
	if !ok {
		logs.Error("缺少ID参数！")
		c.JSON(http.StatusOK,gin.H{
			"message":"缺少ID参数！",
			"errorCode":-1,
		})
		return
	}
	condition := "id = '"+id+"'"
	if err := dal.DeleteDetectConfig(condition); err != nil {
		c.JSON(http.StatusOK,gin.H{
			"message":"数据库删除失败，请重试！",
			"errorCode":-1,
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"message":"success",
		"errorCode":0,
	})
	return
}

