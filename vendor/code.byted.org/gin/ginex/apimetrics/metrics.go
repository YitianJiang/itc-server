package apimetrics

import (
	"fmt"
	"strconv"
	"time"

	"code.byted.org/gin/ginex/internal"
	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/metrics"
	"github.com/gin-gonic/gin"
)

const (
	METRICS_PREFIX           = "toutiao.service.http"
	FRAMEWORK_METRICS_PREFIX = "toutiao.service"
)

var (
	emitter = metrics.NewDefaultMetricsClient(METRICS_PREFIX, true)
)

func Metrics(psm string) gin.HandlerFunc {
	succLatencyMetrics := fmt.Sprintf("%s.calledby.success.latency.us", psm)
	succThroughputMetrics := fmt.Sprintf("%s.calledby.success.throughput", psm)
	errLatencyMetrics := fmt.Sprintf("%s.calledby.error.latency.us", psm)
	errThroughputMetrics := fmt.Sprintf("%s.calledby.error.throughput", psm)
	frameWorkThroughputMetrics := "ginex.throughput"
	frameWorkMetricsTags := map[string]string{
		"version": internal.VERSION,
	}
	return func(c *gin.Context) {
		if psm == "" {
			c.Next()
			return
		}
		defer func() {
			if e := recover(); e != nil {
				EmitPanicCounter(psm)
				panic(e)
			}
		}()

		start := time.Now()
		c.Next()
		end := time.Now()
		latency := end.Sub(start).Nanoseconds() / 1000
		var handleMethod string
		stressTag := "-"
		if v := c.Value(internal.METHODKEY); v != nil {
			// 如果METHODKEY设置了,那它一定是string
			handleMethod = v.(string)
		}
		if v := c.Value(internal.STRESSKEY); v != nil {
			// 如果STRESSKEY被设置了，那他一定是string类型
			stressTag = v.(string)
		}
		tags := map[string]string{
			"status":        strconv.Itoa(c.Writer.Status()),
			"handle_method": handleMethod,
			"from_cluster":  "default",
			"to_cluster":    env.Cluster(),
			"stress_tag":    stressTag,
		}
		// https://wiki.bytedance.net/pages/viewpage.action?pageId=51348664
		code := c.Writer.Status()
		if code >= 200 && code < 400 {
			emitter.EmitTimer(succLatencyMetrics, latency, "", tags)
			emitter.EmitCounter(succThroughputMetrics, 1, "", tags)
		} else {
			emitter.EmitTimer(errLatencyMetrics, latency, "", tags)
			emitter.EmitCounter(errThroughputMetrics, 1, "", tags)
		}
		// emit framework metrics
		emitter.EmitCounter(frameWorkThroughputMetrics, 1, FRAMEWORK_METRICS_PREFIX, frameWorkMetricsTags)
	}
}

// panic埋点, 业务可调用该方法统一埋点.
func EmitPanicCounter(psm string)  {
	panicMetrics := fmt.Sprintf("%s.panic", psm)
	emitter.EmitCounter(panicMetrics, 1, "", make(map[string]string))
}