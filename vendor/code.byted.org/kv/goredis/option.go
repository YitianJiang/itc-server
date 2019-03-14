package goredis

import (
	"time"

	"code.byted.org/kv/redis-v6"
)

const (
	// default timeout
	REDIS_DIAL_TIMEOUT  = 50 * time.Millisecond
	REDIS_READ_TIMEOUT  = 50 * time.Millisecond
	REDIS_WRITE_TIMEOUT = 50 * time.Millisecond

	// default pool timeout
	REDIS_POOL_SIZE      = 100
	REDIS_POOL_TIMEOUT   = 50 * time.Millisecond
	REDIS_POOL_INIT_SIZE = 10

	REDIS_IDLE_TIMEOUT = 60 * time.Minute
	REDIS_LIVE_TIMEOUT = 0

	REDIS_IDLE_CHECK_FREQUENCY = time.Minute

	REDIS_AUTO_LOAD_CONF     = true
	REDIS_AUTO_LOAD_INTERVAL = time.Second * 30

	MAX_FAILURE_RATE = 0.6
	MIN_SAMPLE       = 50
	WINDOW_TIME      = time.Millisecond * 10000
)

// Option is used to configure a redis client and should be passed to NewClient.
type Option struct {
	*redis.Options

	// auto load server/proxy conf by consul or file(redis.conf, redis_web.conf), default true
	autoLoadConf bool
	// specified auto load conf interval, default 10s
	autoLoadInterval time.Duration

	// the max failure rate the breaker is allowed before it trip the circuit
	maxFailureRate float64
	// min samples before test the failure rate
	minSample int64
	// sample window
	windowTime time.Duration

	// user defined config file path
	configFilePath string

	// use consul
	useConsul bool

	//init conn num
	PoolInitSize int
}

func NewOption() *Option {
	opts := &redis.Options{
		DialTimeout:  REDIS_DIAL_TIMEOUT,
		ReadTimeout:  REDIS_READ_TIMEOUT,
		WriteTimeout: REDIS_WRITE_TIMEOUT,

		PoolSize:           REDIS_POOL_SIZE,
		PoolTimeout:        REDIS_POOL_TIMEOUT,
		LiveTimeout:        REDIS_LIVE_TIMEOUT,
		IdleTimeout:        REDIS_IDLE_TIMEOUT,
		IdleCheckFrequency: REDIS_IDLE_CHECK_FREQUENCY,
	}
	opt := &Option{
		Options:          opts,
		PoolInitSize:     REDIS_POOL_INIT_SIZE,
		autoLoadConf:     REDIS_AUTO_LOAD_CONF,
		autoLoadInterval: REDIS_AUTO_LOAD_INTERVAL,
		maxFailureRate:   MAX_FAILURE_RATE,
		minSample:        MIN_SAMPLE,
		windowTime:       WINDOW_TIME,
		configFilePath:   "",
		useConsul:        false,
	}
	return opt
}

// NewOption by self specified timeouts
// default auto load conf unless you disable it by DisableAutoLoadConf()
func NewOptionWithTimeout(
	dialTimeout,
	readTimeout,
	writeTimeout,
	poolTimeout,
	idleTimeout,
	liveTimeout time.Duration,
	poolSize int) *Option {
	if dialTimeout == 0 {
		dialTimeout = REDIS_DIAL_TIMEOUT
	}
	if readTimeout == 0 {
		readTimeout = REDIS_READ_TIMEOUT
	}
	if writeTimeout == 0 {
		writeTimeout = REDIS_WRITE_TIMEOUT
	}
	if poolTimeout == 0 {
		poolTimeout = REDIS_POOL_TIMEOUT
	}
	if idleTimeout == 0 {
		idleTimeout = REDIS_IDLE_TIMEOUT
	}

	if poolSize <= 0 {
		poolSize = REDIS_POOL_SIZE
	}
	opts := &redis.Options{
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,

		PoolSize:           poolSize,
		PoolTimeout:        poolTimeout,
		IdleTimeout:        idleTimeout,
		LiveTimeout:        liveTimeout,
		IdleCheckFrequency: REDIS_IDLE_CHECK_FREQUENCY,
	}
	opt := &Option{
		Options:          opts,
		PoolInitSize:     REDIS_POOL_INIT_SIZE,
		autoLoadConf:     REDIS_AUTO_LOAD_CONF,
		autoLoadInterval: REDIS_AUTO_LOAD_INTERVAL,

		maxFailureRate: MAX_FAILURE_RATE,
		minSample:      MIN_SAMPLE,
		windowTime:     WINDOW_TIME,

		configFilePath: "",
		useConsul:      false,
	}
	return opt
}

func (p *Option) DisableAutoLoadConf() {
	p.autoLoadConf = false
}

func (p *Option) SetMaxRetries(maxRetries int) {
	if maxRetries > 0 {
		p.Options.MaxRetries = maxRetries
	}
}

func (p *Option) SetPoolInitSize(poolInitSize int) {
	if poolInitSize <= 0 {
		p.PoolInitSize = REDIS_POOL_INIT_SIZE
	} else {
		p.PoolInitSize = poolInitSize
	}
}

func (p *Option) SetAutoLoadInterval(t time.Duration) {
	if t > time.Second {
		p.autoLoadInterval = t
	}
}

func (p *Option) SetCircuitBreakerParam(maxFailureRate float64, minSample int64, windowTime time.Duration) {
	if maxFailureRate >= 0 && maxFailureRate < 1 {
		p.maxFailureRate = maxFailureRate
	}
	if minSample > 0 {
		p.minSample = minSample
	}
	if windowTime > 0 {
		p.windowTime = windowTime
	}
}

// SetConfigFilePath can set a specified service discovery config file path when you need.
// Default is /opt/tiger/ss_conf/ss/redis.conf and redis_web.conf
func (p *Option) SetConfigFilePath(path string) {
	if len(path) > 0 {
		p.configFilePath = path
	}
}

// SetServiceDiscoveryWithConsul can get service server:port by consul with the redis cluster name.
// You must choose only one from config file and consul, and now default is config file.
// And at the same time you should know the cluster name may be different in consul and config file
// e.g.
// in config file, cluster name may be: redis_cluster_name
// in consul, cluster name may be: rcproxy_redis_cluster_name or twemproxy_redis_cluster_name
func (p *Option) SetServiceDiscoveryWithConsul() {
	p.useConsul = true
}
