package _const

import (
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	SUCCESS        = 0
	ERROR          = -1
	INVALID_PARAMS = -2
	ERROR_AUTH_CHECK_TOKEN_FAIL    = -3
	DB_LOG_MODE = true
)
const (
	TOS_BUCKET_NAME = "tos-itc-server"
	TOS_BUCKET_KEY = "RXFRCE5018AYZNSAUF36"
)
const ROCKETTOKEN  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoiZmFuanVhbi54cXAiLCJGdWxsX25hbWUiOiLmqIrlqJ8iLCJleHAiOjE1OTM0ODg2NDgsImlzcyI6InJvY2tldDMuMCIsIm5iZiI6MTU2MTk1MTY0OH0.sBcaZ2mxvxVYb05Z6yD4Wr1asalEYLErSD2qf06NTNY"
const ROCKET_URL = "https://rocket-api.bytedance.net/api/v1/products/query"
const APPMAP = `{"data":[{"appName":"今日头条","AppId":13,"doc":"https://docs.bytedance.net/sheet/BT1QQoLQHxE4xACsc3rFug#1"},{"appName":"今日头条极速版","AppId":35,"doc":"https://bytedance.feishu.cn/space/doc/doccnNNOjakRlp1BiWcTxg#"},{"appName":"TB/NR/Babe","AppId":1104,"doc":""},{"appName":"海豚股票","AppId":1182,"doc":""},{"appName":"抖音短视频","AppId":1128,"doc":""},{"appName":"懂车帝","AppId":36,"doc":""},{"appName":"火山小视频","AppId":1112,"doc":""},{"appName":"西瓜视频","AppId":32,"doc":""},{"appName":"半次元","AppId":1250,"doc":""},{"appName":"泡芙社区","AppId":1253,"doc":""},{"appName":"时光相册","AppId":33,"doc":""},{"appName":"好好学习","AppId":1207,"doc":""},{"appName":"皮皮虾","AppId":1319,"doc":"https://bytedance.feishu.cn/space/doc/doccnF1Xa0VB41ldyyHVwW#"},{"appName":"火山小视频极速版","AppId":1350,"doc":""},{"appName":"gogokid","AppId":1277,"doc":""},{"appName":"玩不停","AppId":1288,"doc":""},{"appName":"值点","AppId":1331,"doc":""},{"appName":"小西瓜","AppId":1291,"doc":""},{"appName":"EY","AppId":1249,"doc":""},{"appName":"M","AppId":1811,"doc":"https://docs.bytedance.net/doc/5PzMoH4FDHfZgP7roLKObd"},{"appName":"住小帮","AppId":1398,"doc":""},{"appName":"Hively","AppId":1184,"doc":""},{"appName":"EZ","AppId":1335,"doc":""},{"appName":"头条QA客户端","AppId":9998,"doc":""},{"appName":"EV","AppId":1585,"doc":""},{"appName":"HELO","AppId":1342,"doc":"https://bytedance.feishu.cn/space/doc/doccn3HNNRDp3ptNuYyAce"},{"appName":"飞聊","AppId":1394,"doc":""},{"appName":"多闪","AppId":1349,"doc":""},{"appName":"幸福里","AppId":1370,"doc":""},{"appName":"账号服务端","AppId":9997,"doc":""},{"appName":"线上体验项目","AppId":9996,"doc":""},{"appName":"商业化SDK","AppId":9995,"doc":""},{"appName":"内容云平台","AppId":9994,"doc":""},{"appName":"性能测试平台","AppId":9993,"doc":""},{"appName":"diamond","AppId":9992,"doc":""},{"appName":"AT","AppId":9991,"doc":""},{"appName":"Android基础技术","AppId":9990,"doc":""},{"appName":"客户端技术评审","AppId":9989,"doc":""},{"appName":"趣阅","AppId":1505,"doc":""},{"appName":"ET","AppId":9988,"doc":""},{"appName":"钱包支付","AppId":9987,"doc":""},{"appName":"消费金融","AppId":9986,"doc":""},{"appName":"WebRTC","AppId":1303,"doc":""},{"appName":"财经-保险","AppId":9985,"doc":""},{"appName":"面包金融","AppId":9984,"doc":""},{"appName":"小程序技术项目","AppId":9983,"doc":""},{"appName":"头条QA服务端","AppId":9982,"doc":""},{"appName":"无线研发平台","AppId":9981,"doc":""},{"appName":"iOS编译技术","AppId":9980,"doc":""},{"appName":"Faceu","AppId":10001,"doc":""},{"appName":"轻颜相机","AppId":150121,"doc":""},{"appName":"视频云-视频点播","AppId":9979,"doc":""},{"appName":"字节云","AppId":9978,"doc":""},{"appName":"EO","AppId":9977,"doc":""},{"appName":"视频云-视频直播","AppId":9976,"doc":""},{"appName":"BuzzVideo","AppId":1131,"doc":""},{"appName":"审核平台","AppId":10002,"doc":""},{"appName":"lark","AppId":1161,"doc":""},{"appName":"xs","AppId":9975,"doc":""},{"appName":"开言Feed","AppId":1638,"doc":"https://wiki.bytedance.net/display/EZ/Feed+APP"},{"appName":"Tiktok","AppId":1180,"doc":""},{"appName":"musical.ly","AppId":1233,"doc":""},{"appName":"QA基础技术","AppId":9974,"doc":""},{"appName":"组件化示例","AppId":9973,"doc":""},{"appName":"头条小视频","AppId":9972,"doc":""},{"appName":"机器人TODO汇总","AppId":9971,"doc":""},{"appName":"D","AppId":9970,"doc":""},{"appName":"L","AppId":9969,"doc":""},{"appName":"Xplus","AppId":10006,"doc":""},{"appName":"用户中心","AppId":1661,"doc":""},{"appName":"AILab","AppId":10005,"doc":""},{"appName":"H","AppId":1691,"doc":""},{"appName":"直播中台","AppId":99986,"doc":""},{"appName":"EffectSDK","AppId":9999,"doc":""},{"appName":"VESDK","AppId":10000,"doc":""},{"appName":"UG中台","AppId":10003,"doc":""},{"appName":"字节SDK海外版","AppId":1782,"doc":""},{"appName":"字节SDK国内版","AppId":1781,"doc":""},{"appName":"视频云","AppId":10007,"doc":""},{"appName":"安全","AppId":10008,"doc":""},{"appName":"直播内容安全","AppId":10009,"doc":""},{"appName":"鲜时光","AppId":1840,"doc":""},{"appName":"直播底层服务","AppId":10010,"doc":""},{"appName":"财经DL业务","AppId":9968,"doc":""},{"appName":"国际支付","AppId":9967,"doc":""},{"appName":"财经QA","AppId":10011,"doc":""},{"appName":"EA","AppId":1686,"doc":""},{"appName":"EM","AppId":1700,"doc":""},{"appName":"T-game","AppId":1807,"doc":""},{"appName":"F-game","AppId":1865,"doc":""},{"appName":"H1-game","AppId":1870,"doc":""},{"appName":"Y-game","AppId":1875,"doc":""},{"appName":"视频云-实时通信","AppId":10012,"doc":""},{"appName":"Vigo Video","AppId":1145,"doc":""},{"appName":"R项目内容安全","AppId":10013,"doc":""},{"appName":"AILab US","AppId":10014,"doc":""},{"appName":"V","AppId":1873,"doc":""},{"appName":"Pick","AppId":1778,"doc":""},{"appName":"锤子商城","AppId":10000011,"doc":""},{"appName":"Tiktok-Lite","AppId":1339,"doc":""},{"appName":"Musicaly-Lite","AppId":1340,"doc":""},{"appName":"指令平台","AppId":10015,"doc":""},{"appName":"Vigo Lite","AppId":1257,"doc":""},{"appName":"幸福客","AppId":1488,"doc":""},{"appName":"DR","AppId":1967,"doc":""},{"appName":"Enterprise Intelligence","AppId":10016,"doc":""},{"appName":"Automation LarkFlow","AppId":10017,"doc":""},{"appName":"无限火力","AppId":99981,"doc":""},{"appName":"EXO","AppId":1884,"doc":""}],"errorCode":0,"message":"success"}`

