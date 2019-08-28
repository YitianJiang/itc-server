package utils

import (
	"bytes"
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const DETECT_URL_PRO = "10.1.221.188:9527"

//发送http post请求，其中rbody是一个json串
func PostJsonHttp(url string, rbody []byte) (int, []byte) {
	http.DefaultClient.Timeout = 3 * time.Second
	bodyBuffer := bytes.NewBuffer([]byte(rbody))
	resp, err := http.Post(url, "application/json;charset=utf-8", bodyBuffer)
	if err != nil {
		return -1, nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -2, nil
	}
	return 0, body
}
func PostLocalFileWithParams(params map[string]string, postfilename string, fileName string, api_url string) (string, error) {
	currentpath, _ := os.Getwd()
	fullfilepath := currentpath + "/" + fileName

	///////发送文件部分
	//发送文件
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	for key, value := range params {
		bodyWriter.WriteField(key, value)
	}
	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile(postfilename, fullfilepath)
	if err != nil {
		logs.Error("%s", "error writing to buffer："+err.Error())
		return "error writing to buffer", err
	}

	//打开文件句柄操作
	fh, err := os.Open(fullfilepath)
	if err != nil {
		logs.Error("%s", "error opening file："+err.Error())
		return "error opening file", err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err.Error(), err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(api_url, contentType, bodyBuf)

	//结束发送文件部分
	if err != nil {
		return err.Error(), err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	logs.Debug("%s", string(resp_body))
	if err != nil {
		return err.Error(), err
	}
	return string(resp_body), nil
}
func RecordError(message string, err error) {
	if err != nil {
		logs.Error(message+"%v", err)
	}
}

func AssembleJsonResponse(c *gin.Context, errorCode int, message string, data interface{}) {
	logs.Info("结果返回：errorCode：%d; message: %s", errorCode, message)
	c.JSON(http.StatusOK, gin.H{
		"errorCode": errorCode,
		"message":   message,
		"data":      data,
	})
}

func AssembleJsonResponseWithStatusCode(c *gin.Context, statusCode int, message string, data interface{}) {
	logs.Info("结果返回：statusCode：%d; message: %s", statusCode, message)
	c.JSON(statusCode, gin.H{
		"errorNo": statusCode,
		"message": message,
		"data":    data,
	})
}

func NewGetAppMap() map[int]string {
	var mapInfo map[string]interface{}
	appMaps := Get(_const.ROCKET_URL)
	if appMaps == nil {
		return make(map[int]string)
	}
	json.Unmarshal(appMaps, &mapInfo)
	appList := mapInfo["data"].([]interface{})
	var AppIdMap = make(map[int]string)
	for _, appI := range appList {
		app := appI.(map[string]interface{})
		AppIdMap[int(app["AppId"].(float64))] = app["appName"].(string)
	}
	//fmt.Sprint(AppIdMap)
	return AppIdMap
}

//获取get请求
func Get(url string) []byte {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("token", _const.ROCKETTOKEN)
	if err != nil {
		logs.Error("获取rocket项目信息失败,%v", err)
		for _, lark_people := range _const.LowLarkPeople {
			LarkDingOneInner(lark_people, "获取rocket项目信息失败")
		}
		return nil
	}
	resp, err2 := client.Do(request)
	if err2 != nil {
		logs.Error("获取rocket项目信息失败,%v", err)
		for _, lark_people := range _const.LowLarkPeople {
			LarkDingOneInner(lark_people, "获取rocket项目信息失败")
		}
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//logs.Notice("获取app返回信息："+fmt.Sprint(string(body)))
	return body
}

//发送http post请求，其中rbody是一个json串
func PostJsonHttp2(rbody []byte) bool {
	client := &http.Client{}
	//提交请求
	reqest, err := http.NewRequest("POST", _const.LARK_URL, bytes.NewBuffer(rbody))
	//增加header选项
	reqest.Header.Add("token", _const.ROCKETTOKEN)
	if err != nil {
		logs.Error("rocket发送消息失败！", err.Error())
		return false
	}
	//处理返回结果
	response, _ := client.Do(reqest)
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	m := make(map[string]interface{})
	if err := json.Unmarshal(body, &m); err != nil {
		logs.Error("读取返回body出错！", err.Error())
		return false
	}
	if int(m["errorCode"].(float64)) == 0 {
		return true
	} else {
		return false
	}
}
func GetLarkToken() string {
	resp, err := http.PostForm("https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal/",
		url.Values{"app_id": {"cli_9d8a78c3eff61101"}, "app_secret": {"3kYDkS2M0obuzaEWrArGIc6NOJU6ZVeF"}})
	if err != nil {
		logs.Error("请求lark token出错！", err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("lark token读取返回出错！", err.Error())
	}
	m := make(map[string]interface{})
	json.Unmarshal(body, &m)
	return m["tenant_access_token"].(string)
}

//发送http post请求，其中rbody是一个json串
func PostJsonHttp3(rbody []byte, token, url string) (bool, string) {
	client := &http.Client{}
	//提交请求
	reqest, err := http.NewRequest("POST", url, bytes.NewBuffer(rbody))
	//增加header选项
	newToken := "Bearer " + token
	reqest.Header.Add("Authorization", newToken)
	if err != nil {
		logs.Error("lark官方API rocket发送消息失败！", err.Error())
		return false, ""
	}
	//处理返回结果
	response, _ := client.Do(reqest)
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	m := make(map[string]interface{})
	if err := json.Unmarshal(body, &m); err != nil {
		logs.Error("读取返回body出错！", err.Error())
		return false, ""
	}
	//返回bool只适合发检测的lark消息
	if msg, ok := m["msg"]; ok && msg.(string) == "ok" {
		return true, string(body)
	} else {
		return false, string(body)
	}
}

func GetVersionBMInfo(biz, project, version, os_type string) (rd string, qa string) {
	version_arr := strings.Split(version, ".")
	//TikTok这类型版本号：122005 无法获取BM信息
	if len(version_arr) < 3 {
		return "", ""
	}
	new_version := version_arr[0] + "." + version_arr[1] + "." + version_arr[2]
	client := &http.Client{}
	requestUrl := "https://rocket.bytedance.net/api/v1/project/versions"
	reqest, err := http.NewRequest("GET", requestUrl, nil)
	reqest.Header.Add("token", _const.ROCKETTOKEN)
	q := reqest.URL.Query()
	q.Add("project", project)
	q.Add("biz", biz)
	q.Add("achieve_type", os_type)
	q.Add("version_code", new_version)
	q.Add("nextpage", "1")
	reqest.URL.RawQuery = q.Encode()
	resp, _ := client.Do(reqest)
	if err != nil {
		logs.Error("获取version info出错！", err.Error())
		return "", ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("读取version返回出错！", err.Error())
		return "", ""
	}
	m := map[string]interface{}{}
	json.Unmarshal(body, &m)
	if m["data"] == nil {
		logs.Error("读取BM信息出错！")
		return "", ""
	}
	versionInfo := m["data"].(map[string]interface{})["VersionCards"].([]interface{})
	if len(versionInfo) == 0 {
		return "", ""
	}
	versionParam := versionInfo[0].(map[string]interface{})["Param_ext"].(string)
	var l []interface{}
	err = json.Unmarshal([]byte(versionParam), &l)
	if err != nil {
		logs.Error(err.Error())
	}
	var rd_bm, qa_bm string
	for _, bm := range l {
		if bm.(map[string]interface{})["Param_desc"].(string) == "RD BM" || bm.(map[string]interface{})["Param_desc"].(string) == "QA BM" {
			if bm.(map[string]interface{})["Param_desc"].(string) == "RD BM" {
				rd_bm = bm.(map[string]interface{})["Value"].(string)
			} else {
				qa_bm = bm.(map[string]interface{})["Value"].(string)
			}
			if rd_bm != "" && qa_bm != "" {
				break
			} else {
				continue
			}
		}
	}
	return rd_bm, qa_bm
}

//lark api get
func GetLarkInfo(url string, rbody map[string]string) string {
	client := &http.Client{}
	reqest, err := http.NewRequest("GET", url, nil)
	token := "Bearer " + GetLarkToken()
	reqest.Header.Add("Authorization", token)
	q := reqest.URL.Query()
	for k, v := range rbody {
		q.Add(k, v)
	}
	reqest.URL.RawQuery = q.Encode()
	resp, _ := client.Do(reqest)
	if err != nil {
		logs.Error("获取version info出错！", err.Error())
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}

//通用获取文件过期时间方法
func GetFileExpireTime(fileName string, fileType string, fileBytes []byte, userName string) *time.Time {
	getCertExpUrl := "http://" + DETECT_URL_PRO + "/query_certificate_expire_date" //过期日期访问地址
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("certificate", fileName)
	RecordError("访问过期日期POST请求create form file错误！", err)

	_, err = fileWriter.Write(fileBytes)
	RecordError("访问过期日期POST请求复制文件错误！", err)

	_ = writer.WriteField("username", userName)
	_ = writer.WriteField("type", fileType)
	contentType := writer.FormDataContentType()
	err = writer.Close()
	RecordError("关闭writer出错！！", err)

	response, err := http.Post(getCertExpUrl, contentType, body)
	RecordError("获取文件过期信息失败！", err)

	responseByte, err := ioutil.ReadAll(response.Body)

	responseMap := make(map[string]interface{})
	err = json.Unmarshal(responseByte, &responseMap)
	RecordError("文件过期信息结果解析失败！", err)

	if _, ok := responseMap["expire_time"]; !ok {
		return nil
	}

	expTimeStamp := int64(math.Floor(responseMap["expire_time"].(float64)))
	exp := time.Unix(expTimeStamp, 0)

	return &exp
}

//通用发送get请求方法
func SendHttpGet(url string, values map[string]string, withAuthorization bool) (error, []byte) {
	client := &http.Client{}

	//组装请求
	if values != nil {
		if url[len(url)-1] != '?' {
			url += "?"
		}
		for k, v := range values {
			url += k + "=" + v + "&"
		}
		url = strings.TrimSuffix(url, "&")
	}
	logs.Info("%v", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err, nil
	}
	if withAuthorization {
		req.Header.Set("Authorization", "Basic "+_const.KANI_APP_ID_AND_SECRET_BASE64)
	}

	//发送请求
	resp, err := client.Do(req)
	if err != nil {
		logs.Error(err.Error())
		return err, nil
	}
	defer resp.Body.Close()

	//读取响应
	bodyText, err := ioutil.ReadAll(resp.Body)
	return nil, bodyText
}

//通用发送post请求方法，其中postBody是一个json串。withAuthorization表示是否需要在头部携带token信息
func SendHttpPostByJson(url string, postBody []byte, withAuthorization bool) (error, []byte) {
	//组装请求
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(postBody)))
	if err != nil {
		return err, nil
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	if withAuthorization {
		req.Header.Set("Authorization", "Basic "+_const.KANI_APP_ID_AND_SECRET_BASE64)
	}
	//logs.Warn("%s",req)

	//发送请求
	resp, err := client.Do(req)
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()

	//读取响应
	body, err := ioutil.ReadAll(resp.Body)
	logs.Warn("%s", string(body))
	return err, body
}

//http response model
type GetResourceAdminListResponse struct {
	ResourceAdminListData `json:"data"`
	CommonResponse
}

type CommonResponse struct {
	ErrorNo      int    `json:"errno"`
	Message      string `json:"message"`
	ErrorMessage string `json:"errmsg"`
}

type ResourceAdminListData struct {
	OwnerKeys []string `json:"owner_keys"`
}

//通用调用pmc获取资源负责人列表方法
func GetAccountAdminList(teamId string) *[]string {
	teamId = strings.ToLower(teamId) + "_space_account"
	err, responseBytes := SendHttpGet(_const.GET_ACCOUNT_ADMIN_LIST_URL, map[string]string{"resourceKey": teamId}, true)
	RecordError("查询某资源的admin列表请求失败：", err)
	if err != nil {
		return nil
	}

	var responseObject GetResourceAdminListResponse

	err = json.Unmarshal(responseBytes, &responseObject)
	RecordError("查询某资源的admin列表请求结果解析失败：", err)
	if err != nil {
		return nil
	}

	logs.Info("查询某资源的admin列表请求结果：%v", responseObject.OwnerKeys)
	if responseObject.ErrorNo != 0 {
		logs.Error("查询某资源的admin列表请求结果出错：%s", responseObject.ErrorMessage)
		return nil
	}
	return &responseObject.OwnerKeys
}

//通用调用pmc给指定用户添加权限方法
func GiveUsersPermission(userNames *[]string, resourceKey string, actions *[]string) bool {
	params := map[string]interface{}{
		"resourceKey":  resourceKey,
		"actions":      *actions,
		"employeeKeys": *userNames,
		"groupIds":     &[]string{},
	}
	bodyBytes, err := json.Marshal(params)
	RecordError("marshal失败：", err)
	if err != nil {
		return false
	}

	err, responseByte := SendHttpPostByJson(_const.GIVE_PERMISSION_TO_USER_URL, bodyBytes, true)
	RecordError("发送post请求失败：", err)
	if err != nil {
		return false
	}

	var responseObject CommonResponse
	err = json.Unmarshal(responseByte, &responseObject)

	RecordError("post请求结果解析失败：", err)
	if err != nil {
		return false
	}

	if responseObject.ErrorNo != 0 {
		logs.Error("为用户(%v)添加权限(%v)失败：%s", userNames, actions, responseObject.ErrorMessage)
		return false
	}
	return true
}

func GetItcToken(username string) string {
	url := "https://itc.bytedance.net/t/generateToken?username=" + username
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("请求itc token出错！", err.Error())
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("itc token读取返回出错！", err.Error())
		return ""
	}
	m := make(map[string]interface{})
	json.Unmarshal(body, &m)
	return m["data"].(string)
}
