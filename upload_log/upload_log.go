package uploadlog

import (
	"fmt"
	"time"

	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

const emailSuffix = "@bytedance.com"

// UploadViaHTTP uploads log to the specific platform via HTTP.
func UploadViaHTTP(c *gin.Context) {

	logs.Debug("start uploading log via HTTP")

	userName, exist := c.Get("username")
	if !exist {
		logs.Warn("Invalid user: %v", userName)
		return
	}

	msgInfo := map[string]string{
		"source":      "101",                                   // 业务名称，全局唯一，需提交审核，not null
		"psm":         "toutiao.clientqa.itc-server",           // 当前业务的psm
		"user_name":   userName.(string),                       // not null
		"ip":          c.ClientIP(),                            // 客户端IP，not null
		"ua":          c.Request.UserAgent(),                   // User-Agent
		"referer":     c.Request.Header.Get("referer"),         // Referer
		"xff":         c.Request.Header.Get("X-Forward-For"),   // X-Forwarded-For
		"url":         c.Request.Host + c.Request.URL.String(), // 资源地址, not null
		"title":       "预审平台",                                  // 页面title
		"action_type": c.Request.Method,                        // 动作_属性名词，全部小写，not null
		"action_desc": c.Request.Method,                        // 动作描述
		"extra":       "",                                      // 备用字段
	}
	fmt.Println(msgInfo)
	if err := Publish(msgInfo, 2*time.Second); err != nil {
		logs.Warn("failed to upload log: %v", err)
	}

	logs.Debug("upload log via HTTP success")
	return
}
