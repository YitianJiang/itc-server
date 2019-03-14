package kitc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/thrift"
	"code.byted.org/kite/endpoint"
	"code.byted.org/kite/kitc/circuitbreaker"
	"code.byted.org/kite/kitc/connpool"
	"code.byted.org/kite/kitc/discovery"
	"code.byted.org/kite/kitc/loadbalancer"
	"code.byted.org/kite/kitutil"
	"code.byted.org/kite/kitutil/kitevent"
	"code.byted.org/kite/kitutil/kvstore"
	"code.byted.org/kite/kitutil/logid"
)

type CheckUserError func(resp interface{}) (isFailed bool)

const KITECLIENTKEY = "K_KITC_CLIENT" // KitcClient指针

// KitcClient ...
type KitcClient struct {
	meshMode bool
	name     string
	opts     *Options
	client   Client

	userErrBreaker  *circuit.Panel
	serviceBreaker  *circuit.Panel
	instanceBreaker *circuit.Panel
	pool            connpool.ConnPool
	discoverer      *kitcDiscoverer
	remoteConfiger  *remoteConfiger
	eventQueue      *kitevent.Queue
	loadbalancer    loadbalancer.Loadbalancer
	maxFramedSize   int32

	chain endpoint.Middleware
	once  sync.Once
}

// NewClient ...
func NewClient(name string, ops ...Option) (*KitcClient, error) {
	client, ok := clients[name]
	if !ok {
		return nil, fmt.Errorf("Unknow client name %s, forget import?", name)
	}
	return newWithThriftClient(name, client, ops...)
}

// NewWithThriftClient ...
func NewWithThriftClient(name string, thriftClient Client, ops ...Option) (*KitcClient, error) {
	return newWithThriftClient(name, thriftClient, ops...)
}

func newWithThriftClient(name string, thriftClient Client, ops ...Option) (*KitcClient, error) {
	opts := new(Options)
	for _, do := range ops {
		do.f(opts)
	}

	var pool connpool.ConnPool
	// ServiceMeshMode use LongPool
	if ServiceMeshMode {
		pool = connpool.NewLongPool(100000, 1<<20, 5*time.Second, name)
	} else {
		if opts.UseLongPool {
			pool = connpool.NewLongPool(opts.MaxIdle, opts.MaxIdleGlobal, opts.MaxIdleTimeout, name)
		} else {
			pool = connpool.NewShortPool(name)
		}
	}

	kitclient := &KitcClient{
		name:       name,
		opts:       opts,
		client:     thriftClient,
		pool:       pool,
		eventQueue: kitevent.NewQueue(10),
	}

	if !ServiceMeshMode {
		kitclient.meshMode = false
		// lb
		var lb loadbalancer.Loadbalancer
		if opts.Loadbalancer != nil {
			lb = opts.Loadbalancer
		} else {
			lb = loadbalancer.NewWeightLoadbalancer()
		}
		kitclient.loadbalancer = lb

		// remoteConfiger
		kitclient.remoteConfiger = newRemoteConfiger(kitclient, kvstore.NewETCDStorer())

		// construct service discover
		var discoverer discovery.ServiceDiscoverer
		if opts.Instances != nil {
			discoverer = discovery.NewCustomDiscoverer(opts.Instances)
		} else if opts.Discoverer != nil {
			discoverer = opts.Discoverer
		} else if opts.IKService != nil {
			discoverer = &ikserviceWrapper{opts.IKService}
		} else {
			discoverer = discovery.NewConsulDiscoverer()
		}
		kitclient.discoverer = newKitcDiscoverer(kitclient, discoverer)
		kitclient.discoverer.policy = discovery.NewDiscoveryPolicy()
		if opts.ClusterPolicy != nil {
			kitclient.discoverer.policy.ClusterPolicy = opts.ClusterPolicy
		}
		if opts.EnvPolicy != nil {
			kitclient.discoverer.policy.EnvPolicy = opts.EnvPolicy
		}

		// cb
		kitclient.serviceBreaker, _ = circuit.NewPanel(kitclient.serviceCBChangeHandler, circuit.Options{})
		kitclient.instanceBreaker, _ = circuit.NewPanel(kitclient.instanceCBChangeHandler, circuit.Options{
			ShouldTrip: circuit.RateTripFunc(0.5, 200),
		})

		if opts.CheckUserErrHandlers != nil {
			kitclient.userErrBreaker, _ = circuit.NewPanel(kitclient.userErrCBChangeHandler, circuit.Options{
				ShouldTrip: circuit.RateTripFunc(opts.CBUserErrorRate, opts.CBUserErrorMinSamples),
			})
		}

		kitclient.maxFramedSize = thrift.DEFAULT_MAX_LENGTH
		if opts.MaxFramedSize > 0 {
			kitclient.maxFramedSize = opts.MaxFramedSize
		}
	} else {
		kitclient.remoteConfiger = newRemoteConfiger(kitclient, kvstore.NewETCDStorer())

		kitclient.meshMode = true
	}

	registerKitcClient(kitclient)
	return kitclient, nil
}

