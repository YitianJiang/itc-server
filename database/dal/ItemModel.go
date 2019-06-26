package dal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type ItemStruct struct {
	gorm.Model
	KeyWord       int    `gorm:"column:key_word"        json:"keyWord"`
	FixWay        int    `gorm:"column:fix_way"         json:"fixWay"`
	CheckContent  string `gorm:"column:check_content"   json:"checkContent"`
	Resolution    string `gorm:"column:resolution"      json:"resolution"`
	Regulation    string `gorm:"column:regulation"      json:"regulation"`
	RegulationUrl string `gorm:"column:regulation_url"   json:"regulationUrl"`
	IsGG          int    `gorm:"column:is_gg"           json:"isGg"`
	AppId         string `gorm:"column:app_id"          json:"appId"` //支持多个appId
	Platform      int    `gorm:"column:platform"        json:"platform"`
	Status        int    `gorm:"column:status"          json:"status"`       //没什么用
	QuestionType  int    `gorm:"column:question_type"   json:"questionType"` //数据库使用
}
type MutilitemStruct struct {
	ID              int    `json:"Id"`
	KeyWord         int    `json:"keyWord"`
	FixWay          int    `json:"fixWay"`
	CheckContent    string `json:"checkContent"`
	Resolution      string `json:"resolution"`
	Regulation      string `json:"regulation"`
	RegulationUrl   string `json:"regulationUrl"`
	IsGG            int    `json:"isGg"`
	AppId           string `json:"appId"` //支持多个appId
	Platform        int    `json:"platform"`
	Status          int    `json:"status"`       //没什么用
	QuestionTypeArr string `json:"questionType"` //传入参数支持多种问题类型
}
type QueryItemStruct struct {
	ID               uint   `json:ID`
	QuestionType     int    `json:"questionType"`
	KeyWord          int    `json:"keyWord"`
	FixWay           int    `json:"fixWay"`
	CheckContent     string `json:"checkContent"`
	Resolution       string `json:"resolution"`
	Regulation       string `json:"regulation"`
	RegulationUrl    string `json:"regulationUrl"`
	IsGG             int    `json:"isGg"`
	AppId            string `json:"appId"`
	Platform         int    `json:"platform"`
	QuestionTypeName string `json:"questionTypeName"`
	KeyWordName      string `json:"keyWordName"`
	FixWayName       string `json:"fixWayName"`
	Remark           string `json:"remark"`
	Status           int    `json:"status"`
	Confirmer        string `json:"confirmer"`
}
type ConfirmCheck struct {
	gorm.Model
	TaskId   int    `json:"taskId"`
	ItemId   int    `json:"itemId"`
	Status   int    `json:"status"`
	Operator string `json:"operator"`
	Remark   string `json:"remark"`
}
type Self struct {
	Status int    `json:"status"`
	Id     int    `json:"id"`
	Remark string `json:"remark"`
}
type Confirm struct {
	TaskId int    `json:"taskId"`
	Data   []Self `json:"data"`
}

type AppSelfItem struct {
	gorm.Model
	AppId     int    `gorm:"column:appId"`
	Platform  int    `gorm:"column:platform"`
	SelfItems string `gorm:"column:selfItem"`
	AppName   string `gorm:"column:appname"`
}
type TaskSelfItem struct {
	gorm.Model
	TaskId    int    `gorm:"column:taskId"`
	ToolId    int    `gorm:"column:toolId"`
	SelfItems string `gorm:"column:selfItem"`
}

func (ConfirmCheck) TableName() string {
	return "tb_confirm_check"
}
func (ItemStruct) TableName() string {
	return "tb_item"
}
func (AppSelfItem) TableName() string {
	return "tb_app_selfItem"
}
func (TaskSelfItem) TableName() string {
	return "tb_task_selfItem"
}

