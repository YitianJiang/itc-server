package devconnmanager

import "code.byted.org/gopkg/gorm"

type RolesInfoRes struct {
	RolesObj map[string]string `json:"roles_obj"`
	RolesIndex []string `json:"roles_index"`
}
//从苹果取user数据
type UserDetailInfoReq struct {
	TeamId string `form:"team_id" binding:"required"`
	UpdateDBControl string `form:"db_update"`
}

type AttributeFromAppleUserInfo struct {
	UserName string `json:"username,omitempty"`
	Email string `json:"email,omitempty"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Roles []string `json:"roles"`
	AllAppsVisible bool `json:"allAppsVisible"`
	ProvisioningAllowed bool `json:"provisioningAllowed"`
	ExpirationDate string `json:"expirationDate,omitempty"`
}

type FromAppleUserItemInfo struct {
	UserIdApple string `json:"id"`
	AttributeUserInfo AttributeFromAppleUserInfo `json:"attributes"`
}

type FromAppleUserInfo struct {
	DataList []FromAppleUserItemInfo `json:"data"`
}

//组合user数据
type RetUsersDataDetailObj struct {
	UserIdApple string `json:"user_id"`
	UserName string `json:"username,omitempty"`
	Email string `json:"email,omitempty"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Roles []string `json:"roles"`
	AllAppsVisible bool `json:"allAppsVisible"`
	ProvisioningAllowed bool `json:"provisioningAllowed"`
	ExpirationDate string `json:"expirationDate,omitempty"`
}

type RetUsersDataObj struct {
	EmployData map[string][]RetUsersDataDetailObj `json:"employ_data,omitempty"`
}

type RetUsersInvitedDataObj struct {
	InvitedData map[string][]RetUsersDataDetailObj `json:"invited_data,omitempty"`
}

//从苹果取账号里所有的app数据
type RetALlVisibleAppsItemAttribute struct {
	AppName string `json:"name"`
	BundleID string `json:"bundleId"`
}

type RetALlVisibleAppsItemFromApple struct {
	AppAppleId string `json:"id"`
	AppsAttribute RetALlVisibleAppsItemAttribute `json:"attributes"`
}

type RetAllVisibleAppsFromApple struct {
	DataList []RetALlVisibleAppsItemFromApple `json:"data"`
}
//组合账号里所有的app数据，返回前端
type RetAllVisibleAppItem struct {
	AppAppleId string `json:"app_apple_id" gorm:"column:app_apple_id"`
	AppName string `json:"name" gorm:"column:app_name"`
	BundleID string `json:"bundleId"gorm:"column:bundle_id"`
}

type AllVisibleAppDB struct {
	gorm.Model
	AppAppleId string `json:"id" gorm:"column:app_apple_id"`
	AppName string `json:"name" gorm:"column:app_name"`
	BundleID string `json:"bundleId" gorm:"column:bundle_id"`
	TeamId string `json:"team_id" gorm:"column:team_id"`
}
func (c AllVisibleAppDB) TableName() string {
	return "tt_apple_app_info"
}

//从苹果取某一位user可见app信息
type UserVisibleAppsReq struct {
	TeamId string `form:"team_id" binding:"required"`
	UserId string `form:"user_id" binding:"required"`
	OrInvited string `form:"or_invited" binding:"required"`
}