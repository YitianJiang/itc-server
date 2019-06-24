package dal

import (
	"fmt"
	"strconv"
	"time"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

//二进制包检测任务
type DetectStruct struct {
	gorm.Model
	Creator         string `gorm:"column:creator"                json:"creator"`
	ToLarker        string `gorm:"column:to_larker"              json:"toLarker"`
	ToGroup         string `gorm:"column:to_group"               json:"toGroup"`
	Platform        int    `gorm:"column:platform"               json:"platform"`
	AppName         string `gorm:"column:app_name"               json:"appName"`
	AppVersion      string `gorm:"column:app_version"            json:"appVersion"`
	AppId           string `gorm:"column:app_id"                 json:"appId"`
	CheckContent    string `gorm:"column:check_content"          json:"checkContent"`
	SelfCheckStatus int    `gorm:"column:self_check_status"      json:"selfCheckStatus"` //0-自查未完成；1-自查完成
	TosUrl          string `gorm:"column:tos_url"                json:"tosUrl"`
	Status          int    `gorm:"column:status"                 json:"status"` //0---未完全确认；1---已完全确认
	ExtraInfo		string `gorm:"column:extra_info" 		     json:"extraInfo"`//其他附加信息
}
type ExtraStruct struct {
	CallBackAddr		string 			`json:"callBackAddr"`
}
type RecordTotal struct {
	Total uint
}

type RetDetectTasks struct {
	GetMore uint
	Total   uint
	NowPage uint
	Tasks   []DetectStruct
}

//包检测工具
type DetectTool struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Platform    int    `json:"platform"`
}

//二进制包检测内容
type DetectContent struct {
	gorm.Model
	TaskId      int    `json:"taskId"`
	ToolId      int    `json:"toolId"`
	HtmlContent string `json:"htmlContent"`
	JsonContent string `json:"jsonContent"`
	Status      int    `json:"status"` //是否确认,0-未确认，1-确认通过，2-确认未通过
	Confirmer   string `json:"confirmer"`
	Remark      string `json:"remark"`
}

//apk检测信息----fj新增
type DetectInfo struct {
	gorm.Model
	TaskId      int    `json:"taskId"`
	ApkName     string `json:"apkName"`
	Version     string `json:"version"`
	Channel     string `json:"channel"`
	Permissions string `json:"permissions"`
	ToolId      int    `json:"toolId"`
	SubIndex    int    `json:"index"`
}

//敏感信息详情---fj新增
type DetectContentDetail struct {
	gorm.Model
	TaskId    int    `json:"taskId"`
	Status    int    `json:"status"` //是否确认,0-未确认，1-确认通过，2-确认未通过
	Remark    string `json:"remark"`
	Confirmer string `json:"confirmer"`
	SensiType int    `json:"sensiType"` //敏感信息类型，1-敏感方法，2-敏感字符串
	KeyInfo   string `json:"key"`
	ClassName string `json:"className"`
	DescInfo  string `json:"desc"`
	CallLoc   string `json:"callLoc"`
	ToolId    int    `json:"toolId"`
	SubIndex  int    `json:"index"`
	//OtherVersion		string			`json:"otherVersion"`
	//Priority 			int				`json:"priority"`//0--常规，1--注意，2--危险，3--非常危险，4--未定义
}

type IgnoreInfoStruct struct {
	gorm.Model
	AppId     int    `json:"appId"`
	Platform  int    `json:"platform"` //0-安卓，1-iOS
	KeysInfo  string `json:"keys"`
	SensiType int    `json:"sensiType"` //敏感信息类型，1-敏感方法，2-敏感字符串
	Version   string `json:"version"`   //app版本
	Remarks   string `json:"remarks"`
	Confirmer string `json:"confirmer"`
	Status    int    `json:"status"` //1-确认通过，2-确认未通过
}

/**
 *安卓检测数据查询返回结构
 */
type DetectQueryStruct struct {
	Index         int           `json:"index"`
	ApkName       string        `json:"apkName"`
	Version       string        `json:"version"`
	Channel       string        `json:"channel"`
	Permissions   string        `json:"permissions"`
	SMethods      []SMethod     `json:"sMethods"`
	SStrs         []SStr        `json:"sStrs"`
	SStrs_new     []SStr        `json:"newStrs"`
	Permissions_2 []Permissions `json:"permissionList"`
}

