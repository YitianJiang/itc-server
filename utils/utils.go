package utils

import (
	"bytes"
	"code.byted.org/gopkg/logs"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

//发送http post请求，其中rbody是一个json串
func PostJsonHttp(url string,rbody []byte ) (int, []byte) {
	http.DefaultClient.Timeout = 3 * time.Second
	bodyBuffer := bytes.NewBuffer([]byte(rbody))
	resp, err := http.Post(url, "application/json;charset=utf-8",bodyBuffer)
	if err != nil {
		return -1, nil;
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -2, nil;
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