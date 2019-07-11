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
	TeamId 			    string      `gorm:"team_id"              form:"team_id"                 json:"team_id"`
	IssueId 			string      `gorm:"issue_id"             form:"issue_id"                json:"issue_id,omitempty"`
	KeyId 				string      `gorm:"key_id"               form:"key_id"                  json:"key_id,omitempty"`
	AccountName 		string      `gorm:"account_name"         form:"account_name"            json:"account_name"`
	AccountType 		string      `gorm:"account_type"         form:"account_type"            json:"account_type"`
	AccountP8fileName   string      `gorm:"account_p8file_name"  form:"account_p8file_name"     json:"account_p8file_name,omitempty"`
	AccountP8file 		string      `gorm:"account_p8file"                                      json:"account_p8file,omitempty"`
	UserName 			string      `gorm:"user_name"            form:"user_name"               json:"user_name"`
	PermissionAction   []string     `gorm:"-"                                                   json:"permission_action"`
}
type DelAccRequest struct {
	TeamId string `json:"team_id"`
}

//创建资源请求
type CreResRequest struct {
	ResourceName    string      `json:"resourceName"`
	ResourceKey     string      `json:"resourceKey"`
	CreatorKey      string      `json:"creatorKey"`
	ResourceType    int         `json:"resourceType"`
}

type CreResResponse struct {
	Errno   int      `json:"errno"`
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

