package dal

import (
	"code.byted.org/gopkg/gorm"
)

/*
pkg config表
*/
type DBPkgConfigStruct struct {
	gorm.Model
	Aid                int    `json:"aid"`
	Platform           string `json:"platform"`
	ViewUrl            string `json:"view_url"`
	JobUrl             string `json:"job_url"`
	StopUrl            string `json:"stop_url"`
	PackageConfigParam string `json:"package_config_param"`
	Content            string `json:"content"`
	EmailPrefix        string `json:"email_prefix"`
	AllowBranch        string `json:"allow_branch"`
	ApiId              string `json:"api_id"`
	ApiToken           string `json:"api_token"`
	UserId             uint   `json:"user_id"`
	PermissionLevel    string `json:"permission_level"`
}

func (DBPkgConfigStruct) TableName() string {
	return "package_config"
}

/*
pkg表
*/
type DBPkgStruct struct {
	gorm.Model
	WorkflowId        uint   `json:"workflow_id"`
	PackageConfigId   int    `json:"package_config_id"`
	Aid               int    `json:"aid"`
	Platform          string `json:"platform"`
	Status            int    `json:"status"`
	Active            int    `json:"active"`
	PkgParam          string `json:"pkg_param"`
	RetPkgInfo        string `json:"ret_pkg_info"`
	ErrLog            string `json:"err_log"`
	ViewUrl           string `json:"view_url"`
	JobUrl            string `json:"job_url"`
	StopUrl           string `json:"stop_url"`
	EmailPrefix       string `json:"email_prefix"`
	Content           string `json:"content"`
	OuterVersion      string `json:"outer_version"`
	InnerVersion      string `json:"inner_version"`
	FiveVersion       string `json:"five_version"`
	UpdateVersionCode string `json:"update_version_code"`
	Describes         string `json:"describes"`
	ApiId             string `json:"api_id"`
	ApiToken          string `json:"api_token"`
	JobsysId          string `json:"jobsys_id"`
	UserId            uint   `json:"user_id"`
	PermissionLevel   string `json:"permission_level"`
	SendParam         string `json:"send_param"`
	AutoPkg           int    `json:"auto_pkg"`
}

func (DBPkgStruct) TableName() string {
	return "package"
}

/*
pub config表
*/
type DBPubConfigStruct struct {
	gorm.Model
	Aid               int    `json:"aid"`
	Platform          string `json:"platform"`
	ViewUrl           string `json:"view_url"`
	JobUrl            string `json:"job_url"`
	StopUrl           string `json:"stop_url"`
	PackageConfigId   uint   `json:"package_config_id"`
	PublicConfigParam string `json:"public_config_param"`
	Content           string `json:"content"`
	EmailPrefix       string `json:"email_prefix"`
	ApiId             string `json:"api_id"`
	ApiToken          string `json:"api_token"`
	InnerJob          int    `json:"inner_job"`
	Describes         string `json:"describes"`
	UserId            uint   `json:"user_id"`
	NeedCallback      int    `json:"need_callback"`
	LabelID           int    `json:"label_id"`
}

func (DBPubConfigStruct) TableName() string {
	return "publish_config"
}

/*
default pub config表
*/
type DBDefaultPubConfigStruct struct {
	gorm.Model
	Aid               int    `json:"aid"`
	Platform          string `json:"platform"`
	ViewUrl           string `json:"view_url"`
	JobUrl            string `json:"job_url"`
	StopUrl           string `json:"stop_url"`
	PublicConfigParam string `json:"public_config_param"`
	Content           string `json:"content"`
	EmailPrefix       string `json:"email_prefix"`
	ApiId             string `json:"api_id"`
	ApiToken          string `json:"api_token"`
	InnerJob          int    `json:"inner_job"`
	UserId            uint   `json:"user_id"`
	Describes         string `json:"describes"`
	NeedCallback      int    `json:"need_callback"`
}

func (DBDefaultPubConfigStruct) TableName() string {
	return "default_publish_config"
}

/*
default action config表
*/
type DBDefaultActionConfigStruct struct {
	gorm.Model
	Aid               int    `json:"aid"`
	Platform          string `json:"platform"`
	ViewUrl           string `json:"view_url"`
	JobUrl            string `json:"job_url"`
	StopUrl           string `json:"stop_url"`
	ActionConfigParam string `json:"action_config_param"`
	Content           string `json:"content"`
	EmailPrefix       string `json:"email_prefix"`
	Level             int    `json:"level"`
	ApiId             string `json:"api_id"`
	ApiToken          string `json:"api_token"`
	InnerJob          int    `json:"inner_job"`
	UserId            uint   `json:"user_id"`
	Describes         string `json:"describes"`
	NeedCallback      int    `json:"need_callback"`
	StageName         string `json:"stage_name"`
}

func (DBDefaultActionConfigStruct) TableName() string {
	return "default_action_config"
}

