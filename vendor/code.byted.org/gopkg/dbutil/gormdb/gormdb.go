package gormdb

import (
	"errors"
	"fmt"
	"sync"

	"code.byted.org/gopkg/dbutil/conf"
	"code.byted.org/gopkg/gorm"
	_ "code.byted.org/gopkg/mysql-driver"
)

/**
 * 数据库handler
 */
type DBHandler struct {
	db        *gorm.DB
	optional  *conf.DBOptional
	connected bool
	mu        sync.Mutex
}

func NewDBHandler() *DBHandler {
	return &DBHandler{
		connected: false,
	}
}

func NewDBHandlerWithOptional(optional *conf.DBOptional) *DBHandler {
	return &DBHandler{
		connected: false,
		optional:  optional,
	}
}

func (handler *DBHandler) ConnectDB(optional *conf.DBOptional) (err error) {
	if handler.connected {
		return nil
	}
	handler.mu.Lock()
	defer handler.mu.Unlock()
	if handler.connected {
		return nil
	}

	if optional == nil {
		return errors.New("connect option is nil")
	}

	handler.optional = optional
	dbconfig := handler.optional.GenerateConfig()

	if handler.optional.DriverName == "mysql" {
		handler.optional.DriverName = "mysql2"
		fmt.Println("DBUtil: mysql driver was hacked to mysql2; mysql2 is the mysql driver for toutiao; please see code.byted.org/gopkg/mysql-driver")
	}

	handler.db, err = gorm.Open(handler.optional.DriverName, dbconfig)
	if err != nil {
		return err
	}

	handler.connected = true
	handler.db.DB().SetMaxIdleConns(handler.optional.MaxIdleConns)
	handler.db.DB().SetMaxOpenConns(handler.optional.MaxOpenConns)
	handler.db.SingularTable(true)
	return nil
}

/**
 * 获取db连接，handler用于管理db连接，通过该接口获取，线程安全
 */
func (handler *DBHandler) GetConnection() (*gorm.DB, error) {
	err := handler.ConnectDB(handler.optional)
	if err != nil {
		return nil, err
	}
	return handler.db, nil
}

/**
 * 关闭连接
 */
func (handler *DBHandler) Close() error {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	if handler.connected {
		handler.connected = false
		err := handler.db.Close()
		return err
	}
	return nil
}
