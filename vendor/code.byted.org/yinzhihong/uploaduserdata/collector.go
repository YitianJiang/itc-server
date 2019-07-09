package uploaduserdata
import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"code.byted.org/gopkg/logs"
)

const (
	URL            = "https://ssa.bytedance.net/v2/"
	TQ_LOG_PSM     = "security.tq.log"
	TQ_METRICS_PSM = "security.tq.databus_log"
	KEY            = "b9036bf2e794889076a66a95a2767d0407998d70"
)

type Callback func(map[string]string) (*Msg, error)

func init() {
	logs.AddProvider(logs.NewAgentProvider())
	logs.DefaultLogger().SetPSM(TQ_LOG_PSM)
	//metrics打点
	initMetrics(TQ_METRICS_PSM)
}

func verifyResult(result Msg) error {
	event := reflect.ValueOf(result.Event)
	for i := 0; i < event.NumField(); i++ {
		if event.Field(i).Interface() == "" {
			return errors.New("Event is incomplete")
		}
	}
	header := reflect.ValueOf(result.Header)
	for i := 0; i < header.NumField(); i++ {
		if header.Field(i).Interface() == "" {
			return errors.New("Header is incomlete")
		}
	}
	if result.User.Ip == "" {
		return errors.New("Ip is incomlete")
	}
	if result.Source == "" {
		return errors.New("Source is incomlete")
	}
	return nil
}

func computeToken(psm string, ts int64) (string, error) {
	st := time.Unix(ts, 0).Format("2006-01-02 15:04:05")
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return "", err
	}
	lt, err := time.ParseInLocation("2006-01-02 15:04:05", st, loc)
	if err != nil {
		return "", err
	}
	ts = lt.Unix()
	h := sha1.New()
	h.Write([]byte(psm))
	h.Write([]byte(KEY))
	h.Write([]byte(strconv.FormatInt(ts-2048, 10)))
	result := fmt.Sprintf("%x", h.Sum(nil))
	return result, nil
}

/*
payload: {
        data: {}
        psm: ""
        token: "xxxx"
        "ts": 4374973
}
token == sha1(psm+key+str(ts-2048))
*/

func Publish(callback Callback, msg map[string]string, psm string, ts int64) error {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("err in publish: %v", err)
			emitError("recover", "")
		}
	}()
	pMsg, err := callback(msg)
	msgMap := *pMsg
	if err != nil {
		logs.Error("err in callback: %v", err)
		emitError("callback", "")
		return err
	}
	if err := verifyResult(msgMap); err != nil {
		emitError("msg_format", "")
		return err
	}
	token, err := computeToken(psm, ts)
	if err != nil {
		logs.Error("err in verifyResult: %v", err)
		emitError("compute_token", "")
		return err
	}
	msgBytes, err := json.Marshal(msgMap)
	if err != nil {
		logs.Error("err in marshal: %v", err)
		emitError("json_marshal", msgMap.Header.Domain)
		return err
	}
	reader := bytes.NewReader(msgBytes)
	request, err := http.NewRequest("POST", URL, reader)
	if err != nil {
		logs.Error("err in create http request: %v", err)
		emitError("create_http_request", msgMap.Header.Domain)
		return err
	}
	request.Header.Set("Content-type", "application/json;charset=UTF-8")
	request.Header.Add("TS", strconv.FormatInt(ts, 10))
	request.Header.Add("PSM", psm)
	request.Header.Add("TOKEN", token)
	// 设置timeout
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		logs.Error("err in send http request: %v", err)
		emitError("send_http_request", msgMap.Header.Domain)
		return err
	}
	return err
}
