package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/gopkg/logs"
)

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
		logs.Error("读取BM信息出错！", err.Error())
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
