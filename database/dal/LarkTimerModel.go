package dal

import (
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"fmt"
	"strconv"
	"time"
)

type LarkMsgTimer struct {
	gorm.Model
	AppId			int			`json:"appId"`
	Type			int			`json:"type"` //0-秒，1-分钟，2-小时，3-天
	MsgInterval		int			`json:"msgInterval"`
	Operator		string		`json:"operator"`
}
type LarkGroupMsg struct {
	gorm.Model
	TimerId			int			`json:"timerId"`
	GroupId			string		`json:"groupId"`
	GroupName		string		`json:"groupName"`
	Operator		string		`json:"operator"`
	Platform		int			`json:"platform"`
}

func (LarkMsgTimer) TableName() string {
	return "tb_lark_msg_timer"
}
func (LarkGroupMsg) TableName() string {
	return "tb_lark_group"
}

func InsertLarkMsgTimer(timer LarkMsgTimer) bool {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	var larkTimer LarkMsgTimer
	condition := "app_id='" + fmt.Sprint(timer.AppId) + "'"
	logs.Info(condition)
	if err = connection.Table(LarkMsgTimer{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Where(condition).Find(&larkTimer).Error; err != nil {
		logs.Error("query lark message timer failed")
		if err := connection.Table(LarkMsgTimer{}.TableName()).LogMode(_const.DB_LOG_MODE).
			Create(&timer).Error; err != nil {
			logs.Error("insert lark message timer failed, %v", err)
			return false
		}
		return true
	}
	if err = connection.Table(LarkMsgTimer{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Where(condition).
		Update(map[string]interface{}{
			"msg_interval" : timer.MsgInterval,
			"updated_at" : time.Now(),
			"operator" : timer.Operator,
			"type" : timer.Type}).
		Error; err != nil {
		logs.Error("update lark message timer failed, %v", err)
		return false
	}
	return true
}
func QueryLarkMsgTimerByAppId(appId int) *LarkMsgTimer {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var larkTimer LarkMsgTimer
	condition := "app_id='" + strconv.Itoa(appId) + "'"
	if err := connection.Table(LarkMsgTimer{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Where(condition).Find(&larkTimer).Error; err != nil {
		logs.Error("query lark message timer failed, %v", err)
		return nil
	}
	return &larkTimer
}
//insert data
func InsertLarkGroup(group LarkGroupMsg) bool {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	db := connection.Table(LarkGroupMsg{}.TableName()).LogMode(_const.DB_LOG_MODE)
	err = db.Create(&group).Error
	if err != nil {
		logs.Error("Insert lark group info failed: %v", err)
		return false
	}
	return true
}
//query data by condition
func QueryLarkGroupByCondition(condition string) *[]LarkGroupMsg {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var group []LarkGroupMsg
	db := connection.Table(LarkGroupMsg{}.TableName()).LogMode(_const.DB_LOG_MODE)
	err = db.Where(condition).Find(&group).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.Error("lark group data is empty")
			return &group
		} else {
			logs.Error("query lark group failed, %v", err)
			return nil
		}
	}
	return &group
}
//update data by id
func UpdateLarkGroupById(larkGroup LarkGroupMsg) bool {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	id := larkGroup.ID
	condition := "id='" + fmt.Sprint(id) + "'"
	db := connection.Table(LarkGroupMsg{}.TableName()).LogMode(_const.DB_LOG_MODE)
	err = db.Where(condition).Update(map[string]interface{}{
		"group_name" : larkGroup.GroupName,
		"group_id" : larkGroup.GroupId,
		"operator" : larkGroup.Operator,
		"platform" : larkGroup.Platform,
		"updated_at" : time.Now()}).Error
	if err != nil {
		logs.Error("update lark group info failed, %v", err)
		return false
	}
	return true
}
//delete data by id
func DeleteLarkGroupById(larkGroup LarkGroupMsg) bool {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	id := larkGroup.ID
	condition := "id='" + fmt.Sprint(id) + "'"
	dbUpdate := connection.Table(LarkGroupMsg{}.TableName()).LogMode(_const.DB_LOG_MODE)
	//记录是谁进行的删除操作
	err = dbUpdate.Where(condition).Update(map[string]interface{}{
		"operator" : larkGroup.Operator,
		"updated_at" : time.Now(),
	}).Error
	db := connection.Table(LarkGroupMsg{}.TableName()).LogMode(_const.DB_LOG_MODE)
	err = db.Where(condition).Delete(&larkGroup).Error
	if err != nil {
		logs.Error("deleted lark group info failed, %v", err)
		return false
	}
	return true
}