/*
pub表
*/
type DBPubStruct struct {
	gorm.Model
	Aid             int    `json:"aid"`
	Platform        string `json:"platform"`
	WorkflowId      uint   `json:"workflow_id"`
	PublishConfigId uint   `json:"publish_config_id"`
	PackageId       uint   `json:"package_id"`
	Status          int    `json:"status"`
	PublishParam    string `json:"publish_param"`
	RetPublishInfo  string `json:"ret_publish_info"`
	ErrLog          string `json:"err_log"`
	ViewUrl         string `json:"view_url"`
	JobUrl          string `json:"job_url"`
	StopUrl         string `json:"stop_url"`
	EmailPrefix     string `json:"email_prefix"`
	Active          int    `json:"active"`
	Content         string `json:"content"`
	ApiId           string `json:"api_id"`
	ApiToken        string `json:"api_token"`
	JobsysId        string `json:"jobsys_id"`
	Describes       string `json:"describes"`
	InnerJob        int    `json:"inner_job"`
	UserId          uint   `json:"user_id"`
	NeedCallback    int    `json:"need_callback"`
	SendParam       string `json:"send_param"`
	LabelID         int    `json:"label_id"`
}

func (DBPubStruct) TableName() string {
	return "publish"
}

/*
act config表
*/
type DBActConfigStruct struct {
	gorm.Model
	Aid               int    `json:"aid"`
	Platform          string `json:"platform"`
	PackageConfigId   uint   `json:"package_config_id"`
	ViewUrl           string `json:"view_url"`
	JobUrl            string `json:"job_url"`
	StopUrl           string `json:"stop_url"`
	ActionConfigParam string `json:"action_config_param"`
	Content           string `json:"content"`
	EmailPrefix       string `json:"email_prefix"`
	Level             int    `json:"level"`
	ApiId             string `json:"api_id"`
	ApiToken          string `json:"api_token"`
	InnerJob          int    `json:"inner_job"`
	Describes         string `json:"describes"`
	UserId            uint   `json:"user_id"`
	NeedCallback      int    `json:"need_callback"`
	StageName         string `json:"stage_name"`
	IsSelected        int    `json:"is_selected"`
	MustSelected      int    `json:"must_selected"`
}

func (DBActConfigStruct) TableName() string {
	return "action_config"
}

/*
act表
*/
type DBActStruct struct {
	gorm.Model
	Aid            int    `json:"aid"`
	Platform       string `json:"platform"`
	WorkflowId     uint   `json:"workflow_id"`
	ActionConfigId uint   `json:"action_config_id"`
	PackageId      uint   `json:"package_id"`
	ActionParam    string `json:"action_param"`
	Active         int    `json:"active"`
	Status         int    `json:"status"`
	RetActionInfo  string `json:"ret_action_info"`
	ErrLog         string `json:"err_log"`
	ViewUrl        string `json:"view_url"`
	JobUrl         string `json:"job_url"`
	StopUrl        string `json:"stop_url"`
	EmailPrefix    string `json:"email_prefix"`
	Content        string `json:"content"`
	Level          int    `json:"level"`
	ApiId          string `json:"api_id"`
	ApiToken       string `json:"api_token"`
	JobsysId       string `json:"jobsys_id"`
	Describes      string `json:"describes"`
	InnerJob       int    `json:"inner_job"`
	UserId         uint   `json:"user_id"`
	NeedCallback   int    `json:"need_callback"`
	SendParam      string `json:"send_param"`
	StageName      string `json:"stage_name"`
}

func (DBActStruct) TableName() string {
	return "action"
}

type SubPackageCommon struct {
	gorm.Model
	PackageId    uint   `json:"package_id"`
	Aid          int    `json:"aid"`
	Platform     string `json:"platform"`
	PkgParam     string `json:"pkg_param"`
	RetPkgInfo   string `json:"ret_pkg_info"`
	DownloadUrl  string `json:"download_url"`
	ViewUrl      string `json:"view_url"`
	EmailPrefix  string `json:"email_prefix"`
	Branch       string `json:"branch"`
	OuterVersion string `json:"outer_version"`
	InnerVersion string `json:"inner_version"`
	FiveVersion  string `json:"five_version"`
	Describes    string `json:"describes"`
	LabelID      int    `json:"label_id"`
}

type SubPackage struct {
	SubPackageCommon
}

func (SubPackage) TableName() string {
	return "sub_pakcage"
}

type SubPackageIndex struct {
	SubPackageCommon
}

func (SubPackageIndex) TableName() string {
	return "sub_pakcage_index"
}

type Label struct {
	gorm.Model
	Aid             int    `json:"aid"`
	Platform        string `json:"platform"`
	PermissionLevel string `json:"permission_level"`
	LabelType       int    `json:"label_type"`
	LabelName       string `json:"label_name"`
}

func (Label) TableName() string {
	return "labels"
}
