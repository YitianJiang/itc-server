package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
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
	fmt.Println(string(body))
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
	if m["errorCode"].(int) == 0 {
		return true
	} else {
		return false
	}
}
