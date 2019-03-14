package http_util

import (
	"code.byted.org/clientQA/pkg/job-processor/dal"
	"code.byted.org/clientQA/pkg/request-processor/request-dal"
)

/*
post startflow请求中接收的数据
*/
type PostStartFlowStruct struct {
	PackageConfigId  int                           `json:"package_config_id"`
	Aid              int                           `json:"aid"`
	Platform         string                        `json:"platform"`
	UseUser          string                        `json:"use_user"`
	User             string                        `json:"user"`
	PkgParam         string                        `json:"pkg_param"`
	VersionNum       string                        `json:"version_num"`
	UpdateVersionNum string                        `json:"update_version_num"`
	Version          int                           `json:"version"`
	PublishDiscribe  string                        `json:"publish_discribe"`
	Active           int                           `json:"active"`
	PublishConfigs   []PostPubAndCheckConfigStruct `json:"publish_configs"`
	CheckConfigs     []PostPubAndCheckConfigStruct `json:"check_configs"`
	ViewUrl          string                        `json:"view_url"`
	FiveVersion      string                        `json:"five_version"`
	OuterVersion     string                        `json:"outer_version"`
	InnerVersion     string                        `json:"inner_version"`
	Channel          string                        `json:"CHANNEL"`
	CommitID         string                        `json:"COMMIT_ID"`
	WorkflowID       string                        `json:"workflow_id"`
	AutoPkg          int                           `json:"auto_pkg"`
}

type PostPubAndCheckConfigStruct struct {
	CommonConfigStruct
	//Id              uint   `json:"id"`
	Param    string `json:"param"`
	ParamNum int    `json:"param_num"`
	//Content         string `json:"content"`
	ViewLink string `json:"view_link"`
	Link     string `json:"link"`
	StopLink string `json:"stop_link"`
	//Level           int    `json:"level"`
	//InnerJob        int    `json:"inner_job"`
	//Describes       string `json:"describes"`
	//ApiId           string `json:"api_id"`
	//ApiToken        string `json:"api_token"`
	//NeedCallback    int    `json:"need_callback"`
	IsSelected      int    `json:"is_selected"`
	MustSelected    int    `json:"must_selected"`
	StageName       string `json:"stage_name"`
	LabelID         int    `json:"label_id"`
	LabelName       string `json:"label_name"`
	LabelType       int    `json:"label_type"`
	PermissionLevel string `json:"permission_level"`
}

type PostCheckConfigStruct struct {
	CommonConfigStruct
	//Id                uint   `json:"id"`
	ActionConfigParam string `json:"action_config_param"`
	//Content           string `json:"content"`
	ViewLink string `json:"view_url"`
	Link     string `json:"job_url"`
	StopLink string `json:"stop_url"`
	//Level             int    `json:"level"`
	//InnerJob          int    `json:"inner_job"`
	//Describes         string `json:"describes"`
	//ApiId             string `json:"api_id"`
	//ApiToken          string `json:"api_token"`
	//NeedCallback      int    `json:"need_callback"`
	StageName    string `json:"stage_name"`
	IsSelected   int    `json:"is_selected"`
	MustSelected int    `json:"must_selected"`
}

type PostPubConfigStruct struct {
	CommonConfigStruct
	//Id                uint   `json:"id"`
	PublicConfigParam string `json:"public_config_param"`
	//Content           string `json:"content"`
	ViewLink string `json:"view_url"`
	Link     string `json:"job_url"`
	StopLink string `json:"stop_url"`
	//Level             int    `json:"level"`
	//InnerJob          int    `json:"inner_job"`
	//Describes         string `json:"describes"`
	//ApiId             string `json:"api_id"`
	//ApiToken          string `json:"api_token"`
	//NeedCallback      int    `json:"need_callback"`
	LabelID         int    `json:"label_id"`
	LabelName       string `json:"label_name"`
	LabelType       int    `json:"label_type"`
	PermissionLevel string `json:"permission_level"`
}

type CommonConfigStruct struct {
	Id           uint   `json:"id"`
	Content      string `json:"content"`
	Level        int    `json:"level"`
	InnerJob     int    `json:"inner_job"`
	Describes    string `json:"describes"`
	ApiId        string `json:"api_id"`
	ApiToken     string `json:"api_token"`
	NeedCallback int    `json:"need_callback"`
}

type CallBackStruct struct {
	CallBackStructV2
	JobId      string `json:"job_id"`
	WorkflowId string `json:"workflow_id"`
}

type CallBackStructV2 struct {
	Status string            `json:"status"`
	Result map[string]string `json:"result"`
}

//配置信息
type Config struct {
	Aid             int    `json:"aid"`
	Platform        string `json:"platform"`
	ViewLink        string `json:"view_link"`
	PkgLink         string `json:"pkg_link"`
	StopLink        string `json:"stop_link"`
	PkgConfigId     uint   `json:"pkg_config_id"`
	PkgContent      string `json:"pkg_content"`
	AllowBranch     string `json:"allow_branch"`
	PkgDefaultParam string `json:"pkg_default_param"`
	ApiId           string `json:"api_id"`
	ApiToken        string `json:"api_token"`
	PermissionLevel string `json:"permission_level"`
	//这个可以考虑不要
	//DefaultParam string `json:"default_param"`
	PublishConfigs []PostPubAndCheckConfigStruct `json:"publish_configs"`
	CheckConfigs   []PostPubAndCheckConfigStruct `json:"check_configs"`
}

