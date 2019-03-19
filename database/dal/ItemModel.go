package dal

import (
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"time"
)

type ItemStruct struct {
	gorm.Model
	QuestionType int		`json:"questionType"`
	QuestionTypeName string	`json:"questionTypeName"`
	KeyWord int				`json:"keyWord"`
	FixWay int				`json:"fixWay"`
	CheckContent string		`json:"checkContent"`
	Resolution string		`json:"resolution"`
	Regulation string		`json:"regulation"`
	RegulationUrl string	`json:"regulationUrl"`
	IsGG int				`json:"isGg"`
	AppId int				`json:"appId"`
	Platform int			`json:"platform"`
	Status int				`json:"status"`
}
type ConfirmCheck struct {
	gorm.Model
	TaskId int				`json:"taskId"`
	ItemId int				`json:"itemId"`
	Status int				`json:"status"`
	Operator string			`json:"operator"`
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
	if err := connection.Create(&itemModel).Error; err != nil{
		return 0
	}
	return itemModel.ID
}
//query data
func QueryItemsByCondition(data map[string]interface{}) *[]ItemStruct {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer connection.Close()
	db := connection.Table(ItemStruct{}.TableName())
	condition := data["condition"]
	logs.Info("query items condition: %s", condition)
	if condition != "" {
		db.Where(condition)
	}
	var items []ItemStruct
	if err := db.Find(&items).Error; err != nil{
		logs.Error("%v", err)
		return nil
	}
	return &items
}
//confirm check
func ConfirmSelfCheck(param map[string]interface{}) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false
	}
	defer connection.Close()
	operator := param["operator"]
	taskId := param["taskId"]
	data := param["data"]
	db := connection.Begin()
	//先更新检测任务的自查状态
	if err = db.Table(DetectStruct{}.TableName()).Where("id=?", taskId).Update("SelfCheckStatus", 1).Error; err != nil{
		logs.Error("%v", err)
		db.Rollback()
		return false
	}
	idArray := data.([]string)
	for i:=0; i<len(idArray); i++ {
		var dat map[string]interface{}
		dat = make(map[string]interface{})
		str := idArray[i]
		err := json.Unmarshal([]byte(str), &dat); if err == nil {
			var check ConfirmCheck
			check.ItemId = dat["id"].(int)
			check.Status = dat["status"].(int)
			check.TaskId = taskId.(int)
			check.Operator = operator.(string)
			check.CreatedAt = time.Now()
			check.UpdatedAt = time.Now()
			if err = db.Table(ConfirmCheck{}.TableName()).Create(&check).Error; err != nil {
				logs.Error("insert tb_confirm_check failed, %v", err)
				db.Rollback()
				return false
			}
		}
	}
	db.Commit()
	return true
}
//根据任务id拿到对应的自查信息
func GetSelfCheckByTaskId(condition string) map[uint]int{
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer connection.Close()
	db := connection.Table(ConfirmCheck{}.TableName())
	var items []ConfirmCheck
	if err = db.Where(condition).Find(&items).Error; err != nil {
		logs.Error("query self check item failed, %v", err)
		return nil
	}
	if len(items) == 0 {
		logs.Info("query self check item empty")
		return nil
	}
	var item map[uint]int
	item = make(map[uint]int)
	for i := 0; i < len(items); i++ {
		it := items[i]
		itemId := it.ItemId
		item[uint(itemId)] = it.Status
	}
	return item
}