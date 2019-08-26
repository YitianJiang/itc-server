package dal

import (
	"fmt"
	"time"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

/**
权限配置数据库表结构
*/
type DetectConfigStruct struct {
	gorm.Model
	KeyInfo      string `json:"key"`
	CheckType    int    `json:"type"`     //0--权限，1--api,2--action,3--其他
	Priority     int    `json:"priority"` //-1--无危险，0--一般，1--低危，2--中危，3--高危
	Ability      string `json:"ability"`
	DescInfo     string `json:"desc"`
	DetectType   string `json:"detectType"`
	Suggestion   string `json:"suggestion"`
	AbilityGroup string `json:"group"`
	Reference    string `json:"refer"`
	Creator      string `json:"creator"`
	Platform     int    `json:"platform"`  //0--安卓，1--iOS
	GpFlag       int    `json:"gpFlag"`    //0--非GP，1--GP
	SensiFlag    int    `json:"sensiFlag"` //0--非隐私，1--隐私
}

/**
权限配置列表返回数据结构
*/
type DetectConfigListInfo struct {
	Id        uint      `json:"id"`
	KeyInfo   string    `json:"key"`
	CreatedAt time.Time `json:"createTime"`
	Priority  int       `json:"priority"` //-1--无危险，0--一般，1--低危，2--中危，3--高危
	Ability   string    `json:"ability"`
	DescInfo  string    `json:"desc"`
	Creator   string    `json:"creator"`
	Platform  int       `json:"platform"` //0--安卓，1--iOS
	Type      int       `json:"type"`
	GpFlag    int       `json:"gpFlag"`    //0--非GP，1--GP
	SensiFlag int       `json:"sensiFlag"` //0--非隐私，1--隐私
}

/**
新增/修改权限数据结构
*/
type DetectConfigInfo struct {
	ID           int         `json:"id"`
	KeyInfo      string      `json:"key"`
	Type         interface{} `json:"type"`
	Priority     interface{} `json:"priority"`
	Ability      string      `json:"ability"`
	DescInfo     string      `json:"desc"`
	DetectType   string      `json:"detectType"`
	Suggestion   string      `json:"suggestion"`
	AbilityGroup string      `json:"abilityGroup"`
	Reference    string      `json:"refer"`
	Platform     interface{} `json:"platform"` //0--安卓，1--iOS
	GpFlag       interface{} `json:"gpFlag"`
	SensiFlag    interface{} `json:"sensiFlag"`
}

/**
权限操作历史数据库表结构
*/
type PermHistory struct {
	gorm.Model
	PermId     int    `json:"permId"`
	Status     int    `json:"status"`
	Remarks    string `json:"remarks"`
	Confirmer  string `json:"confirmer"`
	AppId      int    `json:"appId"`
	AppVersion string `json:"appVersion"`
	TaskId     int    `json:"taskId"`
}

/**
权限和app、task关系数据库表结构
*/
type PermAppRelation struct {
	gorm.Model
	TaskId     int    `json:"taskId"`
	AppId      int    `json:"appId"`
	AppVersion string `json:"appVersion"`
	PermInfos  string `json:"permInfos"`
	SubIndex   int    `json:"index"`
}

/**
根据权限查询返回数据结构
*/
type QueryInfoWithPerm struct {
	AppId      int    `json:"appId"`
	AppVersion string `json:"appVersion"`
	PermId     int    `json:"permId"`
	Ability    string `json:"ability"`
	Priority   int    `json:"priority"`
	KeyInfo    string `json:"key"`
}

type TotalStruct struct {
	Total int `json:"total"`
}

func (DetectConfigStruct) TableName() string {
	return "tb_detect_config"
}

func (PermHistory) TableName() string {
	return "tb_perm_history"
}

func (PermAppRelation) TableName() string {
	return "tb_perm_app_relation"
}

/**
新增权限
*/
func InsertDetectConfig(data DetectConfigStruct) (uint, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db err,%v", err)
		return 0, err
	}
	defer connection.Close()
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	db := connection.Table(DetectConfigStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Create(&data).Error; err != nil {
		logs.Error("insert detectConfig err,%v", err)
		return 0, err
	}
	return data.ID, nil
}

/**
查询权限信息
*/
func QueryDetectConfig(queryData map[string]interface{}) *[]DetectConfigStruct {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db error,%v", err)
		return nil
	}
	defer connection.Close()
	db := connection.Table(DetectConfigStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result []DetectConfigStruct

	if err := db.Where(queryData).Find(&result).Error; err != nil {
		logs.Error("query detectConfig failed,%v", err)
		return nil
	}
	return &result
}

/**
查询权限列表信息
*/
func QueryDetectConfigList(condition string, pageInfo map[string]int) (*[]DetectConfigStruct, int, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db error,%v", err)
		return nil, 0, err
	}
	defer connection.Close()
	db := connection.Table(DetectConfigStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	//db.Begin()
	var condition1 = "deleted_at IS NULL and " + condition
	var total TotalStruct
	if err := db.Select("count(id) as total").Where(condition1).Find(&total).Error; err != nil {
		logs.Error("query detectConfig counts failed,%v", err)
		//db.Rollback()
		return nil, 0, err
	}
	var result []DetectConfigStruct
	pageSize := pageInfo["pageSize"]
	page := pageInfo["page"]
	if err := db.Where(condition).Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&result).Error; err != nil {
		logs.Error("query detectConfig failed,%v", err)
		//db.Rollback()
		return nil, total.Total, err
	}
	//db.Commit()
	return &result, total.Total, nil
}

