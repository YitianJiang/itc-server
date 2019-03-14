package kite

import (
	"code.byted.org/gopkg/env"
	"code.byted.org/kite/endpoint"
	"code.byted.org/trace/trace-client-go"
	kext "code.byted.org/trace/trace-client-go/ext"
	tutils "code.byted.org/trace/trace-client-go/utils"
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	olog "github.com/opentracing/opentracing-go/log"
)

func fillSpanDataBeforeHandler(ctx context.Context, rpcInfo *RPCInfo) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil || !trace.IsJaegerSpan(span) {
		return
	}

	ext.Component.Set(span, "kite")
	ext.SpanKindRPCServer.Set(span)
	ext.PeerService.Set(span, trace.FormatOperationName(rpcInfo.UpstreamService, "-"))
	ext.PeerHostIPv4.Set(span, tutils.InetAtoN(rpcInfo.RemoteIP))
	kext.PeerCluster.Set(span, rpcInfo.UpstreamCluster)

	kext.RPCLogID.Set(span, rpcInfo.LogID)
	kext.LocalIDC.Set(span, env.IDC())
	kext.LocalCluster.Set(span, env.Cluster())
	kext.LocalAddress.Set(span, env.HostIP())
}

func fillSpanDataAfterHandler(ctx context.Context, rpcInfo *RPCInfo, resp interface{}, err error) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil || !trace.IsJaegerSpan(span) {
		return
	}

	ext.Error.Set(span, err != nil)
	if err != nil {
		span.LogFields(olog.String("error.kind", err.Error()))
	}
	if response, ok := resp.(endpoint.KiteResponse); ok {
		if response.GetBaseResp() != nil {
			kext.ReturnCode.Set(span, response.GetBaseResp().GetStatusCode())
			if response.GetBaseResp().GetStatusCode() != 0 {
				span.LogFields(olog.String("message", response.GetBaseResp().GetStatusMessage()))
			}
		}
	}

	trace.FillSpanEvent(ctx, kext.EventKindPkgSendStartEnum)
}