//检查项管理增加自查项
func InsertItemModel(mutilItem MutilitemStruct) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false
	}
	defer connection.Close()
	db := connection.Begin()
	//获取配置
	configMap := make(map[int]string)
	var configs []ItemConfig
	err = db.Table(ItemConfig{}.TableName()).LogMode(_const.DB_LOG_MODE).Find(&configs).Error
	if err != nil {
		logs.Error("查询检查配置项失败！", err.Error())
		db.Rollback()
		return false
	}
	if configs != nil && len(configs) > 0 {
		for i := 0; i < len(configs); i++ {
			config := (configs)[i]
			configMap[int(config.ID)] = config.Name
		}
	}
	//在此之前添加的公共项
	var ggItem []ItemStruct
	if err := db.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Where("is_gg = ?", 1).Find(&ggItem).Error; err != nil {
		logs.Error("查询公共项失败！", err.Error())
		db.Rollback()
		return false
	}
	//之前的公共项
	var perItemList []interface{}
	if len(ggItem) != 0 {
		for _, gg := range ggItem {
			ggJson, _ := json.Marshal(gg)
			ggMap := make(map[string]interface{})
			json.Unmarshal(ggJson, &ggMap)
			delete(ggMap, "status")
			delete(ggMap, "appId")
			delete(ggMap, "ID")
			delete(ggMap, "CreatedAt")
			delete(ggMap, "DeletedAt")
			delete(ggMap, "UpdatedAt")
			keyWord := ggMap["keyWord"].(float64)
			fixWay := ggMap["fixWay"].(float64)
			questionType := ggMap["questionType"].(float64)
			ggMap["keyWord"] = configMap[int(keyWord)]
			ggMap["fixWay"] = configMap[int(fixWay)]
			ggMap["questionType"] = configMap[int(questionType)]
			ggMap["id"] = gg.ID
			perItemList = append(perItemList, ggMap)
		}
	}
	questionTypeArr := strings.Split(mutilItem.QuestionTypeArr, ",")
	for _, questionType := range questionTypeArr {
		//构造itemStruct插入数据库
		var item ItemStruct
		item.ID = uint(mutilItem.ID)
		item.KeyWord = mutilItem.KeyWord
		item.FixWay = mutilItem.FixWay
		item.CheckContent = mutilItem.CheckContent
		item.Resolution = mutilItem.Resolution
		item.Regulation = mutilItem.Regulation
		item.RegulationUrl = mutilItem.RegulationUrl
		item.IsGG = mutilItem.IsGG
		item.AppId = mutilItem.AppId
		item.Platform = mutilItem.Platform
		item.Status = 0
		question_type, _ := strconv.Atoi(questionType)
		item.QuestionType = question_type
		//tb_item 处理
		id := mutilItem.ID
		var is ItemStruct
		if err := db.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Where("id = ?", id).Find(&is).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := connection.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&item).Error; err != nil {
					logs.Error("insert self check item failed, %v", err)
					db.Rollback()
					return false
				}
			} else {
				logs.Error("query self check item failed, %v", err)
				db.Rollback()
				return false
			}
		} else {
			if err = db.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Save(&item).Error; err != nil {
				logs.Error("update self check item failed, %v", err)
				db.Rollback()
				return false
			}
		}

		// struct -> map -> json
		//itemMap := Struct2Map(item)
		itemJson, err := json.Marshal(item)
		if err != nil {
			logs.Error("struct -> json error!", err)
			db.Rollback()
			return false
		}
		itemMap := make(map[string]interface{})
		if err := json.Unmarshal(itemJson, &itemMap); err != nil {
			logs.Error("json -> map error!", err)
			db.Rollback()
			return false
		}
		delete(itemMap, "status")
		delete(itemMap, "appId")
		delete(itemMap, "ID")
		delete(itemMap, "CreatedAt")
		delete(itemMap, "DeletedAt")
		delete(itemMap, "UpdatedAt")
		keyWord := itemMap["keyWord"].(float64)
		fixWay := itemMap["fixWay"].(float64)
		questionType := itemMap["questionType"].(float64)
		itemMap["keyWord"] = configMap[int(keyWord)]
		itemMap["fixWay"] = configMap[int(fixWay)]
		itemMap["questionType"] = configMap[int(questionType)]
		itemMap["id"] = item.ID

		//tb_app_selfItem 处理
		var appIdArr []string
		//如果是非公共项，只给在appidlist中app添加
		if item.IsGG == 0 {
			appIdArr = strings.Split(item.AppId, ",")
		}
		//如果是公共项，给appItem表中所有app添加
		if item.IsGG == 1 {
			var apps []AppSelfItem
			if err := db.Table(AppSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Select("appId").Where("platform = ?", item.Platform).Find(&apps).Error; err != nil {
				logs.Error("查询appitem出错！", err.Error())
				db.Rollback()
				return false
			}
			isExitApp := false
			for _, a := range apps {
				if strconv.Itoa(a.AppId) == item.AppId {
					isExitApp = true
				}
				appIdArr = append(appIdArr, strconv.Itoa(a.AppId))
			}
			//公共项对应的app还没有在appItem中有记录，则先添加该公共项
			if !isExitApp {
				appIdArr = append(appIdArr, item.AppId)
			}
		}
		fmt.Println("215,", appIdArr)
		for _, appId := range appIdArr {
			var appItem AppSelfItem
			app_id, _ := strconv.Atoi(appId)
			if err := db.Table(AppSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Where("appId = ? AND platform = ?", app_id, item.Platform).Limit(1).Find(&appItem).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					fmt.Println("206，appitem新增")
					appItem.AppId = app_id
					appItem.Platform = item.Platform
					perItemList = append(perItemList, itemMap)
					app_item := map[string]interface{}{
						"item": perItemList,
					}
					tem, err := json.Marshal(app_item)
					if err != nil {
						logs.Error("map to json error!", err.Error())
						db.Rollback()
						return false
					}
					appItem.SelfItems = string(tem)
					if err := db.Table(AppSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&appItem).Error; err != nil {
						logs.Error(err.Error())
						db.Rollback()
						return false
					}
				} else {
					logs.Error(err.Error())
					db.Rollback()
					return false
				}
			} else {
				//update
				tem_item := appItem.SelfItems
				m := make(map[string]interface{})
				err = json.Unmarshal([]byte(tem_item), &m)
				if err != nil {
					logs.Error("Umarshal failed:", err.Error())
					db.Rollback()
					return false
				}
				list := m["item"].([]interface{})
				isExit := false
				for _, l := range list {
					if int(l.(map[string]interface{})["id"].(float64)) == int(item.ID) {
						isExit = true
						l = itemMap
					}
				}
				if !isExit {
					list = append(list, itemMap)
				}
				m["item"] = list
				tem, err := json.Marshal(m)
				if err != nil {
					logs.Error("map to json error!", err.Error())
					db.Rollback()
					return false
				}
				if err := db.Table(AppSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Model(&appItem).Update("selfItem", string(tem)).Error; err != nil {
					logs.Error(err.Error())
					db.Rollback()
					return false
				}
			}
		}

	}
	db.Commit()
	return true
}