//返回的配置信息
type PostConfig struct {
	Aid             int    `json:"aid"`
	Platform        string `json:"platform"`
	ViewLink        string `json:"view_link"`
	PkgLink         string `json:"pkg_link"`
	StopLink        string `json:"stop_link"`
	PkgConfigId     uint   `json:"pkg_config_id"`
	PkgContent      string `json:"pkg_content"`
	AllowBranch     string `json:"allow_branch"`
	PkgDefaultParam string `json:"pkg_default_param"`
	ApiId           string `json:"api_id"`
	ApiToken        string `json:"api_token"`
	PermissionLevel string `json:"permission_level"`

	//这个可以考虑不要
	//DefaultParam string `json:"default_param"`
	PublishConfigs []PostPubConfigStruct   `json:"publish_configs"`
	CheckConfigs   []PostCheckConfigStruct `json:"check_configs"`
}

type NewProducts struct {
	request_dal.ProductInfo
	WorkflowName string                            `json:"workflow_name"`
	FromAid      string                            `json:"from_aid"`
	Publishes    []dal.DBDefaultPubConfigStruct    `json:"publishes"`
	Actions      []dal.DBDefaultActionConfigStruct `json:"actions"`
}

type ProductConfig struct {
	dal.DBPkgConfigStruct
	UserName  string `json:"user_name"`
	Publishes string `json:"publishes"`
	Actions   string `json:"actions"`
}

//gitlab中获取commit的struct
type GitlabCommitStruct struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

//gitlab中获取branch信息的struct，包含了commit信息
type GitlabBranchStruct struct {
	Name   string `json:"name"`
	Commit GitlabCommitStruct
}

type BranchesStruct struct {
	CommitId    string `json:"commit_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

/*
lark返回的user结构提
*/
type LarkUserStruct struct {
	Ok      bool   `json:"ok"`
	User_id string `json:"user_id"`
}

type IMStruct struct {
	Channel IMChannelStruct `json:"channel"`
	Ok      bool            `json:"ok"`
}

type IMChannelStruct struct {
	Id string `json:"id"`
}

//获取已有任务使用
type Build_result struct {
	Id              string   `json:"id"`
	Result          string   `json:"result"`
	Building        bool     `json:"building"`
	Branch          string   `json:"branch"`
	PkgUrl          []string `json:"pkg_url"`
	Commit_id       string   `json:"COMMIT_ID"`
	Commitid_common string   `json:"COMMITID_COMMON"`
	Channel         string   `json:"CHANNEL"`
	ViewUrl         string   `json:"view_url"`
	InnerVersion    string   `json:"inner_version"`
	FiveVersion     string   `json:"five_version"`
	OuterVersion    string   `json:"outer_version"`
	Job_url         string   `json:"job_url"`
}

type RecentPackage_Build_result struct {
	Id              string `json:"id"`
	Branch          string `json:"branch"`
	PkgUrl          string `json:"pkg_url"`
	Commit_id       string `json:"COMMIT_ID"`
	Commitid_common string `json:"COMMITID_COMMON"`
	Channel         string `json:"CHANNEL"`
	ViewUrl         string `json:"view_url"`
	InnerVersion    string `json:"inner_version"`
	FiveVersion     string `json:"five_version"`
	OuterVersion    string `json:"outer_version"`
	PublishAction   string `json:"publish_action"`
	Result          string `json:"result"`
	Building        string `json:"building"`
	Publish         string `json:"publish"`
	Job_url         string `json:"job_url"`
}

type Ci_result_str struct {
	Id          string
	Result      string
	Url         string
	Actions     []Action_str
	Artifacts   []Artifact_str
	Description string
	Building    bool
}

type Action_str struct {
	BuildsByBranchName BuildsByBranchName_str
	LastBuiltRevision  LastBuiltRevision_str
	RemoteUrls         []string
	Parameters         []Parameter_str
}

type Parameter_str struct {
	Name  string
	Value string
}

type BuildsByBranchName_str struct {
}

type LastBuiltRevision_str struct {
	SHA1   string
	Branch []Branch_str
}

type Branch_str struct {
	SHA1 string
	Name string
}

type Artifact_str struct {
	DisplayPath  string
	FileName     string
	RelativePath string
}

type RetUserInfo struct {
	request_dal.Struct_User
	PermissionName string `json:"permission_name"`
}

type Mail struct {
	Content   string   `json:"content"`
	Subject   string   `json:"subject"`
	To_list   []string `json:"to_list"`
	From_name string   `json:"from_name"`
	From_mail string   `json:"from_mail"`
}

type PushEvent struct {
	Object_kind       string                //两种都有，为push or merge_request
	Object_attributes object_attributesStru //merge_request有的字段
	Ref               string                //push有的字段
	Project           projectStru           //push merge_request都有
	Project_id        uint
	Commits           []commitsStu //push才有
	Checkout_sha      string       // commit id
	User_email        string       // 用户邮箱
}

type object_attributesStru struct {
	Target_branch string
	Source_branch string
	State         string
	Merge_status  string
	Last_Commit   last_commitStru
}

type last_commitStru struct {
	Message string
	Author  commitAuthorStru
}

type projectStru struct {
	Name string
}

type commitsStu struct {
	Message  string
	Modified []string
	Url      string
	Author   commitAuthorStru
}

type commitAuthorStru struct {
	Name  string
	Email string
}

/*
/history/pipeline/one返回的接口信息
*/
type PipelineJobInfo struct {
	Aid        int                   `json:"aid"`
	Platform   string                `json:"platform"`
	WorkflowID int                   `json:"workflow_id"`
	JobInfo    []request_dal.JobInfo `json:"job_info"`
	Describe   string                `json:"describe"`
}
