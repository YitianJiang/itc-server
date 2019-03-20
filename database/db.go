package database

import (
	"code.byted.org/clientQA/itc-server/conf"
	"code.byted.org/golf/ssconf"
	dbconf "code.byted.org/gopkg/dbutil/conf"
	"code.byted.org/gopkg/dbutil/gormdb"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"fmt"
)

var(
	dboptional dbconf.DBOptional
)

func InitDB(){
	ssConf, _ := ssconf.LoadSsConfFile(conf.Configuration.MysqlConfigPath)
	dboptional = dbconf.GetDbConf(ssConf, "itcserver", dbconf.Write)
}

func GetConneection()(*gorm.DB, error){
	handler := gormdb.NewDBHandler()
	logs.Info("dboptional: ", dboptional.GenerateConfig())
	err := handler.ConnectDB(&dboptional)
	if err != nil{
		return nil, fmt.Errorf("Connect DB failed: %v", err)
	}
	return handler.GetConnection()
}

