package kite

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"encoding/json"

	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/logs"
	"code.byted.org/kite/endpoint"
	"code.byted.org/kite/kitc"
	"code.byted.org/kite/kite/gls"
	"code.byted.org/kite/kitutil"
	"code.byted.org/kite/kitutil/kiterrno"
	"code.byted.org/kite/kitutil/logid"
	"code.byted.org/trace/trace-client-go"
	"github.com/opentracing/opentracing-go"
)

func empty2Default(str, dft string) string {
	if str == "" {
		return dft
	}
	return str
}

// AccessLogMW print access log
func AccessLogMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if DisableAccessLog {
			return next(ctx, request)
		}

		code := int32(kiterrno.SuccessCode)
		begin := time.Now()
		response, err := next(ctx, request)
		if err != nil {
			code = int32(kiterrno.UserErrorCode)
		}
		if resp, ok := response.(endpoint.KiteResponse); ok {
			if bp := resp.GetBaseResp(); bp != nil {
				code = bp.GetStatusCode()
			}
		}
		cost := time.Since(begin).Nanoseconds() / 1000 //us

		// access log
		r := GetRPCInfo(ctx)
		kvs := []interface{}{"method", r.Method,
			"rip", empty2Default(r.RemoteIP, "-"),
			"called", empty2Default(r.UpstreamService, "-"),
			"from_cluster", empty2Default(r.UpstreamCluster, "-"),
			"cost", cost,
			"status", code,
			"env", empty2Default(r.Env, "-")}

		if r.StressTag != "" {
			kvs = append(kvs, "stress_tag", r.StressTag)
		}

		ctx = context.WithValue(ctx, logs.LocationCtxKey{}, "access.go:0")
		if err != nil {
			AccessLogger.CtxErrorKVs(ctx, kvs...)
		} else {
			AccessLogger.CtxTraceKVs(ctx, kvs...)
		}
		return response, err
	}
}

// server site metrics format
const (
	successThroughputFmt string = "service.thrift.%s.%s.calledby.success.throughput"
	errorThroughputFmt   string = "service.thrift.%s.%s.calledby.error.throughput"
	successLatencyFmt    string = "service.thrift.%s.%s.calledby.success.latency.us"
	errorLatencyFmt      string = "service.thrift.%s.%s.calledby.error.latency.us"
	accessTotalFmt       string = "service.request.%s.total"

	statusSuccess string = "success"
	statusFailed  string = "failed"
)

// AccessMetricsMW emit access metrics
func AccessMetricsMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if ctx == nil {
			return next(ctx, request)
		}
		begin := time.Now()
		response, err := next(ctx, request)
		took := time.Since(begin).Nanoseconds() / 1000 // us

		r := GetRPCInfo(ctx)
		upstreamService := empty2Default(r.UpstreamService, "-")
		upstreamCluster := empty2Default(r.UpstreamCluster, "-")

		var tname, lname string
		if err == nil {
			tname = fmt.Sprintf(successThroughputFmt, r.Service, r.Method)
			lname = fmt.Sprintf(successLatencyFmt, r.Service, r.Method)
		} else {
			tname = fmt.Sprintf(errorThroughputFmt, r.Service, r.Method)
			lname = fmt.Sprintf(errorLatencyFmt, r.Service, r.Method)
		}

		// when error is nil, status code is user's status code, otherwise, it's framework's error code.
		statusCode := int32(kiterrno.SuccessCode)
		if err == nil {
			if resp, ok := response.(endpoint.KiteResponse); ok {
				if bp := resp.GetBaseResp(); bp != nil {
					statusCode = bp.GetStatusCode()
				}
			}
		} else {
			if frameworkError, ok := err.(*kiterrno.KitErr); ok {
				statusCode = int32(frameworkError.Errno())
			}
		}

		tags := map[string]string{
			"from":         upstreamService,
			"from_cluster": upstreamCluster,
			"to_cluster":   r.Cluster,
			"status_code":  strconv.Itoa(int(statusCode)),
		}
		if r.StressTag != "" {
			tags["stress_tag"] = r.StressTag
		} else {
			tags["stress_tag"] = "-"
		}
		if ServiceMeshMode {
			tags["mesh"] = "1"
		}

		metricsClient.EmitCounter(tname, 1, "", tags)
		metricsClient.EmitTimer(lname, took, "", tags)

		accessTotal := fmt.Sprintf(accessTotalFmt, r.Service)
		metricsClient.EmitCounter(accessTotal, 1, "", tags)

		return response, err
	}
}

