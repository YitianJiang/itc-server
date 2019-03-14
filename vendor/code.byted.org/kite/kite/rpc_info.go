package kite

import "context"

// RPCMeta .
type RPCMeta struct {
	Service         string
	Cluster         string
	UpstreamService string
	UpstreamCluster string
	Method          string
}

func (r RPCMeta) String() string {
	sum := len(r.UpstreamService) + len(r.UpstreamCluster) +
		len(r.Service) + len(r.Cluster) + len(r.Method) + 4
	buf := make([]byte, 0, sum)
	buf = append(buf, r.UpstreamService...)
	buf = append(buf, '/')
	buf = append(buf, r.UpstreamCluster...)
	buf = append(buf, '/')
	buf = append(buf, r.Service...)
	buf = append(buf, '/')
	buf = append(buf, r.Cluster...)
	buf = append(buf, '/')
	buf = append(buf, r.Method...)
	return string(buf)
}

// RPCConfig .
type RPCConfig struct {
	ACLAllow        bool
	StressBotSwitch bool
}

// RPCInfo .
type RPCInfo struct {
	RPCMeta
	RPCConfig

	// extra info
	LogID       string
	Client      string
	Env         string
	LocalIP     string
	RemoteIP    string // upstream IP
	StressTag   string
	TraceTag    string // opentracing context passed by request-extra
	RingHashKey string
	DDPTag      string
}

type rpcInfoCtxKey struct{}

// GetRPCInfo .
func GetRPCInfo(ctx context.Context) *RPCInfo {
	return ctx.Value(rpcInfoCtxKey{}).(*RPCInfo)
}

func newCtxWithRPCInfo(ctx context.Context, rpcInfo *RPCInfo) context.Context {
	return context.WithValue(ctx, rpcInfoCtxKey{}, rpcInfo)
}
