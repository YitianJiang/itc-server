//tqs使用详见：http://tqs.byted.org/#/help
package client

import (
	"bytes"
	"code.byted.org/dp/gotqs/consts"
	"code.byted.org/dp/gotqs/tools"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/pkg/errors"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var CLUSTER_PSM_MAP = map[string]string{
	consts.CLUSTER_TEST:      consts.TQS_API_SERVER_PSM_TEST,
	consts.CLUSTER_DEFAULT:   consts.TQS_API_SERVER_PSM_CN,
	consts.CLUSTER_AWS:       consts.TQS_API_SERVER_PSM_AWS,
	consts.CLUSTER_MALIVA:    consts.TQS_API_SERVER_PSM_AWS,
	consts.CLUSTER_VA_TEST:   consts.TQS_API_SERVER_PSM_TEST_VA,
	consts.CLUSTER_CN_PRIEST: consts.TQS_API_SERVER_PSM_CN_PRIEST,
	consts.CLUSTER_VA_PRIEST: consts.TQS_API_SERVER_PSM_VA_PRIEST,
}

type serverInfo struct {
	srvAddr  [2]string
	curIndex uint8
}

//tqs client
type TqsClient struct {
	ClientConf *TqsClientConf
	InitStatus int
	ServerInfo *serverInfo
}

//tqs client config
type TqsClientConf struct {
	AppId    string
	AppKey   string
	Timeout  time.Duration
	Cluster  string
	UserName string
}

//任务状态
type JobStatus struct {
	AppId      string `json:"appId"`
	EndTime    string `json:"endTime"`
	EngineType string `json:"engineType"`
	Id         int64  `json:"id"`
	LogUrl     string `json:"logUrl"`
	MaxResults int64  `json:"maxResults"`
	Query      string `json:"query"`
	ResultUrl  string `json:"resultUrl"`
	StartTime  string `json:"startTime"`
	Status     string `json:"status"`
	Username   string `json:"userName"`
}

//任务数据预览
type JobPreview struct {
	Rows   [][]string `json:"rows"`
	Url    string     `json:"url"`
	UrlGBK string     `json:"urlGBK"`
}

func (client *TqsClient) GetServerAddr() string {
	logs.Debug("curAddr=%s", client.ServerInfo.srvAddr[client.ServerInfo.curIndex])
	return client.ServerInfo.srvAddr[client.ServerInfo.curIndex]
}

//获取tqs client config
func (client *TqsClient) getTqsClientConf() *TqsClientConf {
	if nil == client.ClientConf {
		return &TqsClientConf{}
	}
	return client.ClientConf
}

//初始化tqs client
func InitClient(ctx context.Context, appId string, appKey string, userName string, timeout time.Duration, cluster string) (*TqsClient, error) {
	tqsConf := &TqsClientConf{
		AppId:    appId,
		AppKey:   appKey,
		Timeout:  timeout,
		Cluster:  cluster,
		UserName: userName,
	}
	tqsClient := &TqsClient{
		ClientConf: tqsConf,
		ServerInfo: &serverInfo{curIndex: 0},
	}

	addr, err := tqsClient.chooseServer(ctx)
	initStatus := consts.CLIENT_INIT_STATUS_SUCCESS
	if err != nil {
		logs.CtxWarn(ctx, "ChooseServer error, err: %v", err)
		initStatus = consts.CLIENT_INIT_STATUS_FAIL
	}
	tqsClient.InitStatus = initStatus
	tqsClient.ServerInfo.srvAddr[0] = addr
	go tqsClient.serverDiscovery(ctx)
	return tqsClient, nil
}

func (client *TqsClient) serverDiscovery(ctx context.Context) {
	for {
		addr, err := client.chooseServer(ctx)
		if err != nil {
			logs.CtxWarn(ctx, "ChooseServer error, err: %v", err)
			time.Sleep(time.Second)
			continue
		}
		nextIdx := 1 - client.ServerInfo.curIndex
		client.ServerInfo.srvAddr[nextIdx] = addr
		client.ServerInfo.curIndex = nextIdx
		time.Sleep(time.Second * 10)
	}
}

