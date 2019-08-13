package devconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"fmt"
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
		utils.RecordError("删除 tt_app_account_cert失败，删除条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}

/**
更新操作
*/
func UpdateAppAccountCert(queryData, item map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppAccountCert{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err1 := db.Where(queryData).Update(item).Error; err1 != nil {
		utils.RecordError("更新 tt_app_account_cert失败，更新条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}

/**
更新操作
返回记录model
*/
func UpdateAppAccountCertAndGetModelByMap(condition map[string]interface{}, updateInfo map[string]interface{}) (error, *AppAccountCert) {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return err, nil
	}
	defer conn.Close()
	var appAccountCert AppAccountCert
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).Unscoped().Where(condition).Update(updateInfo).Find(&appAccountCert).Error; err != nil {
		logs.Error("Update DB Failed:", err)
		return err, nil
	}
	return nil, &appAccountCert
}

/**
更新操作
不返回model
*/
func UpdateAppAccountCertByMap(condition map[string]interface{}, updateInfo map[string]interface{}) bool {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return false
	}
	defer conn.Close()
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(AppAccountCert{}.TableName()).Unscoped().Where(condition).Update(updateInfo).Error; err != nil {
		logs.Error("Update DB Failed:", err)
		return false
	}
	return true
}

/**
查询操作，
返回nil---查询fail，返回空数组--无相关数据
*/
func QueryAppBundleProfiles(queryData map[string]interface{}) *[]AppBundleProfiles {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	db := conn.Table(AppBundleProfiles{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result = make([]AppBundleProfiles, 0)
	if err := db.Where(queryData).Find(&result).Error; err != nil {
		utils.RecordError("查询 tt_app_bundleId_profiles失败，查询条件："+fmt.Sprint(queryData)+",errInfo：", err)
		return nil
	}
	return &result
}

/**
删除操作
*/
func DeleteAppBundleProfiles(queryData map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppBundleProfiles{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var t AppBundleProfiles
	if err1 := db.Where(queryData).Delete(&t).Error; err1 != nil {
		utils.RecordError("删除 tt_app_bundleId_profiles失败，删除条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}

func UpdateAppBundleProfiles(queryData, item map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppBundleProfiles{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err1 := db.Where(queryData).Update(item).Error; err1 != nil {
		utils.RecordError("更新 tt_app_bundleId_profiles失败，条件："+fmt.Sprint(queryData)+",errInfo：", err1)
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
		utils.RecordError("查询 tt_apple_bundleId 失败，查询条件："+fmt.Sprint(queryData)+",errInfo：", err)
		return nil
	}
	return &result
}

/**
删除操作
*/
func DeleteAppleBundleId(queryData map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppleBundleId{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var t AppleBundleId
	if err1 := db.Where(queryData).Delete(&t).Error; err1 != nil {
		utils.RecordError("删除 tt_apple_bundleId失败，删除条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}

func UpdateAppleBundleId(queryData map[string]interface{}, updateData map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppleBundleId{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err1 := db.Where(queryData).Update(updateData).Error; err1 != nil {
		utils.RecordError("更新 tt_apple_bundleid失败，更新条件："+fmt.Sprint(queryData)+",errInfo：", err1)
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
		utils.RecordError("查询 tt_apple_profile失败，查询条件："+fmt.Sprint(queryData)+",errInfo：", err)
		return nil
	}
	return &result
}
/**
查询操作，多值查询
返回nil---查询fail，返回空数组--无相关数据
*/
func QueryAppleProfileWithList(col string,datas *[]interface{}) *[]AppleProfile {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	db := conn.Table(AppleProfile{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result = make([]AppleProfile, 0)
	if err := db.Where(col+" in (?) and deleted_at IS NULL",*datas).Find(&result).Error; err != nil {
		utils.RecordError("查询 tt_apple_profile失败，查询条件："+fmt.Sprint(col)+fmt.Sprint((*datas)[0])+",errInfo：", err)
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
		utils.RecordError("删除 tt_apple_profile失败，删除条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}

/**
更新操作
*/
func UpdateAppleProfile(queryData map[string]interface{}, updateData map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppleProfile{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err1 := db.Where(queryData).Update(updateData).Error; err1 != nil {
		utils.RecordError("更新 tt_apple_profile失败，更新条件："+fmt.Sprint(queryData)+",errInfo：", err1)
		return err1
	}
	return nil
}
func UpdateAppleProfileBatch(col string,rangeData *[]interface{},updateData map[string]interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	db := conn.Table(AppleProfile{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err1 := db.Where(col +" in (?)",*rangeData).Update(updateData).Error; err1 != nil {
		utils.RecordError("更新 tt_apple_profile失败，更新条件："+fmt.Sprint(col)+",errInfo：", err1)
		return err1
	}
	return nil
}

/**
根据app_id联查，获取cert_section
*/
func QueryWithSql(sql string, result interface{}) error {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return err
	}
	defer conn.Close()
	if err := conn.LogMode(_const.DB_LOG_MODE).Raw(sql).Scan(result).Error; err != nil {
		utils.RecordError("query with sql failed,", err)
		return err
	}
	return nil
}
