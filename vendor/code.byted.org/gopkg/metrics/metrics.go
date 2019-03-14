// Package metrics provides a goroutine safe metrics client package metrics
// if TCE_HOST_IP is setted, will use this env value as host address
package metrics

import (
	"errors"
	"os"
	"strings"
)

type metricsType int

const (
	metricsTypeCounter metricsType = iota
	metricsTypeTimer
	metricsTypeStore
	metricsTypeTsStore
	metricsTypeRateCounter
)

func (t metricsType) String() string {
	switch t {
	case metricsTypeCounter:
		return "counter"
	case metricsTypeStore:
		return "store"
	case metricsTypeTsStore:
		return "ts_store"
	case metricsTypeTimer:
		return "timer"
	case metricsTypeRateCounter:
		return "rate_counter"
	}
	return "unknown"
}

const (
	BlackholeAddr = "blackhole"
)

var (
	DefaultMetricsServer = "127.0.0.1:9123"

	ErrDuplicatedMetrics    = errors.New("duplicated metrics name")
	ErrEmitUndefinedMetrics = errors.New("emit undefined metrics")
	ErrEmitBadMetricsType   = errors.New("emit bad metrics type")
	ErrEmitBufferFull       = errors.New("emit buffer full")
	ErrUnKnowValue          = errors.New("Unkown metrics value")

	extTags = make([]byte, 0, 4096)
)

func AddGlobalTag(name, value string) {
	extTags = appendTags(extTags, []T{{name, value}})
}

func ResetGlobalTag() {
	extTags = extTags[:0]
}

func init() {
	if host := strings.TrimSpace(os.Getenv("TCE_HOST_IP")); host != "" {
		DefaultMetricsServer = host + ":9123"
		AddGlobalTag("env_type", "tce")
		AddGlobalTag("pod_name", os.Getenv("MY_POD_NAME"))
		AddGlobalTag("_psm", os.Getenv("TCE_PSM"))
		AddGlobalTag("deploy_stage", os.Getenv("TCE_STAGE"))
	}
}