type SMethod struct {
	Id           uint             `json:"id"`
	Status       int              `json:"status"`
	Remark       string           `json:"remark"`
	Confirmer    string           `json:"confirmer"`
	MethodName   string           `json:"methodName"`
	ClassName    string           `json:"className"`
	Desc         string           `json:"desc"`
	CallLoc      []MethodCallJson `json:"callLoc"`
	OtherVersion string           `json:"otherVersion"`
}
type MethodCallJson struct {
	MethodName string      `json:"method_name"`
	ClassName  string      `json:"class_name"`
	LineNumber interface{} `json:"line_number"`
}

type SStr struct {
	Id           uint          `json:"id"`
	Status       int           `json:"status"`
	Remark       string        `json:"remark"`
	Confirmer    string        `json:"confirmer"`
	Keys         string        `json:"keys"`
	Desc         string        `json:"desc"`
	CallLoc      []StrCallJson `json:"callLoc"`
	ConfirmInfos []ConfirmInfo `json:"confirmerInfos"`
}

type StrCallJson struct {
	Key        string      `json:"key"`
	MethodName string      `json:"method_name"`
	ClassName  string      `json:"class_name"`
	LineNumber interface{} `json:"line_number"`
}
type ConfirmInfo struct {
	//Id					uint 				`json:"id"`
	Key          string `json:"key"`
	Status       int    `json:"status"`
	Remark       string `json:"remark"`
	Confirmer    string `json:"confirmer"`
	OtherVersion string `json:"otherVersion"`
}

type Permissions struct {
	Id           uint   `json:"id"`
	PermId       int    `json:"permId"`
	Key          string `json:"key"`
	Status       int    `json:"status"`
	Remark       string `json:"remark"`
	Confirmer    string `json:"confirmer"`
	OtherVersion string `json:"otherVersion"`
	Priority     int    `json:"priority"`
	Desc         string `json:"desc"`
}

/**
安卓确认post结构
*/
type PostConfirm struct {
	TaskId int    `json:"taskId"`
	Id     int    `json:"id"`
	Status int    `json:"status"`
	Remark string `json:"remark"`
	ToolId int    `json:"toolId"`
	Type   int    `json:"type"`
	Index  int    `json:"index"`
}

//二进制包检测内容，json内容处理区分后
type IOSDetectContent struct {
	gorm.Model
	TaskId          int    `gorm:"column:taskId"            json:"taskId"`
	ToolId          int    `gorm:"column:toolId"            json:"toolId"`
	JsonContent     string `gorm:"column:jsonContent"      json:"jsonContent"`
	Category        string `gorm:"column:category"          json:"category"`
	CategoryName    string `gorm:"column:categoryName"      json:"categoryName"` //兼容就接口保留，后续全部完成后删除
	CategoryContent string `gorm:"column:categoryContent"   json:"categoryContent"`
	Status          int    `gorm:"column:status"            json:"status"`
	Remark          string `gorm:"column:remark"            json:"remark"`
	Confirmer       string `gorm:"column:confirmer"         json:"confirmer"`
}
type IOSNewDetectContent struct {
	gorm.Model
	TaskId        int    `gorm:"column:taskId"            json:"taskId"`
	ToolId        int    `gorm:"column:toolId"            json:"toolId"`
	AppName       string `gorm:"column:appname"           json:"appName"`
	AppId         int    `gorm:"column:app_id"            json:"appId"`
	Version       string `gorm:"column:app_version"       json:"appVersion"`
	MinVersion    string `gorm:"column:min_version"       json:"minVersion"`
	BundleId      string `gorm:"column:bundle_id"         json:"bundleId"`
	SdkVersion    string `gorm:"column:tar_version"       json:"sdkVersion"`
	JsonContent   string `gorm:"column:json_content"      json:"jsonContent"`
	DetectType    string `gorm:"column:detect_type"       json:"detectType"`
	DetectContent string `gorm:"column:detect_content"    json:"detectContent"`
}

//二进制包权限确认历史
type PrivacyHistory struct {
	gorm.Model
	AppName        string `gorm:"column:appname"            json:"appName"`
	AppId          int    `gorm:"column:app_id"            json:"appId"`
	Platform       int    `gorm:"column:platform"           json:"platform"`
	Permission     string `gorm:"column:permission"          json:"permission"`
	Status         int    `gorm:"column:status"             json:"status"` //是否确认,0-未确认，1-确认通过，2-确认未通过
	Confirmer      string `gorm:"column:confirmer"          json:"confirmer"`
	ConfirmReason  string `gorm:"column:confirm_reason"     json:"confirmReason"`
	ConfirmVersion string `gorm:"column:confirm_version"    json:"confirmversion"`
}

