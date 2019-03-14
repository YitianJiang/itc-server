package kitc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"time"

	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/thrift"
	"code.byted.org/kite/endpoint"
	circuit "code.byted.org/kite/kitc/circuitbreaker"
	"code.byted.org/kite/kitc/connpool"
	"code.byted.org/kite/kitc/discovery"
	"code.byted.org/kite/kitc/loadbalancer"
	"code.byted.org/kite/kitc/rand"
	"code.byted.org/kite/kitutil"
	"code.byted.org/kite/kitutil/kiterrno"
	trace "code.byted.org/trace/trace-client-go"
	kext "code.byted.org/trace/trace-client-go/ext"

	"github.com/cespare/xxhash"
	opentracing "github.com/opentracing/opentracing-go"
)

// clientBase implement endpoint.BaseInterface
type clientBase struct {
	logID   string
	caller  string
	client  string
	addr    string
	env     string
	cluster string
}

// GetLogID return logid
func (cb *clientBase) GetLogID() string {
	return cb.logID
}

// GetCaller return caller
func (cb *clientBase) GetCaller() string {
	return cb.caller
}

// GetClient return client
func (cb *clientBase) GetClient() string {
	return cb.client
}

// GetAddr return addr
func (cb *clientBase) GetAddr() string {
	return cb.addr
}

// GetEnv return this request's env
func (cb *clientBase) GetEnv() string {
	return cb.env
}

// GetCluster return upstream's cluster's name
func (cb *clientBase) GetCluster() string {
	return cb.cluster
}

func (cb *clientBase) GetExtra() map[string]string {
	// no-op
	return nil
}

// BaseWriterMW write base info to request
func BaseWriterMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(endpoint.KitcCallRequest)
		if !ok {
			logs.CtxWarn(ctx, "request %v not implement KitcCallRequest", request)
			return next(ctx, request)
		}
		rpcInfo := GetRPCInfo(ctx)

		req.SetBase(&clientBase{
			logID:   rpcInfo.LogID,
			caller:  rpcInfo.From,
			client:  rpcInfo.Client,
			addr:    rpcInfo.LocalIP,
			cluster: rpcInfo.FromCluster,
			env:     rpcInfo.Env,
		})

		realRequest := req.RealRequest()

		if rpcInfo.StressTag != "" {
			if err := hackExtra(realRequest, "stress_tag", rpcInfo.StressTag); err != nil {
				logs.CtxWarn(ctx, "set extra err: %s", err.Error())
			}
		}

		if rpcInfo.DDPTag != "" {
			if err := hackExtra(realRequest, "ddp_tag", rpcInfo.DDPTag); err != nil {
				logs.CtxWarn(ctx, "set extra err: %s", err.Error())
			}
		}

		if userExtra, ok := kitutil.GetCtxDownstreamUserExtra(ctx); ok && len(userExtra) > 0 {
			if extraStr, err := json.Marshal(userExtra); err != nil {
				logs.CtxError(ctx, "set user extra err: %s", err.Error())
			} else {
				if err := hackExtra(realRequest, "user_extra", string(extraStr)); err != nil {
					logs.CtxError(ctx, "set user extra err: %s", err.Error())
				}
			}
		}

		return next(ctx, req)
	}
}

func hackExtra(req interface{}, key, val string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	ptr := reflect.ValueOf(req)
	v := ptr.Elem()
	basePtr := v.FieldByName("Base")
	base := basePtr.Elem()
	extra := base.FieldByName("Extra")
	extra.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
	return nil
}

func getRespCode(resp interface{}) (int, bool) {
	kitResp, ok := resp.(endpoint.KitcCallResponse)
	if ok == false { // if resp is invalid, return directly
		return 0, false
	}

	baseResp := kitResp.GetBaseResp()
	if baseResp == nil {
		return 0, false
	}

	code := baseResp.GetStatusCode()
	return int(code), true
}

