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

var LowLarkPeople = []string{"kanghuaisong", "fanjuan.xqp", "yinzhihong"}
var MiddleLarkPeople = []string{"kanghuaisong", "fanjuan.xqp", "yinzhihong", "gongrui", "zhangshuai.02"}
var HighLarkPeople = []string{"kanghuaisong", "fanjuan.xqp", "yinzhihong", "gongrui", "zhangshuai.02", "chenyujun"}


//苹果apple connect后台管理常参
const (
	ICLOUD  = "ICLOUD"
	DATA_PROTECTION = "DATA_PROTECTION"
)
var IOSSelectCapabilities = []string{ "ACCESS_WIFI_INFORMATION", "APP_GROUPS", "ASSOCIATED_DOMAINS", "AUTOFILL_CREDENTIAL_PROVIDER", "CLASSKIT", "GAME_CENTER",
	"HEALTHKIT", "HOMEKIT", "HOT_SPOT", "IN_APP_PURCHASE", "INTER_APP_AUDIO", "MULTIPATH", "NETWORK_EXTENSIONS", "NFC_TAG_READING", "PERSONAL_VPN",
	"PUSH_NOTIFICATIONS", "SIRIKIT", "WALLET", "WIRELESS_ACCESSORY_CONFIGURATION" }
var MacSelectCapabilities = []string{ "ASSOCIATED_DOMAINS", "NETWORK_EXTENSIONS", "PERSONAL_VPN", "PUSH_NOTIFICATIONS" }
var CloudSettings = []string{"XCODE_6","XCODE_5"}
var ProtectionSettings = []string{"COMPLETE_PROTECTION","PROTECTED_UNLESS_OPEN","PROTECTED_UNTIL_FIRST_USER_AUTH"}