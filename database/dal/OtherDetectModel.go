package dal

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"time"
)

type OtherDetectModel struct {
	gorm.Model
	Creator         string 										`json:"creator"`
	ToLarker        string 										`json:"toLarker"`
	ToGroup         string 										`json:"toGroup"`
	Platform        int    										`json:"platform"`
	OtherName       string 										`json:"name"`
	OtherVersion    string 										`json:"version"`
	AppId           string 										`json:"appId"`
	Status          int    										`json:"status"` //0---未完全确认；1---已完全确认
	ExtraInfo		string 										`json:"extraInfo"`//其他附加信息,text信息
	FileType        string										`json:"fileType"`//目前只有"aar"
}

type OtherDetailInfoStruct struct {
	gorm.Model
	TaskId				int					`json:"taskId"`
	ToolId				int					`json:"toolId"`
	DetailType 			int					`json:"detailType"`//0--敏感方法，1--敏感字符串，2--权限，3--missInfo,4--基本信息
	DetectInfos			string				`json:"detectInfos"`
	SubIndex   			int					`json:"index"`
}
type OtherBasicInfo struct {
	Name   		string 			`json:"name"`
	Version 	string			`json:"version"`
	Channel 	string			`json:"channel"`
}

func (OtherDetectModel) TableName()  string{
	return "tb_other_detect"
}

func (OtherDetailInfoStruct) TableName() string  {
	return "tb_other_detect_detail"
}


/**
	新增其他检测任务
*/
func InsertOtherDetect(otherInfo OtherDetectModel) int  {
	connection,err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v",err)
		return 0
	}
	defer connection.Close()
	otherInfo.CreatedAt = time.Now()
	otherInfo.UpdatedAt = time.Now()
	db := connection.Table(OtherDetectModel{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Create(&otherInfo).Error; err != nil {
		logs.Error("insert otherDetect_task failed,%v",err)
		return 0
	}
	return int(otherInfo.ID)
}


func QueryOtherDetectModelsByMap(param map[string]interface{}) *[]OtherDetectModel {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var detect []OtherDetectModel
	db := connection.Table(OtherDetectModel{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(param).Find(&detect).Error; err != nil {
		logs.Error("查询aar检测任务信息失败，原因：%v", err)
		return nil
	}
	return &detect
}

func UpdateOtherDetectModelByMap(condition string, param map[string]interface{}) error{
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Table(OtherDetectModel{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Update(param).Error;err != nil {
		logs.Error("update aar detect task failed, taskId :"+condition+"errInfo :%v",err)
		return err
	}
	return nil
}

func InsertOtherDetectDetailBatch(details *[]OtherDetailInfoStruct) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	db := connection.Begin()

	for _, detail := range *details {
		detail.CreatedAt = time.Now()
		detail.UpdatedAt = time.Now()
		if err1 := db.Table(OtherDetailInfoStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&detail).Error; err1 != nil {
			logs.Error("数据库新增aar信息失败,%v，具体任务参数：%s", err1, detail.TaskId)
			db.Rollback()
			return err1
		}
	}
	db.Commit()
	return nil
}

func QueryOtherDetectDetail(param map[string]interface{}) *[]OtherDetailInfoStruct {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var detect []OtherDetailInfoStruct
	db := connection.Table(OtherDetailInfoStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(param).Find(&detect).Error; err != nil {
		logs.Error("查询aar检测结果信息失败，原因：%v", err)
		return nil
	}
	return &detect
}

func UpdateOtherDetailInfoByMap(condition string, param map[string]interface{}) error{
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Table(OtherDetailInfoStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Update(param).Error;err != nil {
		logs.Error("update aar detect detail failed, taskId :"+condition+"errInfo :%v",err)
		return err
	}
	return nil
}




