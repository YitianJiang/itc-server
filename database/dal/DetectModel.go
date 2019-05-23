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

//二进制包检测任务
type DetectStruct struct {
	gorm.Model
	Creator         string `json:"creator"`
	Platform        int    `json:"platform"`
	AppName         string `json:"appName"`
	AppVersion      string `json:"appVersion"`
	AppId           string `json:"appId"`
	CheckContent    string `json:"checkContent"`
	SelfCheckStatus int    `json:"selfCheckStatus"` //0-自查未完成；1-自查完成
	TosUrl          string `json:"tosUrl"`
}
type RecordTotal struct {
	Total uint
}
type RetDetectTasks struct {
	GetMore uint
	Total   uint
	NowPage uint
	Tasks   []DetectStruct
}

//包检测工具
type DetectTool struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Platform    int    `json:"platform"`
}

//二进制包检测内容
type DetectContent struct {
	gorm.Model
	TaskId      int    `json:"taskId"`
	ToolId      int    `json:"toolId"`
	HtmlContent string `json:"htmlContent"`
	JsonContent string `json:"jsonContent"`
	Status      int    `json:"status"` //是否确认,0-未确认，1-确认通过，2-确认未通过
	Confirmer   string `json:"confirmer"`
	Remark      string `json:"remark"`
}

//二进制包检测内容，json内容处理区分后
type IOSDetectContent struct {
	gorm.Model
	TaskId          int    `gorm:"column:taskId"            json:"taskId"`
	ToolId          int    `gorm:"column:toolId"            json:"toolId"`
	JsonContent     string `gorm:"column:jsonContent"       json:"jsonContent"`
	Category        string `gorm:"column:category"          json:"category"`
	CategoryName    string `gorm:"column:categoryName"      json:"categoryName"`
	CategoryContent string `gorm:"column:categoryContent"   json:"categoryContent"`
	Status          int    `gorm:"column:status"            json:"status"` //是否确认,0-未确认，1-确认通过，2-确认未通过
	Confirmer       string `gorm:"column:confirmer"         json:"confirmer"`
	Remark          string `gorm:"column:remark"            json:"remark"`
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
func (IOSDetectContent) TableName() string {
	return "tb_ios_detect_content"
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
	if err := db.Create(&detectModel).Error; err != nil {
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
	if err := db.Where("id = ?", detectModeId).LogMode(_const.DB_LOG_MODE).Delete(&DetectStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return err
	}
	return nil
}

/**
 * 更新tos地址
 */
func UpdateDetectTosUrl(path string, taskId uint) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	condition := "id='" + fmt.Sprint(taskId) + "'"
	db := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Update(map[string]interface{}{"tos_url": path, "updated_at": time.Now()}).Error; err != nil {
		logs.Error("%v", err)
		return false
	}
	return true
}

//query by map
func QueryDetectModelsByMap(param map[string]interface{}) *[]DetectStruct {
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
		db = db.Where(condition).Order("created_at desc")
	}
	pageNo, okpn := data["pageNo"]
	pageSize, okps := data["pageSize"]
	if okpn {
		if !okps {
			pageSize = 10
		}
		page := pageNo.(int)
		size := pageSize.(int)
		db = db.Limit(pageSize)
		if page > 0 {
			db = db.Offset((page - 1) * size)
		}
	}
	var items []DetectStruct
	if err := db.Find(&items).Error; err != nil {
		logs.Error("%v", err)
		return nil, 0
	}
	var total RecordTotal
	if condition == "" {
		condition = " 1=1 "
	}
	connect, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, 0
	}
	defer connect.Close()
	dbCount := connect.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := dbCount.Select("count(id) as total").
		Where(condition).Find(&total).Error; err != nil {
		logs.Error("query total record failed! %v", err)
		return &items, 0
	}
	return &items, total.Total
}

//query by map
func QueryTaskBinaryCheckContent(condition string) *[]DetectContent {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var detect []DetectContent
	db := connection.Table(DetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Find(&detect).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &detect
}

func ConfirmBinaryResult(data map[string]string) bool {
	taskId := data["task_id"]
	toolId := data["tool_id"]
	confirmer := data["confirmer"]
	remark := data["remark"]
	status := data["status"]
	statusInt, _ := strconv.Atoi(status)
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	db := connection.Table(DetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE)
	condition := "task_id=" + taskId + " and tool_id=" + toolId
	if err := db.Where(condition).LogMode(_const.DB_LOG_MODE).
		Update(map[string]interface{}{
			"status":     statusInt,
			"confirmer":  confirmer,
			"remark":     remark,
			"updated_at": time.Now(),
		}).Error; err != nil {
		logs.Error("update db tb_detect_content failed: %v", err)
		//db.Rollback()
		return false
	}
	//db.Commit()
	return true
}

//insert tb_ios_detect_content
func CreateIOSDetectModel(content IOSDetectContent) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Begin()
	//insert detect content
	if err := db.Table(IOSDetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Create(&content).Error; err != nil {
		logs.Error("insert binary check content failed, %v", err)
		db.Rollback()
		return err
	}
	db.Commit()
	return nil
}

//query tb_ios_detect_content
func QueryIOSDetectModel(condition map[string]interface{}) *[]IOSDetectContent {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer connection.Close()

	var iosDetectContent []IOSDetectContent
	if err := connection.Table(IOSDetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Find(&iosDetectContent).Error; err != nil {
		logs.Error("请求iOS静态检测结果出错！！！", err.Error())
		return nil
	}
	return &iosDetectContent
}

//update tb_ios_detect_content
func UpdateIOSDetectModel(id int, updates map[string]interface{}) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false
	}
	defer connection.Close()
	if err := connection.Table(IOSDetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).Model(&IOSDetectContent{}).Where("id = ?", id).Update(updates).Error; err != nil {
		logs.Error("更新iOS静态检测结果出错！！！", err.Error())
		return false
	}
	return true
}
