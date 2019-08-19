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
	if err = connection.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
		Where("team_id=?", teamId).Find(&teamIds).
		Error; err != nil {
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
	if err = conn.LogMode(_const.DB_LOG_MODE).
		Table(AccountInfo{}.TableName()).
		Where("team_id=?", accountInfo.TeamId).Find(&teamIds).
		Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return -1
	}
	if len(teamIds) != 0 {
		return -2
	}
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
		Create(&accountInfo).
		Error; err != nil {
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

func TransferValues2AccInfoWithoutAuth(accountInfo *AccountInfo, perms []string) *AccInfoWithoutAuth {
	accInfoWithoutAuth := AccInfoWithoutAuth{}
	accInfoWithoutAuth.AccountType = accountInfo.AccountType
	accInfoWithoutAuth.AccountName = accountInfo.AccountName
	accInfoWithoutAuth.UserName = accountInfo.UserName
	accInfoWithoutAuth.TeamId = accountInfo.TeamId
	accInfoWithoutAuth.PermissionAction = perms
	return &accInfoWithoutAuth
}

func TransferValues2AccInfoWithAuth(accountInfo *AccountInfo, perms []string) *AccInfoWithAuth {
	accInfoWithAuth := AccInfoWithAuth{}
	accInfoWithAuth.TeamId = accountInfo.TeamId
	accInfoWithAuth.UserName = accountInfo.UserName
	accInfoWithAuth.AccountName = accountInfo.AccountName
	accInfoWithAuth.AccountType = accountInfo.AccountType
	accInfoWithAuth.AccountP8file = accountInfo.AccountP8file
	accInfoWithAuth.AccountP8fileName = accountInfo.AccountP8fileName
	accInfoWithAuth.IssueId = accountInfo.IssueId
	accInfoWithAuth.KeyId = accountInfo.KeyId
	accInfoWithAuth.PermissionAction = perms
	return &accInfoWithAuth
}

func QueryAccInfoWithAuth(resPerms *GetPermsResponse) *[]interface{} {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	//todo null问题
	var accountsInfo []interface{}
	var allAccountInfo []AccountInfo
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Find(&allAccountInfo).Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return nil
	}
	for _, accountInfo := range allAccountInfo {
		perms := resPerms.Data[strings.ToLower(accountInfo.TeamId)+"_space_account"]
		checkResult := CheckAdmin(perms)
		if !checkResult {
			accInfoWithoutAuth := TransferValues2AccInfoWithoutAuth(&accountInfo, perms)
			accountsInfo = append(accountsInfo, accInfoWithoutAuth)
		} else {
			accInfoWithAuth := TransferValues2AccInfoWithAuth(&accountInfo, perms)
			accountsInfo = append(accountsInfo, accInfoWithAuth)
		}
	}
	return &accountsInfo
}

//获取权限信息MAP
func QueryAccInfoMapWithoutAuth(resPerms *GetPermsResponse) *map[string]AccInfoWithoutAuth {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var accountsInfo = make(map[string]AccInfoWithoutAuth)
	var allAccountInfo []AccountInfo
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Find(&allAccountInfo).Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return nil
	}
	for _, accountInfo := range allAccountInfo {
		perms := resPerms.Data[strings.ToLower(accountInfo.TeamId)+"_space_account"]
		accInfoWithoutAuth := TransferValues2AccInfoWithoutAuth(&accountInfo, perms)
		accountsInfo[accInfoWithoutAuth.TeamId] = *accInfoWithoutAuth
	}
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
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
		Where("team_id=?", accountInfo.TeamId).Find(&teamIds).
		Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return -1
	}
	if len(teamIds) == 0 {
		return -2
	}
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).
		Where("team_id=?", accountInfo.TeamId).Update(&accountInfo).
		Error; err != nil {
		logs.Error("Update DB Failed:", err)
		return -1
	}
	return 0
}
