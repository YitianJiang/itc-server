package trace

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/net2"
	kext "code.byted.org/trace/trace-client-go/ext"
	j "code.byted.org/trace/trace-client-go/jaeger-client"
	jm "code.byted.org/trace/trace-client-go/jaeger-lib/metrics"
	"code.byted.org/trace/trace-client-go/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"reflect"
)

var (
	globalMetricsFactory        = jm.NullFactory
	localHostIP          uint64 = 0
)

var NoopSpanTypeName string

func init() {
	if hostip := env.HostIP(); "" != hostip && hostip != net2.UnknownIPAddr {
		localHostIP = uint64(utils.InetAtoN(hostip))
	}
	noopSpan := opentracing.NoopTracer{}.StartSpan("noop")
	NoopSpanTypeName = reflect.TypeOf(noopSpan).Name()
}

func Init(serviceName string) error {
	metricsFactory := NewBytedMetricsFactory(serviceName, nil)
	metrics := j.NewMetrics(metricsFactory, nil)
	sampler := NewEtcdSampler(serviceName,
		SamplerOptions.Metrics(metrics),
		SamplerOptions.Logger(BytedLogger))
	sender, err := NewGosdkTransport(defaultTaskName)

	if err == nil {
		reporter := j.NewRemoteReporter(sender,
			j.ReporterOptions.Metrics(metrics),
			j.ReporterOptions.Logger(BytedLogger),
		)
		tracer, _ := j.NewTracer(
			serviceName, sampler, reporter,
			j.TracerOptions.Metrics(metrics),
			j.TracerOptions.Logger(BytedLogger),
			j.TracerOptions.PoolSpans(false),
			j.TracerOptions.HighTraceIDGenerator(func() uint64 { return localHostIP }),
			j.TracerOptions.Gen128Bit(true),
		)
		btracer := newBytedTracer(tracer)
		opentracing.SetGlobalTracer(btracer)
	}

	return err
}

func Close() error {
	if btracer, ok := opentracing.GlobalTracer().(*bytedTracer); ok {
		return btracer.Close()
	}
	return nil
}

func ForceTrace(span opentracing.Span) bool {
	ext.SamplingPriority.Set(span, uint16(0x1))
	if jctx, ok := span.Context().(j.SpanContext); ok {
		return jctx.IsSampled()
	}

	return false
}

func IsSampled(span opentracing.Span) bool {
	if jctx, ok := span.Context().(j.SpanContext); ok {
		return jctx.IsSampled()
	}
	return false
}

func IsJaegerSpan(span opentracing.Span) bool {
	_, ok := span.(*j.Span)
	return ok
}

func JSpanContextToString(spanCtx opentracing.SpanContext) string {
	if jctx, ok := spanCtx.(j.SpanContext); ok {
		return jctx.String()
	}

	return ""
}

func FormatOperationName(serviceName, operationName string) string {
	if serviceName == "" || serviceName == "-" {
		serviceName = "unknown"
	}
	if operationName == "" || operationName == "unknown_method" {
		operationName = "-"
	}

	return serviceName + "::" + operationName
}

func SetOperationName(span opentracing.Span, operation string) bool {
	if jspan, ok := span.(*j.Span); ok {
		jspan.SetOperationName(operation)
		return true
	}
	return false
}

func FillSpanEvent(ctx context.Context, event kext.EventKindEnum) error {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return fmt.Errorf("span is nil, fill event-%v into span failed", event)
	}
	if reflect.TypeOf(span).Name() == NoopSpanTypeName {
		return nil
	}

	switch event {
	case kext.EventKindConnectStartEnum:
		span.LogFields(kext.EventKindConnectStart)
	case kext.EventKindConnectEndEnum:
		span.LogFields(kext.EventKindConnectEnd)
	case kext.EventKindPkgSendStartEnum:
		span.LogFields(kext.EventKindPkgSendStart)
	case kext.EventKindPkgSendEndEnum:
		span.LogFields(kext.EventKindPkgSendEnd)
	case kext.EventKindPkgRecvStartEnum:
		span.LogFields(kext.EventKindPkgRecvStart)
	case kext.EventKindPkgRecvEndEnum:
		span.LogFields(kext.EventKindPkgRecvEnd)
	default:
		return fmt.Errorf("not supported EventKind: %v", event)
	}

	return nil
}

