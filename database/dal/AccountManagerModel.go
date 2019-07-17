package dal

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"strings"
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

type AccInfoWithAuth struct {
	TeamId 			    string      `gorm:"team_id"`
	AccountName 		string      `gorm:"account_name"`
	AccountType 		string      `gorm:"account_type"`
	AccountP8fileName   string      `gorm:"account_p8file_name"`
	AccountP8file 		string      `gorm:"account_p8file"`
	UserName 			string      `gorm:"user_name"`
	PermissionAction   []string
}

type AccInfoWithoutAuth struct {
	TeamId 			    string      `gorm:"team_id"`
	AccountName 		string      `gorm:"account_name"`
	AccountType 		string      `gorm:"account_type"`
	UserName 			string      `gorm:"user_name"`
	PermissionAction   []string
}

type DelAccRequest struct {
	TeamId string `json:"team_id"`
}

type CreateResourceRequest struct {
	ResourceName    string      `json:"resourceName"`
	ResourceKey     string      `json:"resourceKey"`
	CreatorKey      string      `json:"creatorKey"`
	ResourceType    int         `json:"resourceType"`
}

type CreResResponse struct {
	Errno   int      `json:"errno"`
}

type TeamID struct {
	gorm.Model
	TeamId string `gorm:"team_id"`
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

func InsertAccountInfo(accountInfo AccountInfo) int {
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return -1
	}
	defer conn.Close()
	var teamIds []TeamID
	db:= conn.LogMode(_const.DB_LOG_MODE).
		Table(AccountInfo{}.TableName()).
		Select("team_id").
		Where("team_id=?",accountInfo.TeamId).
		Find(&teamIds)
	utils.RecordError("query DB Failed: ", db.Error)
	if len(teamIds)!=0{
		return -2
	}
	db= conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Create(&accountInfo)
	utils.RecordError("Insert into DB Failed: ", db.Error)
	return 0
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

func QueryAccInfoWithAuth(resPerms *GetPermsResponse) *[]interface{}{
	conn, err := database.GetConneection()
	if err != nil {
		utils.RecordError("Get DB Connection Failed: ", err)
		return nil
	}
	defer conn.Close()
	var accountsInfo []interface{}
	var teamIds []TeamID
	db:=conn.LogMode(_const.DB_LOG_MODE).Table(AccountInfo{}.TableName()).Select("team_id").Find(&teamIds)
	utils.RecordError("Query from DB Failed: ", db.Error)
	for _,teamId:=range teamIds{
		perms:=resPerms.Data[strings.ToLower(teamId.TeamId)+"_space_account"]
		if len(perms)==0{
			accInfoWithoutAuth:=AccInfoWithoutAuth{}
			db:=conn.LogMode(_const.DB_LOG_MODE).
				Table(AccountInfo{}.TableName()).
				Where("team_id = ?",teamId.TeamId).
				Find(&accInfoWithoutAuth)
			utils.RecordError("Query from DB Failed: ", db.Error)
			accInfoWithoutAuth.PermissionAction=[]string{}
			accountsInfo=append(accountsInfo, accInfoWithoutAuth)
		}else{
			accInfoWithAuth:=AccInfoWithAuth{}
			db:=conn.LogMode(_const.DB_LOG_MODE).
				Table(AccountInfo{}.TableName()).
				Where("team_id =?",teamId.TeamId).
				Find(&accInfoWithAuth)
			utils.RecordError("Query from DB Failed: ", db.Error)
			accInfoWithAuth.PermissionAction=perms
			accountsInfo=append(accountsInfo, accInfoWithAuth)
		}
	}
	return &accountsInfo
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
