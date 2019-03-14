package request_dal

import (
	"code.byted.org/gopkg/gorm"
	"time"
)

/*
product_info 表
*/
type ProductInfo struct {
	gorm.Model
	Aid            int    `json:"aid"`
	AppName        string `json:"app_name"`
	AppDisplayName string `json:"app_display_name"`
	Platform       string `json:"platform"`
	GitUrl         string `json:"git_url"`
	ProjectId      string `json:"project_id"`
	PermissionName string `json:"permission_name"`
	GitlabToken    string `json:"gitlab_token"`
	HistoryPkgUrl  string `json:"history_pkg_url"`
	GroupId        int    `json:"group_id"`
	GroupName      string `json:"group_name"`
	BundleId       string `json:"bundle_id"`
	PackageName    string `json:"package_name"`
	IconUrl        string `json:"icon_url"`
}

func (ProductInfo) TableName() string {
	return "product_info"
}

/*
user_concern_product 表
*/
type UserConcernProduct struct {
	gorm.Model
	UserID        uint `json:"user_id"`
	ProductInfoID uint `json:"product_info_id"`
	IsConcern     int  `json:"is_concern"`
}

func (UserConcernProduct) TableName() string {
	return "user_concern_product"
}

/*
user表
*/
type Struct_User struct {
	gorm.Model
	Name            string
	Email           string
	Avatar          string
	Employeenumber  uint
	PermissionLevel string `json:"permission_level"`
	Groupid         int
	Full_name       string
	LarkID          string `json:"lark_id"`
}

func (r Struct_User) TableName() string {
	return "users"
}

/*
permission表
*/
type PkgLevel struct {
	gorm.Model
	Name            string `json:"name"`
	PermissionLevel string `json:"permission_level"`
}

func (r PkgLevel) TableName() string {
	return "package_level"
}

/*
pkg 关联返回的数据
*/

type PkgHistory struct {
	Id           int
	RetPkgInfo   string
	ViewUrl      string
	JobUrl       string
	StopUrl      string
	PkgParam     string
	Describes    string
	OuterVersion string
	CreatedAt    *time.Time
	WorkflowId   int
	EmailPrefix  string
	InnerVersion string
	FiveVersion  string
	Status       int
}

type SearchModel struct {
	PageNumber   int
	Aid          string
	Platform     string
	CommitID     string
	Branch       string
	PkgUrl       string
	OperatorUser string
	Version      string
	Content      string
	Describe     string
	WorkFlowId   string
	Since        string
	Until        string
	LabelID      string
}

//获取单个workflow所需要的信息
type JobInfo struct {
	Content     string `json:"content"`
	Describes   string `json:"describes"`
	InputParam  string `json:"input_param"`
	SendParam   string `json:"send_param"`
	ErrorLog    string `json:"error_log"`
	ViewUrl     string `json:"view_url"`
	JobUrl      string `json:"job_url"`
	StopUrl     string `json:"stop_url"`
	Status      int    `json:"status"`
	ReturnParam string `json:"return_param"`
	ID          uint   `json:"id"`
	Stage       int    `json:"stage"`
	JobType     int    `json:"job_type"`
	StageName   string `json:"stage_name"`
	ConfigID    int    `json:"config_id"`
	Active      int    `json:"active"`
	ErrLog      string `json:"err_log"`
}

type CronPkg struct {
	gorm.Model
	PackageConfigID string `json:"package_config_id"`
	Notified        string `json:"notified"`
}

func (c CronPkg) TableName() string {
	return "cron_pkg"
}
