package http_util

import (
	"bytes"
	"code.byted.org/clientQA/pkg/const"
	"code.byted.org/clientQA/pkg/request-processor/request-dal"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	urllib "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var domains = map[string][]string{
	"toutiao-ci.bytedance.net":  {"maoyu", "2212ea867a907f418ca39733e346a2d9"},
	"ci.bytedance.net":          {"zhangshuai.02", "0a120405ccce80c218ad6c05dcb9870e"},
	"learning-ci.bytedance.net": {"maoyu", "6dc57defbb49c3a4736876f7efb37b05"},
	"kid-ci.bytedance.net":      {"maoyu", "b9e13e6c380df5f3dd36750d51a6932b"},
	"easy-ci.bytedance.net":     {"maoyu", "44648b27c996b37ae5e2bcd5348cf21f"},
	"lite-ci.bytedance.net":     {"maoyu", "1a736f4ee9d41096194fabc000e9df15"},
	"life-ci.bytedance.net":     {"maoyu", "b245655a578ba114c54a26911056db90"},
	"ies-ci.bytedance.net":      {"maoyu", "45975cf40308629a1ce3e85085863ff0"},
	"ios-ci.bytedance.net":      {"maoyu", "d55a9fe9c3ac87f44a5de1e84daf91e2"},
	"top-ci.bytedance.net":      {"maoyu", "ff13bd93fe84acca0f3724ceba651c8e"},
}

/*
使用basic auth的形式触发各种job，如果没有自己的token，统一用张帅的token
*/
func PostWithBasicAuth(url string, values urllib.Values, api_id string, api_token string) (error, []byte) {
	client := &http.Client{}
	api_id, api_token = ConfigToken(url)

	if api_id == "" || api_token == "" {
		api_id = _const.UserID
		api_token = _const.ApiToken
	}
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(values.Encode()))
	req.SetBasicAuth(api_id, api_token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//logs.Warn("%v", req)
	resp, err := client.Do(req)
	if err != nil {
		logs.Error(err.Error())
		return err, nil
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	//logs.Info("%s", string(bodyText))
	//该闭包返回了重试后的bodyText，要么是空，要么是返回值
	a := func() string {
		if strings.Contains(strings.ToLower(string(bodyText)), "404") {
			bodyText = nil
			if api_id == "" || api_token == "" {
				api_id = _const.UserID
				api_token = _const.ApiToken
			}
			req2, err := http.NewRequest("POST", url, bytes.NewBufferString(values.Encode()))
			req2.SetBasicAuth(api_id, api_token)
			req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp2, err := client.Do(req2)
			//logs.Warn("%v", req2)
			if err != nil {
				logs.Error(err.Error())
				return err.Error()
			}
			defer resp2.Body.Close()
			bodyText, err = ioutil.ReadAll(resp2.Body)
			//logs.Info("%s", string(bodyText))
			return string(bodyText)
		} else {
			return ""
		}
	}
	go retry(3, 15*time.Second, a)
	return nil, bodyText
}

func ConfigToken(url string) (string, string) {
	//根据域名增加相应的api id和api token...操碎了心
	u, err := urllib.Parse(url)
	if err != nil {
		logs.Error("%v", err)
		return _const.UserID, _const.ApiToken
	}
	if ok := domains[u.Host]; ok != nil {
		return domains[u.Host][0], domains[u.Host][1]
	}
	return _const.UserID, _const.ApiToken
}

func JudgeToken(url string) bool {
	//根据域名增加相应的api id和api token...操碎了心
	u, err := urllib.Parse(url)
	if err != nil {
		logs.Error("%v", err)
		return false
	}
	if ok := domains[u.Host]; ok != nil {
		return true
	}
	return false
}

func retry(attempts int, sleep time.Duration, fn func() string) string {
	//如果重试的时候，发现str返回值还是有404出现，那么继续重试
	if str := fn(); strings.Contains(str, "404") {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return retry(attempts, 2*sleep, fn)
		}
		return "404"
	}
	return ""
}

//发送http post请求，其中rbody是一个json串
func PostJsonHttp(url string, rbody []byte) (error, []byte) {
	bodyBuffer := bytes.NewBuffer([]byte(rbody))

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bodyBuffer)
	if err != nil {
		return err, nil
	}

	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return err, nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}
	//logs.Warn("%s", string(body))
	return nil, body
}

