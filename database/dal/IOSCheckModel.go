package dal

import (
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"fmt"
	"time"
)

type IOSCheckTask struct {
	gorm.Model
	UserName	string					`json:"userName"`
	RepoUrl	    string					`json:"repoUrl"`
	Branch	    string					`json:"branch"`
	AppId		int						`json:"appId"`
	Result      string					`json:"result"`
}
type RetICTTasks struct {
	GetMore uint
	Total uint
	NowPage uint
	Tasks []IOSCheckTask
}
func (IOSCheckTask) TableName() string{
	return "tb_ios_check_task"
}

//insert data
func InsertICT(ictModel IOSCheckTask) (uint, bool) {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return 0, false
	}
	defer connection.Close()
	if err := connection.Table(IOSCheckTask{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Create(&ictModel).Error; err != nil{
		logs.Error("insert ios check task failed, %v", err)
		return 0, false
	}
	return ictModel.ID, true
}
//query data
func QueryICT(data map[string]interface{}) (*[]IOSCheckTask, uint) {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, 0
	}
	defer connection.Close()
	var ict []IOSCheckTask
	db := connection.Table(IOSCheckTask{}.TableName()).LogMode(_const.DB_LOG_MODE)
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
	if err := db.Find(&ict).Error; err != nil{
		logs.Error("query ict data failed, %v", err)
		return nil, 0
	}
	var total RecordTotal
	if condition == "" {
		condition = " 1=1 "
	}
	connect, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, 0
	}
	defer connect.Close()
	dbCount := connect.Table(IOSCheckTask{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := dbCount.Select("count(id) as total").
		Where(condition).Find(&total).Error; err != nil{
		logs.Error("query total record failed! %v", err)
		return &ict, 0
	}
	return &ict, total.Total
}
//update data
func UpdateICT(ictModel IOSCheckTask) bool {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false
	}
	defer connection.Close()
	id := ictModel.ID
	condition := "id='" + fmt.Sprint(id) + "'"
	if err := connection.Table(IOSCheckTask{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Where(condition).Update(map[string]interface{}{"result":ictModel.Result, "updated_at":time.Now()}).Error; err != nil{
		logs.Error("update ict data failed, %v", err)
		return false
	}
	return true
}
