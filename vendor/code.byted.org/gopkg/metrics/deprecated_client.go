package metrics

import "sync"

/*

deprecated: use clientv2.go (MetricsClientV2)

*/

type MetricsClient struct {
	mu sync.RWMutex

	NamespacePrefix string
	AllMetrics      map[string]map[string]metricsType
	Server          string
	IgnoreCheck     bool

	sender *Sender

	cache tagscache
}

func NewMetricsClient(server, namespacePrefix string, ignoreCheck bool) *MetricsClient {
	client := &MetricsClient{
		NamespacePrefix: namespacePrefix,
		AllMetrics:      make(map[string]map[string]metricsType),
		Server:          server,
		IgnoreCheck:     ignoreCheck,
		sender:          NewSender(server),
	}
	return client
}

func NewDefaultMetricsClient(namespacePrefix string, ignoreCheck bool) *MetricsClient {
	return NewMetricsClient(DefaultMetricsServer, namespacePrefix, ignoreCheck)
}

func (mc *MetricsClient) DefineCounter(name, prefix string) error {
	return mc.defineMetrics(name, prefix, metricsTypeCounter)
}

func (mc *MetricsClient) DefineTimer(name, prefix string) error {
	return mc.defineMetrics(name, prefix, metricsTypeTimer)
}

func (mc *MetricsClient) DefineStore(name, prefix string) error {
	return mc.defineMetrics(name, prefix, metricsTypeStore)
}

func (mc *MetricsClient) defineMetrics(name, prefix string, mt metricsType) error {
	// mc.IgnoreCheck won't be modified, not need lock.
	if mc.IgnoreCheck {
		return nil
	}
	if len(prefix) == 0 {
		prefix = mc.NamespacePrefix
	}
	mc.mu.Lock()
	defer mc.mu.Unlock()
	m := mc.AllMetrics[prefix]
	if m == nil {
		m = make(map[string]metricsType)
		mc.AllMetrics[prefix] = m
	}
	t, ok := m[name]
	if !ok {
		m[name] = mt
		return nil
	}
	if mt != t {
		return ErrDuplicatedMetrics
	}
	return nil
}

func (mc *MetricsClient) EmitCounter(name string, value interface{}, prefix string, tagkv map[string]string) error {
	return mc.emit(metricsTypeCounter, name, value, prefix, tagkv)
}

func (mc *MetricsClient) EmitTimer(name string, value interface{}, prefix string, tagkv map[string]string) error {
	return mc.emit(metricsTypeTimer, name, value, prefix, tagkv)
}

func (mc *MetricsClient) EmitStore(name string, value interface{}, prefix string, tagkv map[string]string) error {
	return mc.emit(metricsTypeStore, name, value, prefix, tagkv)
}

func (m *MetricsClient) emit(mt metricsType, name string, value interface{},
	prefix string, tagkv map[string]string) error {
	if len(prefix) == 0 {
		prefix = m.NamespacePrefix
	}
	if !m.IgnoreCheck {
		m.mu.RLock()
		types, ok1 := m.AllMetrics[prefix]
		t, ok2 := types[name] // read from nil is safe
		m.mu.RUnlock()
		if !ok1 || !ok2 {
			return ErrEmitUndefinedMetrics
		}
		if t != mt {
			return ErrEmitBadMetricsType
		}
	}
	v, err := toFloat64(value)
	if err != nil {
		return err
	}
	if mt == metricsTypeCounter && v == 0 { // meaningless
		return nil
	}
	tags := Map2Tags(tagkv)
	e := metricEntry{}
	e.mt = mt
	e.prefix = prefix
	e.name = name
	e.tt = m.cache.Get(tags, extTags)
	e.v = v
	return m.sender.SendAsync(e)
}

// If you use the default metricsClient, then the NamespacePrefix is "",
// so you can fill in "prefix" when using DefineCounter, DefineTimer etc.
// and EmitCounter, EmitTimer etc.
// default metrics client won't ignore metrics check.
var metricsClient = NewDefaultMetricsClient("", false)

func DefineCounter(name, prefix string) error {
	return metricsClient.DefineCounter(name, prefix)
}

func DefineTimer(name, prefix string) error {
	return metricsClient.DefineStore(name, prefix)
}

func DefineStore(name, prefix string) error {
	return metricsClient.DefineStore(name, prefix)
}

func EmitCounter(name string, value interface{}, prefix string, tagkv map[string]string) error {
	return metricsClient.EmitCounter(name, value, prefix, tagkv)
}

func EmitTimer(name string, value interface{}, prefix string, tagkv map[string]string) error {
	return metricsClient.EmitTimer(name, value, prefix, tagkv)
}

func EmitStore(name string, value interface{}, prefix string, tagkv map[string]string) error {
	return metricsClient.EmitStore(name, value, prefix, tagkv)
}