func (kc *KitcClient) SetChain(chain endpoint.Middleware) {
	kc.chain = chain
}

func (kc *KitcClient) initMWChain() {
	if logger == nil {
		logger = &localLogger{}
	}
	if ctxKVLogger == nil {
		ctxKVLogger = &localLogger{}
	}
	if kc.opts.DisableRPCLog {
		logger = &emptyLogger{}
		ctxKVLogger = &emptyLogger{}
	}

	if kc.chain != nil {
		return
	}

	var mids []endpoint.Middleware

	if len(globalUserDefinedMWs) > 0 {
		mids = append(mids, globalUserDefinedMWs...)
	}

	if len(kc.opts.UserDefinedMWsBeforeBase) > 0 {
		mids = append(mids, kc.opts.UserDefinedMWsBeforeBase...)
	}
	mids = append(mids, BaseWriterMW, OpenTracingMW)
	if len(kc.opts.UserDefinedMWs) > 0 {
		mids = append(mids, kc.opts.UserDefinedMWs...)
	}

	if ServiceMeshMode {
		mids = append(mids,
			// RPC logs
			NewRPCLogMW(logger), // init in the first Call()
			// emit RPC metrics
			MeshRPCMetricsMW,
			// Pool
			NewMeshPoolMW(kc.pool),
			// I/O error handler
			MeshIOErrorHandlerMW,
			// set headers
			MeshSetHeadersMW,
		)
	} else {
		mids = append(mids,
			// RPC logs
			NewRPCLogMW(logger), // init in the first Call()
			// emit RPC metrics
			RPCMetricsMW,
			// read acl config from remote to control this RPC
			RPCACLMW,
			// stress
			StressBotMW,
			// service degradation middleware
			DegradationMW,
			// user(business logic) error circuitbreaker
			NewUserErrorCBMW(kc.userErrBreaker, kc.opts.CheckUserErrHandlers),
			// service breaker
			NewServiceBreakerMW(kc.serviceBreaker),
			// RPC timeout
			RPCTimeoutMW,
			// select IDC
			NewIDCSelectorMW(kc.opts.IDCHashFunc),
			// service discover
			NewServiceDiscoverMW(kc.discoverer),
			// LB and RPC retry
			NewLoadbalanceMW(kc.loadbalancer),
			// Instance CB
			NewInstanceBreakerMW(kc.instanceBreaker),
			// Pool
			NewPoolMW(kc.pool),
			// I/O error handler
			IOErrorHandlerMW,
		)
	}

	kc.chain = endpoint.Chain(mids[0], mids[1:]...)
}