/*
get方法，带上basci auth
*/
func GetWithBasicAuth(url string, values map[string]string, api_id string, api_token string) (error, []byte) {
	client := &http.Client{}
	//根据域名增加相应的api id和api token...操碎了心
	u, err := urllib.Parse(url)
	if err != nil {
		logs.Error("%v", err)
		api_id = _const.UserID
		api_token = _const.ApiToken
	}

	if ok := domains[u.Host]; ok != nil {
		api_id = domains[u.Host][0]
		api_token = domains[u.Host][1]
	}

	if api_id == "" || api_token == "" {
		api_id = _const.UserID
		api_token = _const.ApiToken
	}

	if values != nil {
		url += "?"
		for k, v := range values {
			url += k + "=" + v + "&"
		}
		url = strings.TrimSuffix(url, "&")
	}
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(api_id, api_token)
	resp, err := client.Do(req)
	if err != nil {
		logs.Error(err.Error())
		return err, nil
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	return nil, bodyText
}

/*
自定义的header进行get请求
*/
func GetWithCustomHeader(url string, values map[string]string) (error, []byte) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if values != nil {
		for k, v := range values {
			req.Header.Add(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		logs.Error(err.Error())
		return err, nil
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	return nil, bodyText
}

/*
按下按钮让flow系统继续调度
*/
func PushButton(url string, pipelineID string, content string, buttonType int, isPkg bool) (error, string) {
	//如果含有build,那么想办法分割整个字符串。。。然后找到job路径
	if strings.Contains(strings.ToLower(url), "build") {
		urlstrs := strings.Split(url, "build")
		if len(urlstrs) > 0 {
			url = urlstrs[0]
		}
	}
	url += pipelineID + "/input/"
	if isPkg {
		url += "START_PKG/"
	} else {
		url += "START_" + content + "/"
	}

	switch buttonType {
	case _const.ProceedButton:
		url += "proceedEmpty"
	case _const.AbortButton:
		url += "abort"
	}
	//logs.Warn("%s", url)
	if err, resp := PostWithBasicAuth(url, nil, "", ""); err != nil {
		return err, string(resp)
	}
	return nil, ""
}

func PostFileWithParams(download_url string, params map[string]string, postfilename string, api_url string) (string, error) {
	//提取文件
	client_http := &http.Client{}
	/* Authenticate */
	req, err := http.NewRequest("GET", download_url, nil)
	res, err := client_http.Do(req)
	if err != nil {
		return "下载文件 " + download_url + "失败", err
	}

	//本地先下载这个文件
	//首先，应该对file name进行split，我们这里只需要文件名字
	_, remotefilename := filepath.Split(download_url)
	_, realfilename := filepath.Split(remotefilename)
	f, err := os.Create(realfilename)
	defer os.Remove(realfilename)
	defer f.Close()
	if err != nil {
		return "创建文件 " + realfilename + "失败", err
	}
	io.Copy(f, res.Body)
	currentpath, _ := os.Getwd()
	fullfilepath := currentpath + "/" + realfilename

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

	//////结束发送文件部分

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

	//////结束发送文件部分

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

func GetUserAvatar(email string) request_dal.Struct_User {
	var retstruct request_dal.Struct_User
	ret, body := PostWithBasicAuth(_const.Dingavatar, urllib.Values{"email": {email}}, "", "")
	if ret == nil {
		type struct_avatar struct {
			Origin string
		}
		type ret_struct struct {
			Avatar struct_avatar
			Id     uint
		}
		var result []ret_struct
		json.Unmarshal(body, &result)
		//如果获取到的数据大于0（其实只能获取到一个数据而已）
		if len(result) > 0 {

			//retstring = result[0].Avatar.Origin
			retstruct.Employeenumber = result[0].Id
			retstruct.Avatar = result[0].Avatar.Origin
		}
	}
	return retstruct
}

/*
输入一个人名，判断都有哪些组的权限
*/
func AllPermissionJudger(username string) (error, string) {
	err, body := GetWithBasicAuth(fmt.Sprintf(_const.Permission_all_url, username), nil, _const.Kani_appid, _const.Kani_apppwd)
	if err != nil {
		logs.Error("%s", "访问kani失败，无法读取body")
		return errors.New("访问kani失败，无法读取body"), ""
	}
	var dat map[string][]string
	if err := json.Unmarshal(body, &dat); err != nil {
		return errors.New("访问kani失败，无法解析"), ""
	}

	if len(dat) < 1 {
		return nil, ""
	} else {
		permissions := ""
		for k := range dat {
			permissions += k + ","
		}
		permissions = strings.TrimSuffix(permissions, ",")
		return nil, permissions
	}

}
