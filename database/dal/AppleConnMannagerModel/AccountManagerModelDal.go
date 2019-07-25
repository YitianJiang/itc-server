package devconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
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
	if err= connection.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
		Where("team_id=?", teamId).Find(&teamIds).
		Error;err!=nil{
		logs.Error("Query DB Failed:", err)
		return -1
	}
	if len(teamIds) == 0 {
		return -2
	}
	if err = connection.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
		Where("team_id=?", teamId).Delete(&AccountInfo{}).
		Error; err != nil {
		logs.Error("Delete Record Failed")
		return -1
	}
	return 0
}

func InsertAccountInfo(accountInfo AccountInfo) int {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return -1
	}
	defer conn.Close()
	var teamIds []TeamID
	if err= conn.LogMode(_const.DB_LOG_MODE).
		Table(AccountInfo{}.TableName()).
		Where("team_id=?", accountInfo.TeamId).Find(&teamIds).
		Error;err!=nil{
		logs.Error("Query DB Failed:", err)
		return -1
	}
	if len(teamIds) != 0 {
		return -2
	}
	if err=conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
		Create(&accountInfo).
		Error;err!=nil{
		logs.Error("Insert DB Failed:", err)
		return -1
	}
	return 0
}

func QueryAccountInfo(condition map[string]interface{}) *[]AccountInfo {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var accountInfos []AccountInfo
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
		Where(condition).Find(&accountInfos).
		Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return nil
	}
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
		logs.Error("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var accountsInfo []interface{}
	var teamIds []TeamID
	if err= conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Find(&teamIds).Error;err!=nil{
		logs.Error("Query DB Failed:", err)
		return nil
	}
	conn.Begin()
	for _, teamId := range teamIds {
		perms := resPerms.Data[strings.ToLower(teamId.TeamId)+"_space_account"]
		checkResult := CheckAdmin(perms)
		if !checkResult {
			accInfoWithoutAuth := AccInfoWithoutAuth{}
			if err= conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
				Where("team_id = ?", teamId.TeamId).Find(&accInfoWithoutAuth).
				Error;err!=nil{
				logs.Error("Query DB Failed:", err)
				conn.Rollback()
				return nil
			}
			accInfoWithoutAuth.PermissionAction = []string{}
			accountsInfo = append(accountsInfo, accInfoWithoutAuth)
		} else {
			accInfoWithAuth := AccInfoWithAuth{}
			if err= conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
				Where("team_id =?", teamId.TeamId).Find(&accInfoWithAuth).
				Error;err!=nil{
				logs.Error("Query DB Failed:", err)
				conn.Rollback()
				return nil
			}
			accInfoWithAuth.PermissionAction = perms
			accountsInfo = append(accountsInfo, accInfoWithAuth)
		}
	}
	conn.Commit()
	return &accountsInfo
}

func UpdateAccountInfo(accountInfo AccountInfo) int {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return -1
	}
	defer conn.Close()
	var teamIds []TeamID
	if err= conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
		Where("team_id=?", accountInfo.TeamId).Find(&teamIds).
		Error;err!=nil{
		logs.Error("Query DB Failed:", err)
		return -1
	}
	if len(teamIds) == 0 {
		return -2
	}
	if err= conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
		Where("team_id=?", accountInfo.TeamId).Update(&accountInfo).
		Error;err!=nil{
		logs.Error("Update DB Failed:", err)
		return -1
	}
	return 0
}
