package database

import (
	"code.byted.org/clientQA/pkg/conf"
	"code.byted.org/golf/ssconf"
	dbconf "code.byted.org/gopkg/dbutil/conf"
	"code.byted.org/gopkg/dbutil/gormdb"
	"code.byted.org/gopkg/gorm"
	"fmt"
)

var (
	dboptional dbconf.DBOptional
)

func InitDB() {
	ssConf, _ := ssconf.LoadSsConfFile(conf.ProjectConfig.MysqlConfPath)
	dboptional = dbconf.GetDbConf(ssConf, "tt_pkg", dbconf.Write)
}

func GetConnection() (*gorm.DB, error) {
	handler := gormdb.NewDBHandler()
	err := handler.ConnectDB(&dboptional)
	if err != nil {
		return nil, fmt.Errorf("Connect DB Failed: %v", err)
	}
	return handler.GetConnection()
}
