package ginex

import (
	"context"

	"code.byted.org/gin/ginex/internal"
	"code.byted.org/trace/trace-client-go"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

const (
	// RPCContextKey the key for cache RPC context
	RPCContextKey = "K_RPC_CTX"
)

// MethodContext Deprecated: use RPCContext
func MethodContext(ginCtx *gin.Context) context.Context {
	return RPCContext(ginCtx)
}

// RPCContext returns context information for rpc call
// It's a good choice to call this method at the beginning of handler function to
// avoid concurrent read and write gin.Context
func RPCContext(ginCtx *gin.Context) context.Context {
	ctx := context.Background()
	if ginCtx.Keys != nil {
		for key, val := range ginCtx.Keys {
			ctx = context.WithValue(ctx, key, val)
		}
	}

	// init tracing root span if the trace is enabled
	if appConfig.EnableTracing {
		operation := "-"
		// cause wraps.go::Wrap all registed api-funcs will return the same "Wrap-fn" here.
		// so wraps.go::Wrap need to call SetOperationName to fill real method name into Span
		if method, exist := ginCtx.Get(internal.METHODKEY); exist {
			operation = method.(string)
		}
		normOperation := trace.FormatOperationName(PSM(), operation)
		_, ctx = opentracing.StartSpanFromContext(ctx, normOperation)
	}

	return ctx
}

// RpcContext is depreciated for it has lint error, please use RPCContext.
// This API will be deleted in 2019.06
func RpcContext(ginCtx *gin.Context) context.Context {
	return RPCContext(ginCtx)
}

// CacheRPCContext returns context information for rpc call
// It's a good choice to call this method at the beginning of handler function to
// avoid concurrent read and write gin.Context
func CacheRPCContext(ginCtx *gin.Context) context.Context {
	ctx, exists := ginCtx.Get(RPCContextKey)
	if exists {
		return ctx.(context.Context)
	}

	ctx = RPCContext(ginCtx)
	ginCtx.Set(RPCContextKey, ctx)
	return ctx.(context.Context)
}

// CacheRpcContext depreciated
func CacheRpcContext(ginCtx *gin.Context) context.Context {
	return CacheRPCContext(ginCtx)
}

// GetGinCtxStressTag .
func GetGinCtxStressTag(ginCtx *gin.Context) string {
	return ginCtx.GetString(internal.STRESSKEY)
}

// amend handlerName of cache context
func amendCacheCtxHandlerName(ginCtx *gin.Context, handlerName string) {
	if val, exists := ginCtx.Get(RPCContextKey); exists {
		if ctx, ok := val.(context.Context); ok {
			if originHandlerName, ok := ctx.Value(internal.METHODKEY).(string); ok {
				if originHandlerName != handlerName {
					ginCtx.Set(RPCContextKey, context.WithValue(ctx, internal.METHODKEY, handlerName))
				}
			}
		}
	}
}