// IDCSelectorMW .
func IDCSelectorMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		rpcInfo := GetRPCInfo(ctx)
		idcConfig := rpcInfo.IDCConfig
		rpcInfo.TargetIDC = selectIDC(idcConfig)
		return next(ctx, request)
	}
}

func selectIDC(idcConfig []IDCConfig) string {
	if len(idcConfig) == 0 {
		return env.IDC()
	}

	var sum int
	for _, pl := range idcConfig {
		sum += pl.Percent
	}

	if sum == 0 {
		return env.IDC()
	}

	rd := rand.Intn(sum)
	for _, pl := range idcConfig {
		if rd < pl.Percent {
			return pl.IDC
		}
		rd -= pl.Percent
	}
	return idcConfig[0].IDC
}

func NewIDCSelectorMW(GetKey loadbalancer.KeyFunc) endpoint.Middleware {
	if GetKey == nil {
		return IDCSelectorMW
	}
	return func(next endpoint.EndPoint) endpoint.EndPoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			rpcInfo := GetRPCInfo(ctx)
			idcConfig := rpcInfo.IDCConfig

			var lbRequest = request
			if r := ctx.Value(lbKey); r != nil {
				lbRequest = r
			}

			if hashKey, err := GetKey(ctx, lbRequest); err != nil {
				rpcInfo.TargetIDC = selectIDC(idcConfig)
			} else {
				hashCode := xxhash.Sum64String(hashKey)
				rpcInfo.TargetIDC = selectIDCWithHash(hashCode, idcConfig)
			}
			return next(ctx, request)
		}
	}
}

func selectIDCWithHash(code uint64, idcConfig []IDCConfig) string {
	if len(idcConfig) == 0 {
		return env.IDC()
	}

	var sum int
	for _, pl := range idcConfig {
		sum += pl.Percent
	}

	if sum == 0 {
		return env.IDC()
	}

	sort.Slice(idcConfig, func(i, j int) bool {
		return idcConfig[i].IDC < idcConfig[j].IDC
	})

	rd := int(code % uint64(sum))
	for _, pl := range idcConfig {
		if rd < pl.Percent {
			return pl.IDC
		}
		rd -= pl.Percent
	}
	return idcConfig[0].IDC
}

// IOErrorHandlerMW .
func IOErrorHandlerMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		defer func(begin time.Time) {
			rpcInfo := GetRPCInfo(ctx)
			tags := map[string]string{
				"to":           rpcInfo.To,
				"method":       rpcInfo.Method,
				"from_cluster": rpcInfo.FromCluster,
				"to_cluster":   rpcInfo.ToCluster,
			}
			mname := fmt.Sprintf("service.thrift.%s.call.thrift.latency.us", rpcInfo.From)
			cost := time.Since(begin).Nanoseconds() / 1000 //us
			metricsClient.EmitTimer(mname, cost, "", tags)
		}(time.Now())

		trace.FillSpanEvent(ctx, kext.EventKindPkgSendStartEnum)
		resp, err := next(ctx, request)
		if err == nil {
			trace.FillSpanEvent(ctx, kext.EventKindPkgRecvEndEnum)
			return resp, nil
		}
		if terr, ok := err.(thrift.TApplicationException); ok {
			perr := kiterrno.NewProxyException(terr.TypeId(), terr.Error())
			return kiterrno.NewErrResp(terr.TypeId()), perr
		}
		return kiterrno.ErrRespRemoteOrNet, fmt.Errorf("remote or network err: %s", err.Error())
	}
}

func makeTimeoutErr(rpcInfo *rpcInfo) error {
	timeout := time.Duration(rpcInfo.RPCTimeout) * time.Millisecond
	errMsg := fmt.Sprintf("rpc timeout: timeout=%v, to=%s, to_cluster=%s, method=%s",
		timeout, rpcInfo.To, rpcInfo.ToCluster, rpcInfo.Method)
	target := rpcInfo.TargetInstance
	if target != nil {
		errMsg = fmt.Sprintf("%s, remote=%s:%s", errMsg, target.Host, target.Port)
	}
	return kiterrno.NewKitErr(kiterrno.RPCTimeoutCode, errors.New(errMsg))
}

