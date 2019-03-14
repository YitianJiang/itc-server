package kitc

import (
	"code.byted.org/gopkg/env"
	"code.byted.org/kite/kitc/connpool"
	"code.byted.org/kite/kitutil/kiterrno"
	"code.byted.org/trace/trace-client-go"
	kext "code.byted.org/trace/trace-client-go/ext"
	tutil "code.byted.org/trace/trace-client-go/utils"
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	olog "github.com/opentracing/opentracing-go/log"
	"reflect"
	"strconv"
	"sync/atomic"
)

func fillSpanDataBeforeCall(ctx context.Context, rpcInfo *rpcInfo) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil || !trace.IsJaegerSpan(span) {
		return
	}

	ext.SpanKindRPCClient.Set(span)
	ext.PeerService.Set(span, trace.FormatOperationName(rpcInfo.To, rpcInfo.Method))
	ext.Component.Set(span, "kite")
	kext.RPCLogID.Set(span, rpcInfo.LogID)
	kext.LocalCluster.Set(span, env.Cluster())
	kext.LocalIDC.Set(span, env.IDC())
	kext.LocalAddress.Set(span, env.HostIP())
}

func fillSpanDataAfterCall(ctx context.Context, rpcInfo *rpcInfo, resp interface{}, err error) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil || !trace.IsJaegerSpan(span) {
		return
	}

	ext.Error.Set(span, err != nil)
	if err != nil {
		span.LogFields(olog.String("error.kind", err.Error()))
	}
	if code, exist := getRespCode(resp); exist {
		kext.ReturnCode.Set(span, int32(code))
	}

	if rpcInfo.TargetInstance != nil {
		if port, err := strconv.Atoi(rpcInfo.TargetInstance.Port); err == nil {
			ext.PeerPort.Set(span, uint16(port))
		}
		ext.PeerHostIPv4.Set(span, tutil.InetAtoN(rpcInfo.TargetInstance.Host))
		kext.PeerCluster.Set(span, rpcInfo.ToCluster)
		kext.PeerIDC.Set(span, rpcInfo.TargetIDC)
	}

	// kite-related error may cause rpcInfo.Conn concurrent read-write race condition (eg. RPCTimeout)
	if _, ok := err.(*kiterrno.KitErr); !ok {
		if conn, ok := rpcInfo.Conn.(*connpool.ConnWithPkgSize); ok {
			kext.RequestLength.Set(span, atomic.LoadInt32(&(conn.Written)))
			kext.ResponseLength.Set(span, atomic.LoadInt32(&(conn.Readn)))
		}
	}
}

func injectTraceIntoExtra(req interface{}, ctx context.Context) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	span := opentracing.SpanFromContext(ctx)
	if trace.IsSampled(span) {
		ptr := reflect.ValueOf(req)
		v := ptr.Elem()
		basePtr := v.FieldByName("Base")
		base := basePtr.Elem()
		extra := base.FieldByName("Extra")

		spanCtx := opentracing.SpanFromContext(ctx).Context()
		carrier := opentracing.TextMapCarrier{}
		if err = opentracing.GlobalTracer().Inject(spanCtx,
			opentracing.TextMap, carrier); err != nil {
			return
		}
		err = carrier.ForeachKey(func(key, val string) error {
			extra.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
			return nil
		})
	}

	return
}

func SpanCtxFromTextMap(extra map[string]string) (opentracing.SpanContext, error) {
	return opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(extra))
}
