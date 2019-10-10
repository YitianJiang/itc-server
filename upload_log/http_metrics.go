package uploadlog

import (
	"code.byted.org/gopkg/metrics"
)

var (
	mClient *metrics.MetricsClientV2
)

func initMetrics(psm string) {
	mClient = metrics.NewDefaultMetricsClientV2(psm, true)
}

func emitCounter(name string, count int64, tags map[string]string) {
	if tags == nil {
		mClient.EmitCounter(name, count)
	}
	kv := make([]metrics.T, 0, len(tags))
	for k, v := range tags {
		kv = append(kv, metrics.T{k, v})
	}
	mClient.EmitCounter(name, count, kv...)
}

func emitError(errType, domain string) {
	emitCounter("sdk_err_go", 1, map[string]string{"err_type": errType, "err_domain": domain})
	//emitCounter("error_domain", 1, map[string]string{"err_domain": domain})
	//mClient.EmitStore("error_content", err, metrics.T{"err_domain", domain})
}
