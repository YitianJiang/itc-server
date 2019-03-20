package dal

import (
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"fmt"
	"strconv"
)

type LarkMsgTimer struct {
	gorm.Model
	AppId	int			`json:"appId"`
	Type	int			`json:"type"` //0-秒，1-分钟，2-小时，3-天
	MsgInterval	int		`json:"msgInterval"`
}

func (LarkMsgTimer) TableName() string {
	return "tb_lark_msg_timer"
}

func InsertLarkMsgTimer(timer LarkMsgTimer) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	var larkTimer LarkMsgTimer
	condition := "id='" + fmt.Sprint(timer.ID) + "'"
	if err = connection.Table(LarkMsgTimer{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Where(condition).Find(&larkTimer).Error; err != nil {
		if err := connection.Table(LarkMsgTimer{}.TableName()).LogMode(_const.DB_LOG_MODE).
			Create(&timer).Error; err != nil {
			logs.Error("insert lark message timer failed, %v", err)
			return false
		}
		return true
	}
	if err = connection.Table(LarkMsgTimer{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Where(condition).Update("msg_interval", timer.MsgInterval).Error; err != nil {
		logs.Error("update lark message timer failed, %v", err)
		return false
	}
	return true
}
func QueryLarkMsgTimerByAppId(appId int) *LarkMsgTimer {
	connection, err := database.GetConneection()
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