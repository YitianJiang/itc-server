package _const

const (
	SUCCESS                     = 0
	ERROR                       = -1
	INVALID_PARAMS              = -2
	ERROR_AUTH_CHECK_TOKEN_FAIL = -3
	DB_LOG_MODE                 = true
)
const (
	TOS_BUCKET_NAME = "tos-itc-server"
	TOS_BUCKET_KEY  = "RXFRCE5018AYZNSAUF36"
)
const ROCKETTOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoiZmFuanVhbi54cXAiLCJGdWxsX25hbWUiOiLmqIrlqJ8iLCJleHAiOjE1OTM0ODg2NDgsImlzcyI6InJvY2tldDMuMCIsIm5iZiI6MTU2MTk1MTY0OH0.sBcaZ2mxvxVYb05Z6yD4Wr1asalEYLErSD2qf06NTNY"
const ROCKET_URL = "https://rocket-api.bytedance.net/api/v1/products/query"
const LARK_URL = "https://rocket-api.bytedance.net/api/v1/robot/person/send"
const OFFICE_LARK_URL = "https://open.feishu.cn/open-apis/message/v3/send/"
const LARK_Email2Id_URL = "https://open.feishu.cn/open-apis/user/v3/email2id"

var LowLarkPeople = []string{"kanghuaisong", "fanjuan.xqp", "yinzhihong"}
var MiddleLarkPeople = []string{"kanghuaisong", "fanjuan.xqp", "yinzhihong", "gongrui", "zhangshuai.02"}
var HighLarkPeople = []string{"kanghuaisong", "fanjuan.xqp", "yinzhihong", "gongrui", "zhangshuai.02", "chenyujun"}
var PermLarkPeople = []string{"kanghuaisong", "lirensheng"}
var AppVersionProject = map[string]string{"13": "1", "27": "35", "32": "23", "1128": "375", "1112": "397"}

const (
	APPLE_CERT_DELETE_ADDR   = "https://api.appstoreconnect.apple.com/v1/certificates/"
	APPLE_CREATE_CERT_URL    = "https://api.appstoreconnect.apple.com/v1/certificates"
	APPLE_RECEIVED_DATA_TYPE = "certificates"
	CERT_TYPE_IOS_DEV        = "IOS_DEVELOPMENT"
	CERT_TYPE_IOS_DIST       = "IOS_DISTRIBUTION"
	CERT_TYPE_MAC_DEV        = "MAC_APP_DEVELOPMENT"
	CERT_TYPE_MAC_DIST       = "MAC_APP_DISTRIBUTION"
)
const (
	APPLE_USER_INFO_URL                  = "https://api.appstoreconnect.apple.com/v1/users?limit=200"
	APPLE_USER_INVITED_INFO_URL          = "https://api.appstoreconnect.apple.com/v1/userInvitations?limit=200"
	APPLE_USER_INFO_URL_NO_PARAM         = "https://api.appstoreconnect.apple.com/v1/users"
	APPLE_USER_INFO_URL_GET_HOLDER       = "https://api.appstoreconnect.apple.com/v1/users?filter[roles]=ACCOUNT_HOLDER"
	APPLE_USER_INVITED_INFO_URL_NO_PARAM = "https://api.appstoreconnect.apple.com/v1/userInvitations"
	APPLE_USER_PERM_EDIT_URL             = "https://api.appstoreconnect.apple.com/v1/users"
	APPLE_USER_INVITED_URL               = "https://api.appstoreconnect.apple.com/v1/userInvitations"
)
const (
	APPLE_PROFILE_MANAGER_URL = "https://api.appstoreconnect.apple.com/v1/profiles"
	APPLE_BUNDLE_MANAGER_URL = "https://api.appstoreconnect.apple.com/v1/bundleIds/"
)
const (
	TOS_BUCKET_URL            = "http://tosv.byted.org/obj/staticanalysisresult/"
	TOS_BUCKET_NAME_JYT       = "staticanalysisresult"
	TOS_BUCKET_TOKEN_JYT      = "C5V4TROQGXMCTPXLIJFT"
	TOS_PRIVATE_KEY_URL_DEV   = "http://tosv.byted.org/obj/staticanalysisresult/appleConnectFile/fullFileDevAccept/bytedancett_ies_dev.p12"
	TOS_PRIVATE_KEY_URL_DIST  = "http://tosv.byted.org/obj/staticanalysisresult/appleConnectFile/fullFileAccept/bytedancett_ies_priv.p12"
	TOS_CSR_FILE_FOR_DEV_KEY  = "appleConnectFile/fullFileDevAccept/CertificateSigningRequest.certSigningRequest"
	TOS_CSR_FILE_FOR_DIST_KEY = "appleConnectFile/fullFileAccept/CertificateSigningRequest.certSigningRequest"
	TOS_CSR_FILE_URL_DEV      = "http://tosv.byted.org/obj/staticanalysisresult/appleConnectFile/fullFileDevAccept/CertificateSigningRequest.certSigningRequest"
	TOS_CSR_FILE_URL_DIST     = "http://tosv.byted.org/obj/staticanalysisresult/appleConnectFile/fullFileAccept/CertificateSigningRequest.certSigningRequest"
)

