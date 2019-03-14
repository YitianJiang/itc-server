package kitc

import (
	"context"
	"net"
	"time"

	"code.byted.org/kite/kitc/discovery"
)

type RingHashKeyType string

// IDCConfig .
type IDCConfig struct {
	IDC     string
	Percent int // total is 100
}

// CBConfig .
type CBConfig struct {
	IsOpen         bool
	ErrRate        float64
	MinSample      int
	MaxConcurrency int
}

// RPCConfig .
type RPCConfig struct {
	RPCTimeout      int // ms
	ConnectTimeout  int // ms
	ReadTimeout     int // ms
	WriteTimeout    int // ms
	IDCConfig       []IDCConfig
	ServiceCB       CBConfig
	ACLAllow        bool
	DegraPercent    int
	StressBotSwitch bool
}

// RPCMeta .
type RPCMeta struct {
	From        string
	FromCluster string
	FromMethod  string
	To          string
	ToCluster   string
	Method      string
}

func (r RPCMeta) ConfigKey() string {
	sum := len(r.From) + len(r.FromCluster) + len(r.To) + len(r.ToCluster) + len(r.Method) + 4
	buf := make([]byte, 0, sum)
	buf = append(buf, r.From...)
	buf = append(buf, '/')
	buf = append(buf, r.FromCluster...)
	buf = append(buf, '/')
	buf = append(buf, r.To...)
	buf = append(buf, '/')
	buf = append(buf, r.ToCluster...)
	buf = append(buf, '/')
	buf = append(buf, r.Method...)
	return string(buf)
}

func (r RPCMeta) String() string {
	sum := len(r.From) + len(r.FromCluster) + len(r.FromMethod) + len(r.To) + len(r.ToCluster) + len(r.Method) + 5
	buf := make([]byte, 0, sum)
	buf = append(buf, r.From...)
	buf = append(buf, '/')
	buf = append(buf, r.FromCluster...)
	buf = append(buf, '/')
	buf = append(buf, r.FromMethod...)
	buf = append(buf, '/')
	buf = append(buf, r.To...)
	buf = append(buf, '/')
	buf = append(buf, r.ToCluster...)
	buf = append(buf, '/')
	buf = append(buf, r.Method...)
	return string(buf)
}

type rpcInfo struct {
	RPCMeta
	RPCConfig

	// extra info, modified by middlewares
	LocalIP        string
	Env            string
	LogID          string
	Client         string
	TargetIDC      string
	Instances      []*discovery.Instance
	TargetInstance *discovery.Instance
	Conn           net.Conn
	ConnCost       time.Duration
	StressTag      string
	TraceTag       string

	// ringhash key(FIXME)
	RingHashKey string

	// User-defined routing feature
	DDPTag string
}

type rpcInfoCtxKey struct{}

func GetRPCInfo(ctx context.Context) *rpcInfo {
	return ctx.Value(rpcInfoCtxKey{}).(*rpcInfo)
}

func newCtxWithRPCInfo(ctx context.Context, rpcInfo *rpcInfo) context.Context {
	return context.WithValue(ctx, rpcInfoCtxKey{}, rpcInfo)
}
