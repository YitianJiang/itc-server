package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

// Error code
const (
	FAILURE = -1
	SUCCESS = 0
)

// ReturnMsg return the response to requester.
// If the data is not empty, only the first data will be accept while the rest
// will be abandoned.
func ReturnMsg(c *gin.Context, httpCode int, code int, msg string,
	data ...interface{}) {

	switch code {
	case FAILURE:
		logs.Error(msg)
	case SUCCESS:
		logs.Debug(msg)
	default:
		logs.Notice("unsupport code (%v): %v", code, msg)
	}

	obj := gin.H{"code": code, "message": msg}
	if len(data) > 0 {
		obj["data"] = data[0]
	}

	c.JSON(httpCode, obj)

	return
}

// SendHTTPRequest uses specific method sending data to specific URL
// via HTTP request.
func SendHTTPRequest(method string, url string, params map[string]string, headers map[string]string,
	data []byte) ([]byte, error) {

	// Construct HTTP handler
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		logs.Error("construct HTTP request failed: %v", err)
		return nil, err
	}

	// Add query parameters
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	// Set request header
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	logs.Debug("%v", req)

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("send HTTP request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read HTTP response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("read content from HTTP response failed: %v", err)
		return nil, err
	}
	logs.Debug("%s", body)

	return body, err
}