func (IgnoreInfoStruct) TableName() string {
	return "tb_ignored_info"
}
func (DetectInfo) TableName() string {
	return "tb_detect_info_apk"
}
func (DetectContentDetail) TableName() string {
	return "tb_detect_content_detail"
}
func (DetectStruct) TableName() string {
	return "tb_binary_detect"
}
func (DetectContent) TableName() string {
	return "tb_detect_content"
}
func (DetectTool) TableName() string {
	return "tb_detect_tool"
}
func (IOSDetectContent) TableName() string {
	return "tb_ios_detect_content"
}
func (IOSNewDetectContent) TableName() string {
	return "tb_ios_new_detect_content"
}
func (PrivacyHistory) TableName() string {
	return "tb_privacy_history"
}

//insert data
func InsertDetectModel(detectModel DetectStruct) uint {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return 0
	}
	defer connection.Close()
	db := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Create(&detectModel).Error; err != nil {
		return 0
	}
	return detectModel.ID
}

//update data
func UpdateDetectModel(detectModel DetectStruct, content DetectContent) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Begin()
	taskId := detectModel.ID
	condition := "id=" + fmt.Sprint(taskId)
	if err := db.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Where(condition).Update(&detectModel).Error; err != nil {
		logs.Error("update binary check failed, %v", err)
		db.Rollback()
		return err
	}
	//insert detect content
	if err := db.Table(DetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Create(&content).Error; err != nil {
		logs.Error("insert binary check content failed, %v", err)
		db.Rollback()
		return err
	}
	db.Commit()
	return nil
}

//update data-----fj
func UpdateDetectModelNew(detectModel DetectStruct) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Begin()
	taskId := detectModel.ID
	condition := "id=" + fmt.Sprint(taskId)
	detectModel.UpdatedAt = time.Now()
	if err := db.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).
		Where(condition).Update(&detectModel).Error; err != nil {
		logs.Error("update binary check failed, %v", err)
		db.Rollback()
		return err
	}
	db.Commit()
	return nil
}

//delete data
func DeleteDetectModel(detectModeId string) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return err
	}
	defer connection.Close()
	db := connection.Table(DetectStruct{}.TableName())
	if err := db.Where("id = ?", detectModeId).LogMode(_const.DB_LOG_MODE).Delete(&DetectStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return err
	}
	return nil
}

/**
 * 更新tos地址
 */
func UpdateDetectTosUrl(path string, taskId uint) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	condition := "id='" + fmt.Sprint(taskId) + "'"
	db := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Update(map[string]interface{}{"tos_url": path, "updated_at": time.Now()}).Error; err != nil {
		logs.Error("%v", err)
		return false
	}
	return true
}

//query by map
func QueryDetectModelsByMap(param map[string]interface{}) *[]DetectStruct {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var detect []DetectStruct
	db := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(param).Find(&detect).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &detect
}

//query data
func QueryTasksByCondition(data map[string]interface{}) (*[]DetectStruct, uint) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, 0
	}
	defer connection.Close()
	db := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	condition := data["condition"]
	logs.Info("query tasks condition: %s", condition)
	if condition != "" {
		db = db.Where(condition).Order("created_at desc")
	}
	pageNo, okpn := data["pageNo"]
	pageSize, okps := data["pageSize"]
	if okpn {
		if !okps {
			pageSize = 10
		}
		page := pageNo.(int)
		size := pageSize.(int)
		db = db.Limit(pageSize)
		if page > 0 {
			db = db.Offset((page - 1) * size)
		}
	}
	var items []DetectStruct
	if err := db.Find(&items).Error; err != nil {
		logs.Error("%v", err)
		return nil, 0
	}
	var total RecordTotal
	if condition == "" {
		condition = " 1=1 "
	}
	connect, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, 0
	}
	defer connect.Close()
	dbCount := connect.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := dbCount.Select("count(id) as total").
		Where(condition).Find(&total).Error; err != nil {
		logs.Error("query total record failed! %v", err)
		return &items, 0
	}
	return &items, total.Total
}