const (
	recoverMW = "RecoverMW"
)

// RecoverMW print panic info to
func RecoverMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		defer func() {
			if e := recover(); e != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				logs.CtxError(ctx, "KITE: panic: Request is: %v", request)
				logs.CtxError(ctx, "KITE: panic in handler: %s: %s", e, buf)
				panic(recoverMW)
			}
		}()
		return next(ctx, request)
	}
}

// PushNoticeMW .
func PushNoticeMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx = logs.NewNoticeCtx(ctx)
		defer logs.CtxFlushNotice(ctx)
		return next(ctx, request)
	}
}

// BaseRespCheckMW .
func BaseRespCheckMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		resp, err := next(ctx, request)
		if err != nil {
			return nil, err
		}

		r := GetRPCInfo(ctx)
		response, ok := resp.(endpoint.KiteResponse)
		if !ok {
			panic(fmt.Sprintf("response type error in %s's %s method. The error type is %s.",
				r.Service, r.Method, reflect.TypeOf(resp)))
		}
		if response.GetBaseResp() == nil {
			panic(fmt.Sprintf("response's KiteBaseResp is nil in %s's %s method.", r.Service, r.Method))
		}
		return response, nil
	}
}

type extraBase interface {
	GetExtra() map[string]string
}

// ParserMW init rpcinfo
func ParserMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		rpcInfo := GetRPCInfo(ctx)
		rpcInfo.Service = ServiceName
		rpcInfo.Cluster = ServiceCluster
		rpcInfo.LocalIP = env.HostIP()
		if r, ok := request.(endpoint.KiteRequest); ok && r.IsSetBase() {
			b := r.GetBase()
			if extra, ok := b.(extraBase); ok {
				rpcInfo.StressTag = extra.GetExtra()["stress_tag"]
				rpcInfo.DDPTag = extra.GetExtra()["ddp_tag"]
				if userExtraStr := extra.GetExtra()["user_extra"]; len(userExtraStr) > 0 {
					var userExtra map[string]string
					if err := json.Unmarshal([]byte(userExtraStr), &userExtra); err != nil {
						logs.CtxError(ctx, "unmarshal user extra map err: %s", err.Error())
					} else {
						ctx = kitutil.NewCtxWithUpstreamUserExtra(ctx, userExtra)
					}
				}

			}

			rpcInfo.UpstreamService = b.GetCaller()
			rpcInfo.UpstreamCluster = b.GetCluster()
			rpcInfo.Env = b.GetEnv()
			rpcInfo.Client = b.GetClient()
			rpcInfo.RemoteIP = b.GetAddr()
			rpcInfo.LogID = b.GetLogID()
			if rpcInfo.LogID == "" {
				rpcInfo.LogID = logid.GetNginxID()
			}
		}

		if GetRealIP {
			if addrCode := gls.GetGID(); addrCode > 0 {
				rpcInfo.RemoteIP = decodeAddr(addrCode)
			}
		}

		if rpcInfo.UpstreamService == "" || rpcInfo.UpstreamService == "-" {
			rpcInfo.UpstreamService = "none"
		}
		if rpcInfo.UpstreamCluster == "" || rpcInfo.UpstreamCluster == "-" {
			rpcInfo.UpstreamCluster = "default"
		}

		if rpcInfo.DDPTag != "" {
			ctx = kitutil.NewCtxWithDDPRoutingTag(ctx, rpcInfo.DDPTag)
		}

		rpcInfo.RPCConfig = RPCServer.getRPCConfig(rpcInfo.RPCMeta)
		return next(ctx, request)
	}
}

