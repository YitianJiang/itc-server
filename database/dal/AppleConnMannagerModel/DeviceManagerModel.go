package devconnmanager

import "code.byted.org/gopkg/gorm"

type DeviceInfo struct {
	gorm.Model
	UdId              string   `gorm:"ud_id"                   json:"udId"              binding:"required"`
	DeviceId          string   `gorm:"device_id"               json:"deviceId"          binding:"required"`
	DeviceClass       string   `gorm:"device_class"            json:"deviceClass"       binding:"required"`
	DeviceName        string   `gorm:"device_name"             json:"deviceName"        binding:"required"`
	DeviceModel       string   `gorm:"device_model"            json:"deviceModel"       binding:"required"`
	DevicePlatform    string   `gorm:"device_platform"         json:"devicePlatform"    binding:"required"`
	DeviceStatus      string   `gorm:"device_status"           json:"deviceStatus"      binding:"required"`
	DeviceAddedDate   string   `gorm:"device_added_date"       json:"deviceAddedDate"   binding:"required"`
	OpUser            string   `gorm:"op_user"                 json:"opUser"            binding:"required"`
	TeamId            string   `gorm:"team_id"                 json:"teamId,omitempty"  binding:"required"`
}

type GetDevicesInfoRequest struct {
	TeamId            string       `form:"team_id"      binding:"required"`
	UserName          string       `form:"user_name"    binding:"required"`
}

type AddDeviceInfoRequest struct {
	TeamId            string       `json:"team_id"              binding:"required"`
	DeviceName        string       `json:"device_name"          binding:"required"`
	Udid              string       `json:"udid"                 binding:"required"`
	DevicePlatform    string       `json:"device_platform"      binding:"required"`
	AccountType       string       `json:"account_type"         binding:"required"`
	DevicePrincipal   string       `json:"device_principal"     binding:"required"`

}

//苹果添加设备信息请求
type AppAddDevInfoReq struct {
	Data              AppAddDevInfoReqData                  `json:"data"`
}

type AppAddDevInfoReqData struct {
	Type              string                                `json:"type"`
	Attributes        AppAddDevInfoReqAttributes            `json:"attributes"`
}

type AppAddDevInfoReqAttributes struct {
	Name              string       `json:"name"`
	Udid              string       `json:"udid"`
	Platform          string       `json:"platform"`
}

//接收添加设备信息苹果返回
type AddDevInfoAppRet struct {
	Data              DevicesDataObjRes       `json:"data"`
}

type UpdateDeviceInfoRequest struct {
	TeamId            string       `json:"team_id"              binding:"required"`
	DeviceId          string       `json:"device_id"            binding:"required"`
	DeviceName        string       `json:"device_name"                                  gorm:"device_name"                       `
	DeviceStatus      string       `json:"device_status"                                gorm:"device_status"`
	DevicePrincipal   string       `json:"device_principal"`
	AccountType       string       `json:"account_type"         binding:"required"`
}

//苹果更新设备信息请求
type AppUpdDevInfoReq struct {
	Data              AppUpdDevInfoReqData       `json:"data"`
}

type AppUpdDevInfoReqData struct {
	Id                string                                `json:"id"`
	Type              string                                `json:"type"`
	Attributes        AppUpdDevInfoReqAttributes            `json:"attributes"`
}

type AppUpdDevInfoReqAttributes struct {
	Name              string       `json:"name,omitempty"`
	Status            string       `json:"status,omitempty"`
}

type UpdateDeviceFeedback struct {
	FeedBackJson UpdateDeviceFeedbackJson `json:"customer_parameter"`
}

type UpdateDeviceFeedbackJson struct {
	DeviceId string     `json:"device_id"        binding:"required"`
	OpUser   string     `json:"op_user"          binding:"required"`
}

func (DeviceInfo) TableName() string {
	return "tt_apple_device"
}