//query by map
func QueryTaskBinaryCheckContent(condition string) *[]DetectContent {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var detect []DetectContent
	db := connection.Table(DetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if err := db.Where(condition).Find(&detect).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &detect
}

func ConfirmBinaryResult(data map[string]string) bool {
	taskId := data["task_id"]
	toolId := data["tool_id"]
	confirmer := data["confirmer"]
	remark := data["remark"]
	status := data["status"]
	statusInt, _ := strconv.Atoi(status)
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	db := connection.Table(DetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE)
	condition := "task_id=" + taskId + " and tool_id=" + toolId
	if err := db.Where(condition).LogMode(_const.DB_LOG_MODE).
		Update(map[string]interface{}{
			"status":     statusInt,
			"confirmer":  confirmer,
			"remark":     remark,
			"updated_at": time.Now(),
		}).Error; err != nil {
		logs.Error("update db tb_detect_content failed: %v", err)
		//db.Rollback()
		return false
	}
	//db.Commit()
	return true
}

/**
确认安卓二进制结果----------fj
*/
func ConfirmApkBinaryResultNew(data map[string]string) bool {
	id := data["id"]
	//toolId := data["tool_id"]
	confirmer := data["confirmer"]
	remark := data["remark"]
	statusInt, _ := strconv.Atoi(data["status"])
	//statusInt, _ := strconv.Atoi(status)
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	db := connection.Table(DetectContentDetail{}.TableName()).LogMode(_const.DB_LOG_MODE)
	condition := "id=" + id
	if err := db.Where(condition).
		Update(map[string]interface{}{
			"status":     statusInt,
			"confirmer":  confirmer,
			"remark":     remark,
			"updated_at": time.Now(),
		}).Error; err != nil {
		logs.Error("update db tb_detect_content failed: %v", err)
		//db.Rollback()
		return false
	}
	//db.Commit()
	return true
}

/**
检测信息insert-----fj
*/
func InsertDetectInfo(info DetectInfo) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()

	db := connection.Table(DetectInfo{}.TableName()).LogMode(_const.DB_LOG_MODE)

	info.CreatedAt = time.Now()
	info.UpdatedAt = time.Now()

	if err1 := db.Create(&info).Error; err1 != nil {
		logs.Error("数据库新增检测信息失败,%v", err1)
		return err1
	}
	return nil

}

/**
敏感信息详情insert------fj
*/
func InsertDetectDetail(detail DetectContentDetail) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()

	db := connection.Table(DetectContentDetail{}.TableName()).LogMode(_const.DB_LOG_MODE)

	detail.CreatedAt = time.Now()
	detail.UpdatedAt = time.Now()

	if err1 := db.Create(&detail).Error; err1 != nil {
		logs.Error("数据库新增敏感信息失败,%v，敏感信息具体key参数：%s", err1, detail.KeyInfo)
		return err1
	}
	return nil
}

func InsertDetectDetailBatch(details *[]DetectContentDetail) error {
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
		if err1 := db.Table(DetectContentDetail{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&detail).Error; err1 != nil {
			logs.Error("数据库新增敏感信息失败,%v，敏感信息具体key参数：%s", err1, detail.KeyInfo)
			db.Rollback()
			return err1
		}
	}
	db.Commit()
	return nil
}

/**
未确认敏感信息数据量查询-----fj
*/
func QueryUnConfirmDetectContent(condition string) int {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return -1
	}
	defer connection.Close()

	db := connection.Table(DetectContentDetail{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var total RecordTotal
	if err = db.Select("count(id) as total").Where(condition).Find(&total).Error; err != nil {
		logs.Error("query sensitive infos total record failed! %v", err)
		return -1
	}
	return int(total.Total)

}

/**
查询apk检测info-----fj
*/
func QueryDetectInfo(condition string) (*DetectInfo, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil, err
	}
	defer connection.Close()

	db := connection.Table(DetectInfo{}.TableName()).LogMode(_const.DB_LOG_MODE)

	var detectInfo DetectInfo
	if err1 := db.Where(condition).Find(&detectInfo).Error; err1 != nil {
		logs.Error("query detectInfo failed! %v", err)
		return nil, err1
	}
	return &detectInfo, nil

}

/**
兼容.aab查询内容
*/
func QueryDetectInfo_2(condition string) (*[]DetectInfo, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil, err
	}
	defer connection.Close()

	db := connection.Table(DetectInfo{}.TableName()).LogMode(_const.DB_LOG_MODE)

	var detectInfo []DetectInfo
	if err1 := db.Where(condition).Find(&detectInfo).Error; err1 != nil {
		logs.Error("query detectInfo failed! %v", err)
		return nil, err1
	}
	return &detectInfo, nil

}

/**
查询apk敏感信息----fj
*/
func QueryDetectContentDetail(condition string) (*[]DetectContentDetail, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil, err
	}
	defer connection.Close()

	db := connection.Table(DetectContentDetail{}.TableName()).LogMode(_const.DB_LOG_MODE)

	var result []DetectContentDetail

	if err1 := db.Where(condition).Order("status ASC").Find(&result).Error; err1 != nil {
		logs.Error("query detectDetailInfos failed! %v", err)
		return nil, err1
	}
	return &result, nil

}

