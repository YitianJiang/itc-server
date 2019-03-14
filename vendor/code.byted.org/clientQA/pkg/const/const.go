package _const

const (
	Db_error = 0
)

const (
	Deactive_num = 0
	Active_num   = 1
)

const (
	Ret_success = 0
)

const (
	Status_unknown = 0
	Status_fail    = 1
	Status_success = 2
)

const (
	Platform_Android = "android"
	Platform_iOS     = "ios"
)

const (
	Use_User = "1"
)

//job的类型
const (
	Flow_Job = 0
	Pkg_Job  = 1
	Pub_Job  = 2
	Act_Job  = 3
)

//按jenkins按钮的类型
const (
	AbortButton   = 0
	ProceedButton = 1
)

//内置的pub和act为0
const (
	InnerJob = 1
	OtherJob = 0
)

//内部回调发送给job helper的字符串
const (
	Pub_callback_name = "pubcallback"
	Act_callback_name = "actcallback"
)

const (
	To_tos = 0
	To_cdn = 1
)

const (
	Tos_bucket_name = "iOSPackageBackUp"
	Tos_bucket_key  = "V1IIUCA4LME9VU4NGQY3"
)

const (
	Cdn_bucket_name = "rocketpackagebackup"
	Cdn_bucket_key  = "VZ4Q5IK200IZXMVY55R5"
)

const (
	ArchBucketName = "toutiao.ios.arch"
	ArchBucketKey  = "MJMETJODXZF7FZLFY3VT"
)

const (
	Concern    = 1
	NotConcern = 0
)

const (
	Show_on_page    = "1"
	Notshow_on_page = "0"
)

const (
	PkgPubHistoryPageSize = 10
)

//const (
//	PrePkgStage = 0
//	PkgStage    = 1
//	PrePubStage = 2
//	PubStage    = 3
//)

const (
	EnableAutoPkg = 1
)

const (
	MinPrePkgStage = 0
	MaxPrePkgStage = 9
	PkgStage       = 10
	MinPrePubStage = 11
	MaxPrePubStage = 19
	PubStage       = 20
	PkgStageName   = "打包阶段"
	PubStageName   = "发布阶段"
)

const (
	PreBuildPkg = "0"
	FeaturePkg  = "3"
	RDGreyPkg   = "5"
	GreyPkg     = "8"
	OfficialPkg = "10"
)

const (
	AggregateNone  = 1 //不聚合
	AggregateByVer = 2 //通过版本聚合
	AggregateAll   = 3 //聚合成1个

)

//给产品线添加group概念
const (
	TOUTIAO = 1 //toutiao/toutiaolite/tt_small_game
	I18N    = 2
	XIGUA   = 3
	AWEME   = 4
	HOTSOON = 5
	CAR     = 6
	LEARN   = 7 //好好学习
	EX      = 8
	AIKID   = 9
	EZ      = 10
	S       = 11
	F       = 12
	J       = 13
)

const (
	UnecessaryCallback = 0
)

const Inner_job_helper_url = "https://ci.bytedance.net/job/inner_job_callback_helper/buildWithParameters"

//jenkins实名调用
const UserID = "zhangshuai.02"
const ApiToken = "0a120405ccce80c218ad6c05dcb9870e"

//svn的basic auth
const Svn_basic_auth = "Z291aG9uZ3l1QGJ5dGVkYW5jZS5jb206R291aG9uZ3l1XzE5OTE="

//灰度后台app admin的token
const Gray_token string = "YjMxMmVlYTE1OWUyNzBiNzI4YTg4MzJjODVkMTBjYzg="

//gitlab查询branch信息
const Gitlab_branch_url string = "https://code.byted.org/api/v4/projects/%s/repository/branches?search=%s"

//gitlab查询commit信息
const Gitlab_commit_url string = "https://code.byted.org/api/v4/projects/%s/repository/commits"

//gitlab的private token
const Gitlab_private_token = "eJtCr29dQZZEEhHhGP18"

//plist文件存放的地址
const Plist_url string = "http://10.8.78.136:8011/uploadplist/"

//钉钉接口拿到个人私密信息
const Dingavatar string = "http://10.8.78.136:8011/dingquery/"

//判断个人是否有秘密项目的权限
const Permission_all_url string = "https://ee.byted.org/ratak/employees/%s/permissions/"

//user token
const Lark_user_token = "u-725a4ebd-483d-4175-9ca1-16920f94dc17"

//lark发消息的地址
const Larknotice_url = "https://oapi.zjurl.cn/open-apis/api/v2/message"

//lark查询user id
const Lark_finduser_url = "https://oapi.zjurl.cn/open-apis/api/v1/user.user_id"

const Lark_openim_url = "https://oapi.zjurl.cn/open-apis/api/v1/im.open"

const Lark_upload_img = "https://oapi.zjurl.cn/open-apis/api/v2/message/image/upload"

//机器人的token
const Robot_token = "b-f1a766be-c32f-44ed-a784-081f83df2f43"

//查看灰度的地址
const GreyViewUrl = "https://app-admin.bytedance.net/static/main/index.html#/gray_release/update_gray_release?id=%d"

//获取已有包使用的git地址
const TTMAIN_GIT_PATH = "git@code.byted.org:tt_android/ttmain.git"
const COMMON_GIT_PATH = "git@code.byted.org:tt_android/common_business.git"
const CI_URL = "https://ci.bytedance.net/"

const (
	Kani_appid  string = "319"
	Kani_apppwd string = "58D3855FEEAF44C2A0BA4DAD11D7E5A0"
)

const (
	Pkg_callback string = "http://101.bytedance.net/v1/dispatch/pkgcallback"
	Pub_callback string = "http://101.bytedance.net/v1/dispatch/pubcallback"
	Act_callback string = "http://101.bytedance.net/v1/dispatch/actcallback"
	//Act_callback string = "http://10.8.163.168:9872/v1/dispatch/actcallback"
)

const (
	LarkGroupOne  = "0"
	LarkGroupAll  = "1"
	LarkGroupNone = "2"
)

const (
	InnerTestName = "内测稳定包"
)
