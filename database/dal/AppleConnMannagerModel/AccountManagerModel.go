package devconnmanager

import "code.byted.org/gopkg/gorm"

type AccountInfo struct {
	gorm.Model
	TeamId 			    string      `gorm:"team_id"              form:"team_id"                 json:"team_id"`
	IssueId 			string      `gorm:"issue_id"             form:"issue_id"                json:"issue_id,omitempty"`
	KeyId 				string      `gorm:"key_id"               form:"key_id"                  json:"key_id,omitempty"`
	AccountName 		string      `gorm:"account_name"         form:"account_name"            json:"account_name"`
	AccountType 		string      `gorm:"account_type"         form:"account_type"            json:"account_type"`
	AccountP8fileName   string      `gorm:"account_p8file_name"  form:"account_p8file_name"     json:"account_p8file_name,omitempty"`
	AccountP8file 		string      `gorm:"account_p8file"                                      json:"account_p8file,omitempty"`
	UserName 			string      `gorm:"user_name"            form:"user_name"               json:"user_name"`
	PermissionAction   []string     `gorm:"-"                                                   json:"permission_action"`
}

type AccInfoWithAuth struct {
	TeamId 			    string      `gorm:"team_id"`
	AccountName 		string      `gorm:"account_name"`
	AccountType 		string      `gorm:"account_type"`
	AccountP8fileName   string      `gorm:"account_p8file_name"`
	AccountP8file 		string      `gorm:"account_p8file"`
	IssueId 			string      `gorm:"issue_id"`
	KeyId 				string      `gorm:"key_id"               `
	UserName 			string      `gorm:"user_name"`
	PermissionAction   []string
}

type AccInfoWithoutAuth struct {
	TeamId 			    string      `gorm:"team_id"`
	AccountName 		string      `gorm:"account_name"`
	AccountType 		string      `gorm:"account_type"`
	UserName 			string      `gorm:"user_name"`
	PermissionAction   []string
}

type DelAccRequest struct {
	TeamId string `form:"team_id"   binding:"required"`
}

type CreateResourceRequest struct {
	ResourceName    string      `json:"resourceName"`
	ResourceKey     string      `json:"resourceKey"`
	CreatorKey      string      `json:"creatorKey"`
	ResourceType    int         `json:"resourceType"`
}

type CreResResponse struct {
	Errno   int      `json:"errno"`
}

type TeamID struct {
	gorm.Model
	TeamId string `gorm:"team_id"`
}

type CrePermRequest struct {
	PermissionName          string      `json:"permissionName"`
	PermissionAction        string      `json:"permissionAction"`
	CreatorKey              string      `json:"creatorKey"`
	ResourceKey             string      `json:"resourceKey"`
}

func (AccountInfo) TableName() string{
	return  "tt_apple_conn_account"
}
