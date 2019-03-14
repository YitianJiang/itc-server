package ginex

import (
	"net/http"
	"strconv"

	"code.byted.org/gin/ginex/internal"
	"code.byted.org/gopkg/env"
	"code.byted.org/trace/trace-client-go"
	kext "code.byted.org/trace/trace-client-go/ext"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func OpentracingHandler() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		ctx := CacheRPCContext(ginCtx)
		span := opentracing.SpanFromContext(ctx)
		if span != nil {
			defer span.Finish()

			ext.SpanKindRPCServer.Set(span)
			ext.Component.Set(span, "ginex")
			if ginCtx.Request != nil && ginCtx.Request.URL != nil {
				ext.HTTPMethod.Set(span, ginCtx.Request.Method)
				ext.HTTPUrl.Set(span, ginCtx.Request.URL.Path)
				span.SetTag("http.params", ginCtx.Request.URL.RawQuery)
			}

			kext.LocalAddress.Set(span, LocalIP())
			kext.LocalCluster.Set(span, LocalCluster())
			kext.LocalIDC.Set(span, env.IDC())
			if logID, exist := ginCtx.Get(internal.LOGIDKEY); exist {
				kext.RPCLogID.Set(span, logID.(string))
			}
			span.SetTag("podName", env.PodName())
		}

		ginCtx.Next()
		if span != nil {
			statusCode := ginCtx.Writer.Status()
			ext.HTTPStatusCode.Set(span, uint16(statusCode))
			if statusCode < http.StatusOK || statusCode >= http.StatusBadRequest {
				ext.Error.Set(span, true)
			}
			kext.ResponseLength.Set(span, int32(ginCtx.Writer.Size()))
		}
	}
}

func DyeForceTraceHandler() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		ctx := CacheRPCContext(ginCtx)
		span := opentracing.SpanFromContext(ctx)
		if span != nil {
			var deviceID int64 = 0
			var err error
			if did, exist := ginCtx.GetQuery(appConfig.DeviceIDParamKey); exist {
				if deviceID, err = strconv.ParseInt(did, 10, 64); err != nil {
					deviceID = 0
				}
			}
			if appConfig.TraceTagFromTLB {
				if _, exist := ginCtx.Get(internal.TT_TRACE_TAG); exist {
					trace.ForceTrace(span)
				}
			} else if deviceID != 0 && trace.DyeUUID(deviceID) {
				trace.ForceTrace(span)
			}
			if deviceID != 0 {
				span.SetTag("device_id", deviceID)
			}
		}
		ginCtx.Next()
	}
}