func DyeUUID(uuid int64) bool {
	if btracer, ok := opentracing.GlobalTracer().(*bytedTracer); ok {
		if btracer.dyeUUIDSet != nil {
			btracer.RLock()
			defer btracer.RUnlock()
			if btracer.dyeUUIDSet != nil {
				_, exist := btracer.dyeUUIDSet[uuid]
				return exist
			}
		}
	}
	return false
}

func RootSpanEnable(enable int32, from ConfigSourceType) bool {
	if btracer, ok := opentracing.GlobalTracer().(*bytedTracer); ok {
		return btracer.RootSpanEnable(enable, from)
	}
	return false
}

func UpdateDyeUUIDSet(uuidSet map[int64]bool) bool {
	if btracer, ok := opentracing.GlobalTracer().(*bytedTracer); ok {
		return btracer.updateDyeUUIDSet(uuidSet)
	}
	return false
}

// ---------------
// following is bytedTracer releated impl

type ConfigSourceType int32

const (
	CONFIG_FROM_FUNCTION ConfigSourceType = iota
	CONFIG_FROM_REMOTECENTER
	CONFIG_FROM_CONFFILE
	CONFIG_FROM_DEFAULT
)

type bytedTracer struct {
	sync.RWMutex
	opentracing.Tracer
	closed         int64
	rootSpanEnable int32
	dyeUUIDSet     map[int64]bool
	rseSetBy       ConfigSourceType // root span enable config set by
}

func newBytedTracer(tracer opentracing.Tracer) *bytedTracer {
	return &bytedTracer{
		closed:         0,
		Tracer:         tracer,
		rootSpanEnable: 0,
		dyeUUIDSet:     nil,
		rseSetBy:       CONFIG_FROM_DEFAULT,
	}
}

func (t *bytedTracer) Close() error {
	if swapped := atomic.CompareAndSwapInt64(&t.closed, 0, 1); !swapped {
		return fmt.Errorf("repeated attempt to close the sender is ignored")
	}
	if jt, ok := t.Tracer.(*j.Tracer); ok {
		return jt.Close()
	}
	return nil
}

// configure priority: function > config file > remote config center > default
func (t *bytedTracer) RootSpanEnable(enable int32, from ConfigSourceType) bool {
	if from <= t.rseSetBy {
		t.Lock()
		defer t.Unlock()
		if from <= t.rseSetBy {
			// double check when holding lock
			atomic.StoreInt32(&t.rootSpanEnable, enable)
			t.rseSetBy = from
			return true
		}
	}
	return false
}

func (t *bytedTracer) updateDyeUUIDSet(uuidSet map[int64]bool) bool {
	t.Lock()
	defer t.Unlock()
	t.dyeUUIDSet = uuidSet
	return true
}

func (t *bytedTracer) StartSpan(operationName string,
	opts ...opentracing.StartSpanOption) opentracing.Span {

	rootSpanEnable := atomic.LoadInt32(&t.rootSpanEnable)
	var span opentracing.Span
	if rootSpanEnable != 0 {
		span = t.Tracer.StartSpan(operationName, opts...)
	} else {
		childof := false
		for _, opt := range opts {
			if ref, ok := opt.(opentracing.SpanReference); ok {
				if ref.Type == opentracing.ChildOfRef || ref.Type == opentracing.FollowsFromRef {
					if _, ok := ref.ReferencedContext.(j.SpanContext); ok {
						childof = true
						break
					}
				}
			}
		}
		if childof {
			span = t.Tracer.StartSpan(operationName, opts...)
		} else {
			span = (opentracing.NoopTracer{}).StartSpan(operationName, opts...)
		}
	}
	return span
}
