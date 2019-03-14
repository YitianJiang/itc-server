// Package ctx(context) adds extra context informations in gin.Context, which will be used in log, metrics, and thrift rpc call
//
// Added context keys includes:
//   - log id (K_LOGID)
//   - local service name (K_SNAME)
//   - local ip (K_LOCALIP)
//   - local cluster (K_CLUSTER)
//   - method (K_METHOD)
package ctx

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	"code.byted.org/gin/ginex/internal"
	internal_util "code.byted.org/gin/ginex/internal/util"
	"github.com/gin-gonic/gin"
)

var (
	localIP           string
	fullLengthLocalIP []byte
)

func Ctx() gin.HandlerFunc {
	psm := os.Getenv(internal.GINEX_PSM)
	cluster := internal_util.LocalCluster()
	return func(c *gin.Context) {
		if logID := c.Request.Header.Get(internal.TT_LOGID_HEADER_KEY); logID != "" {
			c.Set(internal.LOGIDKEY, logID)
		} else if logID = c.Request.Header.Get(internal.TT_LOGID_HEADER_FALLBACK_KEY); logID != "" {
			c.Set(internal.LOGIDKEY, logID)
		} else {
			logID = genLogId()
			c.Set(internal.LOGIDKEY, logID)
			c.Header(internal.TT_LOGID_HEADER_KEY, logID)
		}
		if env := c.Request.Header.Get(internal.TT_ENV_KEY); env != "" {
			c.Set(internal.ENVKEY, env)
		} else {
			c.Set(internal.ENVKEY, "prod")
		}
		if stressTag := c.Request.Header.Get(internal.TT_STRESS_KEY); stressTag != "" {
			c.Set(internal.STRESSKEY, stressTag)
		}
		if traceTag := c.Request.Header.Get(internal.TT_TRACE_TAG); traceTag != "" {
			c.Set(internal.TT_TRACE_TAG, traceTag)
		}
		c.Set(internal.SNAMEKEY, psm)
		c.Set(internal.LOCALIPKEY, localIP)
		c.Set(internal.CLUSTERKEY, cluster)
		method := internal.GetHandlerNameByPath(c.Request.URL.Path)
		if method == "" {
			method = internal.GetHandlerName(c.Handler())
		}
		if method == "" {
			method = c.HandlerName()
			pos := strings.LastIndexByte(method, '.')
			if pos != -1 {
				method = c.HandlerName()[pos+1:]
			}
		}
		c.Set(internal.METHODKEY, method)
	}
}

// genLogId generates a global unique log id for request
// format: %Y%m%d%H%M%S + ip + 5位随机数
// python runtime使用的random uuid, 这里简单使用random产生一个5位数字随机串
func genLogId() string {
	buf := make([]byte, 0, 64)
	buf = time.Now().AppendFormat(buf, "20060102150405")
	buf = append(buf, fullLengthLocalIP...)

	uuidBuf := make([]byte, 4)
	_, err := rand.Read(uuidBuf)
	if err != nil {
		panic(err)
	}
	uuidNum := binary.BigEndian.Uint32(uuidBuf)
	buf = append(buf, fmt.Sprintf("%05d", uuidNum)[:5]...)
	return string(buf)
}

func init() {
	localIP = os.Getenv(internal.HOST_IP_ADDR)
	if localIP == "" {
		localIP = internal_util.LocalIP()
	}
	elements := strings.Split(localIP, ".")
	for i := 0; i < len(elements); i++ {
		elements[i] = fmt.Sprintf("%03s", elements[i])
	}
	fullLengthLocalIP = []byte(strings.Join(elements, ""))
}