// Call do some remote calling
func (kc *KitcClient) Call(method string, ctx context.Context, request interface{}) (endpoint.KitcCallResponse, error) {
	metricsClient.EmitCounter("kite.request.throughput", 1, "", nil)

	// MWs依赖的某些外部组件, 需要在运行时才能确定, 所以在第一次Call时, 进行初始化
	kc.once.Do(kc.initMWChain)

	// 因为request会被生成代码包裹,
	// 如果用户使用了自定义的LB, 需要保证直接把最原始的req传递给用户
	if !ServiceMeshMode {
		if kc.opts.Loadbalancer != nil || kc.opts.IDCHashFunc != nil {
			ctx = context.WithValue(ctx, lbKey, request)
		}
	}

	// set ringhash key(FIXME)
	if ServiceMeshMode {
		if consist, ok := kc.opts.Loadbalancer.(*loadbalancer.ConsistBalancer); ok {
			if ringhashKey, err := consist.GetKey(ctx, request); err == nil && len(ringhashKey) > 0 {
				ctx = context.WithValue(ctx, RingHashKeyType(":CH"), ringhashKey)
			}
		}
	}

	ctx, err := kc.initRPCInfo(method, ctx)
	if err != nil {
		return nil, err
	}

	caller := kc.client.New(kc)
	next, request := caller.Call(method, request)
	if next == nil || request == nil {
		return nil, fmt.Errorf("service=%s method=%s  unknow method return nil", kc.name, method)
	}

	ctx = context.WithValue(ctx, KITECLIENTKEY, kc)
	resp, err := kc.chain(next)(ctx, request)
	if _, ok := resp.(endpoint.KitcCallResponse); !ok {
		return nil, err
	}
	return resp.(endpoint.KitcCallResponse), err
}

func (kc *KitcClient) initRPCInfo(method string, ctx context.Context) (context.Context, error) {
	// construct RPCMeta
	to := kc.name
	toCluster, _ := kitutil.GetCtxTargetClusterName(ctx)
	if toCluster == "" {
		toCluster = kc.opts.TargetCluster
	}
	if toCluster == "" && !ServiceMeshMode {
		toCluster = "default"
	}
	fromCluster := kitutil.GetCtxWithDefault(kitutil.GetCtxCluster, ctx, env.Cluster())
	from := kitutil.GetCtxWithDefault(kitutil.GetCtxServiceName, ctx, env.PSM())
	fromMethod := kitutil.GetCtxWithDefault(kitutil.GetCtxMethod, ctx, "unknown_method")
	if from == "" {
		return nil, errors.New("no service's name for rpc call, you can use kitutil.NewCtxWithServiceName(ctx, xxxx) to set ctx")
	}
	rpcMeta := RPCMeta{
		From:        from,
		FromCluster: fromCluster,
		FromMethod:  fromMethod,
		To:          to,
		ToCluster:   toCluster,
		Method:      method,
	}

	// get RPCConf
	rpcConf, err := kc.GetRPCConfig(rpcMeta)
	if err != nil {
		return nil, fmt.Errorf("get RPC config err: %s", err.Error())
	}

	// TODO(zhangyuanjia): 移除该兼容性字段
	rpcTimeout, _ := kitutil.GetCtxRPCTimeout(ctx)
	if rpcTimeout > 0 {
		rpcConf.RPCTimeout = int(rpcTimeout / time.Millisecond)
	}

	// prepare some extra fields
	logID, _ := kitutil.GetCtxLogID(ctx)
	if logID == "" {
		logID = logid.GetNginxID()
		ctx = kitutil.NewCtxWithLogID(ctx, logID)
	}
	localIP, _ := kitutil.GetCtxLocalIP(ctx)
	if localIP == "" {
		localIP = env.HostIP()
	}
	env, _ := kitutil.GetCtxEnv(ctx)
	stressTag, _ := kitutil.GetCtxStressTag(ctx)
	traceTag, _ := kitutil.GetCtxTraceTag(ctx)
	instances, _ := kitutil.GetCtxRPCInstances(ctx)

	// construct RPCInfo
	rpcInfo := &rpcInfo{
		RPCMeta:   rpcMeta,
		RPCConfig: rpcConf,
		LogID:     logID,
		LocalIP:   localIP,
		Instances: instances,
		Env:       env,
		StressTag: stressTag,
		TraceTag:  traceTag,
	}

	if ServiceMeshMode {
		// get MeshAddr
		meshAddr := strings.Split(ServiceMeshEgressAddr, ":")
		var host, port string
		if len(meshAddr) == 2 {
			host = meshAddr[0]
			port = meshAddr[1]
		} else {
			host = ServiceMeshEgressAddr // domain socket mode
		}
		rpcInfo.TargetInstance = &discovery.Instance{
			Host: host,
			Port: port,
		}

		// dest_address
		rpcInfo.Instances = kc.opts.Instances
		rpcInfo.TargetIDC = kc.opts.TargetIDC
		if ringHashKey, ok := ctx.Value(RingHashKeyType(":CH")).(string); ok {
			rpcInfo.RingHashKey = ringHashKey
		}

	}

	// Pass DDP ID to Mesh Proxy
	if tag, ok := kitutil.GetCtxDDPRoutingTag(ctx); ok {
		rpcInfo.DDPTag = tag
	} else {
		// Calculate DDP ID if needed
		if WithUIDAndDIDSet(ctx) {
			if tag, ok := kc.checkDDPRoutingRules(ctx, rpcMeta); ok {
				ctx = kitutil.NewCtxWithDDPRoutingTag(ctx, tag)
				rpcInfo.DDPTag = tag
			}
		}
	}

	return newCtxWithRPCInfo(ctx, rpcInfo), nil
}