//确定tqs api server
func (client *TqsClient) chooseServer(ctx context.Context) (string, error) {
	conf := client.getTqsClientConf()
	psm, ok := CLUSTER_PSM_MAP[conf.Cluster]
	if !ok {
		logs.CtxWarn(ctx, "cluster not found, use TQS_API_SERVER_PSM_TEST: %s", consts.TQS_API_SERVER_PSM_TEST)
		psm = consts.TQS_API_SERVER_PSM_CN
	}
	addr, err := tools.GetServerAddr(ctx, psm)
	if err != nil || addr == "" {
		logs.CtxWarn(ctx, "GetServerAddr error, err: %v", err)
		return "", errors.New("GetServerAddr error")
	}
	return addr, nil
}

//http request:body为json
func (client *TqsClient) requestJson(ctx context.Context, method string, url string, posts map[string]string, headers map[string]string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		logs.CtxError(ctx, "http.NewRequest error")
		return nil, errors.New("http.NewRequest error")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		logs.CtxError(ctx, "httpDo error, err: %v", err)
		return nil, errors.New("httpDo error")
	}

	return resp, nil
}

//创建tqs query，返回jobId
// curl -X POST \
//  -H "X-TQS-AppId: your_app_id" \
//  -H "X-TQS-AppKey: your_app_key	" \
//  -H "Content-Type: application/json" \
//  -d '{"user": "user_name", "query": "your_hql", "dryRun": false}' \
//  http://tqs.byted.org/api/v1/queries
// dryRun: false表示服务端除了会完成 query 的解析，还会进一步做执行操作。反之，则只做解析。
func (client *TqsClient) CreateQueryJob(ctx context.Context, query string, dryRun bool, skipCostAnalysis bool) (int64, error) {
	if client == nil || client.InitStatus == consts.CLIENT_INIT_STATUS_INIT || client.ClientConf == nil {
		logs.CtxError(ctx, "TqsClient not init")
		return 0, errors.New("TqsClient not init")
	}
	url := fmt.Sprintf("http://%s/api/v1/queries", client.GetServerAddr())
	clientConf := client.getTqsClientConf()
	body := map[string]interface{}{
		"user":             clientConf.UserName,
		"query":            query,
		"dryRun":           dryRun,
		"skipCostAnalysis": skipCostAnalysis,
	}
	jsBody, err := json.Marshal(body)
	if err != nil {
		logs.CtxError(ctx, "json.Marshal error, err: %v", err)
		return 0, errors.New("json.Marshal error")
	}
	headers := map[string]string{
		"X-TQS-AppId":  clientConf.AppId,
		"X-TQS-AppKey": clientConf.AppKey,
		"Content-Type": "application/json",
	}
	resp, err := client.requestJson(ctx, "POST", url, nil, headers, jsBody)
	if err != nil {
		logs.CtxError(ctx, "request error, err: %v", err)
		return 0, errors.New("request error")
	}
	statusCode, jobId, errCode, errMsg, parseErr := parseCreateQueryResp(resp)
	if parseErr != nil {
		logs.CtxError(ctx, "parseCreateQueryResp error, err: %v", parseErr)
		return 0, errors.New("parseCreateQueryResp error")
	}
	if statusCode >= 400 {
		logs.CtxError(ctx, "HTTP request error, statusCode: %v", statusCode)
		return 0, errors.New("HTTP request error")
	}
	if jobId <= 0 {
		logs.CtxError(ctx, "CreateQueryJob failed, errCode: %v, errMsg: %v", errCode, errMsg)
		return 0, errors.New("CreateQueryJob failed")
	}
	return jobId, nil
}

func parseCreateQueryResp(resp *http.Response) (status int, jobId int64, errCode int64, errMsg string, parseErr error) {
	if resp == nil {
		parseErr = errors.New("resp nil")
		return
	}
	status = resp.StatusCode
	respBody, err := ioutil.ReadAll(resp.Body)
	logs.Debug("response body: %s", respBody)
	if err != nil {
		parseErr = errors.New("read response body error")
		return
	}
	//尝试解json
	type stSuccess struct {
		JobId int64 `json:"jobId"`
	}
	type stError struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	}
	var objSuccess stSuccess
	err = json.Unmarshal(respBody, &objSuccess)
	if err == nil {
		jobId = objSuccess.JobId
		return
	}
	var objError stError
	err = json.Unmarshal(respBody, &objError)
	if err == nil {
		errCode = objError.Code
		errMsg = objError.Message
		return
	}
	err = errors.New("parse response body error")
	return
}

