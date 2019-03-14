package metrics

import (
	"sync"
	"time"
)

type MetricsClientV2 struct {
	mu sync.RWMutex

	server  string
	prefix  string
	nocheck bool

	sender *Sender

	metrictypes map[string]metricsType

	cache tagscache

	block bool
}

var (
	clients   map[string]*MetricsClientV2
	clientsMu sync.Mutex
)

func NewMetricsClientV2(server, prefix string, nocheck bool) *MetricsClientV2 {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if cli := clients[prefix]; cli != nil {
		return cli
	}

	cli := &MetricsClientV2{server: server, prefix: prefix, nocheck: nocheck}
	cli.metrictypes = make(map[string]metricsType)
	cli.sender = NewSender(server)

	if clients == nil {
		clients = make(map[string]*MetricsClientV2)
	}
	clients[prefix] = cli
	return cli
}

func NewDefaultMetricsClientV2(prefix string, nocheck bool) *MetricsClientV2 {
	return NewMetricsClientV2(DefaultMetricsServer, prefix, nocheck)
}

// SetBlock sets whether Emit* should be block if channel full
func (m *MetricsClientV2) SetBlock(v bool) {
	m.block = v
}

func (m *MetricsClientV2) DefineCounter(name string) error {
	return m.defineMetrics(name, metricsTypeCounter)
}

func (m *MetricsClientV2) DefineRateCounter(name string) error {
	return m.defineMetrics(name, metricsTypeRateCounter)
}

func (m *MetricsClientV2) DefineTimer(name string) error {
	return m.defineMetrics(name, metricsTypeTimer)
}

func (m *MetricsClientV2) DefineStore(name string) error {
	return m.defineMetrics(name, metricsTypeStore)
}

func (m *MetricsClientV2) defineMetrics(name string, mt metricsType) error {
	if m.nocheck {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.metrictypes[name]
	if !ok {
		m.metrictypes[name] = mt
		return nil
	}
	if mt != t {
		return ErrDuplicatedMetrics
	}
	return nil
}

func (m *MetricsClientV2) EmitCounter(name string, value interface{}, tags ...T) error {
	return m.emit(metricsTypeCounter, name, value, 0, tags...)
}

func (m *MetricsClientV2) EmitRateCounter(name string, value interface{}, tags ...T) error {
	return m.emit(metricsTypeRateCounter, name, value, 0, tags...)
}

func (m *MetricsClientV2) EmitTimer(name string, value interface{}, tags ...T) error {
	return m.emit(metricsTypeTimer, name, value, 0, tags...)
}

func (m *MetricsClientV2) EmitStore(name string, value interface{}, tags ...T) error {
	return m.emit(metricsTypeStore, name, value, 0, tags...)
}

// EmitStoreWithTime is same as EmitStore except it emit store metrics with time
func (m *MetricsClientV2) EmitStoreWithTime(name string, value interface{}, t time.Time, tags ...T) error {
	if t.IsZero() {
		return m.emit(metricsTypeTsStore, name, value, 0, tags...)
	}
	return m.emit(metricsTypeTsStore, name, value, t.Unix(), tags...)
}

func (m *MetricsClientV2) emit(mt metricsType, name string, value interface{}, ts int64, tags ...T) error {
	if !m.nocheck {
		m.mu.RLock()
		t, ok := m.metrictypes[name]
		m.mu.RUnlock()
		if !ok {
			return ErrEmitUndefinedMetrics
		}
		if t != mt {
			// we reuse DefineStore by metricsTypeStore for metricsTypeTsStore
			if t != metricsTypeStore || mt != metricsTypeTsStore {
				return ErrEmitBadMetricsType
			}
		}
	}
	v, err := toFloat64(value)
	if err != nil {
		return err
	}
	if (mt == metricsTypeCounter || mt == metricsTypeRateCounter) && v == 0 { // meaningless
		return nil
	}
	e := metricEntry{mt: mt, prefix: m.prefix, name: name, ts: ts, v: v}
	e.tt = m.cache.Get(tags, extTags)
	e.v = v
	if m.block {
		return m.sender.SendAsyncBlock(e)
	} else {
		return m.sender.SendAsync(e)
	}
}
