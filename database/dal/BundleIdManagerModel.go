package dal

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

//type BundleIdManager struct {
//	gorm.Model
//	BundleId               string    `gorm:"column:bundle_id"                     json:"bundleId"`
//}
////表名
//func (c BundleIdManager) TableName() string {
//	return "tb_bundleid_manager"
//}

type P8fileManager struct {
	gorm.Model
	BundleId               string    `gorm:"column:bundle_id"                     json:"bundleId"`
	P8StringInfo 		   string    `gorm:"column:p8_string"                     json:"p8String"`
}
//表名
func (c P8fileManager) TableName() string {
	return "tb_bundleid_manager"
}

//func InsertBundleId (bundleid BundleIdManager) bool {
//	connection, err := database.GetDBConnection()
//	if err != nil {
//		logs.Error("Connect to Db failed: %v", err)
//		return false
//	}
//	defer connection.Close()
//	err = connection.Table(BundleIdManager{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&bundleid).Error
//	if err != nil {
//		logs.Error("添加bundle id失败!", err)
//		return false
//	}
//	return true
//}

func InsertBundleIdAndP8 (p8info P8fileManager) bool {
	connection, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	err = connection.Table(P8fileManager{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&p8info).Error
	if err != nil {
		logs.Error("添加bundle id失败!", err)
		return false
	}
	return true
}

func SearchBundleIds () (*[]P8fileManager,bool) {
	connection, err := database.GetDBConnection()
	var BundleIdsObjResponse []P8fileManager
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return &BundleIdsObjResponse,false
	}
	defer connection.Close()
	if err := connection.Table(P8fileManager{}.TableName()).LogMode(_const.DB_LOG_MODE).Find(&BundleIdsObjResponse).Error; err != nil {
		logs.Error("查询Bundle ID列表失败!", err)
		return &BundleIdsObjResponse,false
	}
	return &BundleIdsObjResponse,true
}

func SearchP8String(input map[string]interface{}) (*P8fileManager,bool) {
	connection, err := database.GetDBConnection()
	var BundleIdsObjResponse P8fileManager
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return &BundleIdsObjResponse,false
	}
	defer connection.Close()
	if err := connection.Table(P8fileManager{}.TableName()).LogMode(_const.DB_LOG_MODE).Where(input).Find(&BundleIdsObjResponse).Error; err != nil {
		logs.Error("查询Bundle ID列表失败!", err)
		return &BundleIdsObjResponse,false
	}
	return &BundleIdsObjResponse,true
}