//query data
func QueryItemsByCondition(data map[string]interface{}) *[]QueryItemStruct {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer connection.Close()
	db := connection.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	condition := data["condition"]
	logs.Info("query items condition: %s", condition)
	if condition != "" {
		db = db.Where(condition)
	}
	var items []ItemStruct
	if err := db.Order("key_word ASC").Find(&items).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	var configMap map[uint]string
	configMap = make(map[uint]string)
	configs := QueryConfigByCondition("1=1")
	if configs != nil && len(*configs) > 0 {
		for i := 0; i < len(*configs); i++ {
			config := (*configs)[i]
			configMap[config.ID] = config.Name
		}
	}
	var qis []QueryItemStruct
	if items != nil && len(items) > 0 {
		for j := 0; j < len(items); j++ {
			item := items[j]
			var qisItem QueryItemStruct
			qisItem.ID = item.ID
			qisItem.Platform = item.Platform
			qisItem.AppId = item.AppId
			qisItem.IsGG = item.IsGG
			qisItem.RegulationUrl = item.RegulationUrl
			qisItem.Regulation = item.Regulation
			qisItem.Resolution = item.Resolution
			qisItem.FixWay = item.FixWay
			qisItem.CheckContent = item.CheckContent
			qisItem.KeyWord = item.KeyWord
			qisItem.QuestionType = item.QuestionType
			qisItem.FixWayName = configMap[uint(item.FixWay)]
			qisItem.QuestionTypeName = configMap[uint(item.QuestionType)]
			qisItem.KeyWordName = configMap[uint(item.KeyWord)]
			qis = append(qis, qisItem)
		}
	}
	return &qis
}

