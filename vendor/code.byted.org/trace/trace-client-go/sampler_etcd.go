package trace

import (
	"sync"
	"sync/atomic"
	"time"

	"crypto/md5"
	"encoding/json"
	"math/big"

	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tccclient"
	"code.byted.org/trace/trace-client-go/jaeger-client"
	"code.byted.org/trace/trace-client-go/jaeger-client/thrift-gen/sampling"
	"github.com/pkg/errors"
)

const (
	defaultSamplingRate              = 0.001
	defaultSamplingRefreshInterval   = 3 * time.Minute
	defaultMaxOperations             = 2000
	defaultLowerBoundTracesPerSecond = 1.0 / 300
	defaultSamplingStrategy          = `
{ 
	"service_name" : "*", 
	"default_strategy" : { 
		"operation" : "*", 
		"sampling_rate" : 0.001, 
		"rate_limit" : 1
	}
}
`
)

var lastSamplingDigest = md5sum([]byte(defaultSamplingStrategy))

// SamplerOption is a function that sets some option on the sampler
type SamplerOption func(options *samplerOptions)

// SamplerOptions is a factory for all available SamplerOption's
var SamplerOptions samplerOptions

type samplerOptions struct {
	metrics                 *jaeger.Metrics
	maxOperations           int
	sampler                 jaeger.Sampler
	logger                  jaeger.Logger
	samplingRefreshInterval time.Duration
}

// Metrics creates a SamplerOption that initializes Metrics on the sampler,
// which is used to emit statistics.
func (samplerOptions) Metrics(m *jaeger.Metrics) SamplerOption {
	return func(o *samplerOptions) {
		o.metrics = m
	}
}

// MaxOperations creates a SamplerOption that sets the maximum number of
// operations the sampler will keep track of.
func (samplerOptions) MaxOperations(maxOperations int) SamplerOption {
	return func(o *samplerOptions) {
		o.maxOperations = maxOperations
	}
}

// InitialSampler creates a SamplerOption that sets the initial sampler
// to use before a remote sampler is created and used.
func (samplerOptions) InitialSampler(sampler jaeger.Sampler) SamplerOption {
	return func(o *samplerOptions) {
		o.sampler = sampler
	}
}

// Logger creates a SamplerOption that sets the logger used by the sampler.
func (samplerOptions) Logger(logger jaeger.Logger) SamplerOption {
	return func(o *samplerOptions) {
		o.logger = logger
	}
}

// SamplingRefreshInterval creates a SamplerOption that sets how often the
// sampler will poll local agent for the appropriate sampling strategy.
func (samplerOptions) SamplingRefreshInterval(samplingRefreshInterval time.Duration) SamplerOption {
	return func(o *samplerOptions) {
		o.samplingRefreshInterval = samplingRefreshInterval
	}
}

// -----------------------

// EtcdSampler is a delegating sampler that polls from etcd configure center
// for the appropriate sampling strategy, constructs a corresponding sampler and
// delegates to it for sampling decisions.
type EtcdSampler struct {
	// These fields must be first in the struct because `sync/atomic` expects 64-bit alignment.
	// Cf. https://github.com/golang/go/issues/13868
	closed int64 // 0 - not closed, 1 - closed

	sync.RWMutex
	samplerOptions
	serviceName string

	doneChan chan *sync.WaitGroup
	//kvstore  kvstore.KVStorer
	kvstore *tccclient.Client
}

// NewEtcdSampler creates a sampler that periodically pulls
// the sampling strategy from an etcd configure center
func NewEtcdSampler(
	serviceName string,
	opts ...SamplerOption,
) *EtcdSampler {
	options := applySamplerOptions(opts...)
	kvstore, err := tccclient.NewClient(serviceName, tccclient.NewConfig())
	if err != nil {
		kvstore = nil
		logs.Errorf("init tcc client failed. ErrMsg: %s", err.Error())
	}

	sampler := &EtcdSampler{
		serviceName:    serviceName,
		samplerOptions: options,
		doneChan:       make(chan *sync.WaitGroup),
		kvstore:        kvstore,
	}

	go sampler.pollController()
	return sampler
}

func applySamplerOptions(opts ...SamplerOption) samplerOptions {
	options := samplerOptions{}
	for _, option := range opts {
		option(&options)
	}
	if options.logger == nil {
		options.logger = jaeger.NullLogger
	}
	if options.metrics == nil {
		options.metrics = jaeger.NewNullMetrics()
	}
	if options.maxOperations <= 0 {
		options.maxOperations = defaultMaxOperations
	}
	if options.samplingRefreshInterval <= 0 {
		options.samplingRefreshInterval = defaultSamplingRefreshInterval
	}
	if options.sampler == nil {
		options.sampler, _ = jaeger.NewProbabilisticSampler(defaultSamplingRate)
	}

	return options
}

// IsSampled implements IsSampled() of Sampler.
func (s *EtcdSampler) IsSampled(id jaeger.TraceID, operation string) (bool, []jaeger.Tag) {
	s.RLock()
	defer s.RUnlock()
	return s.sampler.IsSampled(id, operation)
}

