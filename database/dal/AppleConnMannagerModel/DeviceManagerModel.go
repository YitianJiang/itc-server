package devconnmanager

import "code.byted.org/gopkg/gorm"

type DeviceInfo struct {
	gorm.Model
	UdId              string   `gorm:"ud_id"                   json:"ud_id"              binding:"required"`
	DeviceId          string   `gorm:"device_id"               json:"device_id"`
	DeviceClass       string   `gorm:"device_class"            json:"device_class"`
	DeviceName        string   `gorm:"device_name"             json:"device_name"        binding:"required"`
	DeviceModel       string   `gorm:"device_model"            json:"device_model"       binding:"required"`
	DevicePlatform    string   `gorm:"device_platform"         json:"device_platform"    binding:"required"`
	DeviceStatus      string   `gorm:"device_status"           json:"device_status"`
	DeviceAddedDate   string   `gorm:"device_added_date"       json:"device_added_date"`
	OpUser            string   `gorm:"op_user"                 json:"op_user"            binding:"required"`
	TeamId            string   `gorm:"team_id"                 json:"team_id"            binding:"required"`
}

type GetDevicesInfoRequest struct {
	TeamId            string       `form:"team_id"      binding:"required"`
}

type AddDeviceInfoRequest struct {
	OpUser            string       `json:"op_user"              binding:"required"`
	TeamId            string       `json:"team_id"              binding:"required"`
	DeviceName        string       `json:"device_name"          binding:"required"`
	Udid              string       `json:"udid"                 binding:"required"`
	DevicePlatform    string       `json:"device_platform"      binding:"required"`
	AccountType       string       `json:"account_type"         binding:"required"`
	DevicePrincipal   string       `json:"device_principal"`
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
	OpUser            string       `json:"op_user"              binding:"required"`
	TeamId            string       `json:"team_id"              binding:"required"`
	DeviceId          string       `json:"device_id"            binding:"required"`
	UdId              string       `json:"ud_id"                binding:"required"`
	DeviceName        string       `json:"device_name"`
	DeviceStatus      string       `json:"device_status"`
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
	TeamId   string     `json:"team_id"          binding:"required"`
	UdId     string     `json:"ud_id"            binding:"required"`
	OpUser   string     `json:"op_user"          binding:"required"`
}

func (DeviceInfo) TableName() string {
	return "tt_apple_device"
}