const (
	KANI_APP_ID_AND_SECRET_BASE64  = "MzE5OjU4RDM4NTVGRUVBRjQ0QzJBMEJBNERBRDExRDdFNUEw" //kani 应用id:主secret 对的base64编码
	Certain_Resource_All_PERMS_URL = "http://pmc.bytedance.net/v1/config/permission/query/use/resource?"
	USER_ALL_RESOURCES_PERMS_URL   = "http://pmc.bytedance.net/v1/config/permission/query/user/all?"
	GET_ACCOUNT_ADMIN_LIST_URL     = "http://pmc.bytedance.net/v1/config/resource/query/info?"
	Create_RESOURCE_URL            = "http://pmc.bytedance.net/v1/config/resource/create"
	GIVE_PERMISSION_TO_USER_URL    = "http://pmc.bytedance.net/v1/config/permission/authorize"
)

const APPMAP = `{"data":[{"appName":"今日头条","AppId":13,"doc":"https://docs.bytedance.net/sheet/BT1QQoLQHxE4xACsc3rFug#1"},{"appName":"今日头条极速版","AppId":35,"doc":"https://bytedance.feishu.cn/space/doc/doccnNNOjakRlp1BiWcTxg#"},{"appName":"TB/NR/Babe","AppId":1104,"doc":""},{"appName":"海豚股票","AppId":1182,"doc":""},{"appName":"抖音短视频","AppId":1128,"doc":""},{"appName":"懂车帝","AppId":36,"doc":""},{"appName":"火山小视频","AppId":1112,"doc":""},{"appName":"西瓜视频","AppId":32,"doc":""},{"appName":"半次元","AppId":1250,"doc":""},{"appName":"泡芙社区","AppId":1253,"doc":""},{"appName":"时光相册","AppId":33,"doc":""},{"appName":"好好学习","AppId":1207,"doc":""},{"appName":"皮皮虾","AppId":1319,"doc":"https://bytedance.feishu.cn/space/doc/doccnF1Xa0VB41ldyyHVwW#"},{"appName":"火山小视频极速版","AppId":1350,"doc":""},{"appName":"gogokid","AppId":1277,"doc":""},{"appName":"玩不停","AppId":1288,"doc":""},{"appName":"值点","AppId":1331,"doc":""},{"appName":"小西瓜","AppId":1291,"doc":""},{"appName":"EY","AppId":1249,"doc":""},{"appName":"M","AppId":1811,"doc":"https://docs.bytedance.net/doc/5PzMoH4FDHfZgP7roLKObd"},{"appName":"住小帮","AppId":1398,"doc":""},{"appName":"Hively","AppId":1184,"doc":""},{"appName":"EZ","AppId":1335,"doc":""},{"appName":"头条QA客户端","AppId":9998,"doc":""},{"appName":"EV","AppId":1585,"doc":""},{"appName":"HELO","AppId":1342,"doc":"https://bytedance.feishu.cn/space/doc/doccn3HNNRDp3ptNuYyAce"},{"appName":"飞聊","AppId":1394,"doc":""},{"appName":"多闪","AppId":1349,"doc":""},{"appName":"幸福里","AppId":1370,"doc":""},{"appName":"账号服务端","AppId":9997,"doc":""},{"appName":"线上体验项目","AppId":9996,"doc":""},{"appName":"商业化SDK","AppId":9995,"doc":""},{"appName":"内容云平台","AppId":9994,"doc":""},{"appName":"性能测试平台","AppId":9993,"doc":""},{"appName":"diamond","AppId":9992,"doc":""},{"appName":"AT","AppId":9991,"doc":""},{"appName":"Android基础技术","AppId":9990,"doc":""},{"appName":"客户端技术评审","AppId":9989,"doc":""},{"appName":"趣阅","AppId":1505,"doc":""},{"appName":"ET","AppId":9988,"doc":""},{"appName":"钱包支付","AppId":9987,"doc":""},{"appName":"消费金融","AppId":9986,"doc":""},{"appName":"WebRTC","AppId":1303,"doc":""},{"appName":"财经-保险","AppId":9985,"doc":""},{"appName":"面包金融","AppId":9984,"doc":""},{"appName":"小程序技术项目","AppId":9983,"doc":""},{"appName":"头条QA服务端","AppId":9982,"doc":""},{"appName":"无线研发平台","AppId":9981,"doc":""},{"appName":"iOS编译技术","AppId":9980,"doc":""},{"appName":"Faceu","AppId":10001,"doc":""},{"appName":"轻颜相机","AppId":150121,"doc":""},{"appName":"视频云-视频点播","AppId":9979,"doc":""},{"appName":"字节云","AppId":9978,"doc":""},{"appName":"EO","AppId":9977,"doc":""},{"appName":"视频云-视频直播","AppId":9976,"doc":""},{"appName":"BuzzVideo","AppId":1131,"doc":""},{"appName":"审核平台","AppId":10002,"doc":""},{"appName":"lark","AppId":1161,"doc":""},{"appName":"xs","AppId":9975,"doc":""},{"appName":"开言Feed","AppId":1638,"doc":"https://wiki.bytedance.net/display/EZ/Feed+APP"},{"appName":"Tiktok","AppId":1180,"doc":""},{"appName":"musical.ly","AppId":1233,"doc":""},{"appName":"QA基础技术","AppId":9974,"doc":""},{"appName":"组件化示例","AppId":9973,"doc":""},{"appName":"头条小视频","AppId":9972,"doc":""},{"appName":"机器人TODO汇总","AppId":9971,"doc":""},{"appName":"D","AppId":9970,"doc":""},{"appName":"L","AppId":9969,"doc":""},{"appName":"Xplus","AppId":10006,"doc":""},{"appName":"用户中心","AppId":1661,"doc":""},{"appName":"AILab","AppId":10005,"doc":""},{"appName":"H","AppId":1691,"doc":""},{"appName":"直播中台","AppId":99986,"doc":""},{"appName":"EffectSDK","AppId":9999,"doc":""},{"appName":"VESDK","AppId":10000,"doc":""},{"appName":"UG中台","AppId":10003,"doc":""},{"appName":"字节SDK海外版","AppId":1782,"doc":""},{"appName":"字节SDK国内版","AppId":1781,"doc":""},{"appName":"视频云","AppId":10007,"doc":""},{"appName":"安全","AppId":10008,"doc":""},{"appName":"直播内容安全","AppId":10009,"doc":""},{"appName":"鲜时光","AppId":1840,"doc":""},{"appName":"直播底层服务","AppId":10010,"doc":""},{"appName":"财经DL业务","AppId":9968,"doc":""},{"appName":"国际支付","AppId":9967,"doc":""},{"appName":"财经QA","AppId":10011,"doc":""},{"appName":"EA","AppId":1686,"doc":""},{"appName":"EM","AppId":1700,"doc":""},{"appName":"T-game","AppId":1807,"doc":""},{"appName":"F-game","AppId":1865,"doc":""},{"appName":"H1-game","AppId":1870,"doc":""},{"appName":"Y-game","AppId":1875,"doc":""},{"appName":"视频云-实时通信","AppId":10012,"doc":""},{"appName":"Vigo Video","AppId":1145,"doc":""},{"appName":"R项目内容安全","AppId":10013,"doc":""},{"appName":"AILab US","AppId":10014,"doc":""},{"appName":"V","AppId":1873,"doc":""},{"appName":"Pick","AppId":1778,"doc":""},{"appName":"锤子商城","AppId":10000011,"doc":""},{"appName":"Tiktok-Lite","AppId":1339,"doc":""},{"appName":"Musicaly-Lite","AppId":1340,"doc":""},{"appName":"指令平台","AppId":10015,"doc":""},{"appName":"Vigo Lite","AppId":1257,"doc":""},{"appName":"幸福客","AppId":1488,"doc":""},{"appName":"DR","AppId":1967,"doc":""},{"appName":"Enterprise Intelligence","AppId":10016,"doc":""},{"appName":"Automation LarkFlow","AppId":10017,"doc":""},{"appName":"无限火力","AppId":99981,"doc":""},{"appName":"EXO","AppId":1884,"doc":""}],"errorCode":0,"message":"success"}`