// Close implements Close() of Sampler.
func (s *EtcdSampler) Close() {
	if swapped := atomic.CompareAndSwapInt64(&s.closed, 0, 1); !swapped {
		s.logger.Error("Repeated attempt to close the sampler is ignored")
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	s.doneChan <- &wg
	wg.Wait()
}

// Equal implements Equal() of Sampler.
func (s *EtcdSampler) Equal(other jaeger.Sampler) bool {
	// NB The Equal() function is expensive and will be removed. See adaptiveSampler.Equal() for
	// more information.
	if o, ok := other.(*EtcdSampler); ok {
		s.RLock()
		o.RLock()
		defer s.RUnlock()
		defer o.RUnlock()
		return s.sampler.Equal(o.sampler)
	}

	return false
}

func (s *EtcdSampler) getSampler() jaeger.Sampler {
	s.Lock()
	defer s.Unlock()
	return s.sampler
}

func (s *EtcdSampler) setSampler(sampler jaeger.Sampler) {
	s.Lock()
	defer s.Unlock()
	s.sampler = sampler
}

func (s *EtcdSampler) updateSampler() {
	if s.serviceName == "" || s.kvstore == nil {
		return
	}

	var digest int64 = 0
	val, err := s.kvstore.Get("opentracing_sampling")
	if err != nil || val == "" {
		s.metrics.SamplerQueryFailure.Inc(1)
		val = defaultSamplingStrategy
		digest = lastSamplingDigest
	} else {
		digest = md5sum([]byte(val))
		s.metrics.SamplerRetrieved.Inc(1)
	}

	if digest != lastSamplingDigest {
		s.logger.Infof("digest:%x last-digest:%x json-val:\n%s\n", digest, lastSamplingDigest, val)
		serviceStrategies, err := jsonToSamplingStrategies(s.serviceName, val)
		if err != nil || serviceStrategies == nil {
			logs.Error("unmarshal failed. ErrMsg: %s", err.Error())
			s.metrics.SamplerUpdateFailure.Inc(1)
			return
		}

		if err := s.updateAdaptiveSampler(serviceStrategies); err != nil {
			s.logger.Error("update adaptive sampler failed. ErrMsg: " + err.Error())
			s.metrics.SamplerUpdateFailure.Inc(1)
		} else {
			s.metrics.SamplerUpdated.Inc(1)
			lastSamplingDigest = digest
		}
	}
}

// NB: this function should only be called while holding a Write lock
func (s *EtcdSampler) updateAdaptiveSampler(strategies *sampling.PerOperationSamplingStrategies) error {
	sampler, err := jaeger.NewAdaptiveSampler(strategies, s.maxOperations)
	if err == nil {
		s.Lock()
		defer s.Unlock()
		s.sampler = sampler
	}
	return err
}

func (s *EtcdSampler) pollController() {
	ticker := time.NewTicker(s.samplingRefreshInterval)
	defer ticker.Stop()
	s.pollControllerWithTicker(ticker)
}

func (s *EtcdSampler) pollControllerWithTicker(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			s.updateSampler()
		case wg := <-s.doneChan:
			wg.Done()
			return
		}
	}
}

type ServiceSamplingStrategy struct {
	ServiceName         string                    `json:"service_name"`
	DefaultStrategy     OperationSamplingParam    `json:"default_strategy"`
	OperationStrategies []*OperationSamplingParam `json:"operation_strategies"`
	RootSpanEnable      *int32                    `json:"root_span_enable"`
	InnerUUIDList       []int64                   `json:"inner_uuid_list"`
	DyeUUIDList         []int64                   `json:"dye_uuid_list"`
}

type OperationSamplingParam struct {
	Operation                 string  `json:"operation"`
	SamplingRate              float64 `json:"sampling_rate"`
	UpperBoundTracesPerSecond float64 `json:"rate_limit"`
}

func jsonToSamplingStrategies(serviceName, val string) (res *sampling.PerOperationSamplingStrategies, err error) {
	res, err = nil, nil

	var param ServiceSamplingStrategy
	if err = json.Unmarshal([]byte(val), &param); err != nil {
		err = errors.Errorf("json unmarshal failed for ServiceSamplingStrategy")
		return
	}
	if param.ServiceName != "*" && param.ServiceName != serviceName {
		err = errors.Errorf("service name not match between etcd and local service")
		return
	}

	opsLimiter := make(map[string]*upperBoundSender)
	// construct OperationSamplingStrategy for updating adaptiveSampler
	res = &sampling.PerOperationSamplingStrategies{
		DefaultSamplingProbability:       param.DefaultStrategy.SamplingRate,
		DefaultLowerBoundTracesPerSecond: defaultLowerBoundTracesPerSecond,
		DefaultUpperBoundTracesPerSecond: &(param.DefaultStrategy.UpperBoundTracesPerSecond),
	}
	for _, ops := range param.OperationStrategies {
		normOperation := FormatOperationName(serviceName, ops.Operation)
		res.PerOperationStrategies = append(res.PerOperationStrategies,
			&sampling.OperationSamplingStrategy{
				Operation:             normOperation,
				ProbabilisticSampling: &sampling.ProbabilisticSamplingStrategy{ops.SamplingRate},
			})
		if ops.UpperBoundTracesPerSecond > 0.0 {
			opsLimiter[normOperation] = newUpperBoundSender(ops.UpperBoundTracesPerSecond)
		}
	}

	// update global rate limiter
	globalLimiter.update(param.DefaultStrategy.UpperBoundTracesPerSecond, opsLimiter)

	// update root span enable & dye uuid list
	if param.RootSpanEnable != nil {
		RootSpanEnable(*param.RootSpanEnable, CONFIG_FROM_REMOTECENTER)
	}
	if len(param.InnerUUIDList) != 0 || len(param.DyeUUIDList) != 0 {
		dyeUUIDSet := make(map[int64]bool)
		for _, uuid := range param.InnerUUIDList {
			dyeUUIDSet[uuid] = true
		}
		for _, uuid := range param.DyeUUIDList {
			dyeUUIDSet[uuid] = true
		}
		UpdateDyeUUIDSet(dyeUUIDSet)
	}

	return
}

func md5sum(data []byte) int64 {
	h := md5.New()
	h.Write(data)
	ret := big.NewInt(0)
	ret.SetBytes(h.Sum(nil)[0:8])
	return ret.Int64()
}
