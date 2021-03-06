package dal

import (
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type ItemConfig struct {
	gorm.Model
	ConfigType 		int				`json:"configType"` //0-问题类型；1-关键词；2-修复方式
	Name 			string			`json:"name"`
	Platform		int				`json:"platform"`	//所属平台，0-安卓；1-iOS
}

func (ItemConfig) TableName() string{
	return "tb_config"
}
//insert
func InsertItemConfig(config ItemConfig) bool {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	err = connection.Table(ItemConfig{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&config).Error
	if err != nil {
		logs.Error("新增检查项配置失败！", err)
		return false
	}
	return true
}
//query by condition
func QueryConfigByCondition(condition string) *[]ItemConfig {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var configs []ItemConfig
	err = connection.Table(ItemConfig{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Find(&configs).Error
	if err != nil {
		logs.Error("查询检查配置项失败！", err)
		return nil
	}
	return &configs
}
//update
func UpdateConfigByCondition(condition string, config ItemConfig) bool {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	err = connection.Table(ItemConfig{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Update(config).Error
	if err != nil {
		logs.Error("配置项更新失败", err)
		return false
	}
	return true
}
