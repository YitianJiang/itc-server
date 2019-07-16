package devconnmanager

type RolesInfoRes struct {
	RolesObj map[string]string `json:"roles_obj"`
	RolesIndex []string `json:"roles_index"`
}
//从苹果取user数据
type UserDetailInfoReq struct {
	TeamId string `form:"team_id" binding:"required"`
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

//组合数据
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
	InvitedData map[string][]RetUsersDataDetailObj `json:"invited_data,omitempty"`
}