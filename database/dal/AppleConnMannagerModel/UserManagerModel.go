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

//编辑人员权限的上传参数Struct
type UserPermEditReq struct {
	OperateUserName string `binding:"required" json:"operate_user_name"`
	TeamId string `json:"team_id" binding:"required"`
	UserId string `json:"user_id" binding:"required"`
	AppleId string `json:"apple_id" binding:"required"`
	ProvisioningAllowedResult bool `json:"provisioning_allowed_result" binding:"exists"`
	ProvisioningChangeSign string `json:"provisioning_change_sign" binding:"required"`
	RolesResult []string `json:"roles_result" binding:"required"`
	RolesAdd []string `json:"roles_add"`
	RolesMin []string `json:"roles_min"`
	AllAppsVisibleResult bool `json:"all_apps_visible_result" binding:"exists"`
	AllappsVisibleChangeSign string `json:"allapps_visible_change_sign" binding:"required"`
	VisibleAppsResult []string `json:"visible_apps_result"`
	VisibleAppsAdd []string `json:"visible_apps_add"`
	VisibleAppsMin []string `json:"visible_apps_min"`
}

type VisibleAppItemReqOfApple struct {
	AppAppleId string `json:"id" binding:"required"`
	AppType string `json:"type" binding:"required"`
}

type VisibleAppsReqOfApple struct {
	DataList []VisibleAppItemReqOfApple `json:"data,omitempty"`
}

type VisibleAppObjReqOfApple struct {
	VisibleApps *VisibleAppsReqOfApple `json:"visibleApps,omitempty"`
}

type UserAttributesEditReqOfApple struct {
	AllAppsVisible bool `json:"allAppsVisible" binding:"required"`
	ProvisioningAllowed bool `json:"provisioningAllowed" binding:"required"`
	Roles []string `json:"roles" binding:"required"`
}

type UserPermEditReqOfApple struct {
	Attributes UserAttributesEditReqOfApple `json:"attributes" binding:"required"`
	Id string `json:"id" binding:"required"`
	Relationships *VisibleAppObjReqOfApple `json:"relationships,omitempty"`
	Type string `json:"type" binding:"required"`
}
type UserPermEditReqOfAppleObj struct {
	DataObj UserPermEditReqOfApple `json:"data" binding:"required"`
}

type InsertUserPermEditHistoryDBModel struct {
	gorm.Model
	OperateUserName string `gorm:"column:operate_user_name"`
	TeamId string `gorm:"column:team_id"`
	UserId string `gorm:"column:user_id"`
	AppleId string `gorm:"column:apple_id"`
	ProvisioningChange string `gorm:"column:provisioning_change"`
	RolesAdd string `gorm:"column:roles_add"`
	RolesMin string `gorm:"column:roles_min"`
	AllappsVisibleChange string `gorm:"column:allapps_visible_change"`
	VisibleAppsAdd string `gorm:"column:visible_apps_add"`
	VisibleAppsMin string `gorm:"column:visible_apps_min"`
}

func (c InsertUserPermEditHistoryDBModel) TableName() string {
	return "tt_apple_user_edit_history"
}

//邀请人员的上传参数Struct
type UserInvitedReq struct {
	OperateUserName string `binding:"required" json:"operate_user_name"`
	TeamId string `json:"team_id" binding:"required"`
	AppleId string `json:"apple_id" binding:"required"`
	FirstName string `json:"firstName" binding:"required"`
	LastName string `json:"lastName" binding:"required"`
	ProvisioningAllowedResult bool `json:"provisioning_allowed_result" binding:"exists"`
	RolesResult []string `json:"roles_result" binding:"required"`
	AllAppsVisibleResult bool `json:"all_apps_visible_result" binding:"exists"`
	VisibleAppsResult []string `json:"visible_apps_result"`
}

type UserInvitedAttributesReqOfApple struct {
	AllAppsVisible bool `json:"allAppsVisible" binding:"required"`
	ProvisioningAllowed bool `json:"provisioningAllowed" binding:"required"`
	Roles []string `json:"roles" binding:"required"`
	Email string `json:"email" binding:"required"`
	FirstName string `json:"firstName" binding:"required"`
	LastName string `json:"lastName" binding:"required"`
}

type UserInvitedReqOfApple struct {
	Attributes UserInvitedAttributesReqOfApple `json:"attributes" binding:"required"`
	Relationships *VisibleAppObjReqOfApple `json:"relationships,omitempty"`
	Type string `json:"type" binding:"required"`
}
type UserInvitedReqOfAppleObj struct {
	DataObj UserInvitedReqOfApple `json:"data" binding:"required"`
}

type InsertUserInvitedHistoryDBModel struct {
	gorm.Model
	OperateUserName string `gorm:"column:operate_user_name"`
	TeamId string `gorm:"column:team_id"`
	AppleId string `gorm:"column:apple_id"`
	InvitedOrCancel string `gorm:"column:invited_or_cancel"`
}

func (c InsertUserInvitedHistoryDBModel) TableName() string {
	return "tt_apple_user_invited_history"
}