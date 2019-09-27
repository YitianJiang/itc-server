package database

import (
	"fmt"

	"code.byted.org/clientQA/itc-server/conf"
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/golf/ssconf"
	dbconf "code.byted.org/gopkg/dbutil/conf"
	"code.byted.org/gopkg/dbutil/gormdb"
	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

var (
	dboptional dbconf.DBOptional
	// DB is the handler of database.
	DB *gorm.DB
)

func InitDB() {

	if env.IsBoe() {
		//boedb
		ssConf, _ := ssconf.LoadSsConfFile(conf.Configuration.MysqlConfigPath)
		dboptional = dbconf.GetDbConf(ssConf, "itcserver", dbconf.Write)
	} else {
		//online
		//线上采用mysql gdpr
		var err error
		dboptional, err = dbconf.GetDBOptionalByConsulName("toutiao.mysql.itcserver_write")
		if err != nil {
			logs.Error("mysql gdpr failed,%v", err)
			for _, lark_people := range _const.LowLarkPeople {
				utils.LarkDingOneInner(lark_people, "mysql gdpr failed！")
			}
		}
	}

	//test
	// ssConf, _ := ssconf.LoadSsConfFile(conf.Configuration.MysqlConfigPath)
	// dboptional = dbconf.GetDbConf(ssConf, "qa_ee", dbconf.Write)
}

// GetDBConnection returns database handler if connect to database successfully.
func GetDBConnection() (*gorm.DB, error) {

	handler := gormdb.NewDBHandler()
	if err := handler.ConnectDB(&dboptional); err != nil {
		return nil, fmt.Errorf("Connect DB failed: %v", err)
	}

	return handler.GetConnection()
}
