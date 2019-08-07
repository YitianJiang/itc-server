package dal

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"time"
)

/*aar检测相关数据库表操作和结构
 */

/**
aar检测任务表结构
*/
type OtherDetectModel struct {
	gorm.Model
	Creator      string  `json:"creator"`
	ToLarker     string  `json:"toLarker"`
	ToGroup      string  `json:"toGroup"`
	Platform     int     `json:"platform"`
	OtherName    string  `json:"name"`
	OtherVersion string  `json:"version"`
	AppId        string  `json:"appId"`
	Status       int     `json:"status"`    //0---未完全确认；1---已完全确认
	ExtraInfo    string  `json:"extraInfo"` //其他附加信息,text信息
	FileType     string  `json:"fileType"`  //目前只有"aar"
	ErrInfo      *string `json:"errInfo"`   //检测任务错误信息收集
}

/**
aar检测结果信息表结构
*/
type OtherDetailInfoStruct struct {
	gorm.Model
	TaskId      int    `json:"taskId"`
	ToolId      int    `json:"toolId"`
	DetailType  int    `json:"detailType"` //0--敏感方法，1--敏感字符串，2--权限，3--missInfo,4--基本信息
	DetectInfos string `json:"detectInfos"`
	SubIndex    int    `json:"index"`
}

/**
aar基本信息结构
*/
type OtherBasicInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Channel string `json:"channel"`
}

/**
组件平台结果展示结构
*/
type DetectQueryStructAarPlatform struct {
	Index         int                      `json:"index"`
	ApkName       string                   `json:"aarName"`
	Version       string                   `json:"version"`
	Channel       string                   `json:"channel"`
	SMethods      []SMethodAarPlatform     `json:"sMethods"`
	SStrs_new     []SStrAarPlatform        `json:"sStrs"`
	Permissions_2 []PermissionsAarPlatform `json:"permissionList"`
}
type SMethodAarPlatform struct {
	Id         uint             `json:"id"`
	MethodName string           `json:"methodName"`
	ClassName  string           `json:"className"`
	Desc       string           `json:"desc"`
	CallLoc    []MethodCallJson `json:"callLoc"`
	RiskLevel  string           `json:"risk_level"`
	GPFlag     int              `json:"gpFlag"`
}
type SStrAarPlatform struct {
	Id      uint          `json:"id"`
	Keys    string        `json:"keys"`
	Desc    string        `json:"desc"`
	CallLoc []StrCallJson `json:"callLoc"`
	GPFlag  int           `json:"gpFlag"`
}
type PermissionsAarPlatform struct {
	Id       uint   `json:"id"`
	Key      string `json:"key"`
	Priority int    `json:"priority"`
	Desc     string `json:"desc"`
}

func (OtherDetectModel) TableName() string {
	return "tb_other_detect"
}

func (OtherDetailInfoStruct) TableName() string {
	return "tb_other_detect_detail"
}

/**
新增其他检测任务
*/
func InsertOtherDetect(otherInfo OtherDetectModel) int {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v", err)
		return 0
	}
	defer connection.Close()
	otherInfo.CreatedAt = time.Now()
	otherInfo.UpdatedAt = time.Now()
	db := connection.Table(OtherDetectModel{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Create(&otherInfo).Error; err != nil {
		logs.Error("insert otherDetect_task failed,%v", err)
		return 0
	}
	return int(otherInfo.ID)
}

/**
查询aar检测任务
*/
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

/**
查询检测任务列表
*/
func QueryOtherDetctModelOfList(pageInfo map[string]int, condition string) (*[]OtherDetectModel, int, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil, 0, err
	}
	defer connection.Close()
	db := connection.Table(OtherDetectModel{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var condition1 = "deleted_at IS NULL and " + condition
	var total TotalStruct
	if err := db.Select("count(id) as total").Where(condition1).Find(&total).Error; err != nil {
		logs.Error("query aar task counts failed,%v", err)
		return nil, 0, err
	}
	var detect []OtherDetectModel

	if err := db.Where(condition).Offset((pageInfo["page"] - 1) * pageInfo["pageSize"]).Limit(pageInfo["pageSize"]).Order("id DESC").Find(&detect).Error; err != nil {
		logs.Error("query aar detect List failed,%v", err)
		return nil, 0, err
	}
	return &detect, total.Total, nil
}

/**
更新检测任务信息
*/
func UpdateOtherDetectModelByMap(condition string, param map[string]interface{}) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Table(OtherDetectModel{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Update(param).Error; err != nil {
		logs.Error("update aar detect task failed, taskId :"+condition+"errInfo :%v", err)
		return err
	}
	return nil
}

/**
批量插入aar检测信息到数据库
*/
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

/**
查询aar检测结果详情
*/
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

/**
更新aar检测结果
*/
func UpdateOtherDetailInfoByMap(condition string, param map[string]interface{}) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Table(OtherDetailInfoStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Update(param).Error; err != nil {
		logs.Error("update aar detect detail failed, taskId :"+condition+"errInfo :%v", err)
		return err
	}
	return nil
}

/**
批量更新aar旧报告内容
*/
func UpdateOldAArMethods(ids *[]string, infos *[]string) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Table(OtherDetailInfoStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	db.Begin()
	for i := 0; i < len(*ids); i++ {
		condition := "id=" + (*ids)[i]
		if err := db.Where(condition).Update(map[string]interface{}{
			"detect_infos": (*infos)[i],
		}).Error; err != nil {
			db.Rollback()
			return err
		}
	}
	db.Commit()
	return nil
}
