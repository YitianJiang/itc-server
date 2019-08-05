package service

import (
	"bytes"
	"code.byted.org/yuyilei/bot-api/form"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func PostRequest(urlStr string, header map[string]string, body interface{}) (map[string]interface{}, error) {
	bodyArray, err := json.Marshal(body)
	fmt.Printf("%s",string(bodyArray))
	if err != nil {
		return nil, err
	}
	respByte, err := sendRequest(urlStr, header, bytes.NewReader(bodyArray))
	if err != nil {
		return nil, err
	}
	resultDic := make(map[string]interface{})
	err = json.Unmarshal(respByte, &resultDic)
	if err != nil {
		return nil, fmt.Errorf("response json Unmarshall fail, http response body= %s", respByte)
	}
	return resultDic, nil
}

func PostRequestForm(urlStr string, header map[string]string, body io.Reader) (map[string]interface{}, error) {
	respByte, err := sendRequest(urlStr, header, body)
	if err != nil {
		return nil, err
	}
	resultDic := make(map[string]interface{})
	err = json.Unmarshal(respByte, &resultDic)
	if err != nil {
		return nil, fmt.Errorf("json Unmarshall fail, %v", err)
	}
	return resultDic, nil
}

func PostRequestToken(urlStr string, header map[string]string, body interface{}) (*form.AccessTokenResp, error) {
	bodyArray, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	respByte, err := sendRequest(urlStr, header, bytes.NewReader(bodyArray))
	if err != nil {
		return nil, err
	}
	var resp form.AccessTokenResp
	err = json.Unmarshal(respByte, &resp)
	if err != nil {
		return nil, fmt.Errorf("json Unmarshall fail, %v", err)
	}
	return &resp, nil
}

func sendRequest(urlStr string, header map[string]string, body io.Reader) ([]byte, error) {
	request, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
		return nil, fmt.Errorf("wewRequest fail, %v", err)
	}
	for k, v := range header {
		request.Header.Set(k, v)
	}
	client := http.Client{}
	resp, err := client.Do(request)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("doReqeust fail, %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http StatusCode is not in 2XX range, StatusCode= %d, Body= %s", resp.StatusCode, resp.Body)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("readAll fail, %v", err)
	}
	jsonStr := string(respBytes)
	jsonByte := []byte(jsonStr)
	jsonByteLen := len(jsonByte)
	return jsonByte[:jsonByteLen], nil
}