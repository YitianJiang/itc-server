package kitc

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"code.byted.org/gopkg/logs"
	"code.byted.org/kite/endpoint"
	"code.byted.org/kite/kitutil/kiterrno"
)

// TraceLogger .
type TraceLogger interface {
	Trace(format string, v ...interface{})
	Error(format string, v ...interface{})
}

// CtxTraceKVLogger .
type CtxTraceKVLogger interface {
	CtxTraceKVs(ctx context.Context, kvs ...interface{})
	CtxErrorKVs(ctx context.Context, kvs ...interface{})
}

var logger TraceLogger
var ctxKVLogger CtxTraceKVLogger

// SetCallLog which logger for logging calling logs
func SetCallLog(lg TraceLogger) {
	logger = lg
	if ctxKV, ok := lg.(CtxTraceKVLogger); ok {
		ctxKVLogger = ctxKV
	}
}

// localLogger implement kitware.TraceLogger interface
type localLogger struct{}

func (l *localLogger) Trace(format string, v ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", v...)
}
func (l *localLogger) Error(format string, v ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", v...)
}
func (l *localLogger) CtxTraceKVs(ctx context.Context, kvs ...interface{}) {
	fmt.Fprintln(os.Stdout, kvs...)
}
func (l *localLogger) CtxErrorKVs(ctx context.Context, kvs ...interface{}) {
	fmt.Fprintln(os.Stdout, kvs...)
}

type emptyLogger struct{}

func (l *emptyLogger) Trace(format string, v ...interface{})               {}
func (l *emptyLogger) Error(format string, v ...interface{})               {}
func (l *emptyLogger) CtxTraceKVs(ctx context.Context, kvs ...interface{}) {}
func (l *emptyLogger) CtxErrorKVs(ctx context.Context, kvs ...interface{}) {}

// NewRPCLogMW return a middleware for logging RPC record
func NewRPCLogMW(logger TraceLogger) endpoint.Middleware {
	return func(next endpoint.EndPoint) endpoint.EndPoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			code := int32(kiterrno.SuccessCode)
			begin := time.Now()
			response, err := next(ctx, request)
			if err != nil {
				code = int32(kiterrno.UserErrorCode)
			}
			if resp, ok := response.(endpoint.KitcCallResponse); ok {
				if bp := resp.GetBaseResp(); bp != nil {
					code = bp.GetStatusCode()
				}
			}
			cost := time.Since(begin).Nanoseconds() / 1000 // us

			rpcInfo := GetRPCInfo(ctx)
			remoteIP := ""
			if rpcInfo.TargetInstance != nil {
				remoteIP = rpcInfo.TargetInstance.Host
			}

			if ctxKVLogger != nil {
				kvs := []interface{}{"method", rpcInfo.Method,
					"rip", remoteIP,
					"called", rpcInfo.To,
					"to_cluster", rpcInfo.ToCluster,
					"from_method", rpcInfo.FromMethod,
					"cost", cost,
					"status", code,
					"conn_cost", int(rpcInfo.ConnCost),
					"env", rpcInfo.Env}
				if rpcInfo.StressTag != "" {
					kvs = append(kvs, "stress_tag", rpcInfo.StressTag)
				}
				ctx = context.WithValue(ctx, logs.LocationCtxKey{}, "call.go:0")
				if err != nil {
					kvs = append(kvs, "err", err)
					ctxKVLogger.CtxErrorKVs(ctx, kvs...)
				} else {
					ctxKVLogger.CtxTraceKVs(ctx, kvs...)
				}
			} else {
				ss := formatLog(rpcInfo.LocalIP, rpcInfo.From, rpcInfo.LogID, rpcInfo.FromCluster,
					rpcInfo.Method, remoteIP, rpcInfo.To, rpcInfo.ToCluster,
					cost, int64(code), rpcInfo.Env)
				if rpcInfo.StressTag != "" {
					ss += " stress_tag=" + rpcInfo.StressTag
				}
				if err != nil {
					logger.Error("%s, err=%s", ss, err)
				} else {
					logger.Trace("%s", ss)
				}
			}
			return response, err
		}
	}
}

func formatLog(ip, psm, logid, cluster, method, rip, rname, rcluster string, cost, code int64, env string) string {
	b := make([]byte, 0, 4096)
	b = append(b, ip...)
	b = append(b, ' ')
	b = append(b, psm...)
	b = append(b, ' ')
	b = append(b, logid...)
	b = append(b, ' ')
	b = append(b, cluster...)
	b = append(b, " method="...)
	b = append(b, method...)
	b = append(b, " rip="...)
	b = append(b, rip...)
	b = append(b, " called="...)
	b = append(b, rname...)
	b = append(b, " cluster="...)
	b = append(b, rcluster...)
	b = append(b, " cost="...)
	b = strconv.AppendInt(b, cost, 10)
	b = append(b, " status="...)
	b = strconv.AppendInt(b, code, 10)
	b = append(b, " env="...)
	b = append(b, env...)
	return string(b)
}
