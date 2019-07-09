package dal

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type AccountInfo struct {
	gorm.Model
	TeamId 			    string `gorm:"team_id"              form:"team_id"`
	IssueId 			string `gorm:"issue_id"             form:"issue_id"`
	KeyId 				string `gorm:"key_id"               form:"key_id"`
	AccountName 		string `gorm:"account_name"         form:"account_name"`
	AccountType 		string `gorm:"account_type"         form:"account_type"`
	AccountP8fileName   string `gorm:"account_p8file_name"  form:"account_p8file_name"`
	AccountP8file 		string `gorm:"account_p8file"`
	UserName 			string `gorm:"user_name"            form:"user_name"`
}

type RetValueWithP8 struct {
	TeamId 			    string      `json:"team_id"`
	AccountName 		string      `json:"account_name"`
	AccountType 		string      `json:"account_type"`
	UserName 			string      `json:"user_name"`
	AccountP8fileName   string      `json:"account_p8file_name"`
	AccountP8file 		string      `json:"account_p8file"`
	PermissionAction    []string    `json:"permission_action"`
}

type RetValueWithoutP8 struct {
	TeamId 			    string      `json:"team_id"`
	AccountName 		string      `json:"account_name"`
	AccountType 		string      `json:"account_type"`
	UserName 			string      `json:"user_name"`
	PermissionAction    []string    `json:"permission_action"`
}

type AccountExistRel struct {
	IsExisted   bool //标识是否带权限
	AccountInfo AccountInfo
	Permissions []string
}

type AccountQueryRet struct {
	Data []interface{} `json:"data"`
}

func (AccountInfo) TableName() string{
	return  "tt_apple_conn_account"
}
func DeleteAccountInfo(teamId string) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	if err=connection.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Where("team_id=?",teamId).Delete(&AccountInfo{}).Error;err!=nil{
		logs.Error("Delete Record Failed")
		return  false
	}
	return true
}

func InsertAccountInfo(accountInfo AccountInfo) bool {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return false
	}
	defer conn.Close()
	db:= conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Create(&accountInfo)
	utils.RecordError("Insert into DB Failed: ", db.Error)
	return true
}

func QueryAccountInfo(condition map[string]interface{} ) *[]AccountInfo {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var accountInfos []AccountInfo
	db:=conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Where(condition).Find(&accountInfos)
	utils.RecordError("Query from DB Failed: ", db.Error)
	return &accountInfos
}

func UpdateAccountInfo(accountInfo AccountInfo) bool {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return false
	}
	defer conn.Close()
	db:= conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Where("team_id=?",accountInfo.TeamId).Update(&accountInfo)
	utils.RecordError("Update DB Failed: ", db.Error)
	return true
}