//苹果apple connect后台管理常参
const (
	ICLOUD          = "ICLOUD"
	DATA_PROTECTION = "DATA_PROTECTION"
)
const (
	Organization = "Organization"
	Enterprise   = "Enterprise"
)
const (
	IOS_APP_STORE       = "IOS_APP_STORE"
	IOS_APP_INHOUSE     = "IOS_APP_INHOUSE"
	MAC_APP_STORE       = "MAC_APP_STORE"
	IOS_APP_ADHOC       = "IOS_APP_ADHOC"
	IOS_APP_DEVELOPMENT = "IOS_APP_DEVELOPMENT"
	MAC_APP_DEVELOPMENT = "MAC_APP_DEVELOPMENT"
)

var IOSSelectCapabilities = []string{"ACCESS_WIFI_INFORMATION", "APP_GROUPS", "ASSOCIATED_DOMAINS", "AUTOFILL_CREDENTIAL_PROVIDER", "CLASSKIT", "GAME_CENTER",
	"HEALTHKIT", "HOMEKIT", "HOT_SPOT", "IN_APP_PURCHASE", "INTER_APP_AUDIO", "MULTIPATH", "NETWORK_EXTENSIONS", "NFC_TAG_READING", "PERSONAL_VPN",
	"PUSH_NOTIFICATIONS", "SIRIKIT", "WALLET", "WIRELESS_ACCESSORY_CONFIGURATION"}
