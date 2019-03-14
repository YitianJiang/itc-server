package goredis

import (
	"context"
	"os"
	"strings"
	"sync"

	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/metrics"
	redis "code.byted.org/kv/redis-v6"
)

const (
	CALLSTATUS_THROUGHPUT = "throughput"
	CALLSTATUS_LATENCY    = "latency"
	CALLSTATUS_SUCCESS    = "success"
	CALLSTATUS_MISS       = "miss"
	CALLSTATUS_ERROR      = "error"
)

const (
	METRICSNAME_PREFIX   = "toutiao.service.thrift."
	METRICSCLIENT_PREFIX = "redis.client"
	REDIS_PSM_PREFIX     = "toutiao.redis."
	REDIS_PREFIX         = "redis_"
	ABASE_PSM_PREFIX     = "toutiao.abase."
	ABASE_PREFIX         = "abase_"
)

const (
	SuccessThroughput = "call.success.throughput"
	MissThroughput    = "call.miss.throughput"
	ErrThrouphput     = "call.error.throughput"
	SuccessLatency    = "call.success.latency.us"
	MissLatency       = "call.miss.latency.us"
	ErrLatency        = "call.error.latency.us"
	ClusterDefault    = "default"
)

var metricsClient *metrics.MetricsClient = metrics.NewDefaultMetricsClient(METRICSCLIENT_PREFIX, true)
var metricsClientWithPsm *metrics.MetricsClient = metrics.NewDefaultMetricsClient(METRICSNAME_PREFIX+env.PSM(), true)

func init() {
	metricsClient.DefineCounter(CALLSTATUS_THROUGHPUT, "")
	metricsClient.DefineCounter(CALLSTATUS_ERROR, "")
	metricsClient.DefineCounter(CALLSTATUS_MISS, "")
	metricsClient.DefineTimer(CALLSTATUS_LATENCY, "") //us

	metricsClientWithPsm.DefineCounter(SuccessThroughput, "")
	metricsClientWithPsm.DefineCounter(ErrThrouphput, "")
	metricsClientWithPsm.DefineCounter(MissThroughput, "")
	metricsClientWithPsm.DefineTimer(SuccessLatency, "")
	metricsClientWithPsm.DefineTimer(ErrLatency, "")
	metricsClientWithPsm.DefineTimer(MissLatency, "")
}

var mpool = &sync.Pool{New: func() interface{} { return make(map[string]string, 5) }}

func addCallMetrics(ctx context.Context, cmd string, latency int64, err error, cluster string, psm string, redisPsm string, counter int) {
	tagkvGoredis := mpool.Get().(map[string]string)
	tagkvGoredis["cluster"] = cluster
	tagkvGoredis["caller"] = psm
	tagkvGoredis["cmd"] = cmd
	tagkvGoredis["lang"] = "go"

	tagkvThrift := mpool.Get().(map[string]string)
	tagkvThrift["method"] = cmd
	tagkvThrift["to"] = redisPsm
	tagkvThrift["from_cluster"] = env.Cluster()
	tagkvThrift["to_cluster"] = ClusterDefault

	if stressTag, ok := getStressTag(ctx); ok {
		tagkvThrift["stress_tag"] = stressTag
	} else {
		tagkvThrift["stress_tag"] = "-"
	}
	switch {
	case err == nil:
		metricsClient.EmitCounter(CALLSTATUS_THROUGHPUT, counter, "", tagkvGoredis)
		metricsClientWithPsm.EmitCounter(SuccessThroughput, counter, "", tagkvThrift)
		if latency != -1 {
			metricsClientWithPsm.EmitTimer(SuccessLatency, latency, "", tagkvThrift)
		}
	case err == redis.Nil:
		metricsClient.EmitCounter(CALLSTATUS_MISS, counter, "", tagkvGoredis)
		metricsClientWithPsm.EmitCounter(MissThroughput, counter, "", tagkvThrift)
		if latency != -1 {
			metricsClientWithPsm.EmitTimer(MissLatency, latency, "", tagkvThrift)
		}
	default:
		metricsClient.EmitCounter(CALLSTATUS_ERROR, counter, "", tagkvGoredis)
		metricsClientWithPsm.EmitCounter(ErrThrouphput, counter, "", tagkvThrift)
		if latency != -1 {
			metricsClientWithPsm.EmitTimer(ErrLatency, latency, "", tagkvThrift)
		}
	}
	if latency != -1 {
		metricsClient.EmitTimer(CALLSTATUS_LATENCY, latency, "", tagkvGoredis)
	}

	mpool.Put(tagkvGoredis)
	mpool.Put(tagkvThrift)
}

// TODO update DC info
func getDcName(ip string) string {
	if ip == "" {
		return "None"
	} else {
		if strings.HasPrefix(ip, "10.4.") {
			return "hy"
		} else if strings.HasPrefix(ip, "10.6.") || strings.HasPrefix(ip, "10.3.") {
			return "lf"
		} else {
			return "Unidentified"
		}
	}
}

func checkPsm() string {
	psm := os.Getenv("TCE_PSM")
	if len(psm) == 0 {
		psm = os.Getenv("PSM")
	}
	if len(psm) == 0 {
		psm = os.Getenv("SVC_NAME")
	}
	if len(psm) == 0 {
		psm = "redis.psm.none"
	}
	return psm
}

// return redis_XXX / abase_XXX
func GetClusterName(str string) string {
	clusterName := str
	if strings.HasPrefix(str, REDIS_PSM_PREFIX) {
		clusterName = REDIS_PREFIX + clusterName[len(REDIS_PSM_PREFIX):]
	} else if strings.HasPrefix(str, ABASE_PSM_PREFIX) {
		clusterName = ABASE_PREFIX + clusterName[len(ABASE_PSM_PREFIX):]
	}
	return clusterName
}

// return toutiao.redis.XXX / toutiao.abase.XXX
func GetPSMClusterName(str string) string {
	PSMClusterName := str
	if strings.HasPrefix(str, REDIS_PREFIX) {
		PSMClusterName = REDIS_PSM_PREFIX + PSMClusterName[len(REDIS_PREFIX):]
	} else if strings.HasPrefix(str, ABASE_PREFIX) {
		PSMClusterName = ABASE_PSM_PREFIX + PSMClusterName[len(ABASE_PREFIX):]
	}
	return PSMClusterName
}
