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
	dbhandler  *gorm.DB
)

// InitDB initialize the dboptional
func InitDB() {

	var err error
	dboptional, err = dbconf.GetDBOptionalByConsulName("toutiao.mysql.itcserver_write")
	if err != nil {
		logs.Error("mysql gdpr failed,%v", err)
		for _, lark_people := range _const.LowLarkPeople {
			utils.LarkDingOneInner(lark_people, "mysql gdpr failed！")
		}
	}
}

// GetDBConnection returns database handler if connect to database successfully.
func GetDBConnection() (*gorm.DB, error) {

	handler := gormdb.NewDBHandler()
	if err := handler.ConnectDB(&dboptional); err != nil {
		return nil, fmt.Errorf("Connect DB failed: %v", err)
	}

	return handler.GetConnection()
}

// InitDBHandler initialize a global database handler.
func InitDBHandler() error {

	var err error
	if dbhandler, err = GetDBConnection(); err != nil {
		logs.Error("failed to get database handler: %v", err)
		return err
	}

	return nil
}

// DB returns the handler of database.
func DB() *gorm.DB {

	return dbhandler
}
