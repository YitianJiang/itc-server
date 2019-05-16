package dal

import (
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"time"
	"strconv"
	"strings"
)

/*
test_rej_cases表结构
 */
type rejCase struct {
	gorm.Model
	AppId  			int 		`json:"appId"`
	AppName 		string 		`json:"appName"`
	RejTime 		time.Time 	`json:"rejTime"`
	RejReason 		string 		`json:"rejRea"`
	Solution 		string 		`json:"solution"`
	PicLoc 			string 		`json:"picLoc"`
	Version 		string		`json:"version"`
}

type totalStruct struct {
	Total 			int 		`json:"total"`
}
/*
struct used when show rejCases
*/
type RejListInfo struct {
	Id 				int				`json:"id"`
	AppId 			int				`json:"appId"`
	AppName 		string 			`json:"appName"`
	RejTime 		time.Time		`json:"rejTime"`
	RejRea			string 			`json:"rejRea"`
	Solution 		string 			`json:"solution"`
	PicLoc 			[]PicInfo 		`json:"picLoc"`
	Version 		string			`json:"version"`
}
type PicInfo struct {
	PicName			string 		`json:"picName"`
	PicUrl			string		`json:"picUrl"`
}

/*
struct used when add a rejCase
*/
type RejInfo struct {
	AppId 			int 		`json:"appId"`
	AppName 		string 		`json:"appName"`
	RejTime 		time.Time 	`json:"rejTime"`
	RejRea 			string 		`json:"rejRea"`
	Solution 		string 		`json:"solution"`
	Version 		string		`json:"version"`
}

func (rejCase) TableName() string {
	return "tb_rej_cases"
}


/*
	query rejCases meeting with conditions in the db，without condition query all rejCases
*/

func QueryByConditions(param map[string]string) (*[]RejListInfo,int,error){
	connection,err :=database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed:%v",err)
		return nil,0,err
	}
	defer connection.Close()
	db := connection.Table(rejCase{}.TableName()).LogMode(_const.DB_LOG_MODE)
	condition := param["condition"]
	logs.Info("query rejCases by Conditions:%s",condition)
	if condition == "" {
		condition = " 1=1 "
	}

	page,err := strconv.Atoi(param["page"])
	pageSize,err := strconv.Atoi(param["pageSize"])
	var infos []rejCase
	logs.Info("after param:%v",param)
	if err := db.Where(condition).Offset(((page - 1)*pageSize)).Limit(pageSize).Order("ID DESC").Find(&infos).Error; err != nil{
		logs.Error("%v", err)
		return nil,0,err
	}
	var result []RejListInfo
	result = make([]RejListInfo,0)
	for _,item := range infos{
		var rejInfo RejListInfo
		rejInfo.Id = int(item.ID)
		rejInfo.AppId = item.AppId
		rejInfo.AppName = item.AppName
		rejInfo.RejRea = item.RejReason
		rejInfo.RejTime = item.RejTime
		rejInfo.Version = item.Version
		rejInfo.PicLoc = picLocTrans(item.PicLoc)
		rejInfo.Solution = item.Solution
		result=append(result, rejInfo)
	}

	var total totalStruct
	connect,err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil,0,err
	}
	defer connect.Close()
	dbCount := connect.Table(rejCase{}.TableName()).LogMode(_const.DB_LOG_MODE)
	logs.Info("query rejCases counts by Conditions:%s",condition)

	countCondition := "deleted_at IS NULL and "+condition

	if err = dbCount.Select("count(id) as total").Where(countCondition).Find(&total).Error; err != nil {
		logs.Error("query total record failed! %v", err)
		return &result, 0, err
	}
	return &result,total.Total,nil
}

/*
	add rejCases
*/
func InsertRejCase(data map[string]interface{}) error {
	connection,err :=database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed:%v",err)
		return err
	}
	defer connection.Close()
	var rejC rejCase
	rejI := data["info"]
	v,ok := rejI.(RejInfo)
	if ok {
		rejC.AppId = v.AppId
		rejC.AppName = v.AppName
		rejC.Solution = v.Solution
		rejC.RejTime = v.RejTime
		rejC.RejReason = v.RejRea
		rejC.Version = v.Version
	}
	s := data["picPath"]
	loc,ok := s.(string)
	if ok {
		rejC.PicLoc = loc
	}
	rejC.CreatedAt = time.Now()

	if err = connection.Table(rejCase{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&rejC).Error; err != nil {
		logs.Error("update self check item failed, %v", err)
		return err
	}
	return nil
}

/*
	logical delete rejCase in the database
 */
func DeleteCase(id int) error  {
	connection,err :=database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed:%v",err)
		return err
	}
	db := connection.Table(rejCase{}.TableName())
	defer connection.Close()
	if err := db.Where("id = ?", id).LogMode(_const.DB_LOG_MODE).Delete(&rejCase{}).Error; err != nil{
		logs.Error("%v", err)
		return err
	}
	return nil

}

func UpdateRejCaseofSolution(data map[string]string) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed:%v", err)
		return err
	}
	db := connection.Table(rejCase{}.TableName())
	defer connection.Close()
	condition := data["condition"]
	solution := data["solution"]
	err1 := db.Where(condition).Update(map[string]interface{}{
		"updated_at": time.Now(),
		"solution":  solution}).Error;
	if err1 != nil {
		logs.Error("update rejCase failes")
		return err1
	}
	return nil
}

func picLocTrans(path string) []PicInfo{
	var result = make([]PicInfo,0)
	if path ==""{
		return result
	}
	paths := strings.Split(path,";")
	for _,subpath := range paths {
		if subpath != ""{
			values := strings.Split(subpath,"--")
			var pic PicInfo
			pic.PicName = values[0]
			pic.PicUrl = values[1]
			result = append(result, pic)
		}
	}
	return result

}