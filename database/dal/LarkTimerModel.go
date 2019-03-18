package dal

import (
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type LarkMsgTimer struct {
	gorm.Model
	AppId	int			`json:"appId"`
	Type	int			`json:"type"` //0-秒，1-分钟，2-小时，3-天
	Interval	int		`json:"interval"`
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
	if err := connection.Table(LarkMsgTimer{}.TableName()).Create(timer).Error; err != nil {
		logs.Error("insert lark message timer failed, %v", err)
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
	if err := connection.Table(LarkMsgTimer{}.TableName()).Where("app_id=", appId).Find(&larkTimer).Error; err != nil {
		logs.Error("query lark message timer failed, %v", err)
		return nil
	}
	return &larkTimer
}