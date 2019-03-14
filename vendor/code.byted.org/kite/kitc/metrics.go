package kitc

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"code.byted.org/gopkg/metrics"
	"code.byted.org/kite/endpoint"
	"code.byted.org/kite/kitutil/kiterrno"
)

var (
	metricsClient *metrics.MetricsClient
)

const (
	namespacePrefix string = "toutiao"
)

func init() {
	metricsClient = metrics.NewDefaultMetricsClient(namespacePrefix, true)
}

// client site metrics format
const (
	successRPCThroughputFmt string = "service.thrift.%s.call.success.throughput"
	errorRPCThroughputFmt   string = "service.thrift.%s.call.error.throughput"
	successRPCLatencyFmt    string = "service.thrift.%s.call.success.latency.us"
	errorRPCLatencyFmt      string = "service.thrift.%s.call.error.latency.us"

	stabilityFmt string = "service.stability.%s.throughput"
)

// RPCMetricsMW .
func RPCMetricsMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		begin := time.Now()
		response, err := next(ctx, request)
		took := time.Since(begin).Nanoseconds() / 1000 //us

		code, ok := getRespCode(response)
		if ok && err != nil {
			switch code {
			case kiterrno.NotAllowedByACLCode,
				kiterrno.ForbiddenByDegradationCode:
				return response, err
			}
		}

		rpcInfo := GetRPCInfo(ctx)

		tname := fmt.Sprintf(successRPCThroughputFmt, rpcInfo.From)
		lname := fmt.Sprintf(successRPCLatencyFmt, rpcInfo.From)
		if err != nil {
			tname = fmt.Sprintf(errorRPCThroughputFmt, rpcInfo.From)
			lname = fmt.Sprintf(errorRPCLatencyFmt, rpcInfo.From)
		}

		tags := map[string]string{
			"mesh":         "0",
			"from":         rpcInfo.From,
			"from_cluster": rpcInfo.FromCluster,
			"to":           rpcInfo.To,
			"method":       rpcInfo.Method,
		}
		if rpcInfo.ToCluster != "" {
			tags["to_cluster"] = rpcInfo.ToCluster
		} else {
			tags["to_cluster"] = "default"
		}
		if err != nil {
			tags["err_code"] = strconv.Itoa(code)
		}
		if rpcInfo.StressTag != "" {
			tags["stress_tag"] = rpcInfo.StressTag
		} else {
			tags["stress_tag"] = "-"
		}
		metricsClient.EmitCounter(tname, 1, "", tags)
		metricsClient.EmitTimer(lname, took, "", tags)

		stabilityMetrics := fmt.Sprintf(stabilityFmt, rpcInfo.To)
		stabilityMap := make(map[string]string, len(tags))
		for k, v := range tags {
			stabilityMap[k] = v
		}
		if code < 0 {
			stabilityMap["label"] = "business_err"
			stabilityMap["err_code"] = strconv.Itoa(code)
		} else if code >= 100 && code < 200 {
			stabilityMap["label"] = "net_err"
			stabilityMap["err_code"] = strconv.Itoa(code)
		} else {
			stabilityMap["label"] = "success"
		}
		metricsClient.EmitCounter(stabilityMetrics, 1, "", stabilityMap)

		return response, err
	}
}

// RPCMetricsMW .
func MeshRPCMetricsMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		begin := time.Now()
		response, err := next(ctx, request)
		took := time.Since(begin).Nanoseconds() / 1000 //us

		code, ok := getRespCode(response)
		if ok && err != nil {
			switch code {
			case kiterrno.NotAllowedByACLCode,
				kiterrno.ForbiddenByDegradationCode:
				return response, err
			}
		}

		rpcInfo := GetRPCInfo(ctx)

		tname := fmt.Sprintf(successRPCThroughputFmt, rpcInfo.From)
		lname := fmt.Sprintf(successRPCLatencyFmt, rpcInfo.From)
		if err != nil {
			tname = fmt.Sprintf(errorRPCThroughputFmt, rpcInfo.From)
			lname = fmt.Sprintf(errorRPCLatencyFmt, rpcInfo.From)
		}

		tags := map[string]string{
			"mesh":         "1",
			"from":         rpcInfo.From,
			"from_cluster": rpcInfo.FromCluster,
			"to":           rpcInfo.To,
			"method":       rpcInfo.Method,
		}
		if rpcInfo.ToCluster != "" {
			tags["to_cluster"] = rpcInfo.ToCluster
		} else {
			tags["to_cluster"] = "default"
		}
		if err != nil {
			tags["err_code"] = strconv.Itoa(code)
		}
		if rpcInfo.StressTag != "" {
			tags["stress_tag"] = rpcInfo.StressTag
		} else {
			tags["stress_tag"] = "-"
		}
		metricsClient.EmitCounter(tname, 1, "", tags)
		metricsClient.EmitTimer(lname, took, "", tags)

		stabilityMetrics := fmt.Sprintf(stabilityFmt, rpcInfo.To)
		stabilityMap := make(map[string]string, len(tags))
		for k, v := range tags {
			stabilityMap[k] = v
		}
		if code < 0 {
			stabilityMap["label"] = "business_err"
			stabilityMap["err_code"] = strconv.Itoa(code)
		} else if code >= 100 && code < 200 {
			stabilityMap["label"] = "net_err"
			stabilityMap["err_code"] = strconv.Itoa(code)
		} else {
			stabilityMap["label"] = "success"
		}
		metricsClient.EmitCounter(stabilityMetrics, 1, "", stabilityMap)

		return response, err
	}
}
