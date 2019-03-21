package dal

import (
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"fmt"
)
//二进制包检测任务
type DetectStruct struct {
	gorm.Model
	Creator string 			`json:"creator"`
	Platform int			`json:"platform"`
	AppName string			`json:"appName"`
	AppVersion string		`json:"appVersion"`
	AppId string			`json:"appId"`
	CheckContent string		`json:"checkContent"`
	SelfCheckStatus int		`json:"selfCheckStatus"` //0-自查未完成；1-自查完成
	TosUrl string			`json:"tosUrl"`
}
type RecordTotal struct {
	Total uint
}
type RetDetectTasks struct {
	GetMore uint
	Total uint
	NowPage uint
	Tasks []DetectStruct
}
//包检测工具
type DetectTool struct {
	gorm.Model
	Name string 			`json:"name"`
	Description string 		`json:"description"`
	Platform int 			`json:"platform"`
}
//二进制包检测内容
type DetectContent struct {
	gorm.Model
	TaskId int				`json:"taskId"`
	ToolId int				`json:"toolId"`
	HtmlContent string		`json:"htmlContent"`
	JsonContent string		`json:"jsonContent"`
	Status int				`json:"status"`//是否确认
}
func (DetectStruct) TableName() string {
	return "tb_binary_detect"
}
func (DetectContent) TableName() string {
	return "tb_detect_content"
}
func (DetectTool) TableName() string {
	return "tb_detect_tool"
}
//insert data
func InsertDetectModel(detectModel DetectStruct) uint {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return 0
	}
	defer connection.Close()
	db := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Create(&detectModel).Error; err != nil{
		return 0
	}
	return detectModel.ID
}
//update data
func UpdateDetectModel(detectModel DetectStruct, content DetectContent) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Begin()
	taskId := detectModel.ID
	condition := "id=" + fmt.Sprint(taskId)
	if err := db.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Where(condition).Update(&detectModel).Error; err != nil {
		logs.Error("update binary check failed, %v", err)
		db.Rollback()
		return err
	}
	//insert detect content
	if err := db.Table(DetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Create(&content).Error; err != nil {
		logs.Error("insert binary check content failed, %v", err)
		db.Rollback()
		return err
	}
	db.Commit()
	return nil
}
//delete data
func DeleteDetectModel(detectModeId string) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Table(DetectStruct{}.TableName())
	if err := db.Where("id = ?", detectModeId).LogMode(_const.DB_LOG_MODE).Delete(&DetectStruct{}).Error; err != nil{
		logs.Error("%v", err)
		return err
	}
	return nil
}
//query by map
func QueryDetectModelsByMap(param map[string]interface{}) *[]DetectStruct{
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var detect []DetectStruct
	db := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(param).Find(&detect).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &detect
}
//query data
func QueryTasksByCondition(data map[string]interface{}) (*[]DetectStruct, uint) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, 0
	}
	defer connection.Close()
	db := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	condition := data["condition"]
	logs.Info("query tasks condition: %s", condition)
	if condition != "" {
		db.Where(condition)
	}
	pageNo, okpn := data["pageNo"]
	pageSize, okps := data["pageSize"]
	if okpn {
		if !okps {
			pageSize = 10
		}
		page := pageNo.(int)
		size := pageSize.(int)
		db.Limit(pageSize)
		if page > 0 {
			db.Offset((page - 1) * size)
		}
	}
	var items []DetectStruct
	if err := db.Find(&items).Error; err != nil{
		logs.Error("%v", err)
		return nil, 0
	}
	var total RecordTotal
	if condition == "" {
		condition = " 1=1 "
	}
	if err := db.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Select("count(id) as total").
		Where(condition).Find(&total).Error; err != nil{
		logs.Error("query total record failed! %v", err)
		return &items, 0
	}
	return &items, total.Total
}
//query by map
func QueryTaskBinaryCheckContent(condition string) *[]DetectContent{
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var detect []DetectContent
	db := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Find(&detect).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &detect
}

func ConfirmBinaryResult(data map[string]string) bool {
	taskId := data["task_id"]
	toolId := data["tool_id"]
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	//var dc DetectContent
	//db := connection.Table(DetectContent{}.TableName()).Begin()
	db := connection.Table(DetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE)
	condition := "task_id=" + taskId + " and tool_id=" + toolId
	/*condition := "task_id=" + taskId + " and tool_id=" + toolId + " for update"
	if err := db.Where(condition).Find(&dc).Error; err != nil {
		logs.Error("query db tb_detect_content failed: %v", err)
		db.Rollback()
		return false
	}
	if dc.Status == 1{
		db.Rollback()
		return false
	}*/
	if err := db.Where(condition).LogMode(_const.DB_LOG_MODE).Update("status", 1).Error; err != nil {
		logs.Error("update db tb_detect_content failed: %v", err)
		//db.Rollback()
		return false
	}
	//db.Commit()
	return true
}