// Name .
func (kc *KitcClient) Name() string {
	return kc.name
}

// GetRPCConfig .
// TODO(zhengjianbo): RPCMeta, RPCConfig使用指针，避免多次copy, 需要注意defaultRPCConfig的deepcopy
func (kc *KitcClient) GetRPCConfig(r RPCMeta) (RPCConfig, error) {
	var c RPCConfig
	if ServiceMeshMode {
		// mesh环境下有两套超时配置，一个是proxy超时配置, 一个是实际rpc超时配置（不指定的话从control plane读取）
		// proxy超时配置固定，rpc超时配置可选，这个函数只处理rpc超时配置， -1表示用户未指定
		c = RPCConfig{
			RPCTimeout:      -1,
			ConnectTimeout:  -1,
			ReadTimeout:     -1,
			WriteTimeout:    -1,
			StressBotSwitch: false,
		}
	} else {
		// get remote config
		var err error
		c, err = kc.remoteConfiger.GetRemoteConfig(r)
		if err != nil {
			logs.Warnf("KITC: get remote config err: %s, default config will be used", err.Error())
			c = defaultRPCConfig
		}
	}

	// merge options' config
	if kc.opts.ReadWriteTimeout > 0 {
		timeoutInMS := int(kc.opts.ReadWriteTimeout / time.Millisecond)
		c.ReadTimeout = timeoutInMS
		c.WriteTimeout = timeoutInMS
		c.RPCTimeout = timeoutInMS
	}
	if kc.opts.ConnTimeout > 0 {
		timeoutInMS := int(kc.opts.ConnTimeout / time.Millisecond)
		c.ConnectTimeout = timeoutInMS
	}
	if kc.opts.RPCTimeout > 0 {
		timeoutInMS := int(kc.opts.RPCTimeout / time.Millisecond)
		c.ReadTimeout = timeoutInMS
		c.WriteTimeout = timeoutInMS
		c.RPCTimeout = timeoutInMS
	}

	// don't care cb when mesh mode
	if !ServiceMeshMode {
		if kc.opts.DisableCircuitBreaker {
			c.ServiceCB.IsOpen = false
		}
		if kc.opts.CircuitBreakerErrorRate > 0 ||
			kc.opts.CircuitBreakerMaxConcurrency > 0 ||
			kc.opts.CircuitBreakerMinSamples > 0 {
			c.ServiceCB.ErrRate = kc.opts.CircuitBreakerErrorRate
			c.ServiceCB.MaxConcurrency = kc.opts.CircuitBreakerMaxConcurrency
			c.ServiceCB.MinSample = kc.opts.CircuitBreakerMinSamples
		}
	}
	if kc.opts.TargetIDC != "" && len(kc.opts.TargetIDCConfig) > 0 {
		c.IDCConfig = kc.opts.TargetIDCConfig
	}

	return c, nil
}