// ExposeCtxMW expose some variable from RPCInfo to the context for compatibility
func ExposeCtxMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := GetRPCInfo(ctx)
		ctx = kitutil.NewCtxWithServiceName(ctx, r.Service)
		ctx = kitutil.NewCtxWithCluster(ctx, r.Cluster)
		ctx = kitutil.NewCtxWithCaller(ctx, r.UpstreamService)
		ctx = kitutil.NewCtxWithCallerCluster(ctx, r.UpstreamCluster)
		ctx = kitutil.NewCtxWithMethod(ctx, r.Method)
		ctx = kitutil.NewCtxWithLogID(ctx, r.LogID)
		ctx = kitutil.NewCtxWithEnv(ctx, r.Env)
		ctx = kitutil.NewCtxWithClient(ctx, r.Client)
		ctx = kitutil.NewCtxWithLocalIP(ctx, r.LocalIP)
		ctx = kitutil.NewCtxWithAddr(ctx, r.RemoteIP)
		ctx = kitutil.NewCtxWithStressTag(ctx, r.StressTag)
		ctx = kitutil.NewCtxWithTraceTag(ctx, r.TraceTag)
		return next(ctx, request)
	}
}

// ACLMW .
func ACLMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := GetRPCInfo(ctx)
		if !r.ACLAllow {
			return nil, kiterrno.NewKitErr(kiterrno.NotAllowedByACLCode,
				fmt.Errorf("upstream service=%s, cluster=%s", r.UpstreamService, r.UpstreamCluster))
		}
		return next(ctx, request)
	}
}

// StressBotMW .
func StressBotMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := GetRPCInfo(ctx)
		if !r.StressBotSwitch && r.StressTag != "" {
			return nil, kiterrno.NewKitErr(kiterrno.StressBotRejectionCode,
				fmt.Errorf("upstream service=%s, cluster=%s", r.UpstreamService, r.UpstreamCluster))
		}
		return next(ctx, request)
	}
}

// EndpointQPSLimitMW control the traffic on Endpoint
func EndpointQPSLimitMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := GetRPCInfo(ctx)

		if !RPCServer.overloader.TakeEndpointQPS(r.Method) {
			return nil, kiterrno.NewKitErr(kiterrno.EndpointQPSLimitRejectCode,
				fmt.Errorf("service=%s, cluster=%s method=%s", r.Service, r.Cluster, r.Method))
		}

		return next(ctx, request)
	}
}

// AdditionMW .
func AdditionMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := GetRPCInfo(ctx)
		method := r.Method
		if method == "" {
			method = "-"
		}
		if mw, ok := mMap[method]; ok {
			next = mw(next)
		}
		if userMW != nil {
			next = userMW(next)
		}
		return next(ctx, request)
	}
}

// OpentracingMW
func OpenTracingMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		rpcInfo := GetRPCInfo(ctx)
		if r, ok := request.(endpoint.KiteRequest); ok && r.IsSetBase() {
			b := r.GetBase()
			if extra, ok := b.(extraBase); ok {
				var span opentracing.Span
				normOperation := trace.FormatOperationName(rpcInfo.Service, rpcInfo.Method)
				if spanCtx, err := kitc.SpanCtxFromTextMap(extra.GetExtra()); err == nil {
					//logs.Info("span-ctx: %s from client-side", trace.JSpanContextToString(spanCtx))
					span = opentracing.StartSpan(normOperation, opentracing.ChildOf(spanCtx))
				} else {
					span = opentracing.StartSpan(normOperation)
				}
				// finishing span. opentracing.StartSpan should not return nil object
				defer span.Finish()
				ctx = opentracing.ContextWithSpan(ctx, span)
				rpcInfo.TraceTag = trace.JSpanContextToString(span.Context())
			}
		}

		fillSpanDataBeforeHandler(ctx, rpcInfo)
		resp, err := next(ctx, request)
		fillSpanDataAfterHandler(ctx, rpcInfo, resp, err)

		return resp, err
	}
}