/**
根据权限信息模糊联查使用该权限的app情况
*/
func QueryDetectConfigWithSql(sql string) (*[]QueryInfoWithPerm, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db error,%v", err)
		return nil, err
	}
	defer connection.Close()

	var result = make([]QueryInfoWithPerm, 0)
	if err := connection.LogMode(_const.DB_LOG_MODE).Raw(sql).Scan(&result).Error; err != nil {
		logs.Error("query detectConfig failed,%v", err)
		return nil, err
	}

	return &result, nil
}

/**
修改权限信息
*/
func UpdataDetectConfig(data map[string]interface{}) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed, %v", err)
		return err
	}
	defer connection.Close()
	condition := data["condition"]
	info := data["update"] //应该是map[string]interface{}形式
	db := connection.Table(DetectConfigStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Update(info).Error; err != nil {
		logs.Error("update detectConfig failed,info: %v;errInfo:%v", data, err)
		return err
	}
	return nil
}

/**
删除权限信息----暂不对外提供接口
*/
func DeleteDetectConfig(condition string) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v", err)
		return err
	}
	defer connection.Close()
	db := connection.Table(DetectConfigStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var t DetectConfigInfo
	if err := db.Where(condition).Delete(&t).Error; err != nil {
		logs.Error("delete detectConfig failed,info:%v;errorInfo:%v", condition, err)
		return err
	}
	return nil
}

/**
新增权限app对应关系
*/
func InsertPermAppRelation(relation PermAppRelation) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v", err)
		return err
	}
	defer connection.Close()
	relation.CreatedAt = time.Now()
	relation.UpdatedAt = time.Now()
	db := connection.Table(PermAppRelation{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Create(&relation).Error; err != nil {
		logs.Error("insert permission-app relationship failed,%v", err)
		return err
	}
	return nil
}

/**
更新权限app对应关系
*/
func UpdataPermAppRelation(data *PermAppRelation) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v", err)
		return err
	}
	defer connection.Close()
	data.UpdatedAt = time.Now()
	id := data.ID
	condition := "id= '" + fmt.Sprint(id) + "'"
	db := connection.Table(PermAppRelation{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Update(&data).Error; err != nil {
		logs.Error("update permission-app relationship failed,%v", err)
		return err
	}
	return nil
}

/**
查询权限app关系
*/
func QueryPermAppRelation(data map[string]interface{}) (*[]PermAppRelation, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v", err)
		return nil, err
	}
	defer connection.Close()
	db := connection.Table(PermAppRelation{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result []PermAppRelation
	if err := db.Where(data).Find(&result).Error; err != nil {
		logs.Error("query permission-app relationship failed,%v", err)
		return nil, err
	}
	return &result, nil
}

/**
查询权限信息withGroup
*/
func QueryPermAppRelationWithGroup(data map[string]interface{}) (*[]PermAppRelation, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v", err)
		return nil, err
	}
	defer connection.Close()
	db := connection.Table(PermAppRelation{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result []PermAppRelation
	if err := db.Where(data).Group("app_version").Order("app_version DESC").Find(&result).Error; err != nil {
		logs.Error("query permission-app relationship failed,%v", err)
		return nil, err
	}
	return &result, nil
}

/**
新增权限操作历史
*/
func InsertPermOperationHistory(data PermHistory) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v", err)
		return err
	}
	defer connection.Close()
	data.UpdatedAt = time.Now()
	data.CreatedAt = time.Now()
	db := connection.Table(PermHistory{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Create(&data).Error; err != nil {
		logs.Error("insert perm history failed,%v", err)
		return err
	}
	return nil
}

/**
批量插入群仙操作历史
*/
func BatchInsertPermHistory(infos *[]PermHistory) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v", err)
		return err
	}
	defer connection.Close()
	db := connection.Table(PermHistory{}.TableName()).LogMode(_const.DB_LOG_MODE)
	db.Begin()
	for _, info := range *infos {
		info.CreatedAt = time.Now()
		info.UpdatedAt = time.Now()
		if err := db.Create(&info).Error; err != nil {
			logs.Error("insert perm history failed,%v", err)
			db.Rollback()
			return err
		}
	}
	db.Commit()
	return nil
}

/**
查询权限操作历史
*/
func QueryPermHistory(queryData map[string]interface{}) (*[]PermHistory, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("connect to db failed,%v", err)
		return nil, err
	}
	defer connection.Close()
	db := connection.Table(PermHistory{}.TableName()).LogMode(_const.DB_LOG_MODE)

	var result []PermHistory
	if err := db.Where(queryData).Order("updated_at DESC").Find(&result).Error; err != nil {
		logs.Error("query perm history failed,%v", err)
		return nil, err
	}
	return &result, nil
}
