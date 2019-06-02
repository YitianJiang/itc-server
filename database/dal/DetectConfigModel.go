package dal

import (
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type DetectConfigStruct struct {
	gorm.Model
	KeyInfo			string			`json:"keyInfo"`
	//Type			int				`json:"type"` 	//1--权限，2--api，3--action，4--monitor
	Priority		int				`json:"priority"`//0--常规，1--注意，2--危险，3--非常危险，4--未定义
	Ability			string			`json:"ability"`
	DescInfo 		string			`json:"desc"`
	DetectType 		string			`json:"detectType"`
	Suggestion		string			`json:"suggestion"`
	AbilityGroup	string			`json:"group"`
	Reference 		string			`json:"refer"`
}

type DetectConfigInfo struct {
	ID 				int			`json:"id"`
	KeyInfo 		string		`json:"key"`
	//Type 			int 		`json:"type"`
	Priority    	int 		`json:"priority"`
	Ability			string		`json:"ability"`
	DescInfo    	string		`json:"desc"`
	DetectType  	string		`json:"detectType"`
	Suggestion  	string 		`json:"suggestion"`
	AbilityGroup	string		`json:"abilityGroup"`
	Reference  		string		`json:"refer"`
}

type PermHistory struct {
	gorm.Model
	PermId				int					`json:"permId"`
	Status				int					`json:"status"`
	Remark				string 				`json:"remark"`
	Confirmer			string				`json:"confirmer"`
	AppId				int					`json:"appId"`
	AppVersion			string				`json:"appVersion"`
}

type PermAppRelation struct {
	gorm.Model
	PermId				int					`json:"permId"`
	AppId				int					`json:"appId"`
	AppVersion			string				`json:"appVersion"`
}


func (DetectConfigStruct) TableName() string {
	return "tb_detect_config"
}

func (PermHistory) TableName() string {
	return "tb_perm_history"
}

func (PermAppRelation) TableName() string {
	return "tb_perm_app_relation"
}


func InsertDetectConfig(data DetectConfigStruct) error{
	connection,err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db err,%v",err)
		return err
	}
	defer connection.Close()
	db := connection.Table(DetectConfigStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Create(&data).Error; err != nil {
		logs.Error("insert detectConfig err,%v",err)
		return err
	}
	return nil
}

func QueryDetectConfig(queryData map[string]interface{}) *[]DetectConfigStruct  {
	connection,err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db error,%v",err)
		return nil
	}
	defer connection.Close()
	db := connection.Table(DetectConfigStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result []DetectConfigStruct
	if err := db.Where(queryData).Find(&result).Error; err != nil {
		logs.Error("query detectConfig failed,%v",err)
		return nil
	}
	return &result
}

func UpdataDetectConfig(data map[string]interface{}) error  {
	connection,err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed, %v",err)
		return err
	}
	defer connection.Close()
	condition := data["condition"]
	info := data["update"]//应该是map[string]interface{}形式
	db := connection.Table(DetectConfigStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Update(info).Error; err != nil {
		logs.Error("update detectConfig failed,info: %v;errInfo:%v",data,err)
		return err
	}
	return nil
}

func DeleteDetectConfig(condition string) error  {
	connection,err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v",err)
		return err
	}
	defer connection.Close()
	db := connection.Table(DetectConfigStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var t DetectConfigInfo
	if err := db.Where(condition).Delete(&t).Error; err != nil {
		logs.Error("delete detectConfig failed,info:%v;errorInfo:%v",condition,err)
		return err
	}
	return nil
}