//获取APPID和APPNAME的转换map
func GetAPPMAP() map[int]string {
	var mapInfo map[string]interface{}
	json.Unmarshal([]byte(APPMAP),&mapInfo)
	appList := mapInfo["data"].([]interface{})
	var AppIdMap = make(map[int]string)
	for _,appI := range appList{
		app := appI.(map[string]interface{})
		AppIdMap[int(app["AppId"].(float64))] = app["appName"].(string)
	}
	return AppIdMap
}

func NewGetAppMap() map[int]string  {
	var mapInfo map[string]interface{}
	appMaps := Get(ROCKET_URL)
	if appMaps == nil {
		return make(map[int]string)
	}
	json.Unmarshal(appMaps,&mapInfo)
	appList := mapInfo["data"].([]interface{})
	var AppIdMap = make(map[int]string)
	for _,appI := range appList{
		app := appI.(map[string]interface{})
		AppIdMap[int(app["AppId"].(float64))] = app["appName"].(string)
	}
	//fmt.Sprint(AppIdMap)
	return AppIdMap
}

//获取get请求
func Get(url string) []byte {
	client := &http.Client{}
	request,err := http.NewRequest("GET",url,nil)
	request.Header.Add("token",ROCKETTOKEN)
	if err != nil {
		logs.Error("获取rocket项目信息失败,%v",err)
		utils.LarkDingOneInner("fanjuan.xqp","获取rocket项目信息失败")
		utils.LarkDingOneInner("kanghuaisong","获取rocket项目信息失败")
		utils.LarkDingOneInner("yinzhihong","获取rocket项目信息失败")
		return nil
	}
	resp,err2 := client.Do(request)
	if err2 != nil {
		logs.Error("获取rocket项目信息失败,%v",err)
		utils.LarkDingOneInner("fanjuan.xqp","获取rocket项目信息失败")
		utils.LarkDingOneInner("kanghuaisong","获取rocket项目信息失败")
		utils.LarkDingOneInner("yinzhihong","获取rocket项目信息失败")
		return nil
	}
	defer resp.Body.Close()
	body,_ := ioutil.ReadAll(resp.Body)
	//logs.Notice("获取app返回信息："+fmt.Sprint(string(body)))
	return  body
}

