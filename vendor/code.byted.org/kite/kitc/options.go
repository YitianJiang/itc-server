package kitc

import (
	"fmt"
	"strings"
	"time"

	"code.byted.org/kite/endpoint"
	"code.byted.org/kite/kitc/discovery"
	"code.byted.org/kite/kitc/loadbalancer"
)

// Option .
type Option struct {
	f func(*Options)
}

type CheckUserErrorPair struct {
	Method  string
	Handler CheckUserError
}

// Options .
type Options struct {
	RPCTimeout       time.Duration
	ReadWriteTimeout time.Duration
	ConnTimeout      time.Duration
	ConnMaxRetryTime time.Duration

	Instances    []*discovery.Instance
	Discoverer   discovery.ServiceDiscoverer
	Loadbalancer loadbalancer.Loadbalancer

	TargetCluster   string
	TargetIDC       string
	TargetIDCConfig []IDCConfig

	UseLongPool    bool
	MaxIdle        int // MaxIdlePerIns
	MaxIdleGlobal  int
	MaxIdleTimeout time.Duration

	DisableCircuitBreaker    bool
	CircuitBreakerErrorRate  float64
	CircuitBreakerMinSamples int

	CBUserErrorRate       float64
	CBUserErrorMinSamples int64
	CheckUserErrHandlers  map[string]CheckUserError

	DisableRPCLog bool

	UserDefinedMWs           []endpoint.Middleware
	UserDefinedMWsBeforeBase []endpoint.Middleware

	// deprecated
	CircuitBreakerMaxConcurrency int
	IKService                    IKService
	// FIXME: almost are binary(default)
	ProtocolType ProtocolType

	IDCHashFunc   loadbalancer.KeyFunc
	ClusterPolicy *discovery.DiscoveryFilterPolicy
	EnvPolicy     *discovery.DiscoveryFilterPolicy

	//custom max size of TFramedTransport
	MaxFramedSize int32
}

var globalUserDefinedMWs []endpoint.Middleware // global middleware list, work for all rpc client

// WithTimeout config read write timeout
func WithTimeout(timeout time.Duration) Option {
	return Option{func(op *Options) {
		op.ReadWriteTimeout = timeout
		op.RPCTimeout = timeout
	}}
}

// WithConnTimeout config connect timeout, deprecated
func WithConnTimeout(timeout time.Duration) Option {
	return Option{func(op *Options) {
		op.ConnTimeout = timeout
	}}
}

// WithConnMaxRetryTime deprecated
func WithConnMaxRetryTime(d time.Duration) Option {
	return Option{func(op *Options) {
		op.ConnMaxRetryTime = d
	}}
}

// WithLongConnection deprecated, use WithLongPool;
func WithLongConnection(maxIdle int, maxIdleTimeout time.Duration) Option {
	return Option{func(op *Options) {
		// TODO(zhangyuanjia): 因为server端有默认的3s连接超时, 如果空闲连接大于3s, 可能会从连接池里面取出错误连接使用, 故暂时做此限制;
		_magic_maxIdleTimeout := time.Millisecond * 2500
		if maxIdleTimeout > _magic_maxIdleTimeout {
			maxIdleTimeout = _magic_maxIdleTimeout
			fmt.Printf("KITC: maxIdleTimeout is set to limit: 2.5s\n")
		}

		op.UseLongPool = true
		op.MaxIdle = maxIdle
		op.MaxIdleGlobal = 1 << 20 // no limit
		op.MaxIdleTimeout = maxIdleTimeout
	}}
}

func WithLongPool(maxIdlePerInstance, maxIdleGlobal int, maxIdleTimeout time.Duration) Option {
	return Option{func(op *Options) {
		// TODO(zhangyuanjia): 因为server端有默认的3s连接超时, 如果空闲连接大于3s, 可能会从连接池里面取出错误连接使用, 故暂时做此限制;
		_magic_maxIdleTimeout := time.Millisecond * 2500
		if maxIdleTimeout > _magic_maxIdleTimeout {
			maxIdleTimeout = _magic_maxIdleTimeout
			fmt.Printf("KITC: maxIdleTimeout is set to limit: 2.5s\n")
		}

		op.UseLongPool = true
		op.MaxIdle = maxIdlePerInstance
		op.MaxIdleGlobal = maxIdleGlobal
		op.MaxIdleTimeout = maxIdleTimeout
	}}
}

