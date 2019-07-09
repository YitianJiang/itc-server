package database

import (
	"fmt"

	"code.byted.org/clientQA/itc-server/conf"
	"code.byted.org/golf/ssconf"
	dbconf "code.byted.org/gopkg/dbutil/conf"
	"code.byted.org/gopkg/dbutil/gormdb"
	"code.byted.org/gopkg/gorm"
)

var (
	dboptional dbconf.DBOptional
)

func InitDB() {
	ssConf, _ := ssconf.LoadSsConfFile(conf.Configuration.MysqlConfigPath)
	// online
	dboptional = dbconf.GetDbConf(ssConf, "itcserver", dbconf.Write)

	//test
	//dboptional = dbconf.GetDbConf(ssConf, "qa_ee", dbconf.Write)
}

func GetConneection() (*gorm.DB, error) {
	handler := gormdb.NewDBHandler()
	err := handler.ConnectDB(&dboptional)
	if err != nil {
		return nil, fmt.Errorf("Connect DB failed: %v", err)
	}
	return handler.GetConnection()
}
