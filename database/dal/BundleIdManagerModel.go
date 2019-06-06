package dal

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type BundleIdManager struct {
	gorm.Model
	BundleId               string    `gorm:"column:bundle_id"                     json:"bundleId"`
}

//表命
func (c BundleIdManager) TableName() string {
	return "tb_bundleid_manager"
}

func InsertBundleId (bundleid BundleIdManager) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	err = connection.Table(BundleIdManager{}.TableName()).LogMode(_const.DB_LOG_MODE).Create(&bundleid).Error
	if err != nil {
		logs.Error("添加bundle id失败!", err)
		return false
	}
	return true
}

func SearchBundleIds () (*[]BundleIdManager,bool) {
	connection, err := database.GetConneection()
	var BundleIdsObjResponse []BundleIdManager
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return &BundleIdsObjResponse,false
	}
	defer connection.Close()
	if err := connection.Table(BundleIdManager{}.TableName()).LogMode(_const.DB_LOG_MODE).Find(&BundleIdsObjResponse).Error; err != nil {
		logs.Error("查询Bundle ID列表失败!", err)
		return &BundleIdsObjResponse,false
	}
	return &BundleIdsObjResponse,true
}