//检查项管理删除自查项
func DeleteItemsByCondition(condition map[string]interface{}) bool {
	itemId, _ := strconv.Atoi(condition["id"].(string))
	isAll, _ := strconv.Atoi(condition["isGG"].(string))
	appId := condition["appId"] //string
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false
	}
	defer connection.Close()
	db := connection.Begin()
	var item ItemStruct
	if err := db.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Where("id = ?", itemId).Limit(1).Find(&item).Error; err != nil {
		logs.Error("查询tb_item出错！", err.Error())
		db.Rollback()
		return false
	}
	platform := item.Platform
	var item_appId []int
	if isAll == 1 {
		var appItem []AppSelfItem
		if err := db.Table(AppSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Select("appId").Where("platform = ?", platform).Find(&appItem).Error; err != nil {
			logs.Error("查询tb_item出错！", err.Error())
			db.Rollback()
			return false
		}
		for _, app := range appItem {
			item_appId = append(item_appId, app.AppId)
		}
	}
	if isAll == 0 {
		app, _ := strconv.Atoi(appId.(string))
		item_appId = append(item_appId, app)
	}
	//更新tb_app_selfItem，删除包含该itemId的app中的自查项
	for _, app_id := range item_appId {
		var appSelf AppSelfItem
		if err := db.Table(AppSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Where("appId = ? AND platform = ?", app_id, platform).Limit(1).Find(&appSelf).Error; err != nil {
			logs.Error("查询tb_app_selfItem出错！", err.Error())
			db.Rollback()
			return false
		}
		itemSelf := appSelf.SelfItems
		m := make(map[string]interface{})
		if err = json.Unmarshal([]byte(itemSelf), &m); err != nil {
			logs.Error("json -> map error!", err.Error())
			db.Rollback()
			return false
		}
		items := m["item"].([]interface{})
		var new_items []interface{}
		for _, im := range items {
			if int(im.(map[string]interface{})["id"].(float64)) != itemId {
				new_items = append(new_items, im)
			}
		}
		m["item"] = new_items
		item_self, err := json.Marshal(m)
		if err != nil {
			logs.Error("map -> json error!", err.Error())
			db.Rollback()
			return false
		}
		if err := db.Table(AppSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Model(&appSelf).Update("selfItem", string(item_self)).Error; err != nil {
			logs.Error("更新tb_app_selfItem出错！", err.Error())
			db.Rollback()
			return false
		}
	}
	//删除tb_item中该自查项
	if isAll == 1 {
		if err := db.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Where("id = ?", itemId).Delete(ItemStruct{}).Error; err != nil {
			logs.Error("delete failed: %v", err)
			db.Rollback()
			return false
		}
	}
	if isAll == 0 {
		var appIds []string
		appTemp := strings.Split(item.AppId, ",")
		for _, app_id := range appTemp {
			if app_id != appId {
				appIds = append(appIds, app_id)
			}
		}
		if len(appIds) == 0 {
			if err := db.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Where("id = ?", itemId).Delete(ItemStruct{}).Error; err != nil {
				logs.Error("delete failed: %v", err)
				db.Rollback()
				return false
			}
		} else {
			if err := db.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Model(&item).Update("app_id", strings.Join(appIds, ",")).Error; err != nil {
				logs.Error("delete failed: %v", err)
				db.Rollback()
				return false
			}
		}
	}
	db.Commit()
	return true
}

//taskID确认自查项
func ConfirmSelfCheck(param map[string]interface{}) bool {
	//获取前端数据
	operator := param["operator"]
	taskId := param["taskId"]
	data := param["data"]
	//连接数据库
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false
	}
	defer connection.Close()
	db := connection.Begin()
	//获取自查记录
	allCheckFlag := true //自查项是否全部确认标志
	var taskSelf TaskSelfItem
	if err := db.Table(TaskSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Where("taskId = ?", taskId).Find(&taskSelf).Error; err != nil {
		logs.Error("查询taskId自查项失败！", err.Error())
		db.Rollback()
		return false
	}
	taskSelfMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(taskSelf.SelfItems), &taskSelfMap); err != nil {
		logs.Error("json map转换失败！", err.Error())
		db.Rollback()
		return false
	}
	taskSelfList := taskSelfMap["item"].([]interface{})
	idArray := data.([]Self)
	for _, self := range taskSelfList {
		id := self.(map[string]interface{})["id"].(int)
		for _, confirmSelf := range idArray {
			if confirmSelf.Id == id {
				self.(map[string]interface{})["status"] = confirmSelf.Status
				self.(map[string]interface{})["remark"] = confirmSelf.Remark
				self.(map[string]interface{})["confirmer"] = operator
			}
		}
		if self.(map[string]interface{})["status"] == 0 {
			allCheckFlag = false
		}
	}
	taskSelfMap["item"] = taskSelfList
	task_self, err := json.Marshal(taskSelfMap)
	if err != nil {
		logs.Error("map json转换失败！", err.Error())
		db.Rollback()
		return false
	}
	taskSelf.SelfItems = string(task_self)
	if err := db.Table(TaskSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Save(taskSelf).Error; err != nil {
		logs.Error(err.Error())
		db.Rollback()
		return false
	}
	//最后更新检测任务的自查状态
	if allCheckFlag {
		var detect DetectStruct
		if err = db.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).
			Where("id=?", taskId).Find(&detect).Error; err != nil {
			logs.Error("%v", err)
			db.Rollback()
			return false
		}
		if &detect == nil {
			logs.Error(err.Error())
			db.Rollback()
			return false
		}
		detect.SelfCheckStatus = 1
		if err = db.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).
			Save(detect).Error; err != nil {
			logs.Error("%v", err)
			db.Rollback()
			return false
		}
		if detect.Status == 1 {
			//回调接口
		}
	}
	db.Commit()
	return true
}