// Options .
func (kc *KitcClient) Options() *Options {
	return kc.opts
}

// RemoteConfigs .
func (kc *KitcClient) RemoteConfigs() map[string]RPCConfig {
	return kc.remoteConfiger.GetAllRemoteConfigs()
}

// ServiceInstances .
func (kc *KitcClient) ServiceInstances() map[string][]*discovery.Instance {
	return kc.discoverer.Dump()
}

// RecentEvents .
func (kc *KitcClient) RecentEvents() []*kitevent.KitEvent {
	return kc.eventQueue.Dump()
}

// ServiceCircuitbreaker .
func (kc *KitcClient) ServiceCircuitbreaker() *circuit.Panel {
	return kc.serviceBreaker
}

// InstanceCircuitbreaker .
func (kc *KitcClient) InstanceCircuitbreaker() *circuit.Panel {
	return kc.instanceBreaker
}

func (kc *KitcClient) pushEvent(e *kitevent.KitEvent) {
	select {
	case kitcEventCh <- e:
	default:
	}
	logs.Infof("KITC: event: name: %s, time: %v, detail: %s", e.Name, e.Time, e.Detail)
	kc.eventQueue.Push(e)
}

func (kc *KitcClient) serviceCBChangeHandler(key string, oldState, newState circuit.State, m circuit.Metricser) {
	e := &kitevent.KitEvent{
		Name: "service_circuitbreaker_change",
		Time: time.Now(),
		Detail: fmt.Sprintf("%s: %s -> %s, (succ: %d, err: %d, tmout: %d, rate: %f)",
			key, oldState, newState, m.Successes(), m.Failures(), m.Timeouts(), m.ErrorRate()),
	}
	kc.pushEvent(e)
}

func (kc *KitcClient) instanceCBChangeHandler(key string, oldState, newState circuit.State, m circuit.Metricser) {
	e := &kitevent.KitEvent{
		Name: "bad_ip_circuitbreaker_change",
		Time: time.Now(),
		Detail: fmt.Sprintf("%s: %s -> %s, (succ: %d, err: %d, tmout: %d, rate: %f)",
			key, oldState, newState, m.Successes(), m.Failures(), m.Timeouts(), m.ErrorRate()),
	}
	kc.pushEvent(e)
}

func (kc *KitcClient) userErrCBChangeHandler(key string, oldState, newState circuit.State, m circuit.Metricser) {
	e := &kitevent.KitEvent{
		Name: "user_error_circuitbreaker_change",
		Time: time.Now(),
		Detail: fmt.Sprintf("%s: %s -> %s, (succ: %d, err: %d, tmout: %d, rate: %f)",
			key, oldState, newState, m.Successes(), m.Failures(), m.Timeouts(), m.ErrorRate()),
	}
	kc.pushEvent(e)
}

func (kc *KitcClient) discoverChangeHandler(key string, oldIns, newIns []*discovery.Instance) {
	e := &kitevent.KitEvent{
		Name:   "service_address_change",
		Time:   time.Now(),
		Detail: fmt.Sprintf("%s: %s -> %s", key, instances2ReadableStr(oldIns), instances2ReadableStr(newIns)),
	}
	kc.pushEvent(e)
	kc.instanceBreaker.RemoveAllBreakers() // remove all old instances cbs
}
