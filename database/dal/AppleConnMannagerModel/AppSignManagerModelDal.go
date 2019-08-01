package devconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
	"fmt"
	"github.com/astaxie/beego/logs"
)

/**
数据库Insert操作
条件：record对应的struct有TableName()方法
*/
func InsertRecord(record interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()

	insertError := conn.Debug().Create(record).Error
	if insertError != nil {
		utils.RecordError("Insert into DB Failed: ", insertError)
		return insertError
	}
	return nil
}

/**
查询操作，
返回nil---查询fail，返回空数组--无相关数据
*/
func QueryAppAccountCert(queryData map[string]interface{}) *[]AppAccountCert {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	db := conn.Table(AppAccountCert{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result = make([]AppAccountCert, 0)
	if err := db.Where(queryData).Find(&result).Error; err != nil {
		utils.RecordError("查询 tt_app_account_cert失败，查询条件："+fmt.Sprint(queryData)+",errInfo：", err)
		return nil
	}
	return &result
}

/**
删除操作
*/
func DeleteAppAccountCert(queryData map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppAccountCert{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var t AppAccountCert
	if err1 := db.Where(queryData).Delete(&t).Error; err1 != nil {
		logs.Error("删除 tt_app_account_cert失败，删除条件："+fmt.Sprint(queryData)+",errInfo：%v", err1)
		utils.RecordError("删除 tt_app_account_cert失败，删除条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}

/**
查询操作，
返回nil---查询fail，返回空数组--无相关数据
*/
func QueryAppBoundleProfiles(queryData map[string]interface{}) *[]AppBoundleProfiles {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	db := conn.Table(AppBoundleProfiles{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result = make([]AppBoundleProfiles, 0)
	if err := db.Where(queryData).Find(&result).Error; err != nil {
		utils.RecordError("查询 tt_app_account_cert失败，查询条件："+fmt.Sprint(queryData)+",errInfo：", err)
		return nil
	}
	return &result
}

/**
删除操作
*/
func DeleteAppBoundleProfiles(queryData map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppBoundleProfiles{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var t AppBoundleProfiles
	if err1 := db.Where(queryData).Delete(&t).Error; err1 != nil {
		logs.Error("删除 tt_app_account_cert失败，删除条件："+fmt.Sprint(queryData)+",errInfo：%v", err1)
		utils.RecordError("删除 tt_app_account_cert失败，删除条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}

/**
查询操作，
返回nil---查询fail，返回空数组--无相关数据
*/
func QueryAppleBundleId(queryData map[string]interface{}) *[]AppleBundleId {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	db := conn.Table(AppleBundleId{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result = make([]AppleBundleId, 0)
	if err := db.Where(queryData).Find(&result).Error; err != nil {
		utils.RecordError("查询 tt_app_account_cert失败，查询条件："+fmt.Sprint(queryData)+",errInfo：", err)
		return nil
	}
	return &result
}

/**
删除操作
*/
func DeleteAppleBoundleId(queryData map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppleBundleId{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var t AppleBundleId
	if err1 := db.Where(queryData).Delete(&t).Error; err1 != nil {
		logs.Error("删除 tt_app_account_cert失败，删除条件："+fmt.Sprint(queryData)+",errInfo：%v", err1)
		utils.RecordError("删除 tt_app_account_cert失败，删除条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}

/**
查询操作，
返回nil---查询fail，返回空数组--无相关数据
*/
func QueryAppleProfile(queryData map[string]interface{}) *[]AppleProfile {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	db := conn.Table(AppleProfile{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result = make([]AppleProfile, 0)
	if err := db.Where(queryData).Find(&result).Error; err != nil {
		utils.RecordError("查询 tt_app_account_cert失败，查询条件："+fmt.Sprint(queryData)+",errInfo：", err)
		return nil
	}
	return &result
}

/**
删除操作
*/
func DeleteAppleProfile(queryData map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppleProfile{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var t AppleProfile
	if err1 := db.Where(queryData).Delete(&t).Error; err1 != nil {
		logs.Error("删除 tt_app_account_cert失败，删除条件："+fmt.Sprint(queryData)+",errInfo：%v", err1)
		utils.RecordError("删除 tt_app_account_cert失败，删除条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}