//查询tqs query状态，返回jobId
//	curl -X GET \
//	-H "X-TQS-AppId: your_app_id" \
//	-H "X-TQS-AppKey: your_app_key" \
//	-H "Content-Type: application/json" \
//	http://tqs.byted.org/api/v1/jobs/1
func (client *TqsClient) GetQueryJobInfo(ctx context.Context, jobId int64) (*JobStatus, error) {
	if client == nil || client.InitStatus == consts.CLIENT_INIT_STATUS_INIT || client.ClientConf == nil {
		logs.CtxError(ctx, "TqsClient not init")
		return nil, errors.New("TqsClient not init")
	}
	url := fmt.Sprintf("http://%s/api/v1/jobs/%d", client.GetServerAddr(), jobId)
	clientConf := client.getTqsClientConf()
	body := map[string]interface{}{
		//
	}
	jsBody, err := json.Marshal(body)
	if err != nil {
		logs.CtxError(ctx, "json.Marshal error, err: %v", err)
		return nil, errors.New("json.Marshal error")
	}
	if len(body) == 0 {
		jsBody = nil
	}
	headers := map[string]string{
		"X-TQS-AppId":  clientConf.AppId,
		"X-TQS-AppKey": clientConf.AppKey,
		"Content-Type": "application/json",
	}
	resp, err := client.requestJson(ctx, "GET", url, nil, headers, jsBody)
	if err != nil {
		logs.CtxError(ctx, "request error, err: %v", err)
		return nil, errors.New("request error")
	}
	statusCode, jobStatus, parseErr := parseGetQueryResp(resp)
	if parseErr != nil {
		logs.CtxError(ctx, "parseGetQueryResp error, err: %v", parseErr)
		return nil, errors.New("parseGetQueryResp error")
	}
	if statusCode >= 400 {
		logs.CtxError(ctx, "HTTP request error, statusCode: %v", statusCode)
		return nil, errors.New("HTTP request error")
	}
	if jobStatus == nil {
		logs.CtxError(ctx, "parseGetQueryResp error, resp: %v", resp)
		return nil, errors.New("GetQueryJobInfo parse error")
	}
	return jobStatus, nil

}

func parseGetQueryResp(resp *http.Response) (status int, jobStatus *JobStatus, parseErr error) {
	if resp == nil {
		parseErr = errors.New("resp nil")
		return
	}
	status = resp.StatusCode
	respBody, err := ioutil.ReadAll(resp.Body)
	logs.Debug("response body: %s", respBody)
	if err != nil {
		parseErr = errors.New("read response body error")
		return
	}
	//尝试解json
	var jobSt JobStatus
	err = json.Unmarshal(respBody, &jobSt)
	if err == nil {
		jobStatus = &jobSt
		return
	}
	err = errors.New("parse response body error, resp: " + string(respBody))
	return
}

//取消tqs query
// curl -X POST \
//  -H "X-TQS-AppId: your_app_id" \
//  -H "X-TQS-AppKey: your_app_key	" \
//  -H "Content-Type: application/json" \
//  http://tqs.byted.org/api/v1/jobs/1/cancel
func (client *TqsClient) CancelueryJob(ctx context.Context, jobId int64) (bool, error) {
	if client == nil || client.InitStatus == consts.CLIENT_INIT_STATUS_INIT || client.ClientConf == nil {
		logs.CtxError(ctx, "TqsClient not init")
		return false, errors.New("TqsClient not init")
	}
	url := fmt.Sprintf("http://%s/api/v1/jobs/%d/cancel", client.GetServerAddr(), jobId)
	clientConf := client.getTqsClientConf()
	body := map[string]interface{}{
		//
	}
	jsBody, err := json.Marshal(body)
	if err != nil {
		logs.CtxError(ctx, "json.Marshal error, err: %v", err)
		return false, errors.New("json.Marshal error")
	}
	headers := map[string]string{
		"X-TQS-AppId":  clientConf.AppId,
		"X-TQS-AppKey": clientConf.AppKey,
		"Content-Type": "application/json",
	}
	resp, err := client.requestJson(ctx, "POST", url, nil, headers, jsBody)
	if err != nil {
		logs.CtxError(ctx, "request error, err: %v", err)
		return false, errors.New("request error")
	}
	statusCode, success, parseErr := parseCancelQueryResp(resp)
	if parseErr != nil {
		logs.CtxError(ctx, "parseCancelQueryResp error, err: %v", parseErr)
		return false, errors.New("parseCancelQueryResp error")
	}
	if statusCode >= 400 {
		logs.CtxError(ctx, "HTTP request error, statusCode: %v", statusCode)
		return false, errors.New("HTTP request error")
	}
	return success, nil
}

