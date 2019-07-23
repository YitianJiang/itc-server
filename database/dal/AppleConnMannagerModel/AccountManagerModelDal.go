package devconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"strings"
)

func DeleteAccountInfo(teamId string) int {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return -1
	}
	defer connection.Close()
	var teamIds []TeamID
	db := connection.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Where("team_id=?", teamId).Find(&teamIds)
	utils.RecordError("query DB Failed: ", db.Error)
	if len(teamIds) == 0 {
		return -2
	}
	if err = connection.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Where("team_id=?", teamId).Delete(&AccountInfo{}).Error; err != nil {
		logs.Error("Delete Record Failed")
		return -3
	}
	return 0
}

func InsertAccountInfo(accountInfo AccountInfo) int {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return -1
	}
	defer conn.Close()
	var teamIds []TeamID
	db := conn.LogMode(_const.DB_LOG_MODE).
		Table(AccountInfo{}.TableName()).
		Where("team_id=?", accountInfo.TeamId).
		Find(&teamIds)
	utils.RecordError("query DB Failed: ", db.Error)
	if len(teamIds) != 0 {
		return -2
	}
	db = conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Create(&accountInfo)
	utils.RecordError("Insert into DB Failed: ", db.Error)
	return 0
}

func QueryAccountInfo(condition map[string]interface{}) *[]AccountInfo {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var accountInfos []AccountInfo
	db := conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Where(condition).Find(&accountInfos)
	utils.RecordError("Query from DB Failed: ", db.Error)
	return &accountInfos
}

func CheckAdmin(perms []string) bool {
	if len(perms) == 0 {
		return false
	}
	checkResult := false
	for _, perm := range perms {
		if perm == "admin" {
			checkResult = true
			break
		}
	}
	return checkResult
}

func QueryAccInfoWithAuth(resPerms *GetPermsResponse) *[]interface{} {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var accountsInfo []interface{}
	var teamIds []TeamID
	db := conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Find(&teamIds)
	utils.RecordError("Query from DB Failed: ", db.Error)
	for _, teamId := range teamIds {
		perms := resPerms.Data[strings.ToLower(teamId.TeamId)+"_space_account"]
		checkResult := CheckAdmin(perms)
		if !checkResult {
			accInfoWithoutAuth := AccInfoWithoutAuth{}
			db := conn.LogMode(_const.DB_LOG_MODE).
				Table(AccountInfo{}.TableName()).
				Where("team_id = ?", teamId.TeamId).
				Find(&accInfoWithoutAuth)
			utils.RecordError("Query from DB Failed: ", db.Error)
			accInfoWithoutAuth.PermissionAction = []string{}
			accountsInfo = append(accountsInfo, accInfoWithoutAuth)
		} else {
			accInfoWithAuth := AccInfoWithAuth{}
			db := conn.LogMode(_const.DB_LOG_MODE).
				Table(AccountInfo{}.TableName()).
				Where("team_id =?", teamId.TeamId).
				Find(&accInfoWithAuth)
			utils.RecordError("Query from DB Failed: ", db.Error)
			accInfoWithAuth.PermissionAction = perms
			accountsInfo = append(accountsInfo, accInfoWithAuth)
		}
	}
	return &accountsInfo
}

func UpdateAccountInfo(accountInfo AccountInfo) int {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return -1
	}
	defer conn.Close()
	var teamIds []TeamID
	db := conn.LogMode(_const.DB_LOG_MODE).
		Table(AccountInfo{}.TableName()).
		Where("team_id=?", accountInfo.TeamId).
		Find(&teamIds)
	utils.RecordError("query DB Failed: ", db.Error)
	if len(teamIds) == 0 {
		return -2
	}
	db = conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Where("team_id=?", accountInfo.TeamId).Update(&accountInfo)
	utils.RecordError("Update DB Failed: ", db.Error)
	return 0
}