// RPCTimeoutMW .
func RPCTimeoutMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		rpcInfo := GetRPCInfo(ctx)
		ctx, cancel := context.WithTimeout(ctx, time.Duration(rpcInfo.RPCTimeout)*time.Millisecond)
		defer cancel()

		var resp interface{}
		var err error
		done := make(chan error, 1)
		go func() {
			defer func() {
				if err := recover(); err != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]

					logs.CtxError(ctx, "KITC: panic: to=%s, toCluster=%s, method=%s, Request: %v, err: %v\n%s",
						rpcInfo.To, rpcInfo.ToCluster, rpcInfo.Method, request, err, buf)

					done <- fmt.Errorf("KITC: panic, %v\n%s", err, buf)
				}
				close(done)
			}()

			resp, err = next(ctx, request)
		}()

		select {
		case panicErr := <-done:
			if panicErr != nil {
				panic(panicErr.Error()) // throws panic error
			}
			return resp, err
		case <-ctx.Done():
			return kiterrno.ErrRespRPCTimeout, makeTimeoutErr(rpcInfo)
		}
	}
}

// DegradationMW .
func DegradationMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		rpcInfo := GetRPCInfo(ctx)
		if rpcInfo.DegraPercent <= 0 {
			return next(ctx, request)
		}
		if rpcInfo.DegraPercent >= 100 {
			kerr := kiterrno.NewKitErr(kiterrno.ForbiddenByDegradationCode, nil)
			return kiterrno.ErrRespForbiddenByDegradation, kerr
		}

		per := rand.Intn(101)
		if per < rpcInfo.DegraPercent {
			kerr := kiterrno.NewKitErr(kiterrno.ForbiddenByDegradationCode, nil)
			return kiterrno.ErrRespForbiddenByDegradation, kerr
		}

		return next(ctx, request)
	}
}

// NewPoolMW .
func NewPoolMW(pooler connpool.ConnPool) endpoint.Middleware {
	return func(next endpoint.EndPoint) endpoint.EndPoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			rpcInfo := GetRPCInfo(ctx)
			begin := time.Now()

			trace.FillSpanEvent(ctx, kext.EventKindConnectStartEnum)
			conn, err := pooler.Get(rpcInfo.TargetInstance.Host,
				rpcInfo.TargetInstance.Port,
				time.Duration(rpcInfo.ConnectTimeout)*time.Millisecond)
			rpcInfo.ConnCost = time.Now().Sub(begin) / time.Microsecond
			trace.FillSpanEvent(ctx, kext.EventKindConnectEndEnum)

			if err != nil {
				kerr := kiterrno.NewKitErr(kiterrno.GetConnErrorCode, err)
				return kiterrno.ErrRespGetConnError, kerr
			}

			rpcInfo.Conn = &connpool.ConnWithPkgSize{
				Conn: conn,
			}
			return next(ctx, request)
		}
	}
}

// RPCACLMW .
func RPCACLMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		rpcInfo := GetRPCInfo(ctx)
		if !rpcInfo.ACLAllow {
			aclErr := kiterrno.NewKitErr(kiterrno.NotAllowedByACLCode, nil)
			return kiterrno.ErrRespNotAllowedByACL, aclErr
		}

		return next(ctx, request)
	}
}

// StressBotMW .
func StressBotMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := GetRPCInfo(ctx)
		if !r.StressBotSwitch && r.StressTag != "" {
			return kiterrno.ErrStressBotRejection, kiterrno.NewKitErr(kiterrno.StressBotRejectionCode,
				fmt.Errorf("to service=%s, cluster=%s, stress_tag=%s", r.To, r.ToCluster, r.StressTag))
		}
		return next(ctx, request)
	}
}