func parseCancelQueryResp(resp *http.Response) (status int, success bool, parseErr error) {
	if resp == nil {
		parseErr = errors.New("resp nil")
		return
	}
	status = resp.StatusCode
	respBody, err := ioutil.ReadAll(resp.Body)
	logs.Debug("response body: %s", respBody)
	if err != nil {
		parseErr = errors.New("read response body error")
		return
	}
	if string(respBody) == "OK" {
		success = true
	}
	return
}

//预览tqs query数据
//	curl -X GET \
//	-H "X-TQS-AppId: your_app_id" \
//	-H "X-TQS-AppKey: your_app_key" \
//	-H "Content-Type: application/json" \
//	http://tqs.byted.org/api/v1/queries/1
func (client *TqsClient) PreviewQuery(ctx context.Context, jobId int64) (*JobPreview, error) {
	if client == nil || client.InitStatus == consts.CLIENT_INIT_STATUS_INIT || client.ClientConf == nil {
		logs.CtxError(ctx, "TqsClient not init")
		return nil, errors.New("TqsClient not init")
	}
	url := fmt.Sprintf("http://%s/api/v1/queries/%d", client.GetServerAddr(), jobId)
	clientConf := client.getTqsClientConf()
	body := map[string]interface{}{
		//
	}
	jsBody, err := json.Marshal(body)
	if err != nil {
		logs.CtxError(ctx, "json.Marshal error, err: %v", err)
		return nil, errors.New("json.Marshal error")
	}
	headers := map[string]string{
		"X-TQS-AppId":  clientConf.AppId,
		"X-TQS-AppKey": clientConf.AppKey,
		"Content-Type": "application/json",
	}
	resp, err := client.requestJson(ctx, "GET", url, nil, headers, jsBody)
	if err != nil {
		logs.CtxError(ctx, "request error, err: %v", err)
		return nil, errors.New("request error")
	}
	statusCode, jobPreview, parseErr := parsePreviewQueryResp(resp)
	if parseErr != nil {
		logs.CtxError(ctx, "parsePreviewQueryResp error, err: %v", parseErr)
		return nil, errors.New("parsePreviewQueryResp error")
	}
	if statusCode >= 400 {
		logs.CtxError(ctx, "HTTP request error, statusCode: %v", statusCode)
		return nil, errors.New("HTTP request error")
	}
	if jobPreview == nil {
		logs.CtxError(ctx, "parsePreviewQueryResp error")
		return nil, errors.New("parsePreviewQueryResp parse error")
	}
	return jobPreview, nil
}

func parsePreviewQueryResp(resp *http.Response) (status int, jobPreview *JobPreview, parseErr error) {
	if resp == nil {
		parseErr = errors.New("resp nil")
		return
	}
	status = resp.StatusCode
	respBody, err := ioutil.ReadAll(resp.Body)
	//logs.Debug("resp: %v", string(respBody))
	if err != nil {
		parseErr = errors.New("read response body error")
		return
	}
	//OK可能是query配置出错
	if string(respBody) == "OK" {
		parseErr = errors.New("perhaps query setting error, try skipAnalysis=false and dryRun=false")
		return
	}
	//尝试解json
	var jobPre JobPreview
	err = json.Unmarshal(respBody, &jobPre)
	if err == nil {
		jobPreview = &jobPre
		return
	}
	err = errors.New("parse response body error, resp: " + string(respBody))
	return
}