/**
可忽略信息insert------fj
*/
func InsertIgnoredInfo(detail IgnoreInfoStruct) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()

	db := connection.Table(IgnoreInfoStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)

	detail.CreatedAt = time.Now()
	detail.UpdatedAt = time.Now()

	if err1 := db.Create(&detail).Error; err1 != nil {
		logs.Error("数据库新增可忽略信息失败,%v，可忽略信息具体key参数：%s", err1, detail.KeysInfo)
		return err1
	}
	return nil
}

/**
查询可忽略信息----fj
*/
func QueryIgnoredInfo(queryInfo map[string]string) (*[]IgnoreInfoStruct, error) {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil, err
	}
	defer connection.Close()

	db := connection.Table(IgnoreInfoStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result []IgnoreInfoStruct
	condition := queryInfo["condition"]
	if err1 := db.Where(condition).Order("updated_at DESC").Find(&result).Error; err1 != nil {
		logs.Error("query ignoredInfos failed! %v", err1)
		return nil, err1
	}
	return &result, nil
}

//query tb_ios_detect_content
func QueryNewIOSDetectModel(condition map[string]interface{}) *[]IOSNewDetectContent {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer connection.Close()

	var iosDetectContent []IOSNewDetectContent
	if err := connection.Table(IOSNewDetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Find(&iosDetectContent).Error; err != nil {
		logs.Error("请求iOS静态检测结果出错！！！", err.Error())
		return nil
	}
	return &iosDetectContent
}

//update tb_ios_detect_content
func UpdateNewIOSDetectModel(model IOSNewDetectContent, updates map[string]interface{}) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false
	}
	defer connection.Close()
	if err := connection.Table(IOSNewDetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).Model(&model).Update(updates).Error; err != nil {
		logs.Error("更新iOS静态检测结果出错！！！", err.Error())
		return false
	}
	return true
}

//iOS 检测结果分类处理
func InsertNewIOSDetect(black, method, privacy IOSNewDetectContent) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return false
	}
	defer connection.Close()
	//insert detect content
	db := connection.Begin()
	if err := db.Table(IOSNewDetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&black).Error; &black != nil && err != nil {
		logs.Error("insert binary check content failed, %v", err)
		db.Rollback()
		return false
	}
	if err := db.Table(IOSNewDetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&method).Error; &method != nil && err != nil {
		logs.Error("insert binary check content failed, %v", err)
		db.Rollback()
		return false
	}
	if err := db.Table(IOSNewDetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&privacy).Error; &privacy != nil && err != nil {
		logs.Error("insert binary check content failed, %v", err)
		db.Rollback()
		return false
	}
	db.Commit()
	return true
}

//insert tb_ios_detect_permission
func CreatePrivacyHistoryModel(permission PrivacyHistory) error {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return err
	}
	defer connection.Close()
	//insert detect content
	if err := connection.Table(PrivacyHistory{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&permission).Error; err != nil {
		logs.Error("插入权限确认信息出错！, %v", err)
		return err
	}
	return nil
}

//query tb_ios_privacy_history
func QueryPrivacyHistoryModel(condition map[string]interface{}) *[]PrivacyHistory {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer connection.Close()
	var confirmHistory []PrivacyHistory
	if err := connection.Table(PrivacyHistory{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Order("id desc", true).Find(&confirmHistory).Error; err != nil {
		logs.Error("请求iOS权限检测结果出错！！！", err.Error())
		return nil
	}

	return &confirmHistory
}

/*
查询新旧中间数据
*/
func QueryIOSDetectContent(condition map[string]interface{}) *[]IOSDetectContent {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer connection.Close()
	var iosMiddleContenct []IOSDetectContent
	if err := connection.Table(IOSDetectContent{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(condition).Find(&iosMiddleContenct).Error; err != nil {
		logs.Error("请求iOS权限检测结果出错,中间数据！！！", err.Error())
		return nil
	}
	return &iosMiddleContenct
}

/*
查询当前taskId对应app的上一次检测taskId
*/
func QueryLastTaskId(taskId int) int {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return -1
	}
	defer connection.Close()
	var detect DetectStruct
	if err := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Select("app_id").Where("id = ?", taskId).Limit(1).Find(&detect).Error; err != nil {
		logs.Error(err.Error())
		return -1
	} else {
		appId := detect.AppId
		var lastDetect DetectStruct
		if err := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE).Select("id").Where("app_id = ? AND platform = 1 AND id < ?", appId, taskId).Order("id desc", true).Limit(1).Find(&lastDetect).Error; err != nil {
			logs.Error(err.Error())
			return -1
		} else {
			return int(lastDetect.ID)
		}
	}
}
