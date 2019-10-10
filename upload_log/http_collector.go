package uploadlog

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"code.byted.org/gopkg/logs"
)

const (
	URL = "https://ssa.bytedance.net/v2/"
	//TQ_LOG_PSM     = "security.tq.log"
	TQ_METRICS_PSM = "security.tq.http_log"
	KEY            = "b9036bf2e794889076a66a95a2767d0407998d70"
)

func init() {
	// logs.AddProvider(logs.NewAgentProvider())
	// logs.DefaultLogger().SetPSM(TQ_LOG_PSM)
	//metrics打点
	initMetrics(TQ_METRICS_PSM)
}

func computeToken(psm string, ts int64) string {
	//token = sha1(psm+key+str(ts-2048))
	h := sha1.New()
	h.Write([]byte(psm))
	h.Write([]byte(KEY))
	h.Write([]byte(strconv.FormatInt(ts-2048, 10)))
	result := fmt.Sprintf("%x", h.Sum(nil))
	return result
}

var checkList = []string{"source", "user_name", "ip", "url", "action_type"}

func validate(msgInfo map[string]string) error {
	for _, key := range checkList {
		if msgInfo[key] == "" {
			return fmt.Errorf("empty %v", key)
		}
	}
	return nil
}

/*
发送uba msg到http接口
headers = {
  'PSM': psm,
  'TS': ts,
  'TOKEN': token,
  'Content-Type': 'application/json',
}
*/
func Publish(msgInfo map[string]string, timeout time.Duration) error {
	if err := validate(msgInfo); err != nil {
		logs.Error("msg format err:%v, msg: %v", err, msgInfo)
		emitError("msg_format", "")
		return err
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	ts := time.Now().In(loc).Unix()
	token := computeToken(msgInfo["psm"], ts)

	msgBytes, err := json.Marshal(msgInfo)
	if err != nil {
		logs.Error("json dump err: %v, msg: %v", err, msgInfo)
		emitError("json_dump", "")
		return err
	}
	reader := bytes.NewReader(msgBytes)
	request, err := http.NewRequest("POST", URL, reader)
	if err != nil {
		logs.Error("new http request err: %v", err)
		emitError("create_http_request", "")
		return err
	}
	request.Header.Set("Content-type", "application/json;charset=UTF-8")
	request.Header.Add("TS", strconv.FormatInt(ts, 10))
	request.Header.Add("PSM", msgInfo["psm"])
	request.Header.Add("TOKEN", token)
	// 设置timeout
	client := http.Client{Timeout: timeout}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		logs.Error("send http request: %v", err)
		emitError("send_http_request", "")
		return err
	}
	return nil
}
