package dal

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/logs"
)

type Account struct {
	Id int
	Team_id int `gorm:"team_id"`
	Issue_id int `gorm:"issue_id"`
	Key_id int `gorm:"key_id"`
	Account_name string `gorm:"account_name"`
	Account_type string `gorm:"account_type"`
	Account_p8file_name string `gorm:"account_p8file_name"`
	Account_p8file string `gorm:"account_p8file"`
	User_name string `gorm:"user_name"`

}

func DeleteAccount(teamId string) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	if err=connection.LogMode(_const.DB_LOG_MODE).Table("tt_account").Where("team_id=?",teamId).Delete(&Account{}).Error;err!=nil{
		logs.Error("从数据库中删除数据失败")
		return  false
	}
	return true
}

//func InsertAccountInfo(account Account) error {
//	conn, err := database.GetConneection()
//	if err != nil {
//		utils.RecordError("Get DB Connection Failed: ", err)
//		return err
//	}
//	defer conn.Close()
//	db:= conn.LogMode(_const.)
//	utils.RecordError("Insert into DB Failed: ", db.Error)
//	return db.Error
//}