//根据任务id拿到对应的自查信息
func GetSelfCheckByTaskId(condition string) (map[uint]int, map[uint]string, map[uint]string) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, nil, nil
	}
	defer connection.Close()
	db := connection.Table(ConfirmCheck{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var items []ConfirmCheck
	if err = db.Where(condition).Find(&items).Error; err != nil {
		logs.Error("query self check item failed, %v", err)
		return nil, nil, nil
	}
	if len(items) == 0 {
		logs.Info("query self check item empty")
		return nil, nil, nil
	}
	var item map[uint]int
	item = make(map[uint]int)
	var remark = make(map[uint]string)
	var confirmer = make(map[uint]string)
	for i := 0; i < len(items); i++ {
		it := items[i]
		itemId := it.ItemId
		item[uint(itemId)] = it.Status
		remark[uint(itemId)] = it.Remark
		confirmer[uint(itemId)] = it.Operator

	}
	return item, remark, confirmer
}
func QueryItem(condition map[string]interface{}) *[]ItemStruct {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer connection.Close()
	var items []ItemStruct
	db := connection.Table(AppSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err = db.Where(condition).Find(&items).Error; err != nil {
		logs.Error("query self check item failed, %v", err)
		return nil
	}
	return &items
}
func InsertAppSelfItem(appItem AppSelfItem) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false
	}
	defer connection.Close()
	db := connection.Table(AppSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err = db.Create(appItem).Error; err != nil {
		logs.Error("query self check item failed, %v", err)
		return false
	}
	return true
}

//查询app对应自查项
func QueryAppSelfItem(condition map[string]interface{}) *[]AppSelfItem {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer connection.Close()
	db := connection.Table(AppSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var appSelf []AppSelfItem
	if err = db.Where(condition).Order("platform", true).Find(&appSelf).Error; err != nil {
		logs.Error("query self check item failed, %v", err)
		return nil
	}
	return &appSelf

}

//插入taskId自查项
func InsertTaskSelfItem(taskItem TaskSelfItem) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false
	}
	defer connection.Close()
	if err := connection.Table(TaskSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(taskItem).Error; err != nil {
		logs.Error(err.Error())
		return false
	}
	return true
}

//查询taskId自查项
func QueryTaskSelfItem(taskId int) (bool, []interface{}) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false, nil
	}
	defer connection.Close()
	var taskItem TaskSelfItem
	if err := connection.Table(TaskSelfItem{}.TableName()).LogMode(_const.DB_LOG_MODE).Where("taskId = ?", taskId).Limit(1).Find(&taskItem).Error; err != nil {
		logs.Error(err.Error())
		return false, nil
	}
	if &taskItem == nil {
		return true, nil
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(taskItem.SelfItems), &m); err != nil {
		logs.Error(err.Error())
		return false, nil
	}
	return true, m["item"].([]interface{})
}
