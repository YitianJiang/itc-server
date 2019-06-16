package detect

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"time"
)




//权限配置页面操作人员
var permToModify = map[string]int {
	"kanghuaisong":1,
	"zhangshuai.02":1,
	"lirensheng":1,
	//测试加入
	"fanjuan.xqp":1,
}


/**
	新增权限
 */
func AddDetectConfig(c *gin.Context)  {
	username,_ := c.Get("username")

	//权限配置页面操作人员判断
	if v,ok := permToModify[username.(string)]; !ok || v!=1 {
		logs.Error("该用户不允许对权限配置进行操作！")
		c.JSON(http.StatusOK,gin.H{
			"message":"该用户不允许对权限配置进行操作！",
			"errorCode":-1,
		})
		return
	}

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
	if t.KeyInfo == "" || t.Priority == nil || t.Platform == nil || t.DescInfo == "" || t.Type == nil{
		logs.Error("缺少关键参数！")
		c.JSON(http.StatusOK,gin.H{
			"message":"缺少关键参数！",
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
	data.Priority = int(t.Priority.(float64))
	data.Reference = t.Reference
	data.CheckType = int(t.Type.(float64))
	data.Suggestion = t.Suggestion
	data.Platform = int(t.Platform.(float64))

	data.Creator = username.(string)

	queryResult := dal.QueryDetectConfig(map[string]interface{}{
		"key_info":data.KeyInfo,
		"platform":data.Platform,
		"check_type":data.CheckType,
	})
	if queryResult != nil && len(*queryResult) != 0{
		logs.Error("平台已存在该权限！")
		c.JSON(http.StatusOK,gin.H{
			"message":"新增权限设置失败，对应平台已存在该权限！！",
			"errorCode":-1,
		})
		return
	}
	if _,err := dal.InsertDetectConfig(data); err != nil {
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


/**
	查询权限列表
 */
func QueryDectecConfig(c *gin.Context)  {

	type queryStruct struct {
		PageSize 		int			`json:"pageSize"`
		Page 			int			`json:"page"`
		Info    		string		`json:"info"`
		Platform		interface{}	`json:"platform"`
		Type 			interface{} `json:"type"`
	}
	var t queryStruct
	param,_ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(param,&t)
	if err != nil {
		logs.Error("query detectConfig 传入参数不合法！%v",err)
		c.JSON(http.StatusOK,gin.H{
			"message":"传入参数不合法！",
			"errorCode":-1,
			"data":"传入参数不合法！",
		})
		return
	}
	if t.Page <= 0 || t.PageSize <= 0 {
		logs.Error("分页参数不合法!")
		c.JSON(http.StatusOK,gin.H{
			"message":"分页参数不合法!",
			"errorCode":-1,
			"data":"分页参数不合法!",
		})
		return
	}
	pageInfo := make(map[string]int)
	pageInfo["pageSize"] = t.PageSize
	pageInfo["page"]= t.Page

	condition := "1=1"
	if t.Info != ""{
		condition += " and (ability like '%"+t.Info+"%' or key_info like '%"+t.Info+"%')"
	}

	if t.Platform != nil {
		condition += " and platform ='"+fmt.Sprint(t.Platform)+"'"
	}
	if t.Type != nil {
		condition += " and check_type = '"+fmt.Sprint(t.Type)+"'"
	}

	result,count,errQ := dal.QueryDetectConfigList(condition,pageInfo)
	if errQ != nil {
		c.JSON(http.StatusOK,gin.H{
			"message":"查询权限数据库操作失败！",
			"errorCode":-1,
			"data":errQ,
		})
		return
	}
	var permList = make([]dal.DetectConfigListInfo,0)
	if result == nil || len(*result)== 0{
		logs.Error("未查询到相关权限信息")
		c.JSON(http.StatusOK,gin.H{
			"message":"未查询到相关权限信息！",
			"errorCode":0,
			"data":map[string]interface{}{
				"count":count,
				"permList":permList,
			},
		})
		return
	}
	var realResult = make(map[string]interface{})
	for _,re := range (*result) {
		var perm dal.DetectConfigListInfo
		perm.Creator = re.Creator
		perm.CreatedAt = re.CreatedAt
		perm.Id = re.ID
		perm.Priority = re.Priority
		perm.KeyInfo = re.KeyInfo
		perm.Ability = re.Ability
		perm.DescInfo = re.DescInfo
		perm.Platform = re.Platform
		perm.Type = re.CheckType
		permList = append(permList,perm)
	}
	realResult["permList"] = permList
	realResult["count"] = count
	logs.Info("query detectConfig success!")
	c.JSON(http.StatusOK,gin.H{
		"message":"success",
		"errorCode":0,
		"data":realResult,
	})
	return
}

/**
	修改权限信息
 */
func EditDectecConfig(c *gin.Context){
	username,_ := c.Get("username")

	//权限配置页面操作人员判断
	if v,ok := permToModify[username.(string)]; !ok || v!=1 {
		logs.Error("该用户不允许对权限配置进行操作！")
		c.JSON(http.StatusOK,gin.H{
			"message":"该用户不允许对权限配置进行操作！",
			"errorCode":-1,
		})
		return
	}

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
	if priority != nil {
		flag = true
		data["priority"] = int(priority.(float64))
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
	if desc != "" {
		flag = true
		data["desc_info"] = desc
	}
	deType := t.DetectType
	if deType != "" {
		flag = true
		data["detect_type"] = deType
	}
	suggestion := t.Suggestion
	if suggestion != "" {
		flag = true
		data["suggestion"] = suggestion
	}
	refer := t.Reference
	if refer != "" {
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

/**
	查询权限详情
 */
func GetPermDetails(c *gin.Context)  {
	pId,ok := c.GetQuery("id")
	if !ok {
		logs.Error("缺少查询参数！")
		c.JSON(http.StatusOK,gin.H{
			"errorCode":-1,
			"message":"缺少查询参数！",
			"data":"缺少查询参数！",
		})
		return
	}
	pIdInt,err := strconv.Atoi(pId)
	if err != nil {
		logs.Error("查询参数格式不正确！")
		c.JSON(http.StatusOK,gin.H{
			"errorCode":-1,
			"message":"查询参数格式不正确！",
			"data":err,
		})
		return
	}

	result := dal.QueryDetectConfig(map[string]interface{}{
		"id": pIdInt,
	})
	if result == nil || len(*result) == 0 {
		logs.Error("未查询到该权限信息")
		c.JSON(http.StatusOK,gin.H{
			"errorCode":-1,
			"message":"未查询到该权限信息",
			"data":"未查询到该权限信息",
		})
		return
	}
	logs.Info("查询权限详情成功！")
	c.JSON(http.StatusOK,gin.H{
		"errorCode":0,
		"message":"success",
		"data":(*result)[0],
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




type PermQuerySlice [] map[string]interface{}

func (a PermQuerySlice) Len() int {         // 重写 Len() 方法
	return len(a)
}
func (a PermQuerySlice) Swap(i, j int){     // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a PermQuerySlice) Less(i, j int) bool {    // 重写 Less() 方法， 从大到小排序
	return a[j]["priority"].(int)< a[i]["priority"].(int)
}

/**
	根据app信息查询权限信息
 */
func QueryPermissionsWithApp(c *gin.Context)  {
	type queryInfo struct {
		AppId			int 		`json:"appId"`
		AppVersion		string		`json:"appVersion"`
	}
	var t queryInfo
	param,_ := ioutil.ReadAll(c.Request.Body)

	err := json.Unmarshal(param,&t)
	if err != nil {
		logs.Error("查询参数错误！")
		c.JSON(http.StatusOK,gin.H{
			"message":"查询参数错误！",
			"errorCode":-1,
			"data":"查询参数错误！",
		})
		return
	}
	//查询某一app下的权限信息
	if t.AppVersion == "" {
		sql := "SELECT h.app_version,c.id AS perm_id,c.key_info,c.ability,c.priority FROM `tb_detect_config` c, `tb_perm_history` h WHERE h.status = 0 AND c.id = h.perm_id AND h.app_id = '"+fmt.Sprint(t.AppId)+"'"
		result := dal.QueryDetectConfigWithSql(sql)
		if result == nil || len(*result) == 0{
			logs.Error("没有查询到App的权限信息")
			c.JSON(http.StatusOK,gin.H{
				"errorCode":-1,
				"message":"没有查询到App的权限信息",
				"data":"没有查询到App的权限信息",
			})
			return
		}
		realResult := make([]map[string]interface{},0)
		for _,one := range (*result) {
			info := map[string]interface{}{
				"permId" : one.PermId,
				"key":one.KeyInfo,
				"priority":one.Priority,
				"ability":one.Ability,
				"firstVersion":one.AppVersion,
			}
			realResult = append(realResult,info)
		}
		sort.Sort(PermQuerySlice(realResult))
		logs.Info("query permission with appId success!")
		c.JSON(http.StatusOK,gin.H{
			"message":"success",
			"errorCode": 0,
			"data": realResult,
		})
		return
	}else{
		queryResult, err1 := dal.QueryPermAppRelation(map[string]interface{}{
			"app_id":t.AppId,
			"app_version":t.AppVersion,
		})
		if err1 != nil || queryResult == nil || len(*queryResult) == 0 {
			logs.Error("未查询到相关信息")
			c.JSON(http.StatusOK,gin.H{
				"message":"未查询到相关信息！",
				"errorCode":-1,
				"data":"未查询到相关信息！",
			})
			return
		}
		permInfo := []byte((*queryResult)[len(*queryResult)-1].PermInfos)
		var infos []interface{}
		if err := json.Unmarshal(permInfo,&infos); err != nil {
			logs.Error("该app下权限信息存储格式出错,%v",err)
			c.JSON(http.StatusOK,gin.H{
				"message":"该app下权限信息存储格式出错！",
				"errorCode":-1,
				"data":"该app下权限信息存储格式出错！",
			})
			return
		}
		allPermList := GetPermList()
		result := make([]map[string]interface{},0)
		for _,info := range infos {
			vInfo := info.(map[string]interface{})
			//if vInfo["state"].(float64) == 0 {
			if v,ok := allPermList[int(vInfo["perm_id"].(float64))];ok {
				info := v.(map[string]interface{})
				vInfo["priority"] = info["priority"].(int)
				vInfo["ability"] = info["ability"].(string)
			}
			//}
			subInfo := map[string]interface{}{
				"permId" : vInfo["perm_id"],
				"key":vInfo["key"],
				"priority":vInfo["priority"],
				"ability":vInfo["ability"],
				"firstVersion":vInfo["first_version"],
			}
			result = append(result,subInfo)
		}
		sort.Sort(PermQuerySlice(result))
		logs.Info("query permission with appId and appVersion success!")
		c.JSON(http.StatusOK,gin.H{
			"message":"success",
			"errorCode":0,
			"data":result,
		})
		return
	}
}

/**
	查看权限在app中的确认信息
 */
func GetRelationsWithPermission(c *gin.Context)  {
	info := c.DefaultPostForm("info","")
	if info == "" {
		logs.Error("无查询信息")
		c.JSON(http.StatusOK,gin.H{
			"errorCode":-1,
			"message":"无查询信息",
			"data":"无查询信息",
		})
		return
	}
	//-----------模糊联查
	sql := "SELECT h.app_id,h.app_version,c.id AS perm_id,c.key_info,c.ability,c.priority FROM `tb_detect_config` c, `tb_perm_history` h WHERE h.status = 0 AND c.id = h.perm_id AND c.key_info LIKE '%"+info+"%'"

	result := dal.QueryDetectConfigWithSql(sql)
	if result == nil || len(*result) == 0{
		logs.Error("没有查询到该权限信息")
		c.JSON(http.StatusOK,gin.H{
			"errorCode":-1,
			"message":"没有查询到该权限信息",
			"data":"没有查询到该权限信息",
		})
		return
	}
	logs.Info("根据权限查询成功！")
	c.JSON(http.StatusOK,gin.H{
		"message":"success",
		"errorCode": 0,
		"data": (*result),
	})
	return
 }