// NewInstanceBreakerMW .
func NewInstanceBreakerMW(breakerPanel *circuit.Panel) endpoint.Middleware {
	return func(next endpoint.EndPoint) endpoint.EndPoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			rpcInfo := GetRPCInfo(ctx)
			key := rpcInfo.TargetInstance.Host + ":" + rpcInfo.TargetInstance.Port

			if !breakerPanel.IsAllowed(key) {
				targetService := rpcInfo.To
				targetMethod := rpcInfo.Method
				ins := rpcInfo.TargetInstance
				err := fmt.Errorf("service=%s method=%s hostport=%s",
					targetService, targetMethod, key)
				kerr := kiterrno.NewKitErr(kiterrno.NotAllowedByInstanceCBCode, err)
				return kiterrno.NewErrRespWithAddr(kiterrno.NotAllowedByInstanceCBCode, ins.Host+":"+ins.Port), kerr
			}

			resp, err := next(ctx, request)
			if err == nil {
				breakerPanel.Succeed(key)
				return resp, err
			}

			// failed and using error code to decide if ignore it
			// if response is nil, regard it as success, because this RPC is done.
			code, ok := getRespCode(resp)
			if !ok { // ignore
				return resp, err
			}

			// only control connection error
			switch code {
			case kiterrno.GetConnErrorCode:
				breakerPanel.Fail(key)
			default:
				breakerPanel.Succeed(key)
			}
			return resp, err
		}
	}
}

func NewUserErrorCBMW(breakerPanel *circuit.Panel, checkUserErrHandlers map[string]CheckUserError) endpoint.Middleware {
	return func(next endpoint.EndPoint) endpoint.EndPoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			if checkUserErrHandlers == nil {
				return next(ctx, request)
			}

			rpcInfo := GetRPCInfo(ctx)
			cbKey := rpcInfo.To + ":" + rpcInfo.ToCluster + ":" + rpcInfo.Method
			if !breakerPanel.IsAllowed(cbKey) {
				err := fmt.Errorf("service:cluter:mthod=%s", cbKey)
				kerr := kiterrno.NewKitErr(kiterrno.NotAllowedByUserErrCBCode, err)
				return kiterrno.ErrRespNotAllowedByUserErrCB, kerr
			}

			resp, err := next(ctx, request)
			if err == nil && checkUserErrHandlers[rpcInfo.Method] != nil {
				if !checkUserErrHandlers[rpcInfo.Method](resp) {
					breakerPanel.Succeed(cbKey)
				} else {
					breakerPanel.Fail(cbKey)
				}
			}
			return resp, err
		}
	}
}

