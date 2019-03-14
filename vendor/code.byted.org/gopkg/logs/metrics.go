package logs

import (
	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/metrics"
)

var (
	metricsClient *metrics.MetricsClient

	metricsTagWarn  = map[string]string{"level": "WARNING", "cluster": env.Cluster()}
	metricsTagError = map[string]string{"level": "ERROR", "cluster": env.Cluster()}
	metricsTagFatal = map[string]string{"level": "CRITICAL", "cluster": env.Cluster()} // 和py统一, 将fatal打成critical
	metricsLim      = 4                                                                //  只打Warn及以上的日志,
)

func init() {
	metricsClient = metrics.NewDefaultMetricsClient("toutiao.service.log", true)
}

// FIXME: it's strange to do metrics in logs SDK
func doMetrics(logLevel int, psm string) {
	if logLevel < metricsLim {
		return
	}

	if len(psm) == 0 {
		return
	}

	if logLevel == 4 { // warning
		metricsClient.EmitCounter(psm+".throughput", 1, "", metricsTagWarn)
	} else if logLevel == 5 { // error
		metricsClient.EmitCounter(psm+".throughput", 1, "", metricsTagError)
	} else if logLevel == 6 { // fatal
		metricsClient.EmitCounter(psm+".throughput", 1, "", metricsTagFatal)
	}
}
