package dal

import (
	"fmt"
	"strconv"
	"time"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type ItemStruct struct {
	gorm.Model
	QuestionType  int    `json:"questionType"`
	KeyWord       int    `json:"keyWord"`
	FixWay        int    `json:"fixWay"`
	CheckContent  string `json:"checkContent"`
	Resolution    string `json:"resolution"`
	Regulation    string `json:"regulation"`
	RegulationUrl string `json:"regulationUrl"`
	IsGG          int    `json:"isGg"`
	AppId         int    `json:"appId"`
	Platform      int    `json:"platform"`
	Status        int    `json:"status"`
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
	AppId            int    `json:"appId"`
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

func (ConfirmCheck) TableName() string {
	return "tb_confirm_check"
}
func (ItemStruct) TableName() string {
	return "tb_item"
}

//insert data
func InsertItemModel(itemModel ItemStruct) uint {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return 0
	}
	defer connection.Close()
	id := itemModel.ID
	condition := "id='" + fmt.Sprint(id) + "'"
	var is ItemStruct
	if err := connection.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Find(&is).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := connection.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&itemModel).Error; err != nil {
				logs.Error("insert self check item failed, %v", err)
				return 0
			}
		} else {
			logs.Error("query self check item failed, %v", err)
			return 0
		}
	}
	if err = connection.Table(ItemStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Save(&itemModel).Error; err != nil {
		logs.Error("update self check item failed, %v", err)
		return 0
	}
	return itemModel.ID
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

//confirm check
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
	condition := " task_id='" + strconv.Itoa(taskId.(int)) + "'"
	dataMap, _, _ := GetSelfCheckByTaskId(condition)

	allCheckFlag := true //自查项是否全部确认标志
	idArray := data.([]Self)
	for i := 0; i < len(idArray); i++ {
		dat := idArray[i]
		var check ConfirmCheck
		check.ItemId = dat.Id
		check.Status = dat.Status
		check.TaskId = taskId.(int)
		check.Operator = operator.(string)
		check.Remark = dat.Remark
		if dat.Status == 0 {
			allCheckFlag = false //确认是否所有自查项全部被确认
		} else {
			status, ok := dataMap[uint(dat.Id)]
			//数据库中不存在该确认记录，创建一条新纪录
			if !ok {
				check.CreatedAt = time.Now()
				check.UpdatedAt = time.Now()
				if err = db.Table(ConfirmCheck{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&check).Error; err != nil {
					logs.Error("insert tb_confirm_check failed, %v", err)
					db.Rollback()
					return false
				}
			} else {
				//数据库中记录为0，无效记录，更新这条记录即可
				if status == 0 {
					condition := "task_id='" + strconv.Itoa(taskId.(int)) + "' and item_id='" + strconv.Itoa(dat.Id) + "'"
					check.UpdatedAt = time.Now()
					if err = db.Table(ConfirmCheck{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).
						Update(map[string]interface{}{"status": dat.Status, "updated_at": time.Now(), "remark": dat.Remark, "operator": operator.(string)}).Error; err != nil {
						logs.Error("insert tb_confirm_check failed, %v", err)
						db.Rollback()
						return false
					}
				}
			}
		}
	}
	//最后更新检测任务的自查状态
	if allCheckFlag {
		if err = db.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).
			Where("id=?", taskId).Update("self_check_status", 1).Error; err != nil {
			logs.Error("%v", err)
			db.Rollback()
			return false
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