// NewServiceBreakerMW .
func NewServiceBreakerMW(breakerPanel *circuit.Panel) endpoint.Middleware {
	return func(next endpoint.EndPoint) endpoint.EndPoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			rpcInfo := GetRPCInfo(ctx)
			conf := rpcInfo.ServiceCB

			if !conf.IsOpen {
				return next(ctx, request)
			}

			// 由于上下游都是分布式, 在上游单点进行下游整体的并发数限制,
			// 几乎没有意义, 因此剔除该功能;

			cbKey := rpcInfo.To + ":" + rpcInfo.ToCluster + ":" + rpcInfo.Method
			if !breakerPanel.IsAllowed(cbKey) {
				err := fmt.Errorf("service:cluster:method=%s", cbKey)
				kerr := kiterrno.NewKitErr(kiterrno.NotAllowedByServiceCBCode, err)
				return kiterrno.ErrRespNotAllowedByServiceCB, kerr
			}

			resp, err := next(ctx, request)
			if err == nil { // succeed
				breakerPanel.Succeed(cbKey)
				return resp, err
			}

			code, ok := getRespCode(resp)
			if !ok {
				code = kiterrno.UnknownErr
				logs.CtxError(ctx, "invalid response without status code: %v", resp)
			}

			switch code {
			// ignore all internal errors(like NoExpectedField, IDCSelectError) and
			// all ACL and degradation errors, and
			// all RPC timeout errors which have already been recored when the MW receive this error;
			case kiterrno.NotAllowedByACLCode,
				kiterrno.ForbiddenByDegradationCode,
				kiterrno.GetDegradationPercentErrorCode,
				kiterrno.BadConnBalancerCode,
				kiterrno.BadConnRetrierCode,
				kiterrno.ServiceDiscoverCode:
			// regard all network errors and relative errors caused by network as failed
			case kiterrno.NotAllowedByServiceCBCode,
				kiterrno.NotAllowedByInstanceCBCode,
				kiterrno.ConnRetryCode,
				kiterrno.GetConnErrorCode,
				kiterrno.RemoteOrNetErrCode:
				breakerPanel.FailWithTrip(cbKey, circuit.RateTripFunc(conf.ErrRate, int64(conf.MinSample)))
			case kiterrno.RPCTimeoutCode:
				breakerPanel.TimeoutWithTrip(cbKey, circuit.RateTripFunc(conf.ErrRate, int64(conf.MinSample)))
			default:
				breakerPanel.Succeed(cbKey)
			}
			return resp, err
		}
	}
}

// NewLoadbalanceMW .
func NewLoadbalanceMW(lb loadbalancer.Loadbalancer) endpoint.Middleware {
	return func(next endpoint.EndPoint) endpoint.EndPoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			rpcInfo := GetRPCInfo(ctx)
			insKey := rpcInfo.To + ":" + rpcInfo.TargetIDC + ":" +
				rpcInfo.ToCluster + ":" + rpcInfo.Env
			var lbRequest = request
			if r := ctx.Value(lbKey); r != nil {
				lbRequest = r
			}
			picker := lb.NewPicker(ctx, lbRequest, insKey, rpcInfo.Instances)

			var errs []error
			var resp interface{}
			var err error
			for {
				select {
				case <-ctx.Done():
					return kiterrno.ErrRespRPCTimeout, makeTimeoutErr(rpcInfo)
				default:
				}

				targetIns, ok := picker.Pick()
				if !ok {
					errs = append(errs, errors.New("No more instances to retry"))
					kerr := kiterrno.NewKitErr(kiterrno.ConnRetryCode, joinErrs(errs))
					return kiterrno.ErrRespConnRetry, kerr
				}

				rpcInfo.TargetInstance = targetIns
				resp, err = next(ctx, request)
				if err == nil {
					return resp, err
				}
				errs = append(errs, newConnErr(targetIns, err))

				code, ok := getRespCode(resp)
				if !ok {
					break
				}

				switch code {
				case kiterrno.NotAllowedByInstanceCBCode:
					continue
				case kiterrno.GetConnErrorCode:
					logs.CtxWarn(ctx, "KITC: get conn for %v:%v err: %v", rpcInfo.To, rpcInfo.ToCluster, err)
					continue
				}
				break
			}

			return resp, err
		}
	}
}

// OpenTracingMW
func OpenTracingMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		if _, ok := opentracing.GlobalTracer().(opentracing.NoopTracer); ok {
			return next(ctx, request)
		}
		r := GetRPCInfo(ctx)
		normOperation := trace.FormatOperationName(r.From, r.FromMethod)
		span, ctx := opentracing.StartSpanFromContext(ctx, normOperation)

		// pass trace context out of process using Base.Extra
		injectTraceIntoExtra(request, ctx)
		fillSpanDataBeforeCall(ctx, r)
		resp, err := next(ctx, request)
		fillSpanDataAfterCall(ctx, r, resp, err)

		// finishing span
		span.Finish()
		return resp, err
	}
}

func newConnErr(ins *discovery.Instance, err error) error {
	return fmt.Errorf("ins=%s:%s err=%s", ins.Host, ins.Port, err)
}