// WithInstances deprecated, use WithHostPort;
func WithInstances(ins ...*Instance) Option {
	return Option{func(op *Options) {
		dins := make([]*discovery.Instance, 0, len(ins))
		for _, i := range ins {
			dins = append(dins, discovery.NewInstance(i.Host(), i.Port(), i.Tags()))
		}
		op.Instances = dins
	}}
}

// WithHostPort .
func WithHostPort(hosts ...string) Option {
	return Option{func(op *Options) {
		var ins []*discovery.Instance
		for _, hostPort := range hosts {
			val := strings.Split(hostPort, ":")
			if len(val) == 2 {
				ins = append(ins, discovery.NewInstance(val[0], val[1], nil))
			}
		}
		op.Instances = ins
	}}
}

// WithIKService deprecated, please use WithHostPort
func WithIKService(ikService IKService) Option {
	return Option{func(op *Options) {
		op.IKService = ikService
	}}
}

// WithDiscover .
func WithDiscover(discoverer discovery.ServiceDiscoverer) Option {
	return Option{func(op *Options) {
		op.Discoverer = discoverer
	}}
}

// WithCluster .
func WithCluster(cluster string) Option {
	return Option{func(op *Options) {
		op.TargetCluster = cluster
	}}
}

// WithIDC .
func WithIDC(idc string) Option {
	return Option{func(op *Options) {
		op.TargetIDC = idc
		op.TargetIDCConfig = []IDCConfig{IDCConfig{idc, 100}}
	}}
}

// WithCircuitBreaker .
func WithCircuitBreaker(errRate float64, minSample int, concurrency int) Option {
	return Option{func(op *Options) {
		op.CircuitBreakerErrorRate = errRate
		op.CircuitBreakerMinSamples = minSample
		op.CircuitBreakerMaxConcurrency = concurrency
	}}
}

// WithRPCTimeout .
func WithRPCTimeout(timeout time.Duration) Option {
	return Option{func(op *Options) {
		op.RPCTimeout = timeout
	}}
}

// WithDisableCB .
func WithDisableCB() Option {
	return Option{func(op *Options) {
		op.DisableCircuitBreaker = true
	}}
}

// WithLoadbalancer .
func WithLoadbalancer(lb loadbalancer.Loadbalancer) Option {
	return Option{func(op *Options) {
		op.Loadbalancer = lb
	}}
}

// WithDisableRPCLog .
func WithDisableRPCLog() Option {
	return Option{func(op *Options) {
		op.DisableRPCLog = true
	}}
}

// WithMiddleWares will add these MWs after BaseWriterMW
func WithMiddleWares(mws ...endpoint.Middleware) Option {
	return Option{func(op *Options) {
		op.UserDefinedMWs = append(op.UserDefinedMWs, mws...)
	}}
}

// WithMiddleWaresBeforeBase .
func WithMiddleWaresBeforeBase(mws ...endpoint.Middleware) Option {
	return Option{func(op *Options) {
		op.UserDefinedMWsBeforeBase = append(op.UserDefinedMWs, mws...)
	}}
}

func AddGlobalMiddleWares(mws ...endpoint.Middleware) {
	globalUserDefinedMWs = append(globalUserDefinedMWs, mws...)
}

func WithCheckUserErrorHandler(errRate float64, minSamples int64, pairs ...CheckUserErrorPair) Option {
	return Option{func(op *Options) {
		op.CBUserErrorRate = errRate
		op.CBUserErrorMinSamples = minSamples
		if len(pairs) > 0 {
			op.CheckUserErrHandlers = make(map[string]CheckUserError, len(pairs))
		}
		for _, pair := range pairs {
			op.CheckUserErrHandlers[pair.Method] = pair.Handler
		}
	}}
}

func WithProtocolType(protocolType ProtocolType) Option {
	return Option{func(op *Options) {
		op.ProtocolType = protocolType
	}}
}

func WithHashOnIDC(getKey loadbalancer.KeyFunc) Option {
	return Option{func(op *Options) {
		op.IDCHashFunc = getKey
	}}
}

func WithClusterPolicy(policy *discovery.DiscoveryFilterPolicy) Option {
	return Option{func(op *Options) {
		op.ClusterPolicy = policy
	}}
}

func WithEnvPolicy(policy *discovery.DiscoveryFilterPolicy) Option {
	return Option{func(op *Options) {
		op.EnvPolicy = policy
	}}
}

func WithMaxFramedSize(maxSize int32) Option {
	return Option{func(op *Options) {
		op.MaxFramedSize = maxSize
	}}
}
