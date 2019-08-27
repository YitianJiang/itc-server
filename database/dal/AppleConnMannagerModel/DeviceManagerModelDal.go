package devconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/logs"
)

func QueryDevicesInfo(condition map[string]interface{}) (*[]DeviceInfo,bool) {
	var devicesInfo []DeviceInfo
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return  &devicesInfo,false
	}
	defer conn.Close()
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(DeviceInfo{}.TableName()).
		Where(condition).Find(&devicesInfo).
		Error; err != nil {
		logs.Error("Query DB Failed:", err)
		return &devicesInfo,false
	}
	return &devicesInfo,true
}

func UpdateDeviceInfoDB(condition map[string]interface{},updateInfo map[string]interface{}) bool {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return false
	}
	defer conn.Close()
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(DeviceInfo{}.TableName()).
		Where(condition).
		Update(updateInfo).
		Error; err != nil {
		logs.Error("Update DB Failed:", err)
		return false
	}
	return true
}

func AddDeviceInfoDB(deviceInfo *DeviceInfo) bool {
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return false
	}
	defer conn.Close()
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(DeviceInfo{}.TableName()).
		Create(&deviceInfo).
		Error; err != nil {
		logs.Error("Insert DB Failed:", err)
		return false
	}
	return true
}

func AddOrUpdateDeviceInfo(condition map[string]interface{},deviceInfo *DeviceInfo) bool{
	conn, err := database.GetConneection()
	if err != nil {
		logs.Error("Get DB Connection Failed: ", err)
		return  false
	}
	defer conn.Close()
	var count int
	if err = conn.LogMode(_const.DB_LOG_MODE).Table(DeviceInfo{}.TableName()).
		Where(condition).
		Count(&count).
		Error; err != nil {
		logs.Error("Insert or Update DB Failed:", err)
		return false
	}
	if count==0{
		if err = conn.LogMode(_const.DB_LOG_MODE).Table(DeviceInfo{}.TableName()).
			Create(&deviceInfo).
			Error; err != nil {
			logs.Error("Insert DB Failed:", err)
			return false
		}
	}else{
		inputs:=map[string]interface{}{
			"device_name":deviceInfo.DeviceName,
			"device_status":deviceInfo.DeviceStatus,
		}
		if err = conn.LogMode(_const.DB_LOG_MODE).Table(DeviceInfo{}.TableName()).
			Where(condition).
			Update(inputs).
			Error; err != nil {
			logs.Error("Update DB Failed:", err)
			return false
		}
	}
	return true
}