var MacSelectCapabilities = []string{"ASSOCIATED_DOMAINS", "NETWORK_EXTENSIONS", "PERSONAL_VPN", "PUSH_NOTIFICATIONS"}
var CloudSettings = []string{"XCODE_6", "XCODE_5"}
var ProtectionSettings = []string{"COMPLETE_PROTECTION", "PROTECTED_UNLESS_OPEN", "PROTECTED_UNTIL_FIRST_USER_AUTH"}

var RolesInfoMap = map[string]string{"账号持有者": "ACCOUNT_HOLDER", "管理": "ADMIN", "财务": "FINANCE", "App管理": "APP_MANAGER", "开发人员": "DEVELOPER", "营销": "MARKETING", "销售": "SALES", "用户支持": "CUSTOMER_SUPPORT"}
var RolesIndexList = []string{"ACCOUNT_HOLDER", "ADMIN", "FINANCE", "APP_MANAGER", "DEVELOPER", "MARKETING", "SALES", "CUSTOMER_SUPPORT"}
var PermsMap = map[string]string{"user_manager": "user_manager", "all_cert_manager": "all_cert_manager", "dev_cert_manager": "dev_cert_manager"}
var IOSSelectCapabilitiesMap = map[string]string{
	"ACCESS_WIFI_INFORMATION":          "ACCESS_WIFI_INFORMATION",
	"APP_GROUPS":                       "APP_GROUPS",
	"ASSOCIATED_DOMAINS":               "ASSOCIATED_DOMAINS",
	"AUTOFILL_CREDENTIAL_PROVIDER":     "AUTOFILL_CREDENTIAL_PROVIDER",
	"CLASSKIT":                         "CLASSKIT",
	"GAME_CENTER":                      "GAME_CENTER",
	"HEALTHKIT":                        "HEALTHKIT",
	"HOMEKIT":                          "HOMEKIT",
	"HOT_SPOT":                         "HOT_SPOT",
	"IN_APP_PURCHASE":                  "IN_APP_PURCHASE",
	"INTER_APP_AUDIO":                  "INTER_APP_AUDIO",
	"MULTIPATH":                        "MULTIPATH",
	"NETWORK_EXTENSIONS":               "NETWORK_EXTENSIONS",
	"NFC_TAG_READING":                  "NFC_TAG_READING",
	"PERSONAL_VPN":                     "PERSONAL_VPN",
	"PUSH_NOTIFICATIONS":               "PUSH_NOTIFICATIONS",
	"SIRIKIT":                          "SIRIKIT",
	"WALLET":                           "WALLET",
	"WIRELESS_ACCESSORY_CONFIGURATION": "WIRELESS_ACCESSORY_CONFIGURATION",
}
