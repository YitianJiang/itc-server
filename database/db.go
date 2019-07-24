package database

import (
	"fmt"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/utils"
	dbconf "code.byted.org/gopkg/dbutil/conf"
	"code.byted.org/gopkg/dbutil/gormdb"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

var (
	dboptional dbconf.DBOptional
)

func InitDB() {

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

	//test
	//ssConf, _ := ssconf.LoadSsConfFile(conf.Configuration.MysqlConfigPath